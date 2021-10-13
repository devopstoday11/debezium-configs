package main

import (
	"encoding/json"
	"fmt"

	"context"
	"database/sql"
	"log"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	_ "github.com/lib/pq"
)

// LambdaHandler will handle the heavy lifting of parsing the cloudevent from sqs and updating the DB
func LambdaHandler(ctx context.Context, event events.SQSEvent) error {
	// Optimally, the configuration options should be pulled from environment variables
	db, err := sql.Open("postgres", "host=REPLACE_WITH_RDS_INSTANCE_NAME port=5432 user=XXXX password=XXXXX dbname=XXXXX sslmode=disable")

	if err != nil {
		log.Println("Error connecting to db", err)
		return err
	}

	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Printf("Error polling db", err)
		return err
	}

	for _, msg := range event.Records {
		event := cloudevents.NewEvent()
		err = json.Unmarshal([]byte(msg.Body), &event)
		if err != nil {
			log.Println("Error parsing event", err)
			return err
		}

		err = updateDatabase(event, db)
		if err != nil {
			return err
		}
	}

	return nil
}

// main starts the lambda service
func main() {
	lambda.Start(LambdaHandler)
}

type FieldTypes struct {
	Type     string  `json:type`
	Optional bool    `json:optional`
	Name     string  `json:name,omitempty`
	Version  float64 `json:version,omitempty`
	Field    string  `json:field`
}

// mapFields will take an array of fields and convert it to a dictionary keyed on the field name
func mapFields(columns []interface{}) map[string]FieldTypes {
	m := make(map[string]FieldTypes)

	for _, v := range columns {
		nv := v.(map[string]interface{})
		log.Printf("%v", v)
		nm := FieldTypes{
			Type:     nv["type"].(string),
			Optional: nv["optional"].(bool),
			Field:    nv["field"].(string),
		}

		if nv["name"] != nil {
			nm.Name = nv["name"].(string)
		}

		if nv["version"] != nil {
			nm.Version = nv["version"].(float64)
		}

		m[nm.Field] = nm
	}

	return m
}

// extractData converts attributes that need some attention such as the date.
func extractData(value interface{}, key string, fields map[string]FieldTypes) interface{} {
	if fields[key].Name == "io.debezium.time.Date" {
		return time.Unix(int64(86400*value.(float64)), 0)
	} else {
		return value
	}
}

func updateDatabase(event cloudevents.Event, db *sql.DB) error {
	extensions := event.Extensions()
	dbTable := fmt.Sprintf("%v", extensions["iodebeziumtable"])

	body := make(map[string]interface{})
	_ = event.DataAs(&body)

	schema := body["schema"].(map[string]interface{})
	outerFields := schema["fields"].([]interface{})
	innerFields := outerFields[0].(map[string]interface{})
	innerFieldElement := innerFields["fields"].([]interface{})

	fields := mapFields(innerFieldElement)

	payload := body["payload"].(map[string]interface{})
	var payloadBefore map[string]interface{}
	var payloadAfter map[string]interface{}
	var query string

	if payload["before"] != nil {
		payloadBefore = payload["before"].(map[string]interface{})
	}

	if payload["after"] != nil {
		payloadAfter = payload["after"].(map[string]interface{})
	}

	if payloadBefore != nil && payloadAfter == nil {
		// A DELETE operation
		query = "DELETE FROM " + dbTable + " WHERE "
		var values []interface{}

		l := len(payloadBefore)
		i := 1
		for k, v := range payloadBefore {
			query = fmt.Sprintf("%s %s = $%d", query, k, i)
			if i < l {
				query = query + " AND "
			}
			i++

			values = append(values, extractData(v, k, fields))
		}

		txn, err := db.Begin()
		if err != nil {
			return err
		}

		_, err = txn.Exec(query, values...)
		if err != nil {
			return err
		}

		txn.Commit()
		log.Printf("Executing delete...")
	} else if payloadBefore != nil && payloadAfter != nil {
		// An UPDATE operation
		query = "UPDATE " + dbTable + " SET "
		var newValues []interface{}
		var oldValues []interface{}

		l := len(payloadAfter)
		i := 1
		for k, v := range payloadAfter {
			query = fmt.Sprintf("%s %s = $%d", query, k, i)
			if i < l {
				query = query + ", "
			}
			i++
			newValues = append(newValues, extractData(v, k, fields))
		}

		query = fmt.Sprintf("%s WHERE ", query)
		l = len(payloadBefore)
		j := 0
		for k, v := range payloadBefore {
			query = fmt.Sprintf("%s %s = $%d", query, k, i+j)
			if j < l-1 {
				query = query + " AND "
			}
			j++
			oldValues = append(oldValues, extractData(v, k, fields))
		}

		txn, err := db.Begin()
		if err != nil {
			return err
		}

		_, err = txn.Exec(query, append(newValues, oldValues...)...)
		if err != nil {
			return err
		}

		txn.Commit()
		log.Printf("Executing update...")
	} else {
		// An INSERT operation
		query = "INSERT INTO " + dbTable + " ("
		var values []interface{}

		l := len(payloadAfter)
		i := 0
		for k, v := range payloadAfter {
			if i < l-1 {
				query = query + k + ", "
			} else {
				query = query + k
			}
			i++
			values = append(values, extractData(v, k, fields))
		}

		query = query + ") VALUES ("
		i = 1
		for i < l {
			query = fmt.Sprintf("%s $%d, ", query, i)
			i++
		}

		query = fmt.Sprintf("%s $%d)", query, i)

		txn, err := db.Begin()
		if err != nil {
			return err
		}

		_, err = txn.Exec(query, values...)
		if err != nil {
			return err
		}

		txn.Commit()
		log.Printf("Executing insert...")
	}

	return nil
}

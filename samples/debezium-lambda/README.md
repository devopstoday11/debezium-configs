## Overview

This is a sample lambda function to assist with capturing db2 based cloudevents
and apply the changes to a PostgreSQL table.

To find the proper mapping between DB2 types, and their resulting cloudevents,
you can consult the Debezium documentation at: https://debezium.io/documentation/reference/1.7/connectors/db2.html

## Building and using

The lambda function can be built using `go build .` to create the `debezium-lambda`
binary. Next, create a zip archive to package the binary using, `zip debezium-lambda.zip debezium-lambda`.

Lastly, upload the function using the AWS CLI. Here I am assuming a function entry
already exists in AWS:
```shell
$ aws lambda update-function-code --function-name debezium-lambda --zip-file fileb://debezium-lambda.zip
```

Now you have a functional lambda -> PostgreSQL bridge!
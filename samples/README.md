# Samples

These YAML files are designed to be invoked via `kubectl` against either your
cluster or the TriggerMesh SaaS cluster at https://cloud.triggermesh.io.

A quick overview of the files:
* [100-ext-services.yaml](100-ext-services.yaml) - The services that will act as a sink for your event data
* [200-db-registration.yaml](200-db-registration.yaml) - The TriggerMesh Koby Event Source Registration configuration
* [300-db2-lambda.yaml](300-db2-lambda.yaml) - The instantiated DB2 Debezium service sinking events to a lambda function
* [300-db2-sqs.yaml](300-db2-sqs.yaml) - The instantiated DB2 Debezium service sinking events to SQS
* [300-oracle-instance.yaml](300-oracle-instance.yaml) - The instantiated Oracle Debezium instance to Sockeye
* [300-oracle-sqs.yaml](300-oracle-sqs.yaml) - The instantiated Oracle Debezium instance to SQS

To use, apply the 100, 200, and one of the 300 based yaml files in sequence.
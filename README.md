# Building Debezium images

Three repos are required:

* https://github.com/debezium/debezium-connector-db2 - DB2 Database Connector
* https://github.com/triggermesh/debezium-configs - Debezium Docker images and db2 support configs
* https://github.com/triggermesh/debezium - Debezium core server with the updates to
  include support for the HTTP Webhook sink

When the repos are cloned, be sure to check out the `triggermesh` branch prior to building the components and images

## Note about the upstream version of the Debezium repos

The documentation available from [debezium](https://debezium.io) provides a collection of tutorials on how to setup and use Debezium with Apache Kafka. Additionally, the [page on DB2](https://debezium.io/documentation/reference/1.6/connectors/db2.html#setting-up-db2) provides the steps on how to configure DB2 to use Debezium.

Additionally, obtaining DB2 and the installation guide can be found at: https://www.ibm.com/docs/en/db2/11.5?topic=servers-db2-installation-methods

Lastly, Debezium does provide a docker image for db2 that they use to run integration tests against, and can be used as a reference.

## Build Requirements

Building Debezium requires jdk-8 and maven 3.8.  Additionally, building the db2 native components for the connector requires DB2 to be installed along with GCC and libaio.  Consult the Debezium DB2 installation guide for details on how to build it.

To build debezium and the debezium db2 adapter, run the following maven command:

    mvn clean install -Passembly -DskipITs -DskipTests

This will build debezium server and connect, and create the tarball archives that
will be required for creating the docker images.

Next, to build the Debezium server docker image (without the DB2 connector):
* Enter the `debezium-docker-images/server/1.7` directory
* copy the newly created tarball from `$(SRC)/debezium-server/debezium-server-dist/target/debezium-server-dist-1.7.0-SNAPSHOT.tar.gz`
* Build the docker image `docker build -t triggermesh-debezium/server:1.7 .`

Then, to include the DB2 Connector components in the docker image:
* Enter the `debezium-examples/tutorial/debezium-db2-init/db2connect` directory
* Copy the `debezium-connector-db2-1.7.0-SNAPSHOT.jar` file from the `debezium-connector-db2/target` directory. Note the filename may be different due to changes in the Debezium version number. Make sure you are using the 1.7.0 tag/branch
* Build the docker image `docker build --build-arg DEBEZIUM_VERSION=1.7 -t gcr.io/triggermesh-private/debezium-server:latest .`

## Optional Build: Including Kinesis

The AWS Kinesis sink should have been built as a part of the earlier build process, but is not included with the default distribution. To use it, be sure to copy the file from `$(SRC)/debezium-server/debezium-server-kinesis/target/debezium-server-kinesis-1.7.0-SNAPSHOT.jar` and uncomment the appropriate `COPY` line from the Dockerfile in `debezium-examples/tutorial/debezium-db2-init/db2connect`

# Running the Debezium Server image

Debezium (and by extension Quarkus as the underlying jee runtime framework) relies on [microprofile](https://microprofile.io/) for it's configuration. The advantage being that a traditional java properties file and environment variables can be used to configure the server. This allows Debezium to be runnable using something like [Koby](https://github.com/triggermesh/koby).

Running Debezium Server using docker, create a `conf` and `data` directory that can be mounted by docker. Note: in Kubernetes, the data directory should exist at a minimum with a volume claim.

## Running with Docker
In the `conf` directory, create a file called `application.properties`, and using the following as a template, update the database connection values to reflect the target DB2 instance.

```properties
debezium.sink.pravega.scope=empty
# This is the sink type to use: http for the webhook or kinesis for AWS Kinesis
debezium.sink.type=http
# The target URL for streaming events. This can be overridden by setting `K_SINK`
#debezium.sink.http.url=http://ptolamaios.cabnetworks.net:7070/rest/dump
# Defines the DB2 Connector as the connector to use
debezium.source.connector.class=io.debezium.connector.db2.Db2Connector
# The hostname or IP address of the DB2 server that is accessible to the Debezium service
debezium.source.database.hostname=triggermesh-db2
# The database port number
debezium.source.database.port=50001
# The database username that has access to the tables to watch
debezium.source.database.user=db2inst1
# The database user's password
debezium.source.database.password=changeme
# The database instance name
debezium.source.database.dbname=SAMPLE
# The database name that should be used for the event stream
debezium.source.database.server.name=sampleinst
# A whitelist of tables to include when streaming. When not set, this will default to all tables
debezium.source.table.include.list=DB2INST1.SALES
# This is used for tracking the last record debezium has processed and sent out
# The MemoryDatabaseHistory is persisted in memory and will disappear when the service is restarted
# The FileDatabaseHistory persists the tracking info in a flat file
#debezium.source.database.history=io.debezium.relational.history.MemoryDatabaseHistory
debezium.source.database.history=io.debezium.relational.history.FileDatabaseHistory
# Required when using FileDatabaseHistory, and should be a location that will be persisted
debezium.source.database.history.file.filename=/data/db2.history
# Also required when using FileDatabaseHistory for tracking which events have been sent
debezium.source.offset.storage.file.filename=/data/offsets.dat

# Undocumented attribute used to convert the Debezium change events to CloudEvents.
debezium.format.value=cloudevents # can be cloudevents or json

# Undocumented attribute for adjusting log level for debug purposes
#quarkus.log.level=DEBUG
```

The docker run command to use will also ensure that the aws credentials are passed along to the service for running Kinesis.

```sh
docker run --rm --it --name debezium-server -p 8083 \
  -v $PWD/conf:debezium/conf \
  -v $PWD/data:/data \
  -v $HOME/.aws:/home/jboss/.aws \
  -e K_SINK=http://localhost:8080 \
  gcr.io/triggermesh-private/debezium-server:latest
```

A runnable version of the Debezium server with the db2 connector can be found using docker.io/cab105/debezium-server:latest

## Running with Koby/Knative Service

A sample Koby configuration is included in the [samples](samples/) directory. Of note is the parameters can be passed into Debezium as environment variables. For example `debezium.source.database.dbname` becomes `DEBEZIUM_SOURCE_DATABASE_DBNAME`.


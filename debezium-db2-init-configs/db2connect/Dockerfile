ARG DEBEZIUM_VERSION
FROM debezium/server:$DEBEZIUM_VERSION

USER root

RUN microdnf install libaio
RUN microdnf install curl


# Deploy db2 client and drivers
RUN curl https://repo1.maven.org/maven2/com/ibm/db2/jcc/11.5.0.0/jcc-11.5.0.0.jar --output /debezium/lib/jcc-11.5.0.0.jar

# Copy debug version of connector (from github.com/debezium/debezium-connector-db2/target)
COPY debezium-connector-db2-1.7.0-SNAPSHOT.jar /debezium/lib/debezium-connector-db2-1.7.0-SNAPSHOT.jar

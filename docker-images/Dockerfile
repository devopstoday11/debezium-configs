FROM registry.access.redhat.com/ubi8/openjdk-11

LABEL maintainer="Debezium Community"

#
# Set the version, home directory, and MD5 hash.
#
ENV DEBEZIUM_VERSION=1.7.0.Alpha1 \
    SERVER_HOME=/debezium \
    MAVEN_REPO_CENTRAL="https://repo1.maven.org/maven2" \
    LD_LIBRARY_PATH="/instantclient_21_3"

#
# Create a directory for Debezium Server
#
USER root
RUN microdnf -y install gzip && \
    microdnf -y install unzip && \
    microdnf -y install curl && \
    microdnf clean all && \
    mkdir $SERVER_HOME && \
    chmod 755 $SERVER_HOME

#
# Change ownership and switch user
#
RUN chown -R jboss $SERVER_HOME && \
    chgrp -R jboss $SERVER_HOME
USER jboss

RUN mkdir $SERVER_HOME/conf && \
    mkdir $SERVER_HOME/data

#
# Download and install Debezium Server
#
COPY debezium-server-dist-1.7.0-SNAPSHOT.tar.gz /tmp/debezium.tar.gz

#
# Verify the contents and then install ...
#
RUN tar xzf /tmp/debezium.tar.gz -C $SERVER_HOME --strip-components 1

USER root
RUN rm -f /tmp/debezium.tar.gz

#
# Copy the pre-built connectors for db2
# (TODO: Ensure that the connector jar files are copied here prior to invoking docker build)
#
RUN curl https://repo1.maven.org/maven2/com/ibm/db2/jcc/11.5.0.0/jcc-11.5.0.0.jar --output /debezium/lib/jcc-11.5.0.0.jar
COPY debezium-connector-db2-1.7.0-SNAPSHOT.jar /debezium/lib/debezium-connector-db2-1.7.0-SNAPSHOT.jar

#
# Copy the pre-built connectors for oracle
# (NOTE: Due to licensing restrictions, you must have the instantclient downloaded prior to use)
#
COPY debezium-connector-oracle-1.7.0-SNAPSHOT.jar /debezium/lib/debezium-connector-oracle-1.7.0-SNAPSHOT.jar

COPY instantclient-basic-linux.x64-21.3.0.0.0.zip /tmp/instantclient-basic-linux.x64-21.3.0.0.0.zip
RUN unzip /tmp/instantclient-basic-linux.x64-21.3.0.0.0.zip -d / && \
    cp /instantclient_21_3/xstreams.jar /debezium/lib/xstreams.jar && \
    cp /instantclient_21_3/ojdbc8.jar /debezium/lib/ojdbc8.jar && \
    rm /tmp/instantclient-basic-linux.x64-21.3.0.0.0.zip

#
# Allow random UID to use Debezium Server
#
RUN chmod -R g+w,o+w $SERVER_HOME

# Set the working directory to the Debezium Server home directory
WORKDIR $SERVER_HOME

#
# Expose the ports and set up volumes for the data, transaction log, and configuration
#
EXPOSE 8080
VOLUME ["/debezium/conf","/debezium/data"]

CMD ["/debezium/run.sh"]

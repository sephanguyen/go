FROM debezium/kafka:1.9

USER root
RUN microdnf -y install unzip jq && microdnf clean all

USER kafka

RUN mkdir /kafka/etc && cd /kafka/etc &&\
    curl -so jmx_prometheus_javaagent.jar \
    https://repo1.maven.org/maven2/io/prometheus/jmx/jmx_prometheus_javaagent/0.15.0/jmx_prometheus_javaagent-0.15.0.jar

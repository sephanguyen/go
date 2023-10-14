
FROM confluentinc/cp-schema-registry:7.1.2

USER root

RUN mkdir /etc/monitoring && cd /etc/monitoring && \
    curl -so jmx_prometheus_javaagent.jar \
    https://repo1.maven.org/maven2/io/prometheus/jmx/jmx_prometheus_javaagent/0.15.0/jmx_prometheus_javaagent-0.15.0.jar

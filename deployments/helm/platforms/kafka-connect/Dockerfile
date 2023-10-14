FROM debezium/connect:1.9.6.Final

USER root
RUN microdnf -y install unzip jq && microdnf clean all

USER kafka
RUN curl -LO https://d1i4a15mxbxib1.cloudfront.net/api/plugins/confluentinc/kafka-connect-jdbc/versions/10.2.5/confluentinc-kafka-connect-jdbc-10.2.5.zip && \
    unzip ./confluentinc-kafka-connect-jdbc-10.2.5.zip -d ~/connect

RUN curl -LO https://d1i4a15mxbxib1.cloudfront.net/api/plugins/wepay/kafka-connect-bigquery/versions/2.1.9/wepay-kafka-connect-bigquery-2.1.9.zip && \
    unzip ./wepay-kafka-connect-bigquery-2.1.9.zip -d ~/connect

RUN curl -LO https://d1i4a15mxbxib1.cloudfront.net/api/plugins/confluentinc/kafka-connect-elasticsearch/versions/11.1.8/confluentinc-kafka-connect-elasticsearch-11.1.8.zip && \
    unzip ./confluentinc-kafka-connect-elasticsearch-11.1.8.zip -d ~/connect

RUN curl -LO https://d1i4a15mxbxib1.cloudfront.net/api/plugins/jcustenborder/kafka-connect-transform-common/versions/0.1.0.54/jcustenborder-kafka-connect-transform-common-0.1.0.54.zip  && \
    unzip ./jcustenborder-kafka-connect-transform-common-0.1.0.54.zip -d ~/connect

RUN curl -LO https://github.com/aiven/transforms-for-apache-kafka-connect/releases/download/v1.3.0/aiven-kafka-connect-transforms-1.3.0.zip && \
    unzip ./aiven-kafka-connect-transforms-1.3.0.zip -d ~/connect

RUN curl -LO https://d1i4a15mxbxib1.cloudfront.net/api/plugins/confluentinc/connect-transforms/versions/1.4.3/confluentinc-connect-transforms-1.4.3.zip && \
    unzip ./confluentinc-connect-transforms-1.4.3.zip -d ~/connect

RUN mkdir /kafka/etc && cd /kafka/etc &&\
    curl -so jmx_prometheus_javaagent.jar \
    https://repo1.maven.org/maven2/io/prometheus/jmx/jmx_prometheus_javaagent/0.15.0/jmx_prometheus_javaagent-0.15.0.jar



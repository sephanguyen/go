Deploy Kafka cluster

using image `debezium/kafka:1.8`

set env var `CLUSTER_ID="xxxx"`. It start Kafka in KRaft mode
set env var `NODE_ID`. We assign value `ordinal` from HOSTNAME
set env var `KAFKA_CONTROLLER_QUORUM_VOTERS`. We assign value `NODE_ID_1@CLUSTER_HOST_1:9093, NODE_ID_2@CLUSTER_HOST_2:9093`

use persistent volume to /kafka/data
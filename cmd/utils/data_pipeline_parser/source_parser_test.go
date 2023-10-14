package dplparser

import (
	"os"
	"strings"
	"testing"

	"go.uber.org/multierr"
)

var sourcetplsourceForTest, _ = os.ReadFile("./pipeline_source_template.txt")

func TestParseSource(t *testing.T) {
	t.Run("happy case", func(t *testing.T) {
		pipeLineConfig := `
envs: [local, stag, prod]
orgs: [e2e, manabie, jprep, aic, ga, renseikai, synersia, tokyo]
database: bob
schema: public
datapipelines:
- name: bob_to_entryexitmgmt_locations
  table: locations
  source:
    deployEnv: [local]
    deployOrg: [manabie]
  sinks:
  - database: entryexitmgmt
    schema: public
    table: locations
    captureDeleteAll: false
    deployEnv: [local]
    deployOrg: [manabie]
- name: bob_to_calendar_location_types
  table: location_types
  source:
    deployEnv: [local]
    deployOrg: [manabie]
  sinks:
  - database: calendar
    schema: public
    table: location_types
    captureDeleteAll: false
    deployEnv: [local]
    deployOrg: [manabie]
			`
		expectConfig := `
{
  "name": "local_manabie_bob_source_connector",
  "config": {
    "connector.class": "io.debezium.connector.postgresql.PostgresConnector",
    "database.password": "${file:/config/kafka-connect-config.properties:password}",
    "database.dbname": "${file:/config/kafka-connect-config.properties:bobdbname}",
    "database.hostname": "${file:/config/kafka-connect-config.properties:hostname}",
    "database.user": "${file:/config/kafka-connect-config.properties:user}",
    "database.port": "5432",
    "database.server.name": "bob",
    "database.sslmode": "disable",
    "plugin.name": "pgoutput",
    "tasks.max": "1",
    "key.converter":"io.confluent.connect.avro.AvroConverter",
    "key.converter.schema.registry.url":"http://cp-schema-registry:8081",
    "key.converter.schemas.enable": "false",
    "value.converter":"io.confluent.connect.avro.AvroConverter",
    "value.converter.schema.registry.url":"http://cp-schema-registry:8081",
    "value.converter.schemas.enable": "false",
    "slot.name": "local_manabie_bob",
    "slot.drop.on.stop": "false",
    "publication.autocreate.mode": "disabled",
    "publication.name": "debezium_publication",
    "snapshot.mode":"never",
    "tombstones.on.delete": "true",
    "heartbeat.interval.ms": "20000",
    "producer.max.request.size": "10485760",
    "schema.include.list": "public",
    "table.include.list": "public.dbz_signals,public.location_types,public.locations",
    "signal.data.collection": "public.dbz_signals",
    "time.precision.mode": "connect",
    "decimal.handling.mode": "double",
    "incremental.snapshot.chunk.size": "512",
    "topic.creation.default.replication.factor": "-1",
    "topic.creation.default.partitions": "10",
    "topic.creation.default.cleanup.policy": "compact",
    "topic.creation.default.compression.type": "lz4",
    "topic.creation.default.segment.bytes": "16777216",
    "topic.creation.default.delete.retention.ms": "6000",
    "transforms": "route",
    "transforms.route.type": "org.apache.kafka.connect.transforms.RegexRouter",
    "transforms.route.regex": "([^.]+).([^.]+).([^.]+)",
    "transforms.route.replacement": "local.manabie.$1.$2.$3"
  }
}
		`
		pipeLineConfig = strings.TrimSpace(pipeLineConfig)

		dpl, err := NewDataPipelineParser("",
			WithTpl(string(sourcetplsourceForTest)),
			WithTableSchemaDir(tableSchemaDirForTest),
			WithCustomReader(func(fileName string) ([]byte, error) {
				return []byte(pipeLineConfig), nil
			}),
		)
		if err != nil {
			t.Error(err)
			return
		}

		result, err := dpl.ParseSource()
		if err != nil {
			t.Errorf("failed to parse data pipeline: %v", err)
			return
		}

		expect := strings.TrimSpace(expectConfig)
		connectorConfig := ConnectorConfig{"bob_source.json", "manabie/local"}
		res := strings.TrimSpace(result[connectorConfig])

		err = multierr.Combine(
			validateConnector(expect, res),
		)

		if err != nil {
			t.Error(err)
		}
	})

}

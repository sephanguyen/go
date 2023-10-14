package hephaestus

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/manabie-com/backend/cmd/server/hephaestus"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/try"
	"github.com/manabie-com/backend/internal/hephaestus/configurations"
)

func (s *suite) tableInDatabaseBobAndFatima(ctx context.Context, tableName string) (context.Context, error) {
	s.TableMetaData = nextTable(tableName)
	err := s.TableMetaData.CreateTableInSourceAndSink(ctx, s.BobPostgresDBTrace, s.FatimaDBTrace)
	if err != nil {
		return ctx, err
	}
	return ctx, nil
}

func (s *suite) insertSeveralRecordsToTable(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ids, err := s.TableMetaData.GenerateSampleRecordsInSource(ctx, s.BobPostgresDB)
	if err != nil {
		return ctx, err
	}

	stepState.TestDebeziumJobRecordIDs = ids
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) recordsIsSyncedInSourceAndSinkTable(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ids := stepState.TestDebeziumJobRecordIDs
	err := try.Do(func(attempt int) (retry bool, err error) {
		defer func() {
			if err != nil {
				time.Sleep(3 * time.Second)
			}
		}()
		es, err := s.TableMetaData.GetSampleRecordsInSink(ctx, s.FatimaDBTrace, ids)
		if err != nil {
			return true, fmt.Errorf("s.TableMetaData.GetSampleRecordsInSink: %w", err)
		}

		if len(es) != len(ids) {
			return true, fmt.Errorf("expect continue to sync data table test_debezium from bob to fatima")
		}
		return false, nil
	})
	if err != nil {
		return ctx, err
	}

	return ctx, nil
}

func (s *suite) sinkConnectorFileForTableInFatima(ctx context.Context, tableName string) (context.Context, error) {
	topic := fmt.Sprintf("local.manabie.bob.public.%s", tableName)
	sinkCfg := fmt.Sprintf(`
{
  "name": "test_bob_to_fatima_%s_sink_connector",
  "config": {
    "connector.class": "io.confluent.connect.jdbc.JdbcSinkConnector",
    "tasks.max": "1",
    "topics": "%s",
    "connection.url": "${file:/config/kafka-connect-config.properties:fatima_url}",
    "transforms": "unwrap,route",
    "transforms.unwrap.type": "io.debezium.transforms.ExtractNewRecordState",
    "transforms.unwrap.drop.tombstones": "true",
    "transforms.unwrap.delete.handling.mode": "drop",
    "transforms.route.type": "org.apache.kafka.connect.transforms.RegexRouter",
    "transforms.route.regex": "([^.]+).([^.]+).([^.]+).([^.]+).([^.]+)",
    "transforms.route.replacement": "$5",
	"table.name.format": "%s",
    "auto.create": "false",
    "insert.mode": "upsert",
    "delete.enabled": "false",
    "pk.mode": "record_value",
    "pk.fields": "id",
    "fields.whitelist": "id,a,b,c,d"
  }
}
`, tableName, topic, tableName)

	if stat, err := os.Stat(s.SinkConnectorDir); err == nil && stat.IsDir() {
		fileName := fmt.Sprintf("test_bob_to_fatima_%s_sink_connector.json", tableName)
		err := os.WriteFile(filepath.Join(s.SinkConnectorDir, fileName), []byte(sinkCfg), fs.FileMode(0o666))
		if err != nil {
			return ctx, err
		}
	} else {
		return ctx, fmt.Errorf("cannot write sink connector config")
	}

	return ctx, nil
}

func (s *suite) sourceConnectorFileForTableInBob(ctx context.Context, tableName string) (context.Context, error) {
	name := fmt.Sprintf("local_manabie_bob_%s_source_connector", tableName)
	databaseServerName := fmt.Sprintf("bob_%s", tableName)
	slotName := fmt.Sprintf("local_manabie_bob_%s", tableName)
	sourceCfg := fmt.Sprintf(`
{
  "name": "%s",
  "config": {
    "connector.class": "io.debezium.connector.postgresql.PostgresConnector",
    "database.password": "${file:/config/kafka-connect-config.properties:password}",
    "database.dbname": "${file:/config/kafka-connect-config.properties:bobdbname}",
    "database.hostname": "${file:/config/kafka-connect-config.properties:hostname}",
    "database.user": "${file:/config/kafka-connect-config.properties:user}",
    "database.port": "5432",
    "database.server.name": "%s",
    "database.sslmode": "disable",
    "plugin.name": "pgoutput",
    "tasks.max": "1",
    "key.converter":"io.confluent.connect.avro.AvroConverter",
    "key.converter.schema.registry.url":"http://cp-schema-registry:8081",
    "key.converter.schemas.enable": "false",
    "value.converter":"io.confluent.connect.avro.AvroConverter",
    "value.converter.schema.registry.url":"http://cp-schema-registry:8081",
    "value.converter.schemas.enable": "false",
    "slot.name": "%s",
    "slot.drop.on.stop": "true",
    "publication.autocreate.mode": "disabled",
    "publication.name": "debezium_publication",
    "snapshot.mode":"initial",
	"tombstones.on.delete": "true",
	"heartbeat.interval.ms": "20000",
	"heartbeat.action.query": "SELECT 1",
	"producer.max.request.size": "10485760",
    "schema.include.list": "public",
    "table.include.list": "public.dbz_signals,public.%s",
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
    "transforms.route.replacement": "local.manabie.bob.$2.$3"
  }
}`, name, databaseServerName, slotName, tableName)

	if stat, err := os.Stat(s.SourceConnectorDir); err == nil && stat.IsDir() {
		s.SourceConnectorFileName = fmt.Sprintf("test_bob_%s.json", idutil.ULIDNow())
		err := os.WriteFile(filepath.Join(s.SourceConnectorDir, s.SourceConnectorFileName), []byte(sourceCfg), fs.FileMode(0o666))
		if err != nil {
			return ctx, err
		}
	} else {
		return ctx, fmt.Errorf("cannot write source connector config")
	}

	return ctx, nil
}

func (s *suite) runJobUpsertKafkaConnector(ctx context.Context) (context.Context, error) {
	hephaestus.SendIncrementalSnapshot = true
	rsc := bootstrap.NewResources().WithNATS(s.JSM).WithDatabase(map[string]*database.DBTrace{"bob_tc": {DB: s.BobDB}})
	defer rsc.Cleanup() //nolint:errcheck
	err := hephaestus.RunUpsertKafkaConnect(ctx, configurations.Config{
		Common: configs.CommonConfig{
			Environment: s.Cfg.Common.Environment,
		},
		Kafka: configs.KafkaConfig{
			Addr: s.Cfg.Kafka.Addr,
			Connect: configs.KafkaConnectConfig{
				Addr:             s.Cfg.Kafka.Connect.Addr,
				SourceConfigDir:  s.SourceConnectorDir,
				SinkConfigDir:    s.SinkConnectorDir,
				GenSinkConfigDir: s.SinkConnectorDir,
			},
		},
		PostgresV2: configs.PostgresConfigV2{
			Databases: map[string]configs.PostgresDatabaseConfig{
				"bob_tc": {},
			},
		},
	}, rsc,
	)
	if err != nil {
		return ctx, err
	}
	time.Sleep(5 * time.Second)
	return ctx, nil
}

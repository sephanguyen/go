package hephaestus

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/try"

	"github.com/go-kafka/connect"
	"github.com/jackc/pgtype"
)

func (s *suite) createConnector(client *connect.Client, newConnector *connect.Connector) error {
	oldConnector := s.GetConnector(client, newConnector.Name)

	if oldConnector == nil {
		resp, err := client.CreateConnector(newConnector)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
	} else {
		_, resp, err := client.UpdateConnectorConfig(newConnector.Name, newConnector.Config)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
	}

	err := try.Do(func(attempt int) (retry bool, err error) {
		// only check replication slot name for source debezium connector
		repName, ok := newConnector.Config["slot.name"]
		if !s.connectorIsReady(client, newConnector.Name) || (ok && !s.replicationSlotIsActive(repName)) {
			time.Sleep(500 * time.Millisecond)
			return true, fmt.Errorf("connector %s is not ready", newConnector.Name)
		}
		return false, nil
	})

	if err != nil {
		return err
	}
	return nil
}

func (s *suite) connectorIsReady(client *connect.Client, name string) bool {
	status, resp, err := client.GetConnectorStatus(name)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	ready := status.Connector.State == "RUNNING"

	for _, task := range status.Tasks {
		ready = ready && (task.State == "RUNNING")
	}

	return ready
}

func (s *suite) replicationSlotIsActive(repName string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	sql := `SELECT active FROM pg_replication_slots WHERE slot_name=$1`
	var active pgtype.Bool
	err := s.BobPostgresDBTrace.QueryRow(ctx, sql, database.Text(repName)).Scan(&active)
	if err != nil {
		return false
	}
	return active.Bool
}

func (s *suite) createDebeziumSourceConnectorForThatTableInBob(ctx context.Context, tableName string) (context.Context, error) {
	name := fmt.Sprintf("local_manabie_bob_%s_source_connector", tableName)
	databaseServerName := fmt.Sprintf("bob_%s", tableName)
	slotName := fmt.Sprintf("local_manabie_bob_%s", tableName)
	srcCfg := fmt.Sprintf(`{
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
}
	`, name, databaseServerName, slotName, tableName)

	stepState := StepStateFromContext(ctx)
	client := connect.NewClient(s.Cfg.KafkaConnectConfig.Addr)
	stepState.ConnectClient = client

	srcConnector := &stepState.SrcConnector

	err := json.Unmarshal([]byte(srcCfg), &srcConnector)
	if err != nil {
		return ctx, err
	}

	srcConnector.Config["slot.drop.on.stop"] = "true"
	err = s.createConnector(client, srcConnector)
	if err != nil {
		return ctx, err
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createSinkConnectorForThatTableInFatima(ctx context.Context, tableName string) (context.Context, error) {
	topic := fmt.Sprintf("local.manabie.bob.public.%s", tableName)
	sinkCfg := fmt.Sprintf(`
{
  "name": "local_bob_to_fatima_%s_sink_connector",
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
    "auto.create": "false",
    "insert.mode": "upsert",
    "delete.enabled": "false",
    "pk.mode": "record_value",
    "pk.fields": "id",
    "fields.whitelist": "id,a,b,c,d"
  }
}
	`, tableName, topic)
	stepState := StepStateFromContext(ctx)
	client := stepState.ConnectClient

	sinkConnector := &stepState.SinkConnector
	err := json.Unmarshal([]byte(sinkCfg), &sinkConnector)
	if err != nil {
		return ctx, err
	}

	err = s.createConnector(client, sinkConnector)
	if err != nil {
		return ctx, err
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) GetConnector(client *connect.Client, name string) *connect.Connector {
	connector, resp, err := client.GetConnector(name)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	return connector
}

func (s *suite) deleteDebeziumSourceConnector(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	client := stepState.ConnectClient
	name := stepState.SrcConnector.Name
	resp, err := client.DeleteConnector(name)
	if err != nil {
		return ctx, err
	}
	defer resp.Body.Close()

	cn := s.GetConnector(client, name)
	// wait until connector is deleted
	for cn != nil {
		time.Sleep(100 * time.Millisecond)
		cn = s.GetConnector(client, name)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theDataInsertBeforeWillNotBeSynced(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ids := stepState.TestDebeziumRecordIDs

	numRetry := 10
	for numRetry > 0 {
		es, err := s.TableMetaData.GetSampleRecordsInSink(ctx, s.FatimaDBTrace, ids)
		if err != nil {
			return ctx, fmt.Errorf("s.TableMetaData.GetSampleRecordsInSink: %w", err)
		}

		if len(es) != 0 {
			return ctx, fmt.Errorf("expect not sync data table test_debezium from bob to fatima")
		}
		numRetry--
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theDataWillBeSynced(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ids := stepState.TestDebeziumRecordIDs
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

package dplparser

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/go-kafka/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/multierr"
)

var tableSchemaDirForTest = "../../../mock/testing/testdata"
var tplForTest, _ = os.ReadFile("./pipeline_template.txt")

func TestParse(t *testing.T) {
	t.Run("happy case", func(t *testing.T) {
		pipeLineConfig := `
database: bob
schema: public
datapipelines:
- name: bob_to_entryexitmgmt_locations
  table: locations
  deployEnv: [local]
  deployOrg: [manabie]
  source:
    deployEnv: [local, stag, prod]
    deployOrg: [e2e, manabie, jprep, aic, ga, renseikai, synersia, tokyo]
  sinks:
  - database: entryexitmgmt
    name: bob_to_entryexitmgmt_locations_sink_connector
    schema: publicbob_to_entryexitmgmt_locations_sink_connector
    deployEnv: [local, stag, prod]
    deployOrg: [e2e, manabie, jprep, aic, ga, renseikai, synersia, tokyo]
    captureDeleteAll: false
			`
		expectConfig := `
{
	"name": "local_manabie_bob_to_entryexitmgmt_locations_sink_connector",
	"config": {
	"connector.class": "io.confluent.connect.jdbc.JdbcSinkConnector",
	"tasks.max": "1",
	"topics": "local.manabie.bob.public.locations",
	"connection.url": "${file:/config/kafka-connect-config.properties:entryexitmgmt_url}",
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
	"pk.fields": "location_id",
   "fields.whitelist": "location_id,name,created_at,updated_at,deleted_at,resource_path,location_type,partner_internal_id,partner_internal_parent_id,parent_location_id,is_archived,access_path"
	}
}
		`
		pipeLineConfig = strings.TrimSpace(pipeLineConfig)

		dpl, err := NewDataPipelineParser("",
			WithTpl(string(tplForTest)),
			WithTableSchemaDir(tableSchemaDirForTest),
			WithCustomReader(func(fileName string) ([]byte, error) {
				return []byte(pipeLineConfig), nil
			}))
		if err != nil {
			t.Error(err)
			return
		}

		result, err := dpl.Parse()
		if err != nil {
			t.Errorf("failed to parse data pipeline: %v", err)
			return
		}

		expect := strings.TrimSpace(expectConfig)
		connectorConfig := ConnectorConfig{"bob_to_entryexitmgmt_locations.json", "manabie/local"}
		res := strings.TrimSpace(result[connectorConfig])

		err = multierr.Combine(
			validateConnector(expect, res),
		)

		if err != nil {
			t.Error(err)
		}
	})

}

func TestParseWithDeleteEnable(t *testing.T) {
	t.Run("sync table with delete enabled", func(t *testing.T) {
		pipeLineConfig := `

database: bob
schema: public
datapipelines:
- name: bob_to_calendar_location_types
  table: location_types
  source:
    deployEnv: [local, stag, prod]
    deployOrg: [e2e, manabie, jprep, aic, ga, renseikai, synersia, tokyo]
  sinks:
  - database: calendar
    name: bob_to_calendar_location_types_sink_connector
    schema: public
    deployEnv: [local, stag, prod]
    deployOrg: [e2e, manabie, jprep, aic, ga, renseikai, synersia, tokyo]
    captureDeleteAll: true
	`
		expectConfig := `
	{
		"name": "local_manabie_bob_to_calendar_location_types_sink_connector",
		"config": {
			"connector.class": "io.confluent.connect.jdbc.JdbcSinkConnector",
			"tasks.max": "1",
			"topics": "local.manabie.bob.public.location_types",
			"connection.url": "${file:/config/kafka-connect-config.properties:calendar_url}",
			"transforms": "unwrap,route",
			"transforms.unwrap.type": "io.debezium.transforms.ExtractNewRecordState",
			"transforms.unwrap.drop.tombstones": "false",
			"transforms.unwrap.delete.handling.mode": "none",
			"transforms.route.type": "org.apache.kafka.connect.transforms.RegexRouter",
			"transforms.route.regex": "([^.]+).([^.]+).([^.]+).([^.]+).([^.]+)",
			"transforms.route.replacement": "$5",
			"auto.create": "false",
			"insert.mode": "upsert",
			"delete.enabled": "true",
			"pk.mode": "record_key",
			"fields.whitelist": "location_type_id,name,display_name,parent_name,parent_location_type_id,updated_at,created_at,deleted_at,resource_path,is_archived"
		}
	}

			`
		pipeLineConfig = strings.TrimSpace(pipeLineConfig)
		dpl, err := NewDataPipelineParser("",
			WithTpl(string(tplForTest)),
			WithTableSchemaDir(tableSchemaDirForTest),
			WithCustomReader(func(fileName string) ([]byte, error) {
				return []byte(pipeLineConfig), nil
			}))
		if err != nil {
			t.Error(err)
			return
		}

		result, err := dpl.Parse()
		if err != nil {
			t.Errorf("failed to parse data pipeline: %v", err)
			return
		}

		expect := strings.TrimSpace(expectConfig)
		connectorConfig := ConnectorConfig{"bob_to_calendar_location_types.json", "manabie/local"}
		res := strings.TrimSpace(result[connectorConfig])
		err = multierr.Combine(
			validateConnector(expect, res),
		)

		if err != nil {
			t.Error(err)
		}
	})

	t.Run("sync table with delete enabled on envs", func(t *testing.T) {
		pipeLineConfig := `

database: bob
schema: public
datapipelines:
- name: bob_to_calendar_location_types
  table: location_types
  source:
    deployEnv: [local, stag, prod]
    deployOrg: [e2e, manabie, jprep, aic, ga, renseikai, synersia, tokyo]
  sinks:
  - database: calendar
    name: bob_to_calendar_location_types_sink_connector
    schema: public
    deployEnv: [local, stag, prod]
    deployOrg: [e2e, manabie, jprep, aic, ga, renseikai, synersia, tokyo]
    captureDeleteEnvs: [local, stag]
`
		expectConfig := `

{
	"name": "local_manabie_bob_to_calendar_location_types_sink_connector",
	"config": {
		"connector.class": "io.confluent.connect.jdbc.JdbcSinkConnector",
		"tasks.max": "1",
		"topics": "local.manabie.bob.public.location_types",
		"connection.url": "${file:/config/kafka-connect-config.properties:calendar_url}",
		"transforms": "unwrap,route",
		"transforms.unwrap.type": "io.debezium.transforms.ExtractNewRecordState",
		"transforms.unwrap.drop.tombstones": "false",
		"transforms.unwrap.delete.handling.mode": "none",
		"transforms.route.type": "org.apache.kafka.connect.transforms.RegexRouter",
		"transforms.route.regex": "([^.]+).([^.]+).([^.]+).([^.]+).([^.]+)",
		"transforms.route.replacement": "$5",
		"auto.create": "false",
		"insert.mode": "upsert",
		"delete.enabled": "true",
		"pk.mode": "record_key",
		"fields.whitelist": "location_type_id,name,display_name,parent_name,parent_location_type_id,updated_at,created_at,deleted_at,resource_path,is_archived"
	}
}

		`
		pipeLineConfig = strings.TrimSpace(pipeLineConfig)
		dpl, err := NewDataPipelineParser("",
			WithTpl(string(tplForTest)),
			WithTableSchemaDir(tableSchemaDirForTest),
			WithCustomReader(func(fileName string) ([]byte, error) {

				return []byte(pipeLineConfig), nil
			}))
		if err != nil {
			t.Error(err)
			return
		}

		result, err := dpl.Parse()
		if err != nil {
			t.Errorf("failed to parse data pipeline: %v", err)
			return
		}

		expect := strings.TrimSpace(expectConfig)
		connectorConfig := ConnectorConfig{"bob_to_calendar_location_types.json", "manabie/local"}
		res := strings.TrimSpace(result[connectorConfig])
		err = multierr.Combine(
			validateConnector(expect, res),
		)

		if err != nil {
			t.Error(err)
		}
	})
}

func TestParseDeployEnv(t *testing.T) {
	t.Run("specify deploy env and org", func(t *testing.T) {
		pipeLineConfig := `
deployEnv: [local, stag, prod]
deployOrg: [e2e, manabie, jprep, aic, ga, renseikai, synersia, tokyo]
database: bob
schema: public
datapipelines:
- name: bob_to_entryexitmgmt_granted_role_access_path
  table: granted_role_access_path
  deployEnv: [prod]
  deployOrg: [jprep]
  source:
    deployEnv: [local, stag, prod]
    deployOrg: [e2e, manabie, jprep, aic, ga, renseikai, synersia, tokyo]
  sinks:
  - database: entryexitmgmt
    name: bob_to_entryexitmgmt_granted_role_access_path
    schema: public
    deployEnv: [local, stag, prod]
    deployOrg: [e2e, manabie, jprep, aic, ga, renseikai, synersia, tokyo]
    captureDeleteAll: false

				`

		expectConfig := `
	{
		"name": "prod_jprep_bob_to_entryexitmgmt_granted_role_access_path",
		"config": {
			"connector.class": "io.confluent.connect.jdbc.JdbcSinkConnector",
			"tasks.max": "1",
			"topics": "prod.jprep.bob.public.granted_role_access_path",
			"connection.url": "${file:/config/kafka-connect-config.properties:entryexitmgmt_url}",
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
			"pk.fields": "granted_role_id,location_id",
			"fields.whitelist": "granted_role_id,location_id,created_at,updated_at,deleted_at,resource_path"
		}
	}

`
		pipeLineConfig = strings.TrimSpace(pipeLineConfig)
		dpl, err := NewDataPipelineParser("",
			WithTpl(string(tplForTest)),
			WithTableSchemaDir(tableSchemaDirForTest),
			WithCustomReader(func(fileName string) ([]byte, error) {
				return []byte(pipeLineConfig), nil
			}))
		if err != nil {
			t.Error(err)
			return
		}

		result, err := dpl.Parse()
		if err != nil {
			t.Errorf("failed to parse data pipeline: %v", err)
			return
		}

		expect := strings.TrimSpace(expectConfig)
		connectorConfig := ConnectorConfig{"bob_to_entryexitmgmt_granted_role_access_path.json", "jprep/prod"}
		res := strings.TrimSpace(result[connectorConfig])
		err = multierr.Combine(
			validateConnector(expect, res),
		)

		if err != nil {
			t.Error(err)
		}
	})

	t.Run("not define deploy env and org, it mean all env and org", func(t *testing.T) {
		pipeLineConfig := `

database: bob
schema: public
datapipelines:
- name: bob_to_entryexitmgmt_granted_role_access_path
  table: granted_role_access_path
  source:
    table: granted_role_access_path
    deployEnv: [prod]
    deployOrg: [jprep]
  sinks:
  - database: entryexitmgmt
    name: bob_to_entryexitmgmt_granted_role_access_path
    schema: public
    deployEnv: [prod]
    deployOrg: [jprep]
    captureDeleteAll: false
			`

		expectConfig := `
{
	"name": "prod_jprep_bob_to_entryexitmgmt_granted_role_access_path",
	"config": {
		"connector.class": "io.confluent.connect.jdbc.JdbcSinkConnector",
		"tasks.max": "1",
		"topics": "prod.jprep.bob.public.granted_role_access_path",
		"connection.url": "${file:/config/kafka-connect-config.properties:entryexitmgmt_url}",
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
		"pk.fields": "granted_role_id,location_id",
		"fields.whitelist": "granted_role_id,location_id,created_at,updated_at,deleted_at,resource_path"
	}
}
		`
		pipeLineConfig = strings.TrimSpace(pipeLineConfig)
		dpl, err := NewDataPipelineParser("",
			WithTpl(string(tplForTest)),
			WithTableSchemaDir(tableSchemaDirForTest),
			WithCustomReader(func(fileName string) ([]byte, error) {
				return []byte(pipeLineConfig), nil
			}))
		if err != nil {
			t.Error(err)
			return
		}

		result, err := dpl.Parse()
		if err != nil {
			t.Errorf("failed to parse data pipeline: %v", err)
			return
		}

		expect := strings.TrimSpace(expectConfig)

		connectorConfig := ConnectorConfig{"bob_to_entryexitmgmt_granted_role_access_path.json", "jprep/prod"}
		res := strings.TrimSpace(result[connectorConfig])
		err = multierr.Combine(
			validateConnector(expect, res),
		)

		if err != nil {
			t.Error(err)
		}
	})

	t.Run("specify deploy env, org and schema", func(t *testing.T) {
		pipeLineConfig := `
deployEnv: [local, stag, prod]
deployOrg: [e2e, manabie, jprep, aic, ga, renseikai, synersia, tokyo]
database: bob
schema: public
datapipelines:
- name: bob_to_entryexitmgmt_granted_role_access_path
  table: granted_role_access_path
  deployEnv: [prod]
  deployOrg: [jprep]
  source:
    deployEnv: [local, stag, prod]
    deployOrg: [e2e, manabie, jprep, aic, ga, renseikai, synersia, tokyo]
  sinks:
  - database: entryexitmgmt
    name: bob_to_entryexitmgmt_granted_role_access_path
    schema: public
    deployEnv: [local, stag, prod]
    deployOrg: [e2e, manabie, jprep, aic, ga, renseikai, synersia, tokyo]
    captureDeleteAll: false
    deploySchema: [public]

				`

		expectConfig := `
	{
		"name": "prod_jprep_bob_to_entryexitmgmt_granted_role_access_path",
		"config": {
			"connector.class": "io.confluent.connect.jdbc.JdbcSinkConnector",
			"tasks.max": "1",
			"topics": "prod.jprep.bob.public.granted_role_access_path",
			"connection.url": "${file:/config/kafka-connect-config.properties:entryexitmgmt_url}",
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
			"pk.fields": "granted_role_id,location_id",
			"fields.whitelist": "granted_role_id,location_id,created_at,updated_at,deleted_at,resource_path"
		}
	}

`
		pipeLineConfig = strings.TrimSpace(pipeLineConfig)
		dpl, err := NewDataPipelineParser("",
			WithTpl(string(tplForTest)),
			WithTableSchemaDir(tableSchemaDirForTest),
			WithCustomReader(func(fileName string) ([]byte, error) {
				return []byte(pipeLineConfig), nil
			}))
		if err != nil {
			t.Error(err)
			return
		}

		result, err := dpl.Parse()
		if err != nil {
			t.Errorf("failed to parse data pipeline: %v", err)
			return
		}

		expect := strings.TrimSpace(expectConfig)
		connectorConfig := ConnectorConfig{"bob_to_entryexitmgmt_granted_role_access_path.json", "jprep/prod"}
		res := strings.TrimSpace(result[connectorConfig])
		err = multierr.Combine(
			validateConnector(expect, res),
		)

		if err != nil {
			t.Error(err)
		}
	})

	t.Run("specify deploy env, org and schema, filterResourcePath", func(t *testing.T) {
		pipeLineConfig := `
deployEnv: [local, stag, prod]
deployOrg: [e2e, manabie, jprep, aic, ga, renseikai, synersia, tokyo]
database: bob
schema: public
datapipelines:
- name: bob_to_entryexitmgmt_granted_role_access_path
  table: granted_role_access_path
  deployEnv: [prod]
  deployOrg: [jprep]
  source:
    deployEnv: [local, stag, prod]
    deployOrg: [e2e, manabie, jprep, aic, ga, renseikai, synersia, tokyo]
  sinks:
  - database: entryexitmgmt
    name: bob_to_entryexitmgmt_granted_role_access_path
    schema: public
    deployEnv: [local, stag, prod]
    deployOrg: [e2e, manabie, jprep, aic, ga, renseikai, synersia, tokyo]
    captureDeleteAll: false
    deploySchema: [public]
    filterResourcePath: 100

				`

		expectConfig := `
	{
		"name": "prod_jprep_bob_to_entryexitmgmt_granted_role_access_path",
		"config": {
			"connector.class": "io.confluent.connect.jdbc.JdbcSinkConnector",
			"tasks.max": "1",
			"topics": "prod.jprep.bob.public.granted_role_access_path",
			"connection.url": "${file:/config/kafka-connect-config.properties:entryexitmgmt_url}",
			"transforms": "unwrap,route,filterResourcePath",
			"transforms.filterResourcePath.type" : "io.confluent.connect.transforms.Filter$Value",
			"transforms.filterResourcePath.filter.condition": "$[?(@.resource_path =~ /100/)]",
			"transforms.filterResourcePath.filter.type" : "include",
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
			"pk.fields": "granted_role_id,location_id",
			"fields.whitelist": "granted_role_id,location_id,created_at,updated_at,deleted_at,resource_path"
		}
	}

`
		pipeLineConfig = strings.TrimSpace(pipeLineConfig)
		dpl, err := NewDataPipelineParser("",
			WithTpl(string(tplForTest)),
			WithTableSchemaDir(tableSchemaDirForTest),
			WithCustomReader(func(fileName string) ([]byte, error) {
				return []byte(pipeLineConfig), nil
			}))
		if err != nil {
			t.Error(err)
			return
		}

		result, err := dpl.Parse()
		if err != nil {
			t.Errorf("failed to parse data pipeline: %v", err)
			return
		}

		expect := strings.TrimSpace(expectConfig)
		connectorConfig := ConnectorConfig{"bob_to_entryexitmgmt_granted_role_access_path.json", "jprep/prod"}
		res := strings.TrimSpace(result[connectorConfig])
		err = multierr.Combine(
			validateConnector(expect, res),
		)

		if err != nil {
			t.Error(err)
		}
	})
}

func TestParseWithExcludeColumn(t *testing.T) {
	t.Run("parse with exclude column field", func(t *testing.T) {
		pipeLineConfig := `

database: bob
schema: public
datapipelines:
- name: bob_to_entryexitmgmt_granted_role_access_path
  table: granted_role_access_path
  source:
    deployEnv: [local]
    deployOrg: [manabie]
  sinks:
  - database: entryexitmgmt
    name: bob_to_entryexitmgmt_granted_role_access_path_sink_connector
    schema: public
    excludeColumns: [created_at, updated_at, deleted_at]
    captureDeleteAll: false
    deployEnv: [local]
    deployOrg: [manabie]
			`

		expectConfig := `
{
	"name": "local_manabie_bob_to_entryexitmgmt_granted_role_access_path_sink_connector",
	"config": {
		"connector.class": "io.confluent.connect.jdbc.JdbcSinkConnector",
		"tasks.max": "1",
		"topics": "local.manabie.bob.public.granted_role_access_path",
		"connection.url": "${file:/config/kafka-connect-config.properties:entryexitmgmt_url}",
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
		"pk.fields": "granted_role_id,location_id",
		"fields.whitelist": "granted_role_id,location_id,resource_path"
	}
}
		`
		pipeLineConfig = strings.TrimSpace(pipeLineConfig)
		dpl, err := NewDataPipelineParser("",
			WithTpl(string(tplForTest)),
			WithTableSchemaDir(tableSchemaDirForTest),
			WithCustomReader(func(fileName string) ([]byte, error) {
				return []byte(pipeLineConfig), nil
			}))
		if err != nil {
			t.Error(err)
			return
		}

		result, err := dpl.Parse()
		if err != nil {
			t.Errorf("failed to parse data pipeline: %v", err)
			return
		}

		expect := strings.TrimSpace(expectConfig)
		connectorConfig := ConnectorConfig{"bob_to_entryexitmgmt_granted_role_access_path.json", "manabie/local"}
		res := strings.TrimSpace(result[connectorConfig])
		err = multierr.Combine(
			validateConnector(expect, res),
		)

		if err != nil {
			t.Error(err)
		}
	})

	t.Run("parse with exclude column field with not existed column", func(t *testing.T) {
		pipeLineConfig := `

database: bob
schema: public
datapipelines:
- name: bob_to_entryexitmgmt_granted_role_access_path
  table: granted_role_access_path
  source:
    deployEnv: [local, stag, prod]
    deployOrg: [e2e, manabie, jprep, aic, ga, renseikai, synersia, tokyo]
  sinks:
  - database: entryexitmgmt
    name: bob_to_entryexitmgmt_granted_role_access_path_sink_connector
    schema: public
    deployEnv: [local, stag, prod]
    deployOrg: [e2e, manabie, jprep, aic, ga, renseikai, synersia, tokyo]
    excludeColumns: [foo, bar]
    captureDeleteAll: false
			`
		expectConfig := `
{
	"name": "local_manabie_bob_to_entryexitmgmt_granted_role_access_path_sink_connector",
	"config": {
		"connector.class": "io.confluent.connect.jdbc.JdbcSinkConnector",
		"tasks.max": "1",
		"topics": "local.manabie.bob.public.granted_role_access_path",
		"connection.url": "${file:/config/kafka-connect-config.properties:entryexitmgmt_url}",
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
		"pk.fields": "granted_role_id,location_id",
		"fields.whitelist": "granted_role_id,location_id,created_at,updated_at,deleted_at,resource_path"
	}
}
		`
		pipeLineConfig = strings.TrimSpace(pipeLineConfig)
		dpl, err := NewDataPipelineParser("",
			WithTpl(string(tplForTest)),
			WithTableSchemaDir(tableSchemaDirForTest),
			WithCustomReader(func(fileName string) ([]byte, error) {
				return []byte(pipeLineConfig), nil
			}))
		if err != nil {
			t.Error(err)
			return
		}

		result, err := dpl.Parse()
		if err != nil {
			t.Errorf("failed to parse data pipeline: %v", err)
			return
		}

		expect := strings.TrimSpace(expectConfig)
		connectorConfig := ConnectorConfig{"bob_to_entryexitmgmt_granted_role_access_path.json", "manabie/local"}
		res := strings.TrimSpace(result[connectorConfig])
		err = multierr.Combine(
			validateConnector(expect, res),
		)

		if err != nil {
			t.Error(err)
		}
	})

	t.Run("parse with exclude all column", func(t *testing.T) {
		pipeLineConfig := `

database: bob
schema: public
datapipelines:
- name: bob_to_entryexitmgmt_granted_role_access_path
  table: granted_role_access_path
  source:
    deployEnv: [local, stag, prod]
    deployOrg: [e2e, manabie, jprep, aic, ga, renseikai, synersia, tokyo]
  sinks:
  - database: entryexitmgmt
    name: bob_to_entryexitmgmt_granted_role_access_path_connector
    schema: public
    deployEnv: [local, stag, prod]
    deployOrg: [e2e, manabie, jprep, aic, ga, renseikai, synersia, tokyo]
    excludeColumns: [granted_role_id,location_id,created_at,updated_at,deleted_at,resource_path]
    captureDeleteAll: false
			`

		pipeLineConfig = strings.TrimSpace(pipeLineConfig)
		_, err := NewDataPipelineParser("",
			WithTpl(string(tplForTest)),
			WithTableSchemaDir(tableSchemaDirForTest),
			WithCustomReader(func(fileName string) ([]byte, error) {
				return []byte(pipeLineConfig), nil
			}))

		assert.Error(t, err, "empty column config")

	})
}

func TestParseWithEmptyName(t *testing.T) {
	t.Run("parse with not specify name", func(t *testing.T) {
		pipeLineConfig := `
database: bob
schema: public
datapipelines:
- name: bob_to_entryexitmgmt_granted_role_access_path
  table: granted_role_access_path
  source:
    deployEnv: [local]
    deployOrg: [manabie]
  sinks:
  - database: entryexitmgmt
    schema: public
    excludeColumns: [created_at, updated_at, deleted_at]
    captureDeleteAll: false
    deployEnv: [local]
    deployOrg: [manabie]
			`

		expectConfig := `
{
	"name": "local_manabie_bob_to_entryexitmgmt_granted_role_access_path_sink_connector",
	"config": {
		"connector.class": "io.confluent.connect.jdbc.JdbcSinkConnector",
		"tasks.max": "1",
		"topics": "local.manabie.bob.public.granted_role_access_path",
		"connection.url": "${file:/config/kafka-connect-config.properties:entryexitmgmt_url}",
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
		"pk.fields": "granted_role_id,location_id",
		"fields.whitelist": "granted_role_id,location_id,resource_path"
	}
}
		`
		pipeLineConfig = strings.TrimSpace(pipeLineConfig)
		dpl, err := NewDataPipelineParser("",
			WithTpl(string(tplForTest)),
			WithTableSchemaDir(tableSchemaDirForTest),
			WithCustomReader(func(fileName string) ([]byte, error) {
				return []byte(pipeLineConfig), nil
			}))
		if err != nil {
			t.Error(err)
			return
		}

		result, err := dpl.Parse()
		if err != nil {
			t.Errorf("failed to parse data pipeline: %v", err)
			return
		}

		expect := strings.TrimSpace(expectConfig)
		connectorConfig := ConnectorConfig{"bob_to_entryexitmgmt_granted_role_access_path.json", "manabie/local"}
		res := strings.TrimSpace(result[connectorConfig])
		err = multierr.Combine(
			validateConnector(expect, res),
		)

		if err != nil {
			t.Error(err)
		}
	})
}

func TestParseWithPipelineConfgis(t *testing.T) {
	t.Run("parse with pipeline config", func(t *testing.T) {
		pipeLineConfig := `
database: bob
schema: public
datapipelines:
- name: bob_to_entryexitmgmt_granted_role_access_path
  table: granted_role_access_path
  source:
    deployEnv: [local]
    deployOrg: [manabie]
  sinks:
  - database: entryexitmgmt
    schema: public
    pipelineConfigs:
    - env: local
      org: manabie
`

		expectConfig := `
{
	"name": "local_manabie_bob_to_entryexitmgmt_granted_role_access_path_sink_connector",
	"config": {
		"connector.class": "io.confluent.connect.jdbc.JdbcSinkConnector",
		"tasks.max": "1",
		"topics": "local.manabie.bob.public.granted_role_access_path",
		"connection.url": "${file:/config/kafka-connect-config.properties:entryexitmgmt_url}",
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
		"pk.fields": "granted_role_id,location_id",
		"fields.whitelist": "created_at,deleted_at,granted_role_id,location_id,resource_path,updated_at"
	}
}
		`
		pipeLineConfig = strings.TrimSpace(pipeLineConfig)
		dpl, err := NewDataPipelineParser("",
			WithTpl(string(tplForTest)),
			WithTableSchemaDir(tableSchemaDirForTest),
			WithCustomReader(func(fileName string) ([]byte, error) {
				return []byte(pipeLineConfig), nil
			}))
		if err != nil {
			t.Error(err)
			return
		}

		result, err := dpl.Parse()
		if err != nil {
			t.Errorf("failed to parse data pipeline: %v", err)
			return
		}

		expect := strings.TrimSpace(expectConfig)
		connectorConfig := ConnectorConfig{"bob_to_entryexitmgmt_granted_role_access_path.json", "manabie/local"}
		res := strings.TrimSpace(result[connectorConfig])
		err = multierr.Combine(
			validateConnector(expect, res),
		)

		if err != nil {
			t.Error(err)
		}
	})
}

// validateEnv get first line of config, it contain the environment information
// Ex:
// {{- if or (eq "local" .Values.global.environment) (eq "stag" .Values.global.environment) (eq "uat" .Values.global.environment) (eq "prod" .Values.global.environment) }}
func validateEnv(expectConnectorConfig, currentConnectorConfig string) (err error) {
	getdeployEnvFunc := func(config string) []string {
		// get first line
		envRaw := strings.Split(config, "\n")[0]

		rg := regexp.MustCompile(`eq "(\w+)+" .Values.global.environment`)
		deployEnv := make([]string, 0)
		for _, res := range rg.FindAllStringSubmatch(envRaw, -1) {
			if len(res) > 1 {
				deployEnv = append(deployEnv, res[1])
			}
		}
		return deployEnv
	}

	expectdeployEnv := getdeployEnvFunc(expectConnectorConfig)
	currentdeployEnv := getdeployEnvFunc(currentConnectorConfig)

	if len(expectdeployEnv) != len(currentdeployEnv) {
		return fmt.Errorf("expected %d environment variables, got %d", len(currentdeployEnv), len(expectdeployEnv))
	}

	sort.Strings(expectdeployEnv)
	sort.Strings(currentdeployEnv)

	for i := range expectdeployEnv {
		if expectdeployEnv[i] != currentdeployEnv[i] {
			return fmt.Errorf("expected environment variable %s, got %s", expectdeployEnv[i], currentdeployEnv[i])
		}
	}
	return nil
}

func validateOrg(expectConnectorConfig, currentConnectorConfig string) (err error) {
	getOrgsFunc := func(config string) []string {
		// get second line
		envRaw := strings.Split(config, "\n")[1]

		rg := regexp.MustCompile(`eq "(\w+)+" .Values.global.vendor`)
		deployEnv := make([]string, 0)
		for _, res := range rg.FindAllStringSubmatch(envRaw, -1) {
			if len(res) > 1 {
				deployEnv = append(deployEnv, res[1])
			}
		}
		return deployEnv
	}

	expectOrgs := getOrgsFunc(expectConnectorConfig)
	currentOrgs := getOrgsFunc(currentConnectorConfig)

	if len(expectOrgs) != len(currentOrgs) {
		return fmt.Errorf("expected %d org variables, got %d", len(currentOrgs), len(expectOrgs))
	}

	sort.Strings(expectOrgs)
	sort.Strings(currentOrgs)

	for i := range expectOrgs {
		if expectOrgs[i] != currentOrgs[i] {
			return fmt.Errorf("expected org variable %s, got %s", expectOrgs[i], currentOrgs[i])
		}
	}
	return nil
}

func validateConnector(expectConfig, currentConfig string) error {
	expectConnector := connect.Connector{}
	err := json.Unmarshal([]byte(expectConfig), &expectConnector)
	if err != nil {
		return err
	}
	currentConnector := connect.Connector{}
	err = json.Unmarshal([]byte(currentConfig), &currentConnector)
	if err != nil {
		return err
	}

	if expectConnector.Name != currentConnector.Name {
		return fmt.Errorf("connector name mismatch: expected %s, got %s", expectConnector.Name, currentConnector.Name)
	}

	for k := range expectConnector.Config {
		expectVal := expectConnector.Config[k]
		currentVal, ok := currentConnector.Config[k]
		if !ok {
			return fmt.Errorf("missing config %s=%s when generate connector config", k, expectVal)
		}

		// for field primary key or pk.field will compare set columns
		if k == "fields.whitelist" || k == "pk.field" {
			if !validateColumns(expectVal, currentVal) {
				return fmt.Errorf("not match connector %s config %s, expect %s but got %s\n", expectConnector.Name, k, expectVal, currentVal)
			}
			continue
		}
		if expectVal != currentVal {
			return fmt.Errorf("not match connector %s config %s, expect %s but got %s\n", expectConnector.Name, k, expectVal, currentVal)
		}
	}

	return nil
}

func validateColumns(expectColumnsRaw, currentColumnsRaw string) bool {
	expectColumns := strings.Split(expectColumnsRaw, ",")
	currentColumns := strings.Split(currentColumnsRaw, ",")

	if len(expectColumns) != len(currentColumns) {
		return false
	}
	sort.Strings(expectColumns)
	sort.Strings(currentColumns)
	for i := range expectColumns {
		if expectColumns[i] != currentColumns[i] {
			return false
		}
	}
	return true
}

func removeHelmIfEnv(s string) string {
	// {{- if or (eq "local" .Values.global.environment) (eq "stag" .Values.global.environment) }}
	s = strings.TrimSpace(s)
	lines := strings.Split(s, "\n")
	n := len(lines)

	for i, l := range lines {
		rg := regexp.MustCompile(`(\(eq "\w+" .Values.global.environment\))`)
		if rg.Match([]byte(l)) {
			// remove [[- end ]]
			arr := append(lines[:i], lines[i+1:n-1]...)
			res := strings.Join(arr, "\n")
			return res
		}
	}

	return s
}

func removeHelmIfOrg(s string) string {
	// {{- if or (eq "manabie" .Values.global.vendor) (eq "jprep" .Values.global.vendor) }}
	s = strings.TrimSpace(s)
	lines := strings.Split(s, "\n")
	n := len(lines)

	for i := range lines {
		rg := regexp.MustCompile(`(\(eq "\w+" .Values.global.vendor\))`)
		if rg.Match([]byte(lines[i])) {
			// remove [[- end ]]
			arr := append(lines[:i], lines[i+1:n-1]...)
			res := strings.Join(arr, "\n")
			return res
		}
	}

	return s
}

// Custom type that implements fs.DirEntry interface
type MockDirEntry struct {
	name     string
	isDir    bool
	typefile fs.FileMode
	modTime  time.Time
	size     int64
}

func (m MockDirEntry) Name() string {
	return m.name
}

func (m MockDirEntry) IsDir() bool {
	return m.isDir
}

func (m MockDirEntry) Type() fs.FileMode {
	return m.typefile
}

func (m MockDirEntry) Info() (fs.FileInfo, error) {
	return nil, fmt.Errorf("not implemented")
}

type MockInteractFile struct{}

func (m MockInteractFile) ReadDir(name string) ([]os.DirEntry, error) {
	fmt.Println("Mock remove called")
	return []fs.DirEntry{
		MockDirEntry{name: "manabie/local_db1_to_db2_table1.json", isDir: false},
		MockDirEntry{name: "manabie/stag_db1_to_db2_table1.json", isDir: false},
		MockDirEntry{name: "manabie/uat_db1_to_db2_table1.json", isDir: false},
	}, nil
}

func (m MockInteractFile) Remove(name string) error {
	fmt.Println("Mock remove called", name)
	if !strings.Contains(name, "manabie/uat_db1_to_db2_table1.json") {
		return fmt.Errorf("Delete incorrect file")
	}
	return nil
}

func TestDeleteFile(t *testing.T) {
	t.Run("Delete file not existed in list write file", func(t *testing.T) {
		dpls := map[string]bool{"/dir/manabie/stag_db1_to_db2_table1.json": true}
		excludes := []ExcludeConfig{{Env: "local", Org: "manabie", SinkDB: "db2", SourceDB: "db1"}}
		mock := MockInteractFile{}

		err := DeleteConnectorNotExisted(mock, dpls, "/dir", excludes)

		require.NoError(t, err)

	})
}

package tests

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/database"
	skaffoldwrapper "github.com/manabie-com/backend/internal/golibs/execwrapper/skaffold"
	vr "github.com/manabie-com/backend/internal/golibs/variants"

	"github.com/go-kafka/connect"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
)

func TestKafkaConnectorPK(t *testing.T) {
	manifestObjects, err := skaffoldwrapper.New().E(vr.EnvLocal).P(vr.PartnerManabie).Filename("skaffold.manaverse.yaml").CachedRender()
	require.NoError(t, err)
	{
		manifestObjects2, err := skaffoldwrapper.New().E(vr.EnvLocal).P(vr.PartnerManabie).Filename("skaffold2.backend.yaml").V2CachedRender()
		require.NoError(t, err)
		manifestObjects = append(manifestObjects, manifestObjects2...)
	}

	t.Parallel()
	kafkaConnectorSink, err := getKafkaConnectSinkConnectors(manifestObjects)
	require.NoError(t, err)

	for _, c := range kafkaConnectorSink {
		actualPK, ok, err, database, table := loadActualPK(c)
		if !ok {
			continue
		}
		require.NoError(t, err)

		if pkList, ok := getPKList(c); ok {
			if len(pkList) != len(actualPK) {
				t.Errorf("mismatched between pk.fields and actual primary key in table %s.%s", database, table)
			}
			for i := range pkList {
				if pkList[i] != actualPK[i] {
					t.Errorf("mismatch primary key in connector config pk.fields and actual pk in table %s.%s", database, table)
				}
			}
		} else {
			t.Errorf("missing config pk.fields in connector %s", c.Name)
		}
	}
}

func getKafkaConnectSinkConfig(manifestObjects []interface{}) (*corev1.ConfigMap, error) {
	const targetName = "kafka-connect-gen-sink-connector"
	for _, o := range manifestObjects {
		switch v := o.(type) {
		case *corev1.ConfigMap:
			if v.ObjectMeta.Name == targetName {
				return v, nil
			}
		}
	}
	return nil, fmt.Errorf("failed to find any configmap named %q", targetName)
}

func getKafkaConnectSinkConnectors(manifestObjects []interface{}) ([]connect.Connector, error) {
	cm, err := getKafkaConnectSinkConfig(manifestObjects)
	if err != nil {
		return nil, err
	}

	res := make([]connect.Connector, 0, len(cm.Data))
	for _, d := range cm.Data {
		connector := connect.Connector{}
		if err := json.Unmarshal([]byte(d), &connector); err != nil {
			return nil, fmt.Errorf("json.Unmarshal: %s", err)
		}
		res = append(res, connector)
	}
	return res, nil
}

func loadActualPK(c connect.Connector) ([]string, bool, error, string, string) {
	reSink := regexp.MustCompile(`_to_([^_]*)_`)

	reTableName := regexp.MustCompile(`_to_[^_]+_((.*)_sink_connector|(.*)_connector|(.*))`)
	sink := reSink.FindStringSubmatch(c.Name)[1]
	tableNameResult := reTableName.FindStringSubmatch(c.Name)
	tableName := ""
	for _, name := range tableNameResult[1:] {
		if name != "" {
			tableName = name
		}
	}
	if val, ok := c.Config["delete.enabled"]; (ok && val == "true") || sink == "elastic" {
		return nil, false, nil, sink, tableName
	}
	path := fmt.Sprintf("../../mock/testing/testdata/%s/%s.json", sink, tableName)
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, false, err, sink, tableName
	}

	schema := &database.TableSchema{}
	err = json.Unmarshal(b, schema)
	if err != nil {
		return nil, false, err, sink, tableName
	}
	actualPK := make([]string, 0)
	for _, c := range schema.Constraint {
		if c.ConstraintType == "PRIMARY KEY" {
			actualPK = append(actualPK, c.ColumName)
		}
	}
	sort.Strings(actualPK)
	return actualPK, true, err, sink, tableName
}

func getPKList(c connect.Connector) ([]string, bool) {
	if pk, ok := c.Config["pk.fields"]; ok {
		pkList := strings.Split(pk, ",")
		sort.Strings(pkList)
		return pkList, true
	}
	return nil, false
}

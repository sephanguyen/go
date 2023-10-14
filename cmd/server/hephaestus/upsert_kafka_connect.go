package hephaestus

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/alert"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/kafkaconnect"
	"github.com/manabie-com/backend/internal/golibs/logger"
	"github.com/manabie-com/backend/internal/hephaestus/configurations"

	"github.com/Shopify/sarama"
	"github.com/go-kafka/connect"
	"go.uber.org/zap"
)

var zapLogger *zap.Logger
var DeployCustomSinkConnector bool
var SendIncrementalSnapshot bool

const (
	JSONExt  = ".json"
	LocalEnv = "local"
)

func buildDBMap(c configurations.Config, rsc *bootstrap.Resources) map[string]database.QueryExecer {
	res := make(map[string]database.QueryExecer)
	dbNames := c.GetDBNames()
	for _, dbName := range dbNames {
		res[dbName] = rsc.DBWith(dbName)
	}
	return res
}

func RunUpsertKafkaConnect(ctx context.Context, c configurations.Config, rsc *bootstrap.Resources) error {
	zapLogger = logger.NewZapLogger("debug", c.Common.Environment == LocalEnv)

	// Set up for kafka connector
	kafkaConnectClient := connect.NewClient(c.Kafka.Connect.Addr)
	jms := rsc.NATS()

	// Set up kafka admin
	var conf *sarama.Config
	if c.Kafka.EnableAC {
		zapLogger.Info(
			"init config sasl for kafka client",
		)
		conf = sarama.NewConfig()
		conf.Net.SASL.Enable = true
		conf.Net.SASL.Mechanism = sarama.SASLTypePlaintext
		conf.Net.SASL.Handshake = true
		conf.Net.SASL.User = *c.Kafka.Username
		conf.Net.SASL.Password = *c.Kafka.Password
	}
	kafkaAdmin, err := sarama.NewClusterAdmin(c.Kafka.Addr, conf)
	if err != nil {
		zapLogger.Error(
			"failed to connect kafka" + err.Error(),
		)
		return err
	}

	dbMap := buildDBMap(c, rsc)
	connectorManagementClient := kafkaconnect.NewConnectorManagement(zapLogger, kafkaAdmin, kafkaConnectClient, jms, SendIncrementalSnapshot, dbMap)

	// Delete generated connectors
	// Delete all connectors not existed in config dir
	// Temporary disable delete connector
	// Will open when finish feature send slack when connector will deleted
	if c.SlackWebhook != nil && *c.SlackWebhook != "" {
		customSinkConnectorDir := ""
		if DeployCustomSinkConnector {
			customSinkConnectorDir = c.Kafka.Connect.SinkConfigDir
		}
		slackClient := &alert.SlackImpl{
			WebHookURL: *c.SlackWebhook,
			HTTPClient: http.Client{Timeout: time.Duration(10) * time.Second},
		}

		noti := NewNotification(slackClient, c.SlackUser, c.SlackChannel)
		deleteConnecter := NewDeleteConnector(&OSImplement{}, noti)
		deletedConnectors, success, failure := deleteConnecter.DeleteConnectors(connectorManagementClient, c.Kafka.Connect.GenSinkConfigDir, customSinkConnectorDir, c.Kafka.Connect.SourceConfigDir)
		if len(failure) > 0 {
			zapLogger.Error(
				"failed to delete generated sink connectors",
				zap.Strings("success_connectors", success),
				zap.Strings("failure_connectors", failure),
				zap.Any("deleted_connectors", deletedConnectors),
			)
			return fmt.Errorf("failed to upsert generated sink connectors")
		}
		zapLogger.Info("successfully delete generated sink connectors")
	}

	// // Upsert generate connectors which will replace user defined connectors
	// // Upsert source connector
	// success, failure := upsert(connectorManagementClient, c.Kafka.Connect.GenSourceConfigDir)
	// if len(failure) > 0 {
	// 	zapLogger.Error(
	// 		"failed to upsert source connectors",
	// 		zap.Strings("success_connectors", success),
	// 		zap.Strings("failure_connectors", failure),
	// 	)
	// 	return fmt.Errorf("failed to upsert source connectors")
	// }
	// zapLogger.Info("successfully upsert source connectors")

	// Upsert sink connector

	// TODO:
	//   - will be removed when finish generate connectors for data pipeline
	// Upsert source connector
	success, failure := upsert(connectorManagementClient, c.Kafka.Connect.SourceConfigDir)
	if len(failure) > 0 {
		zapLogger.Error(
			"failed to upsert source connectors",
			zap.Strings("success_connectors", success),
			zap.Strings("failure_connectors", failure),
		)
		return fmt.Errorf("failed to upsert source connectors")
	}
	zapLogger.Info("successfully upsert source connectors")

	success, failure = upsert(connectorManagementClient, c.Kafka.Connect.GenSinkConfigDir)
	if len(failure) > 0 {
		zapLogger.Error(
			"failed to upsert generated sink connectors",
			zap.Strings("success_connectors", success),
			zap.Strings("failure_connectors", failure),
		)
		return fmt.Errorf("failed to upsert generated sink connectors")
	}
	zapLogger.Info("successfully upsert generated sink connectors")

	if DeployCustomSinkConnector {
		success, failure = upsert(connectorManagementClient, c.Kafka.Connect.SinkConfigDir)
		if len(failure) > 0 {
			zapLogger.Error(
				"failed to upsert sink connectors",
				zap.Strings("success_connectors", success),
				zap.Strings("failure_connectors", failure),
			)
			return fmt.Errorf("failed to upsert sink connectors")
		}
		zapLogger.Info("successfully upsert sink connectors")
	}

	return nil
}

func upsert(client kafkaconnect.ConnectorManagement, confDir string) (success, failure []string) {
	files, err := os.ReadDir(confDir)
	if err != nil && !os.IsNotExist(err) {
		// Fatal when cannot read dir
		zapLogger.Fatal(
			"failed to read config dir",
			zap.String("kafka_connect_config_dir", confDir),
			zap.Error(err),
		)
		return
	}

	connectors := make([]*connect.Connector, 0)
	// If upsert failed, then log the connector failed status code and continue
	for _, file := range files {
		fileExt := filepath.Ext(file.Name())
		path := filepath.Join(confDir, file.Name())
		// ignore directory, not json file and empty file
		if file.IsDir() || fileExt != JSONExt || isEmpty(path) {
			continue
		}
		connector, err := parseConnector(path)
		if err != nil {
			zapLogger.Error(
				"failed to parse connector config file",
				zap.String("kafka_connect_config_file", path),
				zap.Error(err),
			)
			failure = append(failure, path)
			continue
		}

		connectors = append(connectors, connector)
	}

	success, failure = client.Upsert(connectors)
	return
}

func parseConnector(path string) (*connect.Connector, error) {
	connector := &connect.Connector{}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, connector)
	if err != nil {
		return nil, err
	}

	zapLogger.Info(
		"connector config",
		zap.String("connector_name", connector.Name),
		zap.Any("connector_config", connector.Config),
	)

	return connector, err
}

func isEmpty(path string) bool {
	b, _ := os.ReadFile(path)
	return len(strings.TrimSpace(string(b))) == 0
}

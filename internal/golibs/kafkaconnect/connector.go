package kafkaconnect

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/debezium"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/try"

	"github.com/Shopify/sarama"
	"github.com/go-kafka/connect"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"
)

type ConnectClient interface {
	GetConnector(name string) (*connect.Connector, *http.Response, error)
	GetConnectorStatus(name string) (*connect.ConnectorStatus, *http.Response, error)
	CreateConnector(connector *connect.Connector) (*http.Response, error)
	UpdateConnectorConfig(name string, config connect.ConnectorConfig) (*connect.Connector, *http.Response, error)
	Host() string
	DeleteConnector(name string) (*http.Response, error)
	ListConnectors() ([]string, *http.Response, error)
}

type KafkaAdmin interface {
	ListTopics() (map[string]sarama.TopicDetail, error)
}

type ConnectorManagement interface {
	Upsert(connectors []*connect.Connector) (success, failure []string)
	Delete(connectors []string) (deletedConnectors map[string]*connect.Connector, success, failure []string)
	GetListConnectorNames() ([]string, error)
}

type ConnectorManagementImpl struct {
	Logger                  *zap.Logger
	Client                  ConnectClient
	Admin                   KafkaAdmin
	Jms                     nats.JetStreamManagement
	DBMap                   map[string]database.QueryExecer
	SendIncrementalSnapshot bool
}

func NewConnectorManagement(logger *zap.Logger, admin KafkaAdmin, client ConnectClient, jms nats.JetStreamManagement, sendIncrementalSnapshot bool, dbMap map[string]database.QueryExecer) ConnectorManagement {
	return &ConnectorManagementImpl{
		Logger:                  logger,
		Client:                  client,
		Admin:                   admin,
		Jms:                     jms,
		DBMap:                   dbMap,
		SendIncrementalSnapshot: sendIncrementalSnapshot,
	}
}

func (cm *ConnectorManagementImpl) Ping() error {
	kafkaConnectHost := cm.Client.Host()
	resp, err := http.Get(kafkaConnectHost) //nolint:gosec
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (cm *ConnectorManagementImpl) tryPingKafkaConnect() error {
	return try.Do(func(attempt int) (retry bool, err error) {
		err = cm.Ping()
		if err != nil {
			time.Sleep(time.Second)
			return true, err
		}
		return false, nil
	})
}

func (cm *ConnectorManagementImpl) Upsert(connectors []*connect.Connector) (success, failure []string) {
	err := cm.tryPingKafkaConnect()
	if err != nil {
		cm.Logger.Fatal("cannot connect to kafka connect")
	}
	// If upsert failed, then log the connector failed status code and continue
	success = make([]string, 0)
	failure = make([]string, 0)
	for _, connector := range connectors {
		upsertFunc := func() error {
			oldConnector, err := cm.GetConnector(connector.Name)
			if err != nil {
				return err
			}
			if oldConnector == nil {
				return cm.TryCreateConnector(connector)
			}
			return cm.TryUpdateConnector(oldConnector, connector)
		}

		err := try.Do(func(attempt int) (retry bool, err error) {
			err = upsertFunc()
			if err != nil {
				cm.Logger.Error(
					"failed to upsert connector config",
					zap.String("kafka_connect_connector_name", connector.Name),
					zap.Error(err),
				)
				time.Sleep(time.Second)
				return true, err
			}
			return false, nil
		})

		if err != nil {
			failure = append(failure, connector.Name)
			continue
		}

		success = append(success, connector.Name)
	}
	return
}

func (cm *ConnectorManagementImpl) GetListConnectorNames() ([]string, error) {
	connectorNames, resp, err := cm.Client.ListConnectors()
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return connectorNames, nil
}

func (cm *ConnectorManagementImpl) Delete(connectors []string) (deletedConnectors map[string]*connect.Connector, success, failure []string) {
	err := cm.tryPingKafkaConnect()
	if err != nil {
		cm.Logger.Fatal("cannot connect to kafka connect")
	}
	// If upsert failed, then log the connector failed status code and continue
	deletedConnectors = make(map[string]*connect.Connector, 0)
	success = make([]string, 0)
	failure = make([]string, 0)
	for _, connector := range connectors {
		deleteFunc := func() error {
			cm.Logger.Info("delete connector: ", zap.String("name", connector))
			oldConnector, err := cm.GetConnector(connector)
			if err != nil {
				return err
			}
			if oldConnector != nil {
				// we should only delete sink connector
				// due source connector already setup delete publication when connector deleted it will make us lost data
				// we will refactor source connector and delete this condition
				connectorType, found := oldConnector.Config["connector.class"]
				// the connector match condition connector == strings.ToUpper(connector) is ksql connector we will not delete it
				if !found || connector == strings.ToUpper(connector) {
					return nil
				}
				if connectorType == "io.confluent.connect.jdbc.JdbcSinkConnector" {
					err := cm.TryDeleteConnector(connector)
					if err != nil {
						return err
					}
					deletedConnectors[connector] = oldConnector
					return nil
				}
			}
			return nil
		}

		err := try.Do(func(attempt int) (retry bool, err error) {
			err = deleteFunc()
			if err != nil {
				cm.Logger.Error(
					"failed to delete connector config",
					zap.String("kafka_connect_connector_name", connector),
					zap.Error(err),
				)
				time.Sleep(time.Second)
				return true, err
			}
			return false, nil
		})

		if err != nil {
			failure = append(failure, connector)
			continue
		}

		success = append(success, connector)
	}
	return
}

func (cm *ConnectorManagementImpl) GetConnector(name string) (*connect.Connector, error) {
	connector, resp, err := cm.Client.GetConnector(name)
	if err != nil && resp.StatusCode != http.StatusNotFound {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	return connector, nil
}

func (cm *ConnectorManagementImpl) TryCreateConnector(connector *connect.Connector) error {
	resp, err := cm.Client.CreateConnector(connector)

	if err != nil {
		return fmt.Errorf("failed to create connector config %s error %w", connector.Name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to create connector config %s with status %d", connector.Name, resp.StatusCode)
	}

	return nil
}

func (cm *ConnectorManagementImpl) TryDeleteConnector(connector string) error {
	resp, err := cm.Client.DeleteConnector(connector)

	if err != nil {
		return fmt.Errorf("failed to delete connector config %s error %w", connector, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to delete connector config %s with status %d", connector, resp.StatusCode)
	}

	return nil
}

func (cm *ConnectorManagementImpl) TryUpdateConnector(oldConnector, connector *connect.Connector) error {
	if !anyChangeInConfig(oldConnector, connector) {
		return nil
	}

	_, resp, err := cm.Client.UpdateConnectorConfig(connector.Name, connector.Config)
	if err != nil {
		return fmt.Errorf("failed to update connector config %s error %w", connector.Name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to update connector %s with status %d", connector.Name, resp.StatusCode)
	}

	err = try.Do(func(attempt int) (retry bool, err error) {
		if isReady, err := cm.IsConnectorReady(connector.Name); err != nil || !isReady {
			time.Sleep(500 * time.Millisecond)
			return true, fmt.Errorf("connector %s is not ready", connector.Name)
		}
		return false, nil
	})
	if err != nil {
		return fmt.Errorf("timeout waiting for connector %s to be ready", connector.Name)
	}

	if cm.SendIncrementalSnapshot {
		// publish event to trigger incremental snapshot signal to source database
		// connector.Config["database.server.name"] will match with our service name
		if t, err := cm.GetConnectorType(connector); err == nil && t == "source" {
			newCapturedTables := cm.GetNewCapturedCapture(oldConnector, connector)
			snapshotSignalTable := connector.Config["signal.data.collection"]
			cm.Logger.Info("send incremental snapshot to source database with tables", zap.Strings("tables", newCapturedTables))
			// temporary solution for incremental snapshot signal
			// only allow when snapshot signal is public.dbz_signals
			if len(newCapturedTables) > 0 {
				sourceID, err := extractSourceIDFromConnector(connector)
				if err != nil {
					return err
				}
				repName, err := extractReplicationSlotNameFromConnector(connector)
				if err != nil {
					return err
				}
				db, err := cm.DBWith(sourceID)
				if err != nil {
					return err
				}
				data := debezium.DataCollection{
					SourceID: sourceID,
					Tables:   newCapturedTables,
					RepName:  repName,
				}
				err = cm.SendSnapshotSignal(db, snapshotSignalTable, data)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (cm *ConnectorManagementImpl) DBWith(name string) (database.QueryExecer, error) {
	// source name comes with 2 format "bob" or "alloydb_bob"
	name = strings.TrimPrefix(name, "alloydb_")
	db, ok := cm.DBMap[name]
	if !ok {
		return nil, fmt.Errorf("cannot get db connection for %s", name)
	}
	return db, nil
}

func (cm *ConnectorManagementImpl) SendSnapshotSignal(db database.QueryExecer, snapshotSignalTable string, data debezium.DataCollection) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	err := debezium.IncrementalSnapshot(ctx, db, snapshotSignalTable, data)
	if err != nil {
		return err
	}
	return nil
}

func (cm *ConnectorManagementImpl) GetCurrentTopics() []string {
	topicsMp, err := cm.Admin.ListTopics()
	if err != nil {
		return nil
	}

	topics := make([]string, 0, len(topicsMp))
	for t := range topicsMp {
		topics = append(topics, t)
	}
	return topics
}

func (cm *ConnectorManagementImpl) GetNewCapturedCapture(oldConnector, newConnector *connect.Connector) []string {
	newTables := extractNewTableFromConnector(newConnector, oldConnector)
	existedCapturedTable := extractTableFromTopic(cm.GetCurrentTopics())

	result := make(tableInfoList, 0, len(newTables))
	for _, table := range newTables {
		found := false
		for _, existedTable := range existedCapturedTable {
			if table.Equal(existedTable) {
				found = true
			}
		}
		if !found {
			result = append(result, table)
		}
	}
	return result.GetTableList()
}

func (cm *ConnectorManagementImpl) GetConnectorType(connector *connect.Connector) (string, error) {
	hostURL, _ := url.Parse(cm.Client.Host())

	name := connector.Name
	path := "connectors/" + name

	rel, err := url.Parse(path)
	if err != nil {
		return "", err
	}

	url := hostURL.ResolveReference(rel)
	resp, err := http.Get(url.String())
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	b, _ := io.ReadAll(resp.Body)
	res := gjson.Get(string(b), "type").String()

	return res, nil
}

func (cm *ConnectorManagementImpl) IsConnectorReady(name string) (bool, error) {
	state, resp, err := cm.Client.GetConnectorStatus(name)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if state.Connector.State != "RUNNING" {
		return false, nil
	}

	for _, t := range state.Tasks {
		if t.State != "RUNNING" {
			return false, nil
		}
	}

	return true, nil
}

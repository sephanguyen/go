package hephaestus

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/alert"
	"github.com/manabie-com/backend/internal/golibs/kafkaconnect"

	"github.com/go-kafka/connect"
	"go.uber.org/zap"
)

type OSInterface interface {
	ReadDir(dirname string) ([]os.DirEntry, error)
	IsNotExist(err error) bool
	ReadFile(filename string) ([]byte, error)
}

type OSImplement struct{}

func (s *OSImplement) ReadDir(dirname string) ([]os.DirEntry, error) {
	return os.ReadDir(dirname)
}

func (s *OSImplement) IsNotExist(err error) bool {
	return os.IsNotExist(err)
}

func (s *OSImplement) ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

type NotificationInterface interface {
	sendMessage(connectorName, connectorConfig string) error
}

type NotificationImplement struct {
	SlackClient alert.SlackFactory
	User        string
	Channel     string
}

func NewNotification(slack alert.SlackFactory, user, channel string) *NotificationImplement {
	return &NotificationImplement{SlackClient: slack, User: user, Channel: channel}
}

func (s *NotificationImplement) sendMessage(connectorName, connectorConfig string) error {
	configStr := fmt.Sprintf("```%s```", connectorConfig)
	payloadMap := map[string]interface{}{
		"channel":  s.Channel,
		"username": s.User,
		"text":     fmt.Sprintf("*Deleted connector: `%s`*\n*Config*:\n%s", connectorName, configStr),
	}

	payloadJSON, err := json.Marshal(payloadMap)
	if err != nil {
		return fmt.Errorf("error marshaling payload to JSON: %v", err)
	}

	zapLogger.Info("Send message", zap.Any("payload", payloadMap))
	return s.SlackClient.SendByte(payloadJSON)
}

type DeleteConnector struct {
	OSInterface
	NotificationInterface
}

func NewDeleteConnector(os OSInterface, noti NotificationInterface) *DeleteConnector {
	return &DeleteConnector{OSInterface: os, NotificationInterface: noti}
}

func (s *DeleteConnector) parseConnector(path string) (*connect.Connector, error) {
	connector := &connect.Connector{}
	b, err := s.ReadFile(path)
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
func (s *DeleteConnector) isEmpty(path string) bool {
	b, _ := s.OSInterface.ReadFile(path)
	return len(strings.TrimSpace(string(b))) == 0
}

func (s *DeleteConnector) setConnectorsByDir(dir string, mapDirConnectors map[string]*connect.Connector, failure *[]string) {
	files, err := s.OSInterface.ReadDir(dir)
	if err != nil && !s.OSInterface.IsNotExist(err) {
		// Fatal when cannot read dir
		zapLogger.Fatal(
			"failed to read config dir",
			zap.String("kafka_connect_config_dir", dir),
			zap.Error(err),
		)
		return
	}

	for _, file := range files {
		zapLogger.Info("Reading file", zap.String("file_name", file.Name()))
		fileExt := filepath.Ext(file.Name())
		path := filepath.Join(dir, file.Name())
		// ignore directory, not json file and empty file
		if file.IsDir() || fileExt != JSONExt || s.isEmpty(path) {
			continue
		}
		connector, err := s.parseConnector(path)
		if err != nil {
			zapLogger.Error(
				"failed to parse connector config file",
				zap.String("kafka_connect_config_file", path),
				zap.Error(err),
			)
			*failure = append(*failure, path)
			continue
		}

		if connector != nil {
			mapDirConnectors[connector.Name] = connector
		}
	}
}

func (s *DeleteConnector) DeleteConnectors(client kafkaconnect.ConnectorManagement, sinkDir, customSinkDir, sourceDir string) (deletedConnectors map[string]*connect.Connector, success, failure []string) {
	currentConnectors, err := client.GetListConnectorNames()
	if err != nil && !s.OSInterface.IsNotExist(err) {
		zapLogger.Fatal(
			"failed to get current connectors",
			zap.Error(err),
		)
		return
	}

	newConnectors := make(map[string]*connect.Connector)

	s.setConnectorsByDir(sinkDir, newConnectors, &failure)
	s.setConnectorsByDir(sourceDir, newConnectors, &failure)
	s.setConnectorsByDir(customSinkDir, newConnectors, &failure)

	deleteConnectors := []string{}
	for _, connector := range currentConnectors {
		if _, ok := newConnectors[connector]; !ok {
			zapLogger.Info("Will delete connector:", zap.String("connector_name", connector))
			deleteConnectors = append(deleteConnectors, connector)
		}
	}

	deletedConnectors, success, failure = client.Delete(deleteConnectors)
	zapLogger.Info("Deleted connectors", zap.Any("deleted", deletedConnectors), zap.Any("success", success), zap.Any("failure", failure))
	zapLogger.Info("Delete connectors", zap.Any("success", success), zap.Any("failure", failure))
	s.sendNotiDeletedConnectors(deletedConnectors)
	return
}

func parseStringConnector(connector *connect.Connector) (string, error) {
	b, err := json.MarshalIndent(connector, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (s *DeleteConnector) sendNotiDeletedConnectors(deletedConnectors map[string]*connect.Connector) {
	for connectorName, connector := range deletedConnectors {
		connectorStr, err := parseStringConnector(connector)
		if err != nil {
			zapLogger.Error("Failed parse string connector", zap.Error(err))
			continue
		}
		err = s.NotificationInterface.sendMessage(connectorName, connectorStr)
		if err != nil {
			zapLogger.Error("Failed to send notification", zap.Error(err))
		}
	}
}

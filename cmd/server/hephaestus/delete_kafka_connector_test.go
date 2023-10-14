package hephaestus

import (
	"fmt"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-kafka/connect"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/kafkaconnect"
	"github.com/manabie-com/backend/internal/golibs/logger"
	"gotest.tools/assert"
)

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

type MockOSImplement struct{}

func (s *MockOSImplement) ReadDir(dirname string) ([]os.DirEntry, error) {
	return []fs.DirEntry{
		MockDirEntry{name: "local_manabie_bob1_sink_connector.json", isDir: false},
	}, nil
}

func (s *MockOSImplement) IsNotExist(err error) bool {
	fmt.Println("MockOSImplement", err)
	return os.IsNotExist(err)
}

func (s *MockOSImplement) ReadFile(filename string) ([]byte, error) {
	return []byte(fmt.Sprintf("{\"name\": \"%s\"}", strings.ReplaceAll(filename, ".json", ""))), nil
}

type MockNotificationImplement struct {
}

func (s *MockNotificationImplement) sendMessage(connectorName, connectorConfig string) error {
	if connectorName != "local_manabie_bob_sink_connector" {
		return fmt.Errorf("connector name is not match")
	}
	return nil
}

func TestDeleteKafkaConnector(t *testing.T) {
	connectorConfigDir := "./test_sample/upsert_connector_config"
	testCases := []struct {
		name        string
		successCnt  int
		failureCnt  int
		handlerFunc http.HandlerFunc
	}{
		{
			name:       "delete exists connector successfully",
			successCnt: 1,
			failureCnt: 0,
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				fmt.Println("delete exists connector successfully", r.URL.Path)
				if r.URL.Path == "/connectors/local_manabie_bob_sink_connector" {
					if r.Method == http.MethodGet {
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(`{"name":"local_manabie_bob_sink_connector", "config": {"connector.class": "io.confluent.connect.jdbc.JdbcSinkConnector"}}`))
						return
					}
					if r.Method == http.MethodDelete {
						w.WriteHeader(http.StatusNoContent)
						return
					}
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				if r.URL.Path == "/connectors" {
					if r.Method == http.MethodPost {
						// create connector successfully
						w.WriteHeader(http.StatusCreated)
						return
					} else if r.Method == http.MethodGet {
						// create connector successfully
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(`["local_manabie_bob_sink_connector"]`))
						return
					}
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				w.WriteHeader(http.StatusBadGateway)
			},
		}, {
			name:       "no delete any connector",
			successCnt: 0,
			failureCnt: 0,
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				fmt.Println("delete exists connector successfully", r.URL.Path)
				if r.Method == http.MethodDelete {
					// throw error if have any delete request
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				if r.URL.Path == "/connectors/local_manabie_bob1_sink_connector" {
					if r.Method == http.MethodGet {
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(`{"name":"local_manabie_bob1_sink_connector"}`))
						return
					}
					if r.Method == http.MethodDelete {
						w.WriteHeader(http.StatusNoContent)
						return
					}
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				if r.URL.Path == "/connectors" {
					if r.Method == http.MethodGet {
						// create connector successfully
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(`["local_manabie_bob1_sink_connector"]`))
						return
					}
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				w.WriteHeader(http.StatusBadGateway)
			},
		}, {
			name:       "no delete ksql connector",
			successCnt: 0,
			failureCnt: 0,
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				fmt.Println("delete exists connector successfully", r.URL.Path)
				if r.Method == http.MethodDelete {
					// throw error if have any delete request
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				if r.URL.Path == "/connectors/local_manabie_bob1_sink_connector" {
					if r.Method == http.MethodGet {
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(`{"name":"local_manabie_bob1_sink_connector", "SINK_CONNECTOR_1"}`))
						return
					}
					if r.Method == http.MethodDelete {
						w.WriteHeader(http.StatusNoContent)
						return
					}
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				if r.URL.Path == "/connectors" {
					if r.Method == http.MethodGet {
						// create connector successfully
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(`["local_manabie_bob1_sink_connector"]`))
						return
					}
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				w.WriteHeader(http.StatusBadGateway)
			},
		}}
	zapLogger = logger.NewZapLogger("debug", true)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fmt.Println("test case", tc.name)
			server := httptest.NewServer(tc.handlerFunc)
			client := connect.NewClient(server.URL)
			dbMap := make(map[string]database.QueryExecer)
			connectorManagementClient := kafkaconnect.NewConnectorManagement(zapLogger, nil, client, nil, false, dbMap)
			connector := NewDeleteConnector(&MockOSImplement{}, &MockNotificationImplement{})
			deletedConnectors, success, failure := connector.DeleteConnectors(connectorManagementClient, connectorConfigDir, "", "")
			assert.Equal(t, len(success), tc.successCnt)
			assert.Equal(t, len(failure), tc.failureCnt)
			if tc.successCnt > 0 {
				for connectorName := range deletedConnectors {
					assert.Equal(t, "local_manabie_bob_sink_connector", connectorName)
				}
			}
		})
	}
}

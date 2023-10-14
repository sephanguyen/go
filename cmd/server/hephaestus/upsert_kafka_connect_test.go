package hephaestus

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/kafkaconnect"
	"github.com/manabie-com/backend/internal/golibs/logger"

	"github.com/go-kafka/connect"
	"gotest.tools/assert"
)

// https://docs.confluent.io/platform/current/connect/references/restapi.html
func TestUpdateKafkaConnector(t *testing.T) {
	connectorConfigDir := "./test_sample/upsert_connector_config"
	testCases := []struct {
		name        string
		successCnt  int
		failureCnt  int
		handlerFunc http.HandlerFunc
	}{
		{
			name:       "create new connector successfully",
			successCnt: 1,
			failureCnt: 0,
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				// check if connector exists and return not found connector
				if r.URL.Path == "/connectors/local_manabie_bob_source_connector" {
					if r.Method == http.MethodGet {
						w.WriteHeader(http.StatusNotFound)
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
					}
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				w.WriteHeader(http.StatusBadGateway)
			},
		},
		{
			name:       "update existed connector successfully",
			successCnt: 1,
			failureCnt: 0,
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/connectors/local_manabie_bob_source_connector" {
					// check if connector exists and return 200 status code which indicate the connector exists
					if r.Method == http.MethodGet {
						w.WriteHeader(http.StatusOK)
						return
					}
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				if r.URL.Path == "/connectors/local_manabie_bob_source_connector/config" {
					if r.Method == http.MethodPut {
						w.WriteHeader(http.StatusOK)
						return
					}
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				if r.URL.Path == "/connectors/local_manabie_bob_source_connector/status" {
					if r.Method == http.MethodGet {
						w.WriteHeader(http.StatusOK)
						w.Write([]byte("{ \"name\": \"local_manabie_bob_to_entryexitmgmt_permission_connector\", \"connector\": { \"state\": \"RUNNING\", \"worker_id\": \"...:8083\" }, \"tasks\": [ { \"id\": 0, \"state\": \"RUNNING\", \"worker_id\": \"10.244.0.33:8083\" } ], \"type\": \"sink\" }"))
						return
					}
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				w.WriteHeader(http.StatusBadGateway)
			},
		},
	}
	zapLogger = logger.NewZapLogger("debug", true)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(tc.handlerFunc)
			client := connect.NewClient(server.URL)
			dbMap := make(map[string]database.QueryExecer)
			connectorManagementClient := kafkaconnect.NewConnectorManagement(zapLogger, nil, client, nil, false, dbMap)
			success, failure := upsert(connectorManagementClient, connectorConfigDir)
			assert.Equal(t, len(success), tc.successCnt)
			assert.Equal(t, len(failure), tc.failureCnt)
		})
	}
}

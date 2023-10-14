package kafkaconnect

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Shopify/sarama"
	"github.com/go-kafka/connect"
	"github.com/jackc/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/debezium"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_kafkaconnect "github.com/manabie-com/backend/mock/golibs/kafkaconnect"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
)

func TestGetNewCapturedCapture(t *testing.T) {
	testcases := []struct {
		name          string
		connectormgmt *ConnectorManagementImpl
		existedTopic  []string
		oldConnector  *connect.Connector
		newConnector  *connect.Connector
		mockFunc      func(c *ConnectorManagementImpl, oldConnector, newConnector *connect.Connector, topics []string)
		err           error
		res           []string
	}{
		{
			// newCapturedTables: [public.a, public.b]
			// existedTopic: []
			// output: [public.a,public.b]
			name:          "happy case with non existed topic",
			connectormgmt: new(ConnectorManagementImpl),
			oldConnector:  &connect.Connector{},
			newConnector: &connect.Connector{
				Config: map[string]string{
					"database.server.name": "bob",
					"table.include.list":   "public.a,public.b",
				},
			},
			err: nil,
			res: []string{"public.a", "public.b"},
			mockFunc: func(c *ConnectorManagementImpl, oldConnector, newConnector *connect.Connector, topics []string) {
				admin := mock_kafkaconnect.NewKafkaAdmin(t)

				topic := map[string]sarama.TopicDetail{}
				admin.On("ListTopics").Once().Return(topic, nil)
				c.Admin = admin
			},
		},
		{
			// (newConnector[table.include.list] - oldConnector[table.include.list)
			// newCapturedTables: [public.b]
			// existedTopic: []
			// output: [public.b]
			name:          "happy case with some tables",
			connectormgmt: new(ConnectorManagementImpl),
			oldConnector: &connect.Connector{
				Config: map[string]string{
					"database.server.name": "bob",
					"table.include.list":   "public.a",
				},
			},
			newConnector: &connect.Connector{
				Config: map[string]string{
					"database.server.name": "bob",
					"table.include.list":   "public.a,public.b",
				},
			},
			err: nil,
			res: []string{"public.b"},
			mockFunc: func(c *ConnectorManagementImpl, oldConnector, newConnector *connect.Connector, topics []string) {
				admin := mock_kafkaconnect.NewKafkaAdmin(t)

				topic := map[string]sarama.TopicDetail{}
				admin.On("ListTopics").Once().Return(topic, nil)
				c.Admin = admin
			},
		},
		{
			// (newConnector[table.include.list] - oldConnector[table.include.list)
			// newCapturedTables: [public.a,public.b]
			// existedTopic: [local.manabie.bob.public.b]
			// output: [public.a]
			name:          "happy case with remove already captured table",
			connectormgmt: new(ConnectorManagementImpl),
			existedTopic:  []string{"local.manabie.bob.public.b"},
			oldConnector:  &connect.Connector{},
			newConnector: &connect.Connector{
				Config: map[string]string{
					"database.server.name": "bob",
					"table.include.list":   "public.a,public.b",
				},
			},
			err: nil,
			res: []string{"public.a"},
			mockFunc: func(c *ConnectorManagementImpl, oldConnector, newConnector *connect.Connector, topics []string) {
				admin := mock_kafkaconnect.NewKafkaAdmin(t)

				topicMp := map[string]sarama.TopicDetail{}

				for _, t := range topics {
					topicMp[t] = sarama.TopicDetail{}
				}
				admin.On("ListTopics").Once().Return(topicMp, nil)
				c.Admin = admin
			},
		},
		{
			// (newConnector[table.include.list] - oldConnector[table.include.list)
			// newCapturedTables: [public.a,public.b,public.c]
			// existedTopic: [local.manabie.bob.public.a, local.manabie.bob.public.b, local.manabie.bob.public.c]
			// output: []
			name:          "happy case with all table is captured",
			connectormgmt: new(ConnectorManagementImpl),
			existedTopic:  []string{"local.manabie.bob.public.a", "local.manabie.bob.public.b", "local.manabie.bob.public.c"},
			oldConnector:  &connect.Connector{},
			newConnector: &connect.Connector{
				Config: map[string]string{
					"database.server.name": "bob",
					"table.include.list":   "public.a,public.b,public.c",
				},
			},
			err: nil,
			res: []string{},
			mockFunc: func(c *ConnectorManagementImpl, oldConnector, newConnector *connect.Connector, topics []string) {
				admin := mock_kafkaconnect.NewKafkaAdmin(t)

				topicMp := map[string]sarama.TopicDetail{}

				for _, t := range topics {
					topicMp[t] = sarama.TopicDetail{}
				}
				admin.On("ListTopics").Once().Return(topicMp, nil)
				c.Admin = admin
			},
		},
		{
			// (newConnector[table.include.list] - oldConnector[table.include.list)
			// newCapturedTables: []
			// existedTopic: [local.manabie.bob.public.a, local.manabie.bob.public.b, local.manabie.bob.public.c]
			// output: []
			name:          "happy case with no new added tables",
			connectormgmt: new(ConnectorManagementImpl),
			existedTopic:  []string{"local.manabie.bob.public.a", "local.manabie.bob.public.b", "local.manabie.bob.public.c"},
			oldConnector: &connect.Connector{
				Config: map[string]string{
					"database.server.name": "bob",
					"table.include.list":   "public.a,public.b,public.c",
				},
			},
			newConnector: &connect.Connector{
				Config: map[string]string{
					"database.server.name": "bob",
					"table.include.list":   "public.a,public.b,public.c",
				},
			},
			err: nil,
			res: []string{},
			mockFunc: func(c *ConnectorManagementImpl, oldConnector, newConnector *connect.Connector, topics []string) {
				admin := mock_kafkaconnect.NewKafkaAdmin(t)

				topicMp := map[string]sarama.TopicDetail{}

				for _, t := range topics {
					topicMp[t] = sarama.TopicDetail{}
				}
				admin.On("ListTopics").Once().Return(topicMp, nil)
				c.Admin = admin
			},
		},
		{
			// (newConnector[table.include.list] - oldConnector[table.include.list)
			// newCapturedTables: []
			// existedTopic: []
			// output: []
			name:          "happy case with empty new table and empty captured table",
			connectormgmt: new(ConnectorManagementImpl),
			existedTopic:  []string{"local.manabie.bob.public.a", "local.manabie.bob.public.b", "local.manabie.bob.public.c"},
			oldConnector:  &connect.Connector{},
			newConnector:  &connect.Connector{},
			err:           nil,
			res:           []string{},
			mockFunc: func(c *ConnectorManagementImpl, oldConnector, newConnector *connect.Connector, topics []string) {
				admin := mock_kafkaconnect.NewKafkaAdmin(t)

				topicMp := map[string]sarama.TopicDetail{}

				for _, t := range topics {
					topicMp[t] = sarama.TopicDetail{}
				}
				admin.On("ListTopics").Once().Return(topicMp, nil)
				c.Admin = admin
			},
		},
		{
			// (newConnector[table.include.list] - oldConnector[table.include.list)
			// newCapturedTables: []
			// existedTopic: [local.manabie.bob.public.a, local.manabie.bob.public.b, local.manabie.bob.public.c]
			// output: []
			name:          "happy case with empty new table but have already captured tables",
			connectormgmt: new(ConnectorManagementImpl),
			existedTopic:  []string{"local.manabie.bob.public.a"},
			oldConnector: &connect.Connector{
				Config: map[string]string{
					"database.server.name": "bob",
					"table.include.list":   "public.a",
				},
			},
			newConnector: &connect.Connector{
				Config: map[string]string{
					"database.server.name": "bob",
					"table.include.list":   "public.a,public.b,public.c",
				},
			},
			err: nil,
			res: []string{"public.b", "public.c"},
			mockFunc: func(c *ConnectorManagementImpl, oldConnector, newConnector *connect.Connector, topics []string) {
				admin := mock_kafkaconnect.NewKafkaAdmin(t)

				topicMp := map[string]sarama.TopicDetail{}

				for _, t := range topics {
					topicMp[t] = sarama.TopicDetail{}
				}
				admin.On("ListTopics").Once().Return(topicMp, nil)
				c.Admin = admin
			},
		},
	}

	for _, tc := range testcases {
		tc.mockFunc(tc.connectormgmt, tc.oldConnector, tc.newConnector, tc.existedTopic)
		res := tc.connectormgmt.GetNewCapturedCapture(tc.oldConnector, tc.newConnector)
		assert.Equal(t, tc.res, res)
	}
}

func TestTryCreateConnector(t *testing.T) {
	testcases := []struct {
		name          string
		connectormgmt *ConnectorManagementImpl
		connector     *connect.Connector
		mockFunc      func(*ConnectorManagementImpl, *connect.Connector)
		err           error
	}{
		{
			name:          "create connector successfully with status code created",
			connectormgmt: &ConnectorManagementImpl{},
			connector: &connect.Connector{
				Name: "test",
			},
			mockFunc: func(c *ConnectorManagementImpl, connector *connect.Connector) {

				client := mock_kafkaconnect.NewConnectClient(t)
				resp := httptest.NewRecorder()
				resp.WriteHeader(http.StatusCreated)
				client.On("CreateConnector", connector).Once().Return(resp.Result(), nil)

				c.Client = client
				c.Logger = zap.NewExample()
			},
			err: nil,
		},
		{
			name:          "cannot create connector, server response status internal error code",
			connectormgmt: &ConnectorManagementImpl{},
			connector: &connect.Connector{
				Name: "test-2",
			},
			mockFunc: func(c *ConnectorManagementImpl, connector *connect.Connector) {

				client := mock_kafkaconnect.NewConnectClient(t)
				resp := httptest.NewRecorder()
				resp.WriteHeader(http.StatusInternalServerError)
				client.On("CreateConnector", connector).Once().Return(resp.Result(), nil)

				c.Client = client
				c.Logger = zap.NewExample()
			},
			err: fmt.Errorf("failed to create connector config test-2 with status 500"),
		},
	}

	for _, tc := range testcases {
		tc.mockFunc(tc.connectormgmt, tc.connector)

		err := tc.connectormgmt.TryCreateConnector(tc.connector)
		if tc.err != nil {
			assert.Equal(t, tc.err, err)
		} else {
			assert.Nil(t, err)
		}
	}
}

func TestTryUpdateConnector(t *testing.T) {
	testcases := []struct {
		name          string
		connectormgmt *ConnectorManagementImpl
		oldConnector  *connect.Connector
		connector     *connect.Connector
		existedTopic  []string
		mockFunc      func(c *ConnectorManagementImpl, oldConnector *connect.Connector, newConnector *connect.Connector, topic []string)
		err           error
	}{
		{
			name:          "update connector successfully with status code ok when add new field",
			connectormgmt: &ConnectorManagementImpl{},
			oldConnector: &connect.Connector{
				Name:   "test",
				Config: make(map[string]string),
			},
			connector: &connect.Connector{
				Name: "test",
				Config: map[string]string{
					"table.include.list":   "public.a,public.b",
					"database.server.name": "bob",
					"slot.name":            "local_manabie_bob",
				},
			},
			existedTopic: []string{},
			mockFunc: func(c *ConnectorManagementImpl, oldConnector, connector *connect.Connector, topics []string) {
				host := mockKafkaConnectServerGetConnectorType("source")

				// mock kafka connect client
				client := mock_kafkaconnect.NewConnectClient(t)
				resp := httptest.NewRecorder()
				resp.WriteHeader(http.StatusOK)
				client.On("Host").Once().Return(host)
				client.On("UpdateConnectorConfig", connector.Name, connector.Config).Once().Return(connector, resp.Result(), nil)

				state := &connect.ConnectorStatus{
					Name:      "test",
					Connector: connect.ConnectorState{State: "RUNNING"},
				}
				client.On("GetConnectorStatus", connector.Name).Once().Return(state, resp.Result(), nil)
				c.Client = client

				// mock kafka admin
				admin := mock_kafkaconnect.NewKafkaAdmin(t)
				topicRes := make(map[string]sarama.TopicDetail)
				for _, t := range topics {
					topicRes[t] = sarama.TopicDetail{}
				}
				admin.On("ListTopics").Once().Return(topicRes, nil)
				c.Admin = admin

				dbMap := make(map[string]database.QueryExecer)
				mockDB := mock_database.NewQueryExecer(t)
				mockRow := mock_database.NewRow(t)
				mockRow.On("Scan", mock.Anything).Once().Return(nil)
				mockDB.On("QueryRow", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Once().Return(mockRow)
				cmdTag := pgconn.CommandTag([]byte(`1`))
				mockDB.On("Exec", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything).Once().Return(cmdTag, nil)
				dbMap["bob"] = mockDB

				c.DBMap = dbMap
				c.Logger = zap.NewExample()
			},
			err: nil,
		},
		{
			name:          "update source connector successfully and add snapshot new table, remove some table already captured",
			connectormgmt: &ConnectorManagementImpl{},
			oldConnector: &connect.Connector{
				Name: "test",
				Config: map[string]string{
					"database.server.name": "bob",
					"table.include.list":   "public.a,public.b",
					"slot.name":            "local_manabie_bob",
				},
			},
			connector: &connect.Connector{
				Name: "test",
				Config: map[string]string{
					"database.server.name": "bob",
					"table.include.list":   "public.a,public.b,public.c,public.d",
					"slot.name":            "local_manabie_bob",
				},
			},
			existedTopic: []string{"local.manabie.bob.public.c"},
			mockFunc: func(c *ConnectorManagementImpl, oldConnector, connector *connect.Connector, topics []string) {
				host := mockKafkaConnectServerGetConnectorType("source")

				// mock kafka connect client
				client := mock_kafkaconnect.NewConnectClient(t)
				resp := httptest.NewRecorder()
				resp.WriteHeader(http.StatusOK)
				client.On("Host").Once().Return(host)
				client.On("UpdateConnectorConfig", connector.Name, connector.Config).Once().Return(connector, resp.Result(), nil)

				state := &connect.ConnectorStatus{
					Name:      "test",
					Connector: connect.ConnectorState{State: "RUNNING"},
				}
				client.On("GetConnectorStatus", connector.Name).Once().Return(state, resp.Result(), nil)
				c.Client = client

				// mock kafka admin
				admin := mock_kafkaconnect.NewKafkaAdmin(t)
				topicRes := make(map[string]sarama.TopicDetail)
				for _, t := range topics {
					topicRes[t] = sarama.TopicDetail{}
				}
				admin.On("ListTopics").Once().Return(topicRes, nil)
				c.Admin = admin

				dbMap := make(map[string]database.QueryExecer)
				mockDB := mock_database.NewQueryExecer(t)
				mockRow := mock_database.NewRow(t)
				mockRow.On("Scan", mock.Anything).Once().Return(nil)
				mockDB.On("QueryRow", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Once().Return(mockRow)
				cmdTag := pgconn.CommandTag([]byte(`1`))
				mockDB.On("Exec", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything).Once().Return(cmdTag, nil)
				dbMap["bob"] = mockDB

				c.DBMap = dbMap

				c.Logger = zap.NewExample()
			},
			err: nil,
		},
		{
			name:          "update source connector successfully and add snapshot new table",
			connectormgmt: &ConnectorManagementImpl{},
			oldConnector: &connect.Connector{
				Name: "test",
				Config: map[string]string{
					"database.server.name": "bob",
					"slot.name":            "local_manabie_bob",
					"table.include.list":   "public.a,public.b",
				},
			},
			connector: &connect.Connector{
				Name: "test",
				Config: map[string]string{
					"database.server.name": "bob",
					"slot.name":            "local_manabie_bob",
					"table.include.list":   "public.a,public.b,public.c,public.d",
				},
			},
			existedTopic: []string{},
			mockFunc: func(c *ConnectorManagementImpl, oldConnector, connector *connect.Connector, topics []string) {
				host := mockKafkaConnectServerGetConnectorType("source")

				// mock kafka connect client
				client := mock_kafkaconnect.NewConnectClient(t)
				resp := httptest.NewRecorder()
				resp.WriteHeader(http.StatusOK)
				client.On("Host").Once().Return(host)
				client.On("UpdateConnectorConfig", connector.Name, connector.Config).Once().Return(connector, resp.Result(), nil)

				state := &connect.ConnectorStatus{
					Name:      "test",
					Connector: connect.ConnectorState{State: "RUNNING"},
				}
				client.On("GetConnectorStatus", connector.Name).Once().Return(state, resp.Result(), nil)
				c.Client = client

				// mock kafka admin
				admin := mock_kafkaconnect.NewKafkaAdmin(t)
				topicRes := make(map[string]sarama.TopicDetail)
				for _, t := range topics {
					topicRes[t] = sarama.TopicDetail{}
				}
				admin.On("ListTopics").Once().Return(topicRes, nil)
				c.Admin = admin

				dbMap := make(map[string]database.QueryExecer)
				mockDB := mock_database.NewQueryExecer(t)
				mockRow := mock_database.NewRow(t)
				mockRow.On("Scan", mock.Anything).Once().Return(nil)
				mockDB.On("QueryRow", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Once().Return(mockRow)
				cmdTag := pgconn.CommandTag([]byte(`1`))
				mockDB.On("Exec", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything).Once().Return(cmdTag, nil)
				dbMap["bob"] = mockDB

				c.DBMap = dbMap

				c.Logger = zap.NewExample()
			},
			err: nil,
		},
		{
			name:          "update connector successfully with status code ok when update existing config field",
			connectormgmt: &ConnectorManagementImpl{},
			oldConnector: &connect.Connector{
				Name: "test",
				Config: map[string]string{
					"field": "value",
				},
			},
			connector: &connect.Connector{
				Name: "test",
				Config: map[string]string{
					"field": "updated_value",
				},
			},
			mockFunc: func(c *ConnectorManagementImpl, oldConnector, connector *connect.Connector, topics []string) {
				host := mockKafkaConnectServerGetConnectorType("sink")
				client := mock_kafkaconnect.NewConnectClient(t)
				client.On("Host").Once().Return(host)

				resp := httptest.NewRecorder()
				resp.WriteHeader(http.StatusOK)
				client.On("UpdateConnectorConfig", connector.Name, connector.Config).Once().Return(connector, resp.Result(), nil)

				state := &connect.ConnectorStatus{
					Name:      "test",
					Connector: connect.ConnectorState{State: "RUNNING"},
				}
				client.On("GetConnectorStatus", connector.Name).Once().Return(state, resp.Result(), nil)
				c.Client = client
				c.Logger = zap.NewExample()
			},
			err: nil,
		},
		{
			name:          "update connector successfully with status code ok when remove existing config field",
			connectormgmt: &ConnectorManagementImpl{},
			oldConnector: &connect.Connector{
				Name: "test",
				Config: map[string]string{
					"field": "value",
				},
			},
			connector: &connect.Connector{
				Name:   "test",
				Config: make(map[string]string),
			},
			mockFunc: func(c *ConnectorManagementImpl, oldConnector, connector *connect.Connector, topics []string) {
				host := mockKafkaConnectServerGetConnectorType("sink")
				client := mock_kafkaconnect.NewConnectClient(t)
				client.On("Host").Once().Return(host)
				resp := httptest.NewRecorder()
				resp.WriteHeader(http.StatusOK)
				client.On("UpdateConnectorConfig", connector.Name, connector.Config).Once().Return(oldConnector, resp.Result(), nil)

				state := &connect.ConnectorStatus{
					Name:      "test",
					Connector: connect.ConnectorState{State: "RUNNING"},
				}
				client.On("GetConnectorStatus", connector.Name).Once().Return(state, resp.Result(), nil)
				c.Client = client
				c.Logger = zap.NewExample()
			},
			err: nil,
		},
		{
			name:          "skip when config of connector not changed",
			connectormgmt: &ConnectorManagementImpl{},
			oldConnector: &connect.Connector{
				Name: "test",
				Config: map[string]string{
					"unchange_field": "unchange_value",
				},
			},
			connector: &connect.Connector{
				Name: "test",
				Config: map[string]string{
					"unchange_field": "unchange_value",
				},
			},
			mockFunc: func(c *ConnectorManagementImpl, oldConnector, connector *connect.Connector, topics []string) {
				client := mock_kafkaconnect.NewConnectClient(t)
				resp := httptest.NewRecorder()
				resp.WriteHeader(http.StatusOK)
				c.Client = client
				c.Logger = zap.NewExample()
			},
			err: nil,
		},
		{
			name:          "cannot update connector, server response status internal error",
			connectormgmt: &ConnectorManagementImpl{},
			oldConnector: &connect.Connector{
				Name: "test-2",
			},
			connector: &connect.Connector{
				Name: "test-2",
				Config: map[string]string{
					"field": "value",
				},
			},
			mockFunc: func(c *ConnectorManagementImpl, oldConnector, connector *connect.Connector, topics []string) {
				// host := mockKafkaConnectServerGetConnectorType("sink")
				client := mock_kafkaconnect.NewConnectClient(t)
				resp := httptest.NewRecorder()
				resp.WriteHeader(http.StatusInternalServerError)
				client.On("UpdateConnectorConfig", connector.Name, connector.Config).Once().Return(nil, resp.Result(), nil)

				c.Client = client
				c.Logger = zap.NewExample()
			},
			err: fmt.Errorf("failed to update connector test-2 with status 500"),
		},
		{
			name:          "skip update connector with ignore field",
			connectormgmt: &ConnectorManagementImpl{},
			oldConnector: &connect.Connector{
				Name: "test-2",
				Config: map[string]string{
					"name":  "connector_name",
					"field": "value",
				},
			},
			connector: &connect.Connector{
				Name: "test-2",
				Config: map[string]string{
					"field": "value",
				},
			},
			mockFunc: func(c *ConnectorManagementImpl, oldConnector, connector *connect.Connector, topics []string) {
				client := mock_kafkaconnect.NewConnectClient(t)
				resp := httptest.NewRecorder()
				resp.WriteHeader(http.StatusOK)
				c.Client = client
				c.Logger = zap.NewExample()
			},
			err: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.connectormgmt.SendIncrementalSnapshot = true
			tc.mockFunc(tc.connectormgmt, tc.oldConnector, tc.connector, tc.existedTopic)
			err := tc.connectormgmt.TryUpdateConnector(tc.oldConnector, tc.connector)
			if tc.err != nil {
				assert.Equal(t, tc.err, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestUpsert(t *testing.T) {
	testcases := []struct {
		name          string
		connectormgmt *ConnectorManagementImpl
		oldConnector  []*connect.Connector
		connector     []*connect.Connector
		existedTopic  []string
		mockFunc      func(c *ConnectorManagementImpl, oldConnector []*connect.Connector, newConnector []*connect.Connector, topics []string)

		success []string
		failure []string
	}{
		{
			name:          "happy case update list of source connector",
			connectormgmt: new(ConnectorManagementImpl),
			oldConnector: []*connect.Connector{
				{
					Name:   "test",
					Config: make(connect.ConnectorConfig),
				},
			},
			connector: []*connect.Connector{
				{
					Name: "test",
					Config: map[string]string{
						"database.server.name": "bob",
						"slot.name":            "local_manabie_bob",
						"table.include.list":   "public.a,public.b",
					},
				},
			},
			existedTopic: []string{},
			success:      []string{"test"},
			failure:      []string{},
			mockFunc: func(c *ConnectorManagementImpl, oldConnectorList, newConnectorList []*connect.Connector, topics []string) {
				// mock kafka connect client
				client := mock_kafkaconnect.NewConnectClient(t)
				// mock kafka admi
				admin := mock_kafkaconnect.NewKafkaAdmin(t)
				// mock nats jetstream
				jms := new(mock_nats.JetStreamManagement)
				dbMap := make(map[string]database.QueryExecer)
				mockDB := mock_database.NewQueryExecer(t)
				mockRow := mock_database.NewRow(t)
				mockRow.On("Scan", mock.Anything).Once().Return(nil)
				mockDB.On("QueryRow", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Once().Return(mockRow)
				cmdTag := pgconn.CommandTag([]byte(`1`))
				mockDB.On("Exec", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything).Once().Return(cmdTag, nil)
				dbMap["bob"] = mockDB

				for i := 0; i < len(oldConnectorList); i++ {
					oldConnector := oldConnectorList[i]
					newConnector := newConnectorList[i]
					host := mockKafkaConnectServerGetConnectorType("source")
					client.On("Host").Once().Return(host)
					resp := httptest.NewRecorder()
					resp.WriteHeader(http.StatusOK)
					client.On("GetConnector", newConnector.Name).Once().Return(oldConnector, resp.Result(), nil)
					resp = httptest.NewRecorder()
					resp.WriteHeader(http.StatusOK)
					client.On("UpdateConnectorConfig", newConnector.Name, newConnector.Config).Once().Return(oldConnector, resp.Result(), nil)

					state := &connect.ConnectorStatus{
						Name:      "test",
						Connector: connect.ConnectorState{State: "RUNNING"},
					}
					client.On("GetConnectorStatus", newConnector.Name).Once().Return(state, resp.Result(), nil)
					client.On("Host").Once().Return(host)

					topicRes := map[string]sarama.TopicDetail{}
					for _, t := range topics {
						topicRes[t] = sarama.TopicDetail{}
					}
					admin.On("ListTopics").Once().Return(topicRes, nil)

					data := debezium.DataCollection{
						SourceID: "bob",
						Tables:   []string{"public.a", "public.b"},
						RepName:  "local_manabie_bob",
					}

					b, _ := json.Marshal(data)
					jms.On("PublishContext", context.Background(), constants.SubjectDebeziumIncrementalSnapshotSend, b).Once().Return(nil, nil)
				}
				c.Client = client
				c.Admin = admin
				c.Jms = jms
				c.DBMap = dbMap
				c.Logger = zap.NewExample()
			},
		},
	}

	for _, tc := range testcases {
		tc.connectormgmt.SendIncrementalSnapshot = true
		tc.mockFunc(tc.connectormgmt, tc.oldConnector, tc.connector, tc.existedTopic)
		success, failure := tc.connectormgmt.Upsert(tc.connector)
		assert.Equal(t, tc.success, success)
		assert.Equal(t, tc.failure, failure)
	}
}

func mockKafkaConnectServerGetConnectorType(expectedType string) (host string) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/connectors/test" {
			w.Header().Set("Content-Type", "application/json")
			// return as source connector
			w.Write([]byte(fmt.Sprintf(`{"type": "%s"}`, expectedType)))
		}
	}))

	host = server.URL
	return
}

func TestTryDeleteConnector(t *testing.T) {
	testcases := []struct {
		name          string
		connectormgmt *ConnectorManagementImpl
		connector     *connect.Connector
		mockFunc      func(*ConnectorManagementImpl, *connect.Connector)
		err           error
	}{
		{
			name:          "delete connector successfully with status code deleted",
			connectormgmt: &ConnectorManagementImpl{},
			connector: &connect.Connector{
				Name: "test",
			},
			mockFunc: func(c *ConnectorManagementImpl, connector *connect.Connector) {

				client := mock_kafkaconnect.NewConnectClient(t)
				resp := httptest.NewRecorder()
				resp.WriteHeader(http.StatusNoContent)
				client.On("DeleteConnector", connector.Name).Once().Return(resp.Result(), nil)

				c.Client = client
				c.Logger = zap.NewExample()
			},
			err: nil,
		},
		{
			name:          "cannot delete connector, server response status internal error code",
			connectormgmt: &ConnectorManagementImpl{},
			connector: &connect.Connector{
				Name: "test-2",
			},
			mockFunc: func(c *ConnectorManagementImpl, connector *connect.Connector) {

				client := mock_kafkaconnect.NewConnectClient(t)
				resp := httptest.NewRecorder()
				resp.WriteHeader(http.StatusInternalServerError)
				client.On("DeleteConnector", connector.Name).Once().Return(resp.Result(), nil)

				c.Client = client
				c.Logger = zap.NewExample()
			},
			err: fmt.Errorf("failed to delete connector config test-2 with status 500"),
		},
	}

	for _, tc := range testcases {
		tc.mockFunc(tc.connectormgmt, tc.connector)

		err := tc.connectormgmt.TryDeleteConnector(tc.connector.Name)
		if tc.err != nil {
			assert.Equal(t, tc.err, err)
		} else {
			assert.Nil(t, err)
		}
	}
}

func TestGetListConnector(t *testing.T) {
	testcases := []struct {
		name          string
		connectormgmt *ConnectorManagementImpl
		rs            []string
		mockFunc      func(impl *ConnectorManagementImpl, connectors []string)
		err           error
	}{
		{
			name:          "get connector successfully with list connectors",
			connectormgmt: &ConnectorManagementImpl{},
			rs:            []string{"connector1", "connector2"},
			mockFunc: func(c *ConnectorManagementImpl, connectors []string) {
				client := mock_kafkaconnect.NewConnectClient(t)
				resp := httptest.NewRecorder()
				resp.WriteHeader(http.StatusOK)
				data, _ := json.Marshal(connectors)
				resp.Write(data)
				client.On("ListConnectors").Once().Return(connectors, resp.Result(), nil)

				c.Client = client
				c.Logger = zap.NewExample()
			},
			err: nil,
		},
		{
			name:          "get connector successfully with list connectors",
			connectormgmt: &ConnectorManagementImpl{},
			rs:            nil,
			mockFunc: func(c *ConnectorManagementImpl, connectors []string) {
				client := mock_kafkaconnect.NewConnectClient(t)
				resp := httptest.NewRecorder()
				resp.WriteHeader(http.StatusBadRequest)
				err := fmt.Errorf("failed to list connectors")
				client.On("ListConnectors").Once().Return(connectors, resp.Result(), err)

				c.Client = client
				c.Logger = zap.NewExample()
			},
			err: fmt.Errorf("failed to list connectors"),
		},
	}

	for _, tc := range testcases {
		tc.mockFunc(tc.connectormgmt, tc.rs)
		connectors, err := tc.connectormgmt.GetListConnectorNames()
		if err == nil {
			assert.Equal(t, tc.rs, connectors)
			assert.Nil(t, err)
		} else {
			assert.Equal(t, err, tc.err)
		}
	}
}

func TestDelete(t *testing.T) {
	testcases := []struct {
		name          string
		connectormgmt *ConnectorManagementImpl
		oldConnector  []*connect.Connector
		existedTopic  []string
		mockFunc      func(c *ConnectorManagementImpl, oldConnector []*connect.Connector)

		success           []string
		failure           []string
		deletedConnectors map[string]*connect.Connector

		connectors []string
	}{
		{
			name:          "Delete list connectors",
			connectormgmt: new(ConnectorManagementImpl),
			oldConnector: []*connect.Connector{
				{
					Name: "test",
					Config: map[string]string{
						"database.server.name": "bob",
						"slot.name":            "local_manabie_bob",
						"connector.class":      "io.confluent.connect.jdbc.JdbcSinkConnector",
					},
				},
			},
			existedTopic: []string{},
			success:      []string{"test"},
			connectors:   []string{"test"},
			failure:      []string{},
			deletedConnectors: map[string]*connect.Connector{
				"test": {
					Name: "test",
					Config: map[string]string{
						"database.server.name": "bob",
						"slot.name":            "local_manabie_bob",
						"connector.class":      "io.confluent.connect.jdbc.JdbcSinkConnector",
					},
				},
			},
			mockFunc: func(c *ConnectorManagementImpl, oldConnectorList []*connect.Connector) {
				client := mock_kafkaconnect.NewConnectClient(t)
				admin := mock_kafkaconnect.NewKafkaAdmin(t)
				for i := 0; i < len(oldConnectorList); i++ {
					oldConnector := oldConnectorList[i]
					host := mockKafkaConnectServerGetConnectorType("source")
					client.On("Host").Once().Return(host)
					resp := httptest.NewRecorder()
					resp.WriteHeader(http.StatusOK)
					client.On("GetConnector", oldConnector.Name).Once().Return(oldConnector, resp.Result(), nil)
					resp = httptest.NewRecorder()
					resp.WriteHeader(http.StatusNoContent)
					client.On("DeleteConnector", oldConnector.Name).Once().Return(resp.Result(), nil)

				}
				c.Client = client
				c.Admin = admin
				c.Logger = zap.NewExample()
			},
		},
	}

	for _, tc := range testcases {
		tc.mockFunc(tc.connectormgmt, tc.oldConnector)
		deletedConnectors, success, failure := tc.connectormgmt.Delete(tc.connectors)
		assert.Equal(t, tc.success, success)
		assert.Equal(t, tc.failure, failure)
		assert.Equal(t, tc.deletedConnectors, deletedConnectors)
	}
}

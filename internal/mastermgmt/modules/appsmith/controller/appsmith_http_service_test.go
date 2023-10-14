package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/appsmith/application/commands"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/appsmith/domain"
	mock_clients "github.com/manabie-com/backend/mock/lessonmgmt/zoom/clients"

	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

const (
	endpoint        = "/mastermgmt/api/v1/appsmith/track"
	endpointPull    = "/mastermgmt/api/v1/appsmith/pull?branchName=staging"
	endpointPullErr = "/mastermgmt/api/v1/appsmith/pull"
)

func TestAppsmithHTTPService_SaveLog(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock).DatabaseName("appsmith"))
	defer mt.Close()
	logRepo := &MockLogRepo{}
	zcf := configs.AppsmithAPI{}
	s := AppsmithHTTPService{
		ctxzap.Extract(ctx),
		commands.AppsmithCommandHandler{
			DB:      mt.DB,
			LogRepo: logRepo,
		}, zcf}
	gin.SetMode(gin.TestMode)

	logStr := `{
		"context": {
		  "ip": "203.192.213.46",
		  "library": {
			"name": "unknown",
			"version": "unknown"
		  }
		},
		"event": "Instance Active",
		"integrations": {},
		"messageId": "api-1jokIBOkNv8nEmu2fGeNb01G1RC",
		"properties": {
		  "instanceId": "<uuid>"
		},
		"receivedAt": "2020-11-04T08:15:49.537Z",
		"timestamp": "2020-11-04T08:15:49.537Z",
		"type": "track",
		"userId": "203.192.213.46"
	  }
	  `
	var e domain.EventLog
	json.Unmarshal([]byte(logStr), e)

	t.Run("Success", func(t *testing.T) {
		//arrange
		rr := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rr)
		respBody, err := json.Marshal(gin.H{
			"success": true,
		})

		logRepo.On("SaveLog", mock.Anything, mock.Anything, mock.Anything).Once().Return(e, nil)

		request, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(logStr))

		c.Request = request

		//act
		s.Track(c)

		//assert
		assert.NoError(t, err)

		assert.Equal(t, 200, rr.Code)
		assert.Equal(t, respBody, rr.Body.Bytes())
		logRepo.AssertExpectations(t)
	})

	t.Run("Internal error", func(t *testing.T) {
		//arrange
		rr := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rr)
		respBody, err := json.Marshal(gin.H{
			"error": "internal err",
		})

		logRepo.On("SaveLog", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("internal err"))

		request, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(logStr))

		c.Request = request

		//act
		s.Track(c)

		//assert
		assert.NoError(t, err)

		assert.Equal(t, 500, rr.Code)
		assert.Equal(t, respBody, rr.Body.Bytes())
		logRepo.AssertExpectations(t)
	})
}

func TestAppsmithHTTPService_Pull(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	zcf := configs.AppsmithAPI{
		ENDPOINT:      "https://appsmith.staging-green.manabie.io/api/v1",
		ApplicationID: "app_id",
		Authorization: "abc",
	}
	mockHTTPClient := &mock_clients.MockHTTPClient{}
	s := AppsmithHTTPService{
		ctxzap.Extract(ctx),
		commands.AppsmithCommandHandler{
			HTTPClient: mockHTTPClient,
		},
		zcf,
	}
	gin.SetMode(gin.TestMode)

	t.Run("Success", func(t *testing.T) {
		//arrange
		rr := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rr)

		pullResponse := `{}`
		request, err := http.NewRequest(http.MethodGet, endpointPull, nil)
		mockHTTPClient.On("SendRequest", mock.Anything, mock.Anything).
			Return(&http.Response{Body: io.NopCloser(bytes.NewBuffer([]byte(pullResponse)))}, nil).Once()

		mockHTTPClient.On("SendRequest", mock.Anything, mock.Anything).
			Return(&http.Response{Body: io.NopCloser(bytes.NewBuffer([]byte(pullResponse)))}, nil).Once()

		c.Request = request

		//act
		s.PullMetadata(c)

		//assert
		assert.NoError(t, err)

		assert.Equal(t, 200, rr.Code)
	})

	t.Run("Error", func(t *testing.T) {
		//arrange
		rr := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rr)

		pullResponse := `{}`
		request, _ := http.NewRequest(http.MethodGet, endpointPull, nil)
		mockHTTPClient.On("SendRequest", mock.Anything, mock.Anything).
			Return(&http.Response{Body: io.NopCloser(bytes.NewBuffer([]byte(pullResponse)))}, fmt.Errorf("internal err")).Once()

		c.Request = request

		//act
		s.PullMetadata(c)

		assert.Equal(t, 500, rr.Code)
	})

	t.Run("Miss params", func(t *testing.T) {
		//arrange
		rr := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rr)

		pullResponse := `{}`
		request, err := http.NewRequest(http.MethodGet, endpointPullErr, nil)
		mockHTTPClient.On("SendRequest", mock.Anything, mock.Anything).
			Return(&http.Response{Body: io.NopCloser(bytes.NewBuffer([]byte(pullResponse)))}, nil).Once()

		c.Request = request

		//act
		s.PullMetadata(c)

		//assert
		assert.NoError(t, err)

		assert.Equal(t, 500, rr.Code)
	})
}

type MockLogRepo struct {
	mock.Mock
}

func (r *MockLogRepo) SaveLog(arg1 context.Context, arg2 *mongo.Database, arg3 domain.EventLog) (domain.EventLog, error) {
	args := r.Called(arg1, arg2, arg3)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(domain.EventLog), args.Error(1)
}

package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/constants"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/gin-gonic/gin"
)

func Test_DataCleanerController_Handle(t *testing.T) {
	t.Parallel()
	jsm := new(mock_nats.JetStreamManagement)

	t.Run("success send event clean data test", func(t *testing.T) {
		logStr := `{
			"tables": "courses",
			"service": "bob",
			"school_id": "100000",
			"per_batch": 100,
			"before_at": "2023-02-02T13:46:20+07:00",
			"after_at": "2023-02-02T13:45:49+07:00",
			"extra_cond": [
				{
					"table": "courses",
					"condition": " and name like 'text'"
				}
			]
		}`
		dataCleaner := &DataCleanerController{
			JSM: jsm,
		}
		rr := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rr)
		respBody, err := json.Marshal(gin.H{
			"success": true,
		})
		jsm.On("PublishAsyncContext", mock.Anything, constants.SubjectCleanDataTestEventNats, mock.Anything).Once().Return("", nil)
		request, err := http.NewRequest(http.MethodPost, "/draft-http/v1/data_clean/payload", strings.NewReader(logStr))

		c.Request = request
		dataCleaner.Handle(c)
		assert.NoError(t, err)

		assert.Equal(t, 200, rr.Code)
		assert.Equal(t, respBody, rr.Body.Bytes())
	})
	t.Run("fail to parse json send event clean data test", func(t *testing.T) {
		logStr := `{
			aa: "test"
		}`
		dataCleaner := &DataCleanerController{
			JSM: jsm,
		}
		rr := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rr)
		request, err := http.NewRequest(http.MethodPost, "/draft-http/v1/data_clean/payload", strings.NewReader(logStr))

		c.Request = request
		dataCleaner.Handle(c)
		assert.NoError(t, err)

		assert.Equal(t, 400, rr.Code)
	})
}

package controllers

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/manabie-com/backend/internal/enigma/configurations"
	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

var body = map[string]string{
	"event":         "feature-created",
	"createdBy":     "admin",
	"featureToggle": "",
	"timestamp":     "2022-07-27T02:42:10.312Z",
}

const (
	testEmail      = "testemail@manabie.com"
	tesToken       = "test token"
	testDateFormat = "2006-01-02"
	testJiraPath   = "/rest/api/3/version/"
	apiPath        = "/release/update"
)

func TestUpdateReleaseStatus(t *testing.T) {
	t.Parallel()

	t.Run("Error when versionID is missing", func(t *testing.T) {
		server := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
			}),
		)
		defer server.Close()

		resp := httptest.NewRecorder()
		gin.SetMode(gin.TestMode)

		ctx, r := gin.CreateTestContext(resp)
		RegisterUpdateReleaseController(r.Group("release"), zap.NewNop(), &configurations.Config{
			Jira: configs.JiraConfig{
				Email:         testEmail,
				Token:         tesToken,
				APIBaseURL:    server.URL + testJiraPath,
				APITimeFormat: testDateFormat,
			},
		})

		body["featureToggle"] = "xxxx xxxxxxxxxxxxxx"

		jsonValue, _ := json.Marshal(body)
		ctx.Request, _ = http.NewRequest(http.MethodPost, server.URL+apiPath, bytes.NewBuffer(jsonValue))
		r.ServeHTTP(resp, ctx.Request)
		assert.Equal(t, http.StatusBadRequest, resp.Code)

		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		bodyString := string(bodyBytes)
		assert.Equal(t, `{"error":"Can not find version ID in the description"}`, bodyString)
	})

	t.Run("Error when versionID is not found", func(t *testing.T) {
		bodyResponse := `{"errorMessages":["Could not find version for id '9999999'"],"errors":{}}`
		server := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				s := strings.Split(r.URL.Path, "/")
				if s[len(s)-1] != "10000" {
					w.WriteHeader(http.StatusNotFound)
					w.Write([]byte(bodyResponse))
				}
			}),
		)
		defer server.Close()

		resp := httptest.NewRecorder()
		gin.SetMode(gin.TestMode)

		ctx, r := gin.CreateTestContext(resp)
		RegisterUpdateReleaseController(r.Group("release"), zap.NewNop(), &configurations.Config{
			Jira: configs.JiraConfig{
				Email:         testEmail,
				Token:         tesToken,
				APIBaseURL:    server.URL + testJiraPath,
				APITimeFormat: testDateFormat,
			},
		})

		body["featureToggle"] = "VID-9999999 9999999 is a wrong id"

		jsonValue, _ := json.Marshal(body)
		ctx.Request, _ = http.NewRequest(http.MethodPost, server.URL+apiPath, bytes.NewBuffer(jsonValue))
		r.ServeHTTP(resp, ctx.Request)
		assert.Equal(t, http.StatusNotFound, resp.Code)

		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		bodyString := string(bodyBytes)
		assert.Equal(t, `{"errorMessages":["Could not find version for id '9999999'"],"errors":{}}`, bodyString)
	})

	t.Run("Success", func(t *testing.T) {
		server := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				s := strings.Split(r.URL.Path, "/")
				if s[len(s)-1] == "10000" {
					w.WriteHeader(http.StatusOK)
				}
			}),
		)
		defer server.Close()

		resp := httptest.NewRecorder()
		gin.SetMode(gin.TestMode)

		ctx, r := gin.CreateTestContext(resp)
		RegisterUpdateReleaseController(r.Group("release"), zap.NewNop(), &configurations.Config{
			Jira: configs.JiraConfig{
				Email:         testEmail,
				Token:         tesToken,
				APIBaseURL:    server.URL + testJiraPath,
				APITimeFormat: testDateFormat,
			},
		})

		body["featureToggle"] = "VID-10000 10000 is a right id"

		jsonValue, _ := json.Marshal(body)
		ctx.Request, _ = http.NewRequest(http.MethodPost, server.URL+apiPath, bytes.NewBuffer(jsonValue))
		r.ServeHTTP(resp, ctx.Request)
		assert.Equal(t, http.StatusOK, resp.Code)
	})
}

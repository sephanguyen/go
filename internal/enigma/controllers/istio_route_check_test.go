package controllers

// UNCOMMENT ME and run this test locally to verify behavior...

import (
	"io/ioutil"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/manabie-com/backend/internal/enigma/configurations"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type MockEndpoint struct {
	Fail   bool
	Called int32
}

func (i *MockEndpoint) TryEndpoint(host, service string) (ok bool, err error) {
	atomic.AddInt32(&i.Called, 1)
	if i.Fail {
		return false, errors.Errorf("a mock error")
	}
	return true, errors.Errorf("the error we wanted")
}

func TestRouteCheckerGood(t *testing.T) {
	gin.SetMode(gin.TestMode)

	controller := &IstioRouteCheckerController{
		checker: &MockEndpoint{
			Fail:   false,
			Called: 0,
		},
		logger: zap.NewNop(),
		config: &configurations.Config{
			RouteCheckerHosts: []string{
				"web-api.prod.aic.manabie.io",
				"web-api.prod.ga.manabie.io",
			},
			RouteCheckerServices: []string{
				"manabie.bob.UserService",
				"manabie.tom.ChatService",
				"manabie.yasuo.SchoolService",
			},
		},
	}

	r := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(r)
	controller.HandleCheck(c)

	b, err := ioutil.ReadAll(r.Body)
	require.Nil(t, err)
	require.Equal(t, r.Code, 200)
	require.Equal(t, controller.checker.(*MockEndpoint).Called, int32(6))
	require.NotEmpty(t, b)

}

func TestRouteCheckerBad(t *testing.T) {
	gin.SetMode(gin.TestMode)

	controller := &IstioRouteCheckerController{
		checker: &MockEndpoint{
			Fail:   true,
			Called: 0,
		},
		logger: zap.NewNop(),
		config: &configurations.Config{
			RouteCheckerHosts: []string{
				"web-api.prod.aic.manabie.io",
				"web-api.prod.ga.manabie.io",
			},
			RouteCheckerServices: []string{
				"manabie.bob.UserService",
				"manabie.tom.ChatService",
				"manabie.yasuo.SchoolService",
			},
		},
	}

	r := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(r)
	controller.HandleCheck(c)

	b, err := ioutil.ReadAll(r.Body)
	require.Nil(t, err)
	require.Equal(t, r.Code, 502)
	require.Equal(t, controller.checker.(*MockEndpoint).Called, int32(6))
	require.NotEmpty(t, b)

}

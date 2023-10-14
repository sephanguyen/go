package controllers

import (
	"net/http/httptest"
	"testing"

	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestHealthCheckController_Status(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := testutil.NewMockDB()
	row := &mock_database.Row{}
	db.DB.On("QueryRow", mock.Anything, mock.AnythingOfType("string")).Once().Return(row, nil)
	row.On("Scan", mock.Anything).Once().Return(nil)
	controller := &HealthCheckController{
		Logger: zap.NewNop(),
		DB:     db.DB,
	}

	r := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(r)
	controller.HealthCheckStatus(c)
}

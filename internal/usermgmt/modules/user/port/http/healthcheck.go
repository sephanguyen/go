package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthCheckService struct {
}

func (s *HealthCheckService) Status(c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Message: "ok",
	})
}

package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/appsmith/application/commands"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/appsmith/domain"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AppsmithHTTPService struct {
	Logger                 *zap.Logger
	AppsmithCommandHandler commands.AppsmithCommandHandler
	AppsmithAPI            configs.AppsmithAPI
}

// Track events from Appsmith
func (a *AppsmithHTTPService) Track(c *gin.Context) {
	req, logger, err := a.toAppsmithEvent(c)

	if err != nil {
		logger.Error(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	_, err = a.AppsmithCommandHandler.SaveLog(c.Request.Context(), req)

	if err != nil {
		logger.Error(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (a *AppsmithHTTPService) toAppsmithEvent(c *gin.Context) (domain.EventLog, *zap.Logger, error) {
	logger := a.Logger.With(zap.String("service", "appsmith"))

	var ev domain.EventLog
	jsonData, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil, logger, err
	}
	err = json.Unmarshal(jsonData, &ev)

	if err != nil {
		return nil, logger, err
	}

	logger = logger.With(zap.String("event", fmt.Sprint(ev)))
	return ev, logger, nil
}

func (a *AppsmithHTTPService) PullMetadata(c *gin.Context) {
	branchName := c.Query("branchName")
	if len(branchName) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "branchName is required",
		})
		return
	}
	// discard before pull metadata
	_, err := a.AppsmithCommandHandler.DiscardChange(c.Request.Context(), branchName, a.AppsmithAPI)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error ": fmt.Sprintf("Appsmith discard error: %s", err.Error()),
		})
		return
	}
	res, err := a.AppsmithCommandHandler.PullMetadata(c.Request.Context(), branchName, a.AppsmithAPI)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error ": fmt.Sprintf("Appsmith pull metadata error: %s", err.Error()),
		})
		return
	}
	c.JSON(http.StatusOK, res)
}

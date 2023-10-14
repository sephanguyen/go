package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/manabie-com/backend/internal/draft/configurations"
	"github.com/manabie-com/backend/internal/draft/middlewares"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/nats"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/proto"
)

type DataCleanerController struct {
	JSM nats.JetStreamManagement
}

func RegisterDataCleanerController(c *configurations.Config, r *gin.Engine, d *DataCleanerController) {
	r.POST("/draft-http/v1/data_clean/payload", middlewares.VerifySignature(middlewares.HeaderKey, c.DraftAPISecret), d.Handle)
}

func (d *DataCleanerController) Handle(c *gin.Context) {
	ctx := c.Request.Context()
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
	}
	var pl CleanDataPayload
	if err := json.Unmarshal(payload, &pl); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Errorf("unable to unmarshal dat to message: %w", err),
		})
		return
	}

	msg, err := proto.Marshal(&npb.EventDataClean{
		Service:   pl.Service,
		Tables:    pl.Tables,
		SchoolId:  pl.SchoolID,
		BeforeAt:  pl.BeforeAt,
		AfterAt:   pl.AfterAt,
		ExtraCond: toExtraConds(pl.ExtraCond),
		PerBatch:  int32(pl.PerBatch),
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Errorf("unable to marshal dat to message: %w", err),
		})
	}
	_, err = d.JSM.PublishAsyncContext(ctx, constants.SubjectCleanDataTestEventNats, msg)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Errorf("PublishAsyncContext: %w", err),
		})
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func toExtraConds(extraConds []ExtraCond) []*npb.ExtraCond {
	results := make([]*npb.ExtraCond, 0, len(extraConds))
	for _, e := range extraConds {
		results = append(results, &npb.ExtraCond{
			Table:     e.Table,
			Condition: e.Condition,
		})
	}

	return results
}

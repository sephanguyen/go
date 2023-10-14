package controllers

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/enigma/middlewares"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/nats"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type CloudConvertController struct {
	logger *zap.Logger
	JSM    nats.JetStreamManagement
}

func RegisterCloudConvertController(r *gin.RouterGroup, l *zap.Logger, jsm nats.JetStreamManagement) {
	c := &CloudConvertController{
		logger: l,
		JSM:    jsm,
	}

	r.POST("/job-events", c.HandleJobEvents)
}

type cloudConvertJobData struct {
	Event string `json:"event"`
	Job   struct {
		ID        string     `json:"id"`
		Status    string     `json:"status"`
		CreatedAt *time.Time `json:"created_at"`
		StartedAt *time.Time `json:"started_at"`
		EndedAt   *time.Time `json:"ended_at"`
		Tasks     []struct {
			ID        string      `json:"id"`
			Name      string      `json:"name"`
			Operation string      `json:"operation"`
			Status    string      `json:"status"`
			Message   interface{} `json:"message"`
			Percent   float32     `json:"percent"`
			Result    struct {
				Files []struct {
					Dir      string `json:"dir"`
					Filename string `json:"filename"`
				} `json:"files"`
			} `json:"result"`
			CreatedAt *time.Time `json:"created_at"`
			StartedAt *time.Time `json:"started_at"`
			EndedAt   *time.Time `json:"ended_at"`
			Links     struct {
				Self string `json:"self"`
			} `json:"links"`
		} `json:"tasks"`
		Links struct {
			Self string `json:"self"`
		} `json:"links"`
	} `json:"job"`
}

func (c *CloudConvertController) HandleJobEvents(ctx *gin.Context) {
	payload := middlewares.PayloadFromContext(ctx)
	if len(payload) == 0 {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "payload body is empty",
		})
		return
	}

	req := &cloudConvertJobData{}
	if err := json.Unmarshal(payload, req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	data := &npb.CloudConvertJobData{
		JobId:          req.Job.ID,
		JobStatus:      req.Event,
		Signature:      ctx.GetHeader(middlewares.CloudConvertSigKey),
		RawPayload:     payload,
		ConvertedFiles: parseCloudConvertConvertedFiles(req),
	}
	msg, err := proto.Marshal(data)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if _, err := c.JSM.PublishContext(ctx, constants.SubjectCloudConvertJobEventNatsJS, msg); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

func parseCloudConvertConvertedFiles(data *cloudConvertJobData) []string {
	var convertedFiles []string

	tasks := data.Job.Tasks
	for _, task := range tasks {
		// export task name contains the "export-" string.
		if strings.Contains(task.Name, "export-") {
			convertedFiles = make([]string, 0, len(task.Result.Files))

			for _, f := range task.Result.Files {
				// f.Dir is the folder where the converted files exists, e.g. assignments/2021-01-01/random-string/
				// f.Filename is the converted filename, e.g. filename-1.png
				filename := f.Dir + url.PathEscape(f.Filename) // f.Filename may contain special characters
				convertedFiles = append(convertedFiles, filename)
			}
			break
		}
	}

	return convertedFiles
}

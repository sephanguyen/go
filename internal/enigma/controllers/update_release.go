package controllers

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	re "regexp"
	"time"

	"github.com/manabie-com/backend/internal/enigma/configurations"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var (
	versionID string
	r         = re.MustCompile(`(VID-)(\d+)`)
)

type toggleInformation struct {
	Event         string `json:"event"`
	CreatedBy     string `json:"createdBy"`
	FeatureToggle string `json:"featureToggle"`
	Timestamp     string `json:"timestamp"`
}

type UpdateReleaseController struct {
	logger *zap.Logger
	config *configurations.Config
}

func RegisterUpdateReleaseController(r *gin.RouterGroup, zapLogger *zap.Logger, c *configurations.Config) {
	h := &UpdateReleaseController{
		logger: zapLogger,
		config: c,
	}
	r.POST("/update", h.UpdateReleaseStatus)
}

func (h *UpdateReleaseController) ParseTime(timeString string) string {
	parsed, err := time.Parse(time.RFC3339, timeString)

	if err != nil {
		h.logger.Error(err.Error())
	}
	return parsed.Format(h.config.Jira.APITimeFormat)
}

func (h *UpdateReleaseController) ParseVersionID(description string) string {
	groupMatched := r.FindStringSubmatch(description)
	if len(groupMatched) < 2 {
		return ""
	}
	return groupMatched[2]
}

func (h *UpdateReleaseController) UpdateReleaseStatus(c *gin.Context) {
	h.logger.Info("Update release status")

	// Bind request body which receives from unleash
	var toggle toggleInformation
	err := c.BindJSON(&toggle)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("Release description: " + toggle.FeatureToggle)

	// Get versionID from toggle description
	versionID = h.ParseVersionID(toggle.FeatureToggle)
	if versionID == "" {
		h.logger.Info("Can not find version ID in the description")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Can not find version ID in the description"})
		return
	}

	// Make a PUT request to Jira to update version
	client := &http.Client{}
	postBody, _ := json.Marshal(map[string]string{
		"id":          versionID,
		"released":    "true",
		"releaseDate": h.ParseTime(toggle.Timestamp),
	})
	postBodyBuffer := bytes.NewBuffer(postBody)
	jiraAPIUrl := h.config.Jira.APIBaseURL + versionID
	req, err := http.NewRequest("PUT", jiraAPIUrl, postBodyBuffer)

	if err != nil {
		h.logger.Error(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req.SetBasicAuth(h.config.Jira.Email, h.config.Jira.Token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)

	//Handle Error
	if err != nil {
		h.logger.Error(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	//Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		h.logger.Error(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info(string(body))
	c.Data(resp.StatusCode, "application/json", body)
}

package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type HealthCheckController struct {
	Logger *zap.Logger
	DB     database.Ext
}

func RegisterHealthCheckController(r *gin.RouterGroup, zapLogger *zap.Logger, db database.Ext) {
	h := &HealthCheckController{
		Logger: zapLogger,
		DB:     db,
	}

	r.GET("/status", h.HealthCheckStatus)
	r.POST("/jprep", h.JPREPHealthCheck)
}

type Message struct {
	URL          string   `json:"url"`
	Paths        []string `json:"paths"`
	ErrorCode    int      `json:"error_code"`
	ContentMatch string   `json:"content_match"`
}

func parseHTTPRequest(r *http.Request) (*Message, error) {
	// print the request body
	var buf bytes.Buffer
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading request body: %w", err)
	}
	fmt.Printf("Uptime check request body: %s\n", buf.String())

	m := Message{}

	if err := json.NewDecoder(&buf).Decode(&m); err != nil {
		return nil, fmt.Errorf("json.NewDecoder: %w", err)
	}
	if m.URL == "" {
		return nil, fmt.Errorf("url is empty")
	}

	url := html.EscapeString(m.URL)

	return &Message{
		URL:          url,
		ErrorCode:    m.ErrorCode,
		ContentMatch: m.ContentMatch,
		Paths:        m.Paths,
	}, nil
}

func checkPUTMethod(c *gin.Context, message *Message) {
	// Create an HTTP client with a custom timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Create the PUT request
	req, err := http.NewRequest("PUT", message.URL, nil)
	if err != nil {
		// http.Error(w, fmt.Sprintf("error creating request: %v", err), http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": fmt.Sprintf("error creating request %s: %v", message.URL, err),
		})
		return
	}

	// Execute the request
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": fmt.Sprintf("error sending request %s: %v", message.URL, err),
		})

		return
	}
	defer resp.Body.Close()

	if message.ErrorCode != 0 && resp.StatusCode != message.ErrorCode {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": fmt.Sprintf("request %s: status code %d did not match, got %d", message.URL, message.ErrorCode, resp.StatusCode),
		})
		return
	}

	// Read the response body
	var buf bytes.Buffer
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": fmt.Sprintf("request %s: error reading response body: %v", message.URL, err),
		})
		return
	}

	// Check if the response body matches the expected content
	if message.ContentMatch != "" && !strings.Contains(buf.String(), message.ContentMatch) {
		// http.Error(w, fmt.Sprintf("unexpected: content did not match %s", message.ContentMatch), http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": fmt.Sprintf("request %s: content did not match %s", message.URL, message.ContentMatch),
		})
		return
	}
}

func (h *HealthCheckController) JPREPHealthCheck(c *gin.Context) {
	// Parse the request body
	message, err := parseHTTPRequest(c.Request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": fmt.Sprintf("parseHTTPRequest: %v", err),
		})
		return
	}

	// loop message.paths
	for _, path := range message.Paths {
		checkPUTMethod(c, &Message{
			URL:          fmt.Sprintf("%s%s", message.URL, path),
			ErrorCode:    message.ErrorCode,
			ContentMatch: message.ContentMatch,
			Paths:        []string{},
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "OK",
	})
}

func (h *HealthCheckController) HealthCheckStatus(c *gin.Context) {
	claim := &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: fmt.Sprint(constants.JPREPSchool),
		},
	}
	ctx := interceptors.ContextWithJWTClaims(c, claim)
	h.Logger.Info("Healthcheck status")

	var selectResult int
	row := h.DB.QueryRow(ctx, "SELECT 1")
	err := row.Scan(&selectResult)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

package eureka

import (
	"bytes"
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/manabie-com/backend/internal/eureka/configurations"
	repo "github.com/manabie-com/backend/internal/eureka/repositories/learning_history_data_sync"
	service "github.com/manabie-com/backend/internal/eureka/services/learning_history_data_sync"
	"github.com/manabie-com/backend/internal/golibs/alert"
	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
	"google.golang.org/api/idtoken"
)

var organizations = []string{"manabie", "tokyo"}

func init() {
	bootstrap.RegisterJob("eureka_upload_data_file_to_withus", uploadDataFileToWithus).
		Desc("Cmd to upload data file from eureka to Withus")
}

func uploadDataFileToWithus(ctx context.Context, c configurations.Config, rsc *bootstrap.Resources) error {
	if err := uploadDataFileToWithusByTenant(ctx, c, rsc, "-2147483630", "Managara Base"); err != nil {
		return fmt.Errorf("uploadDataFileToWithusByTenant failed for Managara Base: %s", err)
	}
	if err := uploadDataFileToWithusByTenant(ctx, c, rsc, "-2147483629", "Managara HS"); err != nil {
		return fmt.Errorf("uploadDataFileToWithusByTenant failed for Managara HS: %s", err)
	}
	return nil
}

func uploadDataFileToWithusByTenant(ctx context.Context, c configurations.Config, rsc *bootstrap.Resources, schoolID, schoolName string) error {
	// Setup logger
	zapLogger := rsc.Logger()
	ctx = ctxzap.ToContext(ctx, zapLogger)
	szapLogger := zapLogger.Sugar()

	// Skip organizations without tokyo,manabie.
	if !sliceutils.Contains(organizations, c.Common.Organization) {
		zapLogger.Info(fmt.Sprintf("Skipped by organization: %s", c.Common.Organization))
		return nil
	}

	szapLogger.Infof("school name ----: %s", schoolName)
	szapLogger.Infof("school id ----: %s", schoolID)

	// Setup Mastermgmt client & service
	mastermgmtConn := rsc.GRPCDial("mastermgmt")

	configurationClient := mpb.NewInternalServiceClient(mastermgmtConn)
	relayServerCfgKey := "syllabus.study_plan.learning_history.relay_server_url"
	configs, err := configurationClient.GetConfigurations(ctx, &mpb.GetConfigurationsRequest{Keyword: relayServerCfgKey, OrganizationId: schoolID})
	if err != nil {
		return fmt.Errorf("configurationClient.GetConfigurations: %s", err)
	}

	var relayServerURL string
	for _, config := range configs.Items {
		if config.ConfigKey == relayServerCfgKey {
			relayServerURL = config.ConfigValue
			break
		}
	}
	if relayServerURL == "" {
		zapLogger.Info("Relay Server URL is not set in DB. Skipped")
		return nil
	}

	audience := fmt.Sprintf("https://%s", relayServerURL)
	client, err := idtoken.NewClient(ctx, audience)
	if err != nil {
		return fmt.Errorf("idtoken.NewClient: %w", err)
	}

	// for db RLS query
	ctx = auth.InjectFakeJwtToken(ctx, schoolID)

	dbPool := rsc.DBWith("eureka")
	dbTrace := &database.DBTrace{
		DB: dbPool,
	}
	httpClient := http.Client{Timeout: time.Duration(10) * time.Second}
	alertClient := &alert.SlackImpl{
		WebHookURL: c.SyllabusSlackWebHook,
		HTTPClient: httpClient,
	}
	learningHistoryRepo := &repo.LearningHistoryDataSyncRepo{}
	learningHistoryService := &service.LearningHistoryDataSyncService{
		DBTrace:                     dbTrace,
		LearningHistoryDataSyncRepo: learningHistoryRepo,
		Alert:                       alertClient,
	}

	zapLogger.Info(
		"-----Upload data file to Withus-----",
	)

	reportL6FileName, reportM1FileName, learningHistoryData, err := learningHistoryService.ExportLearningHistoryData(ctx)
	if err != nil {
		resErr := fmt.Errorf("failed to export learning history data: %w", err)
		notifySlackError(c, schoolName, zapLogger, alertClient, resErr)
		return resErr
	}

	// Request configs
	var selectedFileName string
	var path string
	switch schoolID {
	// managara-base
	case "-2147483630":
		path = "managara-base"
		selectedFileName = reportL6FileName
	// managara-hs
	case "-2147483629":
		path = "managara-hs"
		selectedFileName = reportM1FileName
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	contentType := writer.FormDataContentType()
	part, err := writer.CreateFormFile("REPORTS", selectedFileName)
	if err != nil {
		zapLogger.Error(
			"failed to create form file",
			zap.Error(err),
		)
		resErr := fmt.Errorf("failed to create form file %s: %w", selectedFileName, err)
		notifySlackError(c, schoolName, zapLogger, alertClient, resErr)
		return resErr
	}
	_, err = part.Write(learningHistoryData)
	if err != nil {
		zapLogger.Error(
			"failed to copy byte to form file",
			zap.Error(err),
		)
		resErr := fmt.Errorf("failed to copy byte to form file %s: %w", selectedFileName, err)
		notifySlackError(c, schoolName, zapLogger, alertClient, resErr)
		return resErr
	}
	writer.Close()

	res, err := client.Post(fmt.Sprintf("https://%s/upload-file/%s", relayServerURL, path), contentType, bytes.NewReader(body.Bytes()))
	if err != nil || res.StatusCode != 200 {
		if err == nil {
			err = fmt.Errorf("request is not success: Status code: %s", res.Status)
		}
		resErr := fmt.Errorf("failed to upload file: %w", err)
		notifySlackError(c, schoolName, zapLogger, alertClient, resErr)
		return resErr
	}
	defer res.Body.Close()
	return nil
}

func notifySlackError(cfg configurations.Config, schoolName string, zapLogger *zap.Logger, slackAlert alert.SlackFactory, err error) {
	att := alert.InitAttachment("error")
	att.AddSourceInfo(schoolName, cfg.Common.Environment)
	att.AddErrorInfo(err)
	err = slackAlert.Send(alert.Payload{
		Text: "Learning history data sync failed",
		Attachments: []alert.IAttachment{
			att,
		},
	})
	if err != nil {
		zapLogger.Error(
			"failed to send slack alert",
			zap.Error(err),
		)
	}
}

package services

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/manabie-com/backend/internal/entryexitmgmt/constant"
	"github.com/manabie-com/backend/internal/entryexitmgmt/entities"
	"github.com/manabie-com/backend/internal/entryexitmgmt/services/uploader"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/try"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	eepb "github.com/manabie-com/backend/pkg/manabuf/entryexitmgmt/v1"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"github.com/jackc/pgconn"
	pgx "github.com/jackc/pgx/v4"
	"github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type QrCodeContent struct {
	QrCode  string           `json:"qrcode"`
	Version constant.Version `json:"version"`
}

func (s *EntryExitModifierService) Generate(ctx context.Context, studentID string) (string, error) {
	studentQRRecord, err := s.StudentQRRepo.FindByID(ctx, s.DB, studentID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return "", fmt.Errorf("err StudentQRRepo.FindByID %w", err)
	}

	// if student has no QR, generate
	if studentQRRecord == nil {
		return s.generateAndUploadQr(ctx, studentID, constant.V2)
	}

	// if student QR is outdated, generate the latest and allowed version.
	latestAllowedQrVersion := constant.V2
	if studentQRRecord.Version.String != string(latestAllowedQrVersion) {
		return s.generateAndUploadQr(ctx, studentID, latestAllowedQrVersion)
	}

	return studentQRRecord.QRURL.String, nil
}

func (s *EntryExitModifierService) GenerateBatchQRCodes(ctx context.Context, req *eepb.GenerateBatchQRCodesRequest) (*eepb.GenerateBatchQRCodesResponse, error) {
	studentIDsLength := len(req.StudentIds)
	if studentIDsLength < 1 {
		return nil, status.Error(codes.InvalidArgument, "student ids cannot be empty")
	}

	chanResps := make(chan *eepb.GenerateBatchQRCodesResponse_GeneratedQRCodesURL, studentIDsLength)
	chanErrs := make(chan *eepb.GenerateBatchQRCodesResponse_GenerateBatchQRCodesError, studentIDsLength)

	var wg sync.WaitGroup

	for _, studentID := range req.StudentIds {
		wg.Add(1)
		go s.generateBatchQRCodes(ctx, studentID, &wg, chanErrs, chanResps)
	}

	go func() {
		wg.Wait()
		close(chanErrs)
		close(chanResps)
	}()

	qrCodes := make([]*eepb.GenerateBatchQRCodesResponse_GeneratedQRCodesURL, 0, studentIDsLength)
	errors := make([]*eepb.GenerateBatchQRCodesResponse_GenerateBatchQRCodesError, 0, studentIDsLength)

	for i := 0; i < studentIDsLength; i++ {
		select {
		case res, ok := <-chanResps:
			if ok {
				qrCodes = append(qrCodes, res)
			}
		case err, ok := <-chanErrs:
			if ok {
				errors = append(errors, err)
			}
		}
	}

	return &eepb.GenerateBatchQRCodesResponse{
		Errors:  errors,
		QrCodes: qrCodes,
	}, nil
}

func (s *EntryExitModifierService) generateBatchQRCodes(
	ctx context.Context,
	studentID string,
	wg *sync.WaitGroup,
	chanErrs chan *eepb.GenerateBatchQRCodesResponse_GenerateBatchQRCodesError,
	chanResps chan *eepb.GenerateBatchQRCodesResponse_GeneratedQRCodesURL) {
	defer wg.Done()

	url, err := s.Generate(ctx, studentID)
	if err != nil {
		chanErrs <- &eepb.GenerateBatchQRCodesResponse_GenerateBatchQRCodesError{
			StudentId: studentID,
			Error:     err.Error(),
		}
		return
	}

	chanResps <- &eepb.GenerateBatchQRCodesResponse_GeneratedQRCodesURL{
		StudentId: studentID,
		Url:       url,
	}
}

func createTempDir() (string, error) {
	dir, err := ioutil.TempDir("", "qrcode")
	if err != nil {
		return "", err
	}

	return dir, nil
}

func (s *EntryExitModifierService) cleanup(objectPath string) {
	if err := os.RemoveAll(objectPath); err != nil {
		s.logger.Warnf("os.RemoveAll: %v", err)
	}
}

func (s *EntryExitModifierService) generateAndUploadQr(ctx context.Context, studentID string, version constant.Version) (string, error) {
	// For future versions, add a condition how to generate the encrypted content here
	// if version == constant.V3 { encryptedContent, err := generateQrContentV3(studentID) }

	encryptedContent, err := s.generateQrContentV2(studentID)
	if err != nil {
		return "", fmt.Errorf("generateQrContentV2: %v", err)
	}

	qrc, err := qrcode.New(encryptedContent)
	if err != nil {
		return "", fmt.Errorf("qrcode.New: %v", err)
	}

	tempDir, err := createTempDir()
	if err != nil {
		return "", fmt.Errorf("createTempDir: %v", err)
	}
	defer s.cleanup(tempDir)

	objectPath := fmt.Sprintf("%v/%v.png", tempDir, encryptedContent)
	objectWriter, err := standard.New(objectPath)
	if err != nil {
		return "", fmt.Errorf("standard.New: %v", err)
	}

	// Save to File
	if err := qrc.Save(objectWriter); err != nil {
		return "", fmt.Errorf("qrc.Save: %v", err)
	}

	uploadService, err := s.UploadServiceSelector.GetUploadService()
	if err != nil {
		return "", fmt.Errorf("s.UploadServiceSelector.GetUploadService: %v", err)
	}

	uploader, err := uploadService.InitUploader(ctx, &uploader.UploadRequest{
		ObjectName:    encryptedContent,
		FileExtension: constant.PNG,
		ContentType:   constant.ContentType,
	})

	if err != nil {
		return "", fmt.Errorf("FileStoreUploadService.InitUploader: %v", err)
	}

	studentQR := &entities.StudentQR{
		StudentID: database.Text(studentID),
		QRURL:     database.Text(uploader.DownloadURL),
		Version:   database.Text(string(version)),
	}

	// To retry saving and uploading of QR if student_id does not exist
	// This is because there is a delay in syncing the students in kafka
	if err := try.Do(func(attempt int) (bool, error) {
		err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) (err error) {
			return s.saveAndUpload(ctx, tx, uploader, studentQR, objectPath)
		})

		// If no error, return and not retry
		if err == nil {
			return false, nil
		}

		// If error is not foreign key violation error or student QR RLS error, return and not retry
		// SQLSTATE 23503 - FOREIGN KEY VIOLATION error
		// SQLSTATE 42501 - RLS error
		if err != nil && err.Error() != constant.PgConnForeignKeyError && err.Error() != constant.StudentQrRLSError {
			return false, err
		}

		// Here, the error is foreign key violation error or student QR RLS error. Retry the insert and upload of QR
		time.Sleep(s.retryQuerySleep)
		log.Printf("Retrying the saving and uploading of QR Code. Attempt: %d \n", attempt)
		return attempt < 10, fmt.Errorf("cannot create student QR, err %v", err)
	}); err != nil {
		if err.Error() == constant.PgConnDuplicateError {
			return "", nil
		}
		return "", err
	}

	return uploader.DownloadURL, nil
}

func (s *EntryExitModifierService) saveAndUpload(ctx context.Context, tx pgx.Tx, uploader *uploader.UploadInfo, studentQR *entities.StudentQR, objectPath string) (err error) {
	if err := s.StudentQRRepo.Upsert(ctx, tx, studentQR); err != nil {
		pgerr, ok := errors.Unwrap(err).(*pgconn.PgError)
		if ok && pgerr.Code == "23505" {
			return errors.New(constant.PgConnDuplicateError)
		}

		if ok && pgerr.Code == "23503" {
			return errors.New(constant.PgConnForeignKeyError)
		}

		if ok && pgerr.Code == "42501" && strings.Contains(pgerr.Error(), "student_qr") {
			return errors.New(constant.StudentQrRLSError)
		}

		return fmt.Errorf("s.StudentQRRepo.Upsert: %v", err)
	}

	if err := uploader.DoUploadFromFile(ctx, objectPath); err != nil {
		return fmt.Errorf("uploader.DoUploadFromFile: %v", err)
	}

	return nil
}

func (s *EntryExitModifierService) CheckAutoGenQRCodeIsEnabled(ctx context.Context) (bool, error) {
	var configValue string
	resourcePath, err := interceptors.ResourcePathFromContext(ctx)
	if err != nil {
		return false, fmt.Errorf("s.MasterConfigurationService.GetConfigurations err: %s", err)
	}
	getConfigurationsRequest := &mpb.GetConfigurationsRequest{
		Paging:         &cpb.Paging{},
		Keyword:        constant.AutoGenQRCodeConfigKey,
		OrganizationId: resourcePath,
	}

	res, err := s.MasterConfigurationService.GetConfigurations(SignCtx(ctx), getConfigurationsRequest)

	if err != nil {
		return false, fmt.Errorf("s.MasterConfigurationService.GetConfigurations err: %s", err)
	}

	if len(res.GetItems()) == 0 {
		return false, fmt.Errorf("s.MasterConfigurationService.GetConfigurations err: no %v config key found", constant.AutoGenQRCodeConfigKey)
	}

	configurationResponse := res.GetItems()[0]
	configValue = configurationResponse.GetConfigValue()

	if configValue == "on" {
		return true, nil
	}

	return false, nil
}

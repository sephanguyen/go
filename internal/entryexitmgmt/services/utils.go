package services

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/entryexitmgmt/constant"
	"github.com/manabie-com/backend/internal/entryexitmgmt/entities"
	"github.com/manabie-com/backend/internal/entryexitmgmt/services/uploader"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/unleashclient"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	eepb "github.com/manabie-com/backend/pkg/manabuf/entryexitmgmt/v1"

	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UploadServiceSelector struct {
	SdkUploadService  uploader.Uploader
	CurlUploadService uploader.Uploader
	UnleashClient     unleashclient.ClientInstance
	Env               string
}

func (selector *UploadServiceSelector) GetUploadService() (uploader.Uploader, error) {
	useSDK, err := selector.UnleashClient.IsFeatureEnabled("BACKEND_EntryExit_EntryExitManagement_SDK_Upload", selector.Env)
	if err != nil {
		return nil, fmt.Errorf("unleashClient.IsFeatureEnabled: %w", err)
	}

	uploadService := selector.CurlUploadService
	if useSDK {
		uploadService = selector.SdkUploadService
	}

	return uploadService, nil
}

func setEntryExitValue(entryExitID int32, studentID string, entryAt, exitAt *timestamppb.Timestamp) (*entities.StudentEntryExitRecords, error) {
	e := new(entities.StudentEntryExitRecords)
	database.AllNullEntity(e)

	errs := []error{}

	if entryExitID != 0 {
		errs = append(errs, e.ID.Set(entryExitID))
	}

	if strings.TrimSpace(studentID) != "" {
		errs = append(errs, e.StudentID.Set(studentID))
	}

	if entryAt != nil {
		errs = append(errs, e.EntryAt.Set(entryAt.AsTime().UTC()))
	}

	if exitAt != nil {
		if !exitAt.AsTime().IsZero() {
			errs = append(errs, e.ExitAt.Set(exitAt.AsTime().UTC()))
		}
	}

	if err := multierr.Combine(errs...); err != nil {
		return nil, fmt.Errorf("multierr.Combine: %w", err)
	}

	return e, nil
}

func generateEntryExitCreateRecord(studentID string, entryAt, exitAt *timestamppb.Timestamp) (*entities.StudentEntryExitRecords, error) {
	return setEntryExitValue(0, studentID, entryAt, exitAt)
}

func generateEntryExitUpdateRecord(entryExitID int32, entryAt, exitAt *timestamppb.Timestamp) (*entities.StudentEntryExitRecords, error) {
	return setEntryExitValue(entryExitID, "", entryAt, exitAt)
}

// get school of the student who scan
func (s *EntryExitModifierService) getParentIDs(ctx context.Context, db database.QueryExecer, studentID string) ([]string, error) {
	var parentIds []string
	parentIds, err := s.StudentParentRepo.GetParentIDsByStudentID(ctx, db, studentID)
	if err != nil {
		return parentIds, err
	}

	return parentIds, nil
}

func generateEntryExitNotifyTitleMessage(notifyDetails *EntryExitNotifyDetails) *EntryExitNotifyDetails {
	studentName := notifyDetails.Student.GetName()
	studentCountry := cpb.Country(cpb.Country_value[notifyDetails.Student.Country.String])
	recordType := notifyDetails.RecordType

	switch studentCountry {
	case cpb.Country_COUNTRY_JP:
		notifyDetails.Title = "入退室記録"
	default:
		notifyDetails.Title = "Entry & Exit Activity"
	}

	switch recordType {
	case eepb.RecordType_QR_CODE_SCAN:
		touchTime := notifyDetails.TouchTime
		formattedDateTime := fmt.Sprintf("%d/%02d/%02d %02d:%02d",
			touchTime.Year(), touchTime.Month(), touchTime.Day(),
			touchTime.Hour(), touchTime.Minute())
		notifyDetails.Message = studentName + " exited the center at " + formattedDateTime
		if studentCountry == cpb.Country_COUNTRY_JP {
			// japanese translation 2021/10/14 11:00:00に<生徒名>が教室から退室しました
			notifyDetails.Message = formattedDateTime + "に" + studentName + "が教室から退室しました"
		}
		if notifyDetails.TouchEvent.String() == constant.TouchEntry {
			notifyDetails.Message = studentName + " entered the center at " + formattedDateTime
			if studentCountry == cpb.Country_COUNTRY_JP {
				// japanese translation 2021/10/14 10:00:00に<生徒名>が教室に入室しました
				notifyDetails.Message = formattedDateTime + "に" + studentName + "が教室に入室しました"
			}
		}
	case eepb.RecordType_CREATE_MANUAL:
		notifyDetails.Message = "There are new records for entry & exit of " + studentName + ". Please review them on your kid history"
		if studentCountry == cpb.Country_COUNTRY_JP {
			// japanese translation <生徒名>の新しい入退室記録が追加されました。入退室記録から確認してください
			notifyDetails.Message = studentName + "の新しい入退室記録が追加されました。入退室記録から確認してください"
		}

	case eepb.RecordType_UPDATE_MANUAL:
		notifyDetails.Message = "There are updated records for entry & exit of " + studentName + ". Please review them on your kid history"
		if studentCountry == cpb.Country_COUNTRY_JP {
			// japanese translation <生徒名>の新しい入退室記録が更新されました。入退室記録から確認してください
			notifyDetails.Message = studentName + "の新しい入退室記録が更新されました。入退室記録から確認してください"
		}
	}
	return notifyDetails
}

func isSameDate(t1, t2 time.Time) bool {
	dateFormat := "2006-01-02 00:00:00"
	return t1.Format(dateFormat) == t2.Format(dateFormat)
}

// isZeroTime checks if the Unix value is zero (a nil *timestamppb.Timestamp) or a zero value time.Time
func isZeroTime(t time.Time) bool {
	return t.Unix() == 0 || t.IsZero()
}

func validateEntryDateTime(entry, exit time.Time) error {
	if isZeroTime(entry) {
		return errors.New("this field is required|date|time")
	}

	// set datetime to compare with the entry datetime
	toCompare := exit
	if isZeroTime(exit) {
		toCompare = time.Now()
	}

	dateTimeLabel := "date"
	if isSameDate(entry, toCompare) {
		dateTimeLabel = "time"
	}

	if entry.After(time.Now()) {
		return errors.New(fmt.Sprintf("entry %[1]s must not be a future %[1]s", dateTimeLabel))
	}

	return nil
}

func validateExitDateTime(entry, exit time.Time) error {
	if isZeroTime(exit) {
		return nil
	}

	dateTimeLabel := "date"
	if isSameDate(entry, exit) {
		dateTimeLabel = "time"
	}

	// check if exit time is a future date
	if exit.After(time.Now()) {
		return errors.New(fmt.Sprintf("exit %[1]s must not be a future %[1]s", dateTimeLabel))
	}

	// check if exit time is earlier than entry time
	if exit.Before(entry) {
		return errors.New(fmt.Sprintf("entry %[1]s must be earlier than exit %[1]s|%[1]s", dateTimeLabel))
	}

	return nil
}

func validateRequest(payload *eepb.EntryExitPayload) error {
	entryTime := payload.EntryDateTime.AsTime()
	exitTime := payload.ExitDateTime.AsTime()

	if strings.TrimSpace(payload.StudentId) == "" {
		return errors.New("student id cannot be empty")
	}

	if err := validateEntryDateTime(entryTime, exitTime); err != nil {
		return err
	}

	if err := validateExitDateTime(entryTime, exitTime); err != nil {
		return err
	}

	return nil
}

func (s *EntryExitModifierService) generateQrContentV2(studentID string) (string, error) {
	secretByte, err := hex.DecodeString(s.encryptSecretKeyV2)
	if err != nil {
		return "", fmt.Errorf("hex.DecodeString: %v", err)
	}

	// encrypt student ID
	encryptedStudentID, err := s.CryptV2.Encrypt(studentID, secretByte)
	if err != nil {
		return "", fmt.Errorf("encrypt: %v", err)
	}

	qrCodeByte, _ := json.Marshal(QrCodeContent{QrCode: encryptedStudentID, Version: constant.V2})

	return base64.URLEncoding.EncodeToString(qrCodeByte), nil
}

func (s *EntryExitModifierService) getStudentIDFromQr(ctx context.Context, content string) (string, error) {
	var studentID string
	decodedQrContent, err := base64.URLEncoding.DecodeString(content)
	if err != nil {
		return "", status.Error(codes.InvalidArgument, err.Error())
	}

	var qrCodeContent QrCodeContent
	_ = json.Unmarshal(decodedQrContent, &qrCodeContent)

	switch qrCodeContent.Version {
	case constant.V1:
		studentID = string(decodedQrContent)
	case constant.V2:
		studentID, err = s.decryptQRV2(ctx, qrCodeContent.QrCode)
		if err != nil {
			return "", err
		}

	}

	return studentID, nil
}

func (s *EntryExitModifierService) decryptQRV2(ctx context.Context, qrCode string) (string, error) {
	encryptionKeys := []string{s.encryptSecretKeyV2}

	// Get the resource_path and check if the resource_path is Synersia
	resourcePath := golibs.ResourcePathFromCtx(ctx)
	if resourcePath == constant.SynersiaResourcePath {
		// If qrcode_encryption_synersia is not empty, add it to list of encryption keys
		if strings.TrimSpace(s.encryptSecretKeySynersiaV2) != "" {
			encryptionKeys = append(encryptionKeys, s.encryptSecretKeySynersiaV2)
		}

		// If qrcode_encryption_tokyo is not empty, add it to list of encryption keys
		if strings.TrimSpace(s.encryptSecretKeyTokyoV2) != "" {
			encryptionKeys = append(encryptionKeys, s.encryptSecretKeyTokyoV2)
		}
	}

	var (
		studentID  string
		secretByte []byte
		err        error
		latestErr  error
	)
	for _, keys := range encryptionKeys {
		// Reset the latestErr every start of the loop. We will return the error of the last key
		// If in the first key there is an error, and on the second key the decryption is successful, the latestErr is nil and we can return the studentID
		if latestErr != nil {
			s.logger.Warnf("the previous key encountered an error. err: %v", latestErr)
		}
		latestErr = nil

		// If error occurs, just go to the next key and assign the error to latestErr
		secretByte, err = hex.DecodeString(keys)
		if err != nil {
			latestErr = status.Error(codes.Internal, err.Error())
			continue
		}

		studentID, err = s.CryptV2.Decrypt(qrCode, secretByte)
		// break if there are no errors
		if err == nil {
			break
		}

		// If error occurs, just go to the next key and assign the error to latestErr
		latestErr = status.Error(codes.Internal, err.Error())
		if strings.Contains(err.Error(), "cipher: message authentication failed") {
			latestErr = status.Error(codes.InvalidArgument, "There is an issue with the QR code. The QR code may be from another organization.")
		}
	}

	// check the latestErr if there are errors during decryption
	if latestErr != nil {
		return "", latestErr
	}

	return studentID, nil
}

// get user profile of student
func (s *EntryExitModifierService) getUserProfileOfStudent(ctx context.Context, student *entities.Student) (*entities.Student, error) {
	existingUser, err := s.UserRepo.FindByID(ctx, s.DB, student.ID.String)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	student.User = *existingUser
	return student, nil
}

func (s *EntryExitModifierService) getStudentWithEEAccess(ctx context.Context, studentID string, withUserProfile bool) (*entities.Student, error) {
	student, err := s.StudentRepo.FindByID(ctx, s.DB, studentID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "student id does not exist")
	}

	// reassign student with user profile
	if withUserProfile {
		student, err = s.getUserProfileOfStudent(ctx, student)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return student, nil
}

func (s *EntryExitModifierService) notifyParents(
	ctx context.Context,
	student *entities.Student,
	notifyDetails *EntryExitNotifyDetails,
) (isNotified bool, err error) {
	getParentIDS, err := s.getParentIDs(ctx, s.DB, student.ID.String)
	if err != nil {
		return false, err
	}

	isNotified = true
	err = s.Notify(ctx, generateEntryExitNotifyTitleMessage(notifyDetails))
	if err != nil || len(getParentIDS) == 0 {
		isNotified = false
	}

	return isNotified, nil
}

func SignCtx(ctx context.Context) context.Context {
	headers, ok := metadata.FromIncomingContext(ctx)
	var pkg, token, version string
	if ok {
		pkg = headers["pkg"][0]
		token = headers["token"][0]
		version = headers["version"][0]
	}
	return metadata.AppendToOutgoingContext(ctx, "pkg", pkg, "version", version, "token", token)
}

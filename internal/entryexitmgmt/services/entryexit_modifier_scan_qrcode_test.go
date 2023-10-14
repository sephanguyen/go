package services

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/entryexitmgmt/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mock_repositories "github.com/manabie-com/backend/mock/entryexitmgmt/repositories"
	mock_services "github.com/manabie-com/backend/mock/entryexitmgmt/services"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	eepb "github.com/manabie-com/backend/pkg/manabuf/entryexitmgmt/v1"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestEntryExitModifierService_Scan(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mockDb := new(mock_database.Ext)
	mockStudentQRRepo := new(mock_repositories.MockStudentQRRepo)
	mockStudentEntryExitRecordsRepo := new(mock_repositories.MockStudentEntryExitRecordsRepo)
	mockEntryExitQueueRepo := new(mock_repositories.MockEntryExitQueueRepo)
	mockStudentRepo := new(mock_repositories.MockStudentRepo)
	mockStudentParentRepo := new(mock_repositories.MockStudentParentRepo)
	mockUserRepo := new(mock_repositories.MockUserRepo)
	mockJsm := new(mock_nats.JetStreamManagement)
	mockCrypt := new(mock_services.MockCrypt)

	mockValidStudent := &entities.Student{
		ID: database.Text("test"),
	}

	mockEntryExit := &entities.StudentEntryExitRecords{}
	mockEntryExit.EntryAt.Set(time.Now().Add(-1 * time.Minute))
	mockIds := []string{"1", "2"}
	user := &entities.User{
		FullName: database.Text("Albert Einstein JR"),
	}

	s := &EntryExitModifierService{
		DB:                          mockDb,
		StudentQRRepo:               mockStudentQRRepo,
		StudentEntryExitRecordsRepo: mockStudentEntryExitRecordsRepo,
		EntryExitQueueRepo:          mockEntryExitQueueRepo,
		StudentRepo:                 mockStudentRepo,
		StudentParentRepo:           mockStudentParentRepo,
		JSM:                         mockJsm,
		UserRepo:                    mockUserRepo,
		encryptSecretKeyV2:          "72d48c2c91e62ce3d0bf5b4bed09afb5",
	}

	const (
		studentID = "student-id-sample"
	)

	validQRCodeContentV1 := base64.URLEncoding.EncodeToString([]byte(studentID))

	// Set the real implementation first to generate v2 of QR Content
	s.CryptV2 = &CryptV2{}
	validQrCodeContentV2, err := s.generateQrContentV2(studentID)
	if err != nil {
		panic(err)
	}

	now := time.Now()

	tokyoLoc, _ := time.LoadLocation("Asia/Tokyo")

	tokyoTouchTime := time.Date(now.Year(), now.Month(), now.Day(), 14, 0, 0, 0, tokyoLoc)
	previousDateTokyoTime := tokyoTouchTime.AddDate(0, 0, -1)

	existingUTCentry := time.Date(now.Year(), now.Month(), now.Day(), 23, 0, 0, 0, time.UTC)
	existingUTCentry = existingUTCentry.AddDate(0, 0, -1)
	existingUTCexit := existingUTCentry.Add(5 * time.Hour)

	latestRecordInTokyo := &entities.StudentEntryExitRecords{
		StudentID: database.Text(studentID),
		EntryAt:   database.Timestamptz(existingUTCentry),
	}

	latestRecordInTokyoWithExit := &entities.StudentEntryExitRecords{
		StudentID: database.Text(studentID),
		EntryAt:   database.Timestamptz(existingUTCentry),
		ExitAt:    database.Timestamptz(existingUTCexit),
	}

	// Set to mockCrypt
	s.CryptV2 = mockCrypt

	testcases := []TestCase{
		{
			name:        "happy case for creating entry record with qr code version 1",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: nil,
			req: &eepb.ScanRequest{
				QrcodeContent: validQRCodeContentV1,
				TouchTime:     timestamppb.Now(),
			},
			expectedResp: &eepb.ScanResponse{
				Successful:     true,
				ParentNotified: true,
			},
			setup: func(ctx context.Context) {
				mockStudentQRRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(nil, nil)
				mockStudentRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(mockValidStudent, nil)
				mockUserRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(user, nil)
				mockStudentEntryExitRecordsRepo.On("LockAdvisoryByStudentID", ctx, mockDb, mock.Anything).Once().Return(true, nil)
				mockStudentEntryExitRecordsRepo.On("GetLatestRecordByID", ctx, mockDb, mock.Anything).Once().Return(nil, nil)
				mockEntryExitQueueRepo.On("Create", ctx, mockDb, mock.Anything).Once().Return(nil)
				mockStudentEntryExitRecordsRepo.On("Create", ctx, mockDb, mock.Anything).Once().Return(nil)
				mockStudentParentRepo.On("GetParentIDsByStudentID", ctx, mockDb, mock.Anything).Once().Return(mockIds, nil)
				mockJsm.On("PublishContext", mock.Anything, "Notification.Created", mock.Anything, mock.Anything).Once().Return(nil, nil)
				mockStudentEntryExitRecordsRepo.On("UnLockAdvisoryByStudentID", ctx, mockDb, mock.Anything).Return(nil)
			},
		},
		{
			name:        "happy case for creating entry record with qr code version2",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: nil,
			req: &eepb.ScanRequest{
				QrcodeContent: validQrCodeContentV2,
				TouchTime:     timestamppb.Now(),
			},
			expectedResp: &eepb.ScanResponse{
				Successful:     true,
				ParentNotified: true,
			},
			setup: func(ctx context.Context) {
				mockCrypt.On("Decrypt", mock.Anything, mock.Anything).Once().Return(studentID, nil)
				mockStudentQRRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(nil, nil)
				mockStudentRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(mockValidStudent, nil)
				mockUserRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(user, nil)
				mockStudentEntryExitRecordsRepo.On("LockAdvisoryByStudentID", ctx, mockDb, mock.Anything).Once().Return(true, nil)
				mockStudentEntryExitRecordsRepo.On("GetLatestRecordByID", ctx, mockDb, mock.Anything).Once().Return(nil, nil)
				mockEntryExitQueueRepo.On("Create", ctx, mockDb, mock.Anything).Once().Return(nil)
				mockStudentEntryExitRecordsRepo.On("Create", ctx, mockDb, mock.Anything).Once().Return(nil)
				mockStudentParentRepo.On("GetParentIDsByStudentID", ctx, mockDb, mock.Anything).Once().Return(mockIds, nil)
				mockJsm.On("PublishContext", mock.Anything, "Notification.Created", mock.Anything, mock.Anything).Once().Return(nil, nil)
				mockStudentEntryExitRecordsRepo.On("UnLockAdvisoryByStudentID", ctx, mockDb, mock.Anything).Return(nil)
			},
		},
		{
			name:        "happy case for creating entry record with Asia/Tokyo timezone",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: nil,
			req: &eepb.ScanRequest{
				QrcodeContent: validQrCodeContentV2,
				TouchTime:     timestamppb.Now(),
				Timezone:      "Asia/Tokyo",
			},
			expectedResp: &eepb.ScanResponse{
				Successful:     true,
				ParentNotified: true,
			},
			setup: func(ctx context.Context) {
				mockCrypt.On("Decrypt", mock.Anything, mock.Anything).Once().Return(studentID, nil)
				mockStudentQRRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(nil, nil)
				mockStudentRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(mockValidStudent, nil)
				mockUserRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(user, nil)
				mockStudentEntryExitRecordsRepo.On("LockAdvisoryByStudentID", ctx, mockDb, mock.Anything).Once().Return(true, nil)
				mockStudentEntryExitRecordsRepo.On("GetLatestRecordByID", ctx, mockDb, mock.Anything).Once().Return(nil, nil)
				mockEntryExitQueueRepo.On("Create", ctx, mockDb, mock.Anything).Once().Return(nil)
				mockStudentEntryExitRecordsRepo.On("Create", ctx, mockDb, mock.Anything).Once().Return(nil)
				mockStudentParentRepo.On("GetParentIDsByStudentID", ctx, mockDb, mock.Anything).Once().Return(mockIds, nil)
				mockJsm.On("PublishContext", mock.Anything, "Notification.Created", mock.Anything, mock.Anything).Once().Return(nil, nil)
				mockStudentEntryExitRecordsRepo.On("UnLockAdvisoryByStudentID", ctx, mockDb, mock.Anything).Return(nil)
			},
		},
		{
			name:        "happy case for touch exit record with Asia/Tokyo timezone",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: nil,
			req: &eepb.ScanRequest{
				QrcodeContent: validQrCodeContentV2,
				TouchTime:     timestamppb.New(tokyoTouchTime),
				Timezone:      "Asia/Tokyo",
			},
			expectedResp: &eepb.ScanResponse{
				Successful:     true,
				ParentNotified: true,
			},
			setup: func(ctx context.Context) {
				mockCrypt.On("Decrypt", mock.Anything, mock.Anything).Once().Return(studentID, nil)
				mockStudentQRRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(nil, nil)
				mockStudentRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(mockValidStudent, nil)
				mockUserRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(user, nil)
				mockStudentEntryExitRecordsRepo.On("LockAdvisoryByStudentID", ctx, mockDb, mock.Anything).Once().Return(true, nil)
				mockStudentEntryExitRecordsRepo.On("GetLatestRecordByID", ctx, mockDb, mock.Anything).Once().Return(latestRecordInTokyo, nil)
				mockEntryExitQueueRepo.On("Create", ctx, mockDb, mock.Anything).Once().Return(nil)
				mockStudentEntryExitRecordsRepo.On("Update", ctx, mockDb, mock.Anything).Once().Return(nil)
				mockStudentParentRepo.On("GetParentIDsByStudentID", ctx, mockDb, mock.Anything).Once().Return(mockIds, nil)
				mockJsm.On("PublishContext", mock.Anything, "Notification.Created", mock.Anything, mock.Anything).Once().Return(nil, nil)
				mockStudentEntryExitRecordsRepo.On("UnLockAdvisoryByStudentID", ctx, mockDb, mock.Anything).Return(nil)
			},
		},
		{
			name:        "happy case for touch entry record with Asia/Tokyo if the latest record is previous date",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: nil,
			req: &eepb.ScanRequest{
				QrcodeContent: validQrCodeContentV2,
				TouchTime:     timestamppb.New(tokyoTouchTime),
				Timezone:      "Asia/Tokyo",
			},
			expectedResp: &eepb.ScanResponse{
				Successful:     true,
				ParentNotified: true,
			},
			setup: func(ctx context.Context) {
				mockCrypt.On("Decrypt", mock.Anything, mock.Anything).Once().Return(studentID, nil)
				mockStudentQRRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(nil, nil)
				mockStudentRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(mockValidStudent, nil)
				mockUserRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(user, nil)
				mockStudentEntryExitRecordsRepo.On("LockAdvisoryByStudentID", ctx, mockDb, mock.Anything).Once().Return(true, nil)
				mockStudentEntryExitRecordsRepo.On("GetLatestRecordByID", ctx, mockDb, mock.Anything).Once().Return(&entities.StudentEntryExitRecords{
					StudentID: database.Text(studentID),
					EntryAt:   database.Timestamptz(previousDateTokyoTime),
				}, nil)
				mockEntryExitQueueRepo.On("Create", ctx, mockDb, mock.Anything).Once().Return(nil)
				mockStudentEntryExitRecordsRepo.On("Create", ctx, mockDb, mock.Anything).Once().Return(nil)
				mockStudentParentRepo.On("GetParentIDsByStudentID", ctx, mockDb, mock.Anything).Once().Return(mockIds, nil)
				mockJsm.On("PublishContext", mock.Anything, "Notification.Created", mock.Anything, mock.Anything).Once().Return(nil, nil)
				mockStudentEntryExitRecordsRepo.On("UnLockAdvisoryByStudentID", ctx, mockDb, mock.Anything).Return(nil)
			},
		},
		{
			name:        "happy case for touch entry record with Asia/Tokyo with existing entry and exit latest recoerd",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: nil,
			req: &eepb.ScanRequest{
				QrcodeContent: validQrCodeContentV2,
				TouchTime:     timestamppb.New(tokyoTouchTime),
				Timezone:      "Asia/Tokyo",
			},
			expectedResp: &eepb.ScanResponse{
				Successful:     true,
				ParentNotified: true,
			},
			setup: func(ctx context.Context) {
				mockCrypt.On("Decrypt", mock.Anything, mock.Anything).Once().Return(studentID, nil)
				mockStudentQRRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(nil, nil)
				mockStudentRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(mockValidStudent, nil)
				mockUserRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(user, nil)
				mockStudentEntryExitRecordsRepo.On("LockAdvisoryByStudentID", ctx, mockDb, mock.Anything).Once().Return(true, nil)
				mockStudentEntryExitRecordsRepo.On("GetLatestRecordByID", ctx, mockDb, mock.Anything).Once().Return(latestRecordInTokyoWithExit, nil)
				mockEntryExitQueueRepo.On("Create", ctx, mockDb, mock.Anything).Once().Return(nil)
				mockStudentEntryExitRecordsRepo.On("Create", ctx, mockDb, mock.Anything).Once().Return(nil)
				mockStudentParentRepo.On("GetParentIDsByStudentID", ctx, mockDb, mock.Anything).Once().Return(mockIds, nil)
				mockJsm.On("PublishContext", mock.Anything, "Notification.Created", mock.Anything, mock.Anything).Once().Return(nil, nil)
				mockStudentEntryExitRecordsRepo.On("UnLockAdvisoryByStudentID", ctx, mockDb, mock.Anything).Return(nil)
			},
		},
		{
			name:        "happy case for creating entry record",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: nil,
			req: &eepb.ScanRequest{
				QrcodeContent: validQrCodeContentV2,
				TouchTime:     timestamppb.Now(),
			},
			expectedResp: &eepb.ScanResponse{
				Successful:     true,
				ParentNotified: true,
			},
			setup: func(ctx context.Context) {
				mockCrypt.On("Decrypt", mock.Anything, mock.Anything).Once().Return(studentID, nil)
				mockStudentQRRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(nil, nil)
				mockStudentRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(mockValidStudent, nil)
				mockUserRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(user, nil)
				mockStudentEntryExitRecordsRepo.On("LockAdvisoryByStudentID", ctx, mockDb, mock.Anything).Once().Return(true, nil)
				mockStudentEntryExitRecordsRepo.On("GetLatestRecordByID", ctx, mockDb, mock.Anything).Once().Return(nil, nil)
				mockEntryExitQueueRepo.On("Create", ctx, mockDb, mock.Anything).Once().Return(nil)
				mockStudentEntryExitRecordsRepo.On("Create", ctx, mockDb, mock.Anything).Once().Return(nil)
				mockStudentParentRepo.On("GetParentIDsByStudentID", ctx, mockDb, mock.Anything).Once().Return(mockIds, nil)
				mockJsm.On("PublishContext", mock.Anything, "Notification.Created", mock.Anything, mock.Anything).Once().Return(nil, nil)
				mockStudentEntryExitRecordsRepo.On("UnLockAdvisoryByStudentID", ctx, mockDb, mock.Anything).Return(nil)
			},
		},
		{
			name:        "happy case for updating exit record",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: nil,
			req: &eepb.ScanRequest{
				QrcodeContent: validQrCodeContentV2,
				TouchTime:     timestamppb.Now(),
			},
			setup: func(ctx context.Context) {
				mockCrypt.On("Decrypt", mock.Anything, mock.Anything).Once().Return(studentID, nil)
				mockStudentQRRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(nil, nil)
				mockStudentRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(mockValidStudent, nil)
				mockUserRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(user, nil)
				mockStudentEntryExitRecordsRepo.On("LockAdvisoryByStudentID", ctx, mockDb, mock.Anything).Once().Return(true, nil)
				mockStudentEntryExitRecordsRepo.On("GetLatestRecordByID", ctx, mockDb, mock.Anything).Once().Return(mockEntryExit, nil)
				mockEntryExitQueueRepo.On("Create", ctx, mockDb, mock.Anything).Once().Return(nil)
				mockStudentEntryExitRecordsRepo.On("Update", ctx, mockDb, mock.Anything).Once().Return(nil)
				mockStudentParentRepo.On("GetParentIDsByStudentID", ctx, mockDb, mock.Anything).Once().Return(mockIds, nil)
				mockJsm.On("PublishContext", mock.Anything, "Notification.Created", mock.Anything, mock.Anything).Once().Return(nil, nil)
				mockStudentEntryExitRecordsRepo.On("UnLockAdvisoryByStudentID", ctx, mockDb, mock.Anything).Return(nil)
			},
		},
		{
			name:        "decryption error on invalid encryption key",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "There is an issue with the QR code. The QR code may be from another organization."),
			req: &eepb.ScanRequest{
				QrcodeContent: validQrCodeContentV2,
				TouchTime:     timestamppb.Now(),
			},
			setup: func(ctx context.Context) {
				mockCrypt.On("Decrypt", mock.Anything, mock.Anything).Once().Return("", fmt.Errorf("cipher: message authentication failed"))
			},
		},
		{
			name:        "decryption error",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.Internal, "error Decrypt"),
			req: &eepb.ScanRequest{
				QrcodeContent: validQrCodeContentV2,
				TouchTime:     timestamppb.Now(),
			},
			setup: func(ctx context.Context) {
				mockCrypt.On("Decrypt", mock.Anything, mock.Anything).Once().Return("", fmt.Errorf("error Decrypt"))
			},
		},
		{
			name:        "no qrcode content in scan",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "qrcode content cannot be empty"),
			req:         &eepb.ScanRequest{},
			setup: func(ctx context.Context) {
			},
		},
		{
			name:        "no touch time content in scan",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "touch time cannot be empty"),
			req: &eepb.ScanRequest{
				QrcodeContent: validQrCodeContentV2,
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name:        "touch time is past date",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "touch time should not be past date"),
			req: &eepb.ScanRequest{
				QrcodeContent: validQrCodeContentV2,
				TouchTime:     timestamppb.New(time.Now().AddDate(0, -1, 10)),
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name:        "student qrcode content not found",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.Internal, "no rows in result set"),
			req: &eepb.ScanRequest{
				QrcodeContent: validQrCodeContentV2,
				TouchTime:     timestamppb.Now(),
			},
			setup: func(ctx context.Context) {
				mockCrypt.On("Decrypt", mock.Anything, mock.Anything).Once().Return(studentID, nil)
				mockStudentQRRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name:        "create entryexit record failed",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.Internal, "tx is closed"),
			req: &eepb.ScanRequest{
				QrcodeContent: validQrCodeContentV2,
				TouchTime:     timestamppb.Now(),
			},
			setup: func(ctx context.Context) {
				mockCrypt.On("Decrypt", mock.Anything, mock.Anything).Once().Return(studentID, nil)
				mockStudentQRRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(nil, nil)
				mockStudentRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(mockValidStudent, nil)
				mockUserRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(user, nil)
				mockStudentEntryExitRecordsRepo.On("LockAdvisoryByStudentID", ctx, mockDb, mock.Anything).Once().Return(true, nil)
				mockStudentEntryExitRecordsRepo.On("GetLatestRecordByID", ctx, mockDb, mock.Anything).Once().Return(nil, nil)
				mockEntryExitQueueRepo.On("Create", ctx, mockDb, mock.Anything).Once().Return(nil)
				mockStudentEntryExitRecordsRepo.On("Create", ctx, mockDb, mock.Anything).Once().Return(pgx.ErrTxClosed)
				mockStudentEntryExitRecordsRepo.On("UnLockAdvisoryByStudentID", ctx, mockDb, mock.Anything).Return(nil)
			},
		},
		{
			name:        "failed scan due 1 minute limit",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.PermissionDenied, "please wait after 1 min to scan again"),
			req: &eepb.ScanRequest{
				QrcodeContent: validQrCodeContentV2,
				TouchTime:     timestamppb.New(time.Now().Add(-30 * time.Second)),
			},
			setup: func(ctx context.Context) {
				mockCrypt.On("Decrypt", mock.Anything, mock.Anything).Once().Return(studentID, nil)
				mockStudentQRRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(nil, nil)
				mockStudentRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(mockValidStudent, nil)
				mockUserRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(user, nil)
				mockStudentEntryExitRecordsRepo.On("LockAdvisoryByStudentID", ctx, mockDb, mock.Anything).Once().Return(true, nil)
				mockStudentEntryExitRecordsRepo.On("GetLatestRecordByID", ctx, mockDb, mock.Anything).Once().Return(mockEntryExit, nil)
				mockStudentEntryExitRecordsRepo.On("UnLockAdvisoryByStudentID", ctx, mockDb, mock.Anything).Return(nil)
			},
		},
		{
			name:        "student id returns no rows result set",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "student id does not exist"),
			req: &eepb.ScanRequest{
				QrcodeContent: validQrCodeContentV2,
				TouchTime:     timestamppb.Now(),
			},
			setup: func(ctx context.Context) {
				mockCrypt.On("Decrypt", mock.Anything, mock.Anything).Once().Return(studentID, nil)
				mockStudentQRRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(nil, nil)
				mockStudentRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(mockValidStudent, pgx.ErrNoRows)
			},
		},
		{
			name:        "get latest record no rows result",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.Internal, "no rows in result set"),
			req: &eepb.ScanRequest{
				QrcodeContent: validQrCodeContentV2,
				TouchTime:     timestamppb.Now(),
			},
			setup: func(ctx context.Context) {
				mockCrypt.On("Decrypt", mock.Anything, mock.Anything).Once().Return(studentID, nil)
				mockStudentQRRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(nil, nil)
				mockStudentRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(mockValidStudent, nil)
				mockUserRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(user, nil)
				mockStudentEntryExitRecordsRepo.On("LockAdvisoryByStudentID", ctx, mockDb, mock.Anything).Once().Return(true, nil)
				mockStudentEntryExitRecordsRepo.On("UnLockAdvisoryByStudentID", ctx, mockDb, mock.Anything).Return(nil)
				mockStudentEntryExitRecordsRepo.On("GetLatestRecordByID", ctx, mockDb, mock.Anything).Once().Return(mockEntryExit, pgx.ErrNoRows)
			},
		},
		{
			name:        "get latest record failed",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.Internal, "tx is closed"),
			req: &eepb.ScanRequest{
				QrcodeContent: validQrCodeContentV2,
				TouchTime:     timestamppb.Now(),
			},
			setup: func(ctx context.Context) {
				mockCrypt.On("Decrypt", mock.Anything, mock.Anything).Once().Return(studentID, nil)
				mockStudentQRRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(nil, nil)
				mockStudentRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(mockValidStudent, nil)
				mockUserRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(user, nil)
				mockStudentEntryExitRecordsRepo.On("LockAdvisoryByStudentID", ctx, mockDb, mock.Anything).Once().Return(true, nil)
				mockStudentEntryExitRecordsRepo.On("UnLockAdvisoryByStudentID", ctx, mockDb, mock.Anything).Return(nil)
				mockStudentEntryExitRecordsRepo.On("GetLatestRecordByID", ctx, mockDb, mock.Anything).Once().Return(mockEntryExit, pgx.ErrTxClosed)
			},
		},
		{
			name:        "create entry record failed",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.Internal, "tx is closed"),
			req: &eepb.ScanRequest{
				QrcodeContent: validQrCodeContentV2,
				TouchTime:     timestamppb.Now(),
			},
			setup: func(ctx context.Context) {
				mockCrypt.On("Decrypt", mock.Anything, mock.Anything).Once().Return(studentID, nil)
				mockStudentQRRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(nil, nil)
				mockStudentRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(mockValidStudent, nil)
				mockUserRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(user, nil)
				mockStudentEntryExitRecordsRepo.On("LockAdvisoryByStudentID", ctx, mockDb, mock.Anything).Once().Return(true, nil)
				mockStudentEntryExitRecordsRepo.On("UnLockAdvisoryByStudentID", ctx, mockDb, mock.Anything).Return(nil)
				mockStudentEntryExitRecordsRepo.On("GetLatestRecordByID", ctx, mockDb, mock.Anything).Once().Return(nil, nil)
				mockEntryExitQueueRepo.On("Create", ctx, mockDb, mock.Anything).Once().Return(nil)
				mockStudentEntryExitRecordsRepo.On("Create", ctx, mockDb, mock.Anything).Once().Return(pgx.ErrTxClosed)
			},
		},
		{
			name:        "update entryexit record failed",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.Internal, "tx is closed"),
			req: &eepb.ScanRequest{
				QrcodeContent: validQrCodeContentV2,
				TouchTime:     timestamppb.Now(),
			},
			setup: func(ctx context.Context) {
				mockCrypt.On("Decrypt", mock.Anything, mock.Anything).Once().Return(studentID, nil)
				mockStudentQRRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(nil, nil)
				mockStudentRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(mockValidStudent, nil)
				mockUserRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(user, nil)
				mockStudentEntryExitRecordsRepo.On("LockAdvisoryByStudentID", ctx, mockDb, mock.Anything).Once().Return(true, nil)
				mockStudentEntryExitRecordsRepo.On("UnLockAdvisoryByStudentID", ctx, mockDb, mock.Anything).Return(nil)
				mockStudentEntryExitRecordsRepo.On("GetLatestRecordByID", ctx, mockDb, mock.Anything).Once().Return(mockEntryExit, nil)
				mockEntryExitQueueRepo.On("Create", ctx, mockDb, mock.Anything).Once().Return(nil)
				mockStudentEntryExitRecordsRepo.On("Update", ctx, mockDb, mock.Anything).Once().Return(pgx.ErrTxClosed)
			},
		},
		{
			name:        "failed parent notification send",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: nil,
			req: &eepb.ScanRequest{
				QrcodeContent: validQrCodeContentV2,
				TouchTime:     timestamppb.Now(),
			},
			expectedResp: &eepb.ScanResponse{
				Successful:     true,
				ParentNotified: false,
			},
			setup: func(ctx context.Context) {
				mockCrypt.On("Decrypt", mock.Anything, mock.Anything).Once().Return(studentID, nil)
				mockStudentQRRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(nil, nil)
				mockStudentRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(mockValidStudent, nil)
				mockUserRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(user, nil)
				mockStudentEntryExitRecordsRepo.On("LockAdvisoryByStudentID", ctx, mockDb, mock.Anything).Once().Return(true, nil)
				mockStudentEntryExitRecordsRepo.On("UnLockAdvisoryByStudentID", ctx, mockDb, mock.Anything).Return(nil)
				mockStudentEntryExitRecordsRepo.On("GetLatestRecordByID", ctx, mockDb, mock.Anything).Once().Return(nil, nil)
				mockEntryExitQueueRepo.On("Create", ctx, mockDb, mock.Anything).Once().Return(nil)
				mockStudentEntryExitRecordsRepo.On("Create", ctx, mockDb, mock.Anything).Once().Return(nil)
				mockStudentParentRepo.On("GetParentIDsByStudentID", ctx, mockDb, mock.Anything).Once().Return(mockIds, nil)
				for i := 0; i < 4; i++ {
					mockJsm.On("PublishContext", mock.Anything, "Notification.Created", mock.Anything, mock.Anything).Once().Return(nil, errors.New("publish error"))
				}
			},
		},
		{
			name:        "failed to get parent ids closed db pool",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.Internal, "closed pool"),
			req: &eepb.ScanRequest{
				QrcodeContent: validQrCodeContentV2,
				TouchTime:     timestamppb.Now(),
			},
			setup: func(ctx context.Context) {
				mockCrypt.On("Decrypt", mock.Anything, mock.Anything).Once().Return(studentID, nil)
				mockStudentQRRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(nil, nil)
				mockStudentRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(mockValidStudent, nil)
				mockUserRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(user, nil)
				mockStudentEntryExitRecordsRepo.On("LockAdvisoryByStudentID", ctx, mockDb, mock.Anything).Once().Return(true, nil)
				mockStudentEntryExitRecordsRepo.On("UnLockAdvisoryByStudentID", ctx, mockDb, mock.Anything).Return(nil)
				mockStudentEntryExitRecordsRepo.On("GetLatestRecordByID", ctx, mockDb, mock.Anything).Once().Return(nil, nil)
				mockEntryExitQueueRepo.On("Create", ctx, mockDb, mock.Anything).Once().Return(nil)
				mockStudentEntryExitRecordsRepo.On("Create", ctx, mockDb, mock.Anything).Once().Return(nil)
				mockStudentParentRepo.On("GetParentIDsByStudentID", ctx, mockDb, mock.Anything).Once().Return(mockIds, (puddle.ErrClosedPool))
			},
		},
		{
			name:        "failed to get parent ids err no rows",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.Internal, "no rows in result set"),
			req: &eepb.ScanRequest{
				QrcodeContent: validQrCodeContentV2,
				TouchTime:     timestamppb.Now(),
			},
			setup: func(ctx context.Context) {
				mockCrypt.On("Decrypt", mock.Anything, mock.Anything).Once().Return(studentID, nil)
				mockStudentQRRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(nil, nil)
				mockStudentRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(mockValidStudent, nil)
				mockUserRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(user, nil)
				mockStudentEntryExitRecordsRepo.On("LockAdvisoryByStudentID", ctx, mockDb, mock.Anything).Once().Return(true, nil)
				mockStudentEntryExitRecordsRepo.On("UnLockAdvisoryByStudentID", ctx, mockDb, mock.Anything).Return(nil)
				mockStudentEntryExitRecordsRepo.On("GetLatestRecordByID", ctx, mockDb, mock.Anything).Once().Return(nil, nil)
				mockEntryExitQueueRepo.On("Create", ctx, mockDb, mock.Anything).Once().Return(nil)
				mockStudentEntryExitRecordsRepo.On("Create", ctx, mockDb, mock.Anything).Once().Return(nil)
				mockStudentParentRepo.On("GetParentIDsByStudentID", ctx, mockDb, mock.Anything).Once().Return(make([]string, 0), pgx.ErrNoRows)
			},
		},
		{
			name:        "user id returns no rows result set",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.Internal, "no rows in result set"),
			req: &eepb.ScanRequest{
				QrcodeContent: validQrCodeContentV2,
				TouchTime:     timestamppb.Now(),
			},
			setup: func(ctx context.Context) {
				mockCrypt.On("Decrypt", mock.Anything, mock.Anything).Once().Return(studentID, nil)
				mockStudentQRRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(nil, nil)
				mockStudentRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(mockValidStudent, nil)
				mockUserRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name:        "timezone is invalid",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, fmt.Sprintf("Error on initializing timezone location: err %v", "unknown time zone INVALID")),
			req: &eepb.ScanRequest{
				QrcodeContent: validQrCodeContentV2,
				TouchTime:     timestamppb.Now(),
				Timezone:      "INVALID",
			},
			setup: func(ctx context.Context) {
				mockCrypt.On("Decrypt", mock.Anything, mock.Anything).Once().Return(studentID, nil)
				mockStudentQRRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(nil, nil)
				mockStudentRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(mockValidStudent, nil)
				mockUserRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(user, nil)
				mockStudentEntryExitRecordsRepo.On("GetLatestRecordByID", ctx, mockDb, mock.Anything).Once().Return(nil, nil)
				mockStudentEntryExitRecordsRepo.On("LockAdvisoryByStudentID", ctx, mockDb, mock.Anything).Once().Return(true, nil)
				mockStudentEntryExitRecordsRepo.On("UnLockAdvisoryByStudentID", ctx, mockDb, mock.Anything).Return(nil)
			},
		},
		{
			name:        "acquire lock failed",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.Aborted, fmt.Sprintf("%s already processing. skipped", studentID)),
			req: &eepb.ScanRequest{
				QrcodeContent: validQrCodeContentV2,
				TouchTime:     timestamppb.Now(),
			},
			setup: func(ctx context.Context) {
				mockCrypt.On("Decrypt", mock.Anything, mock.Anything).Once().Return(studentID, nil)
				mockStudentQRRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(nil, nil)
				mockStudentRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(mockValidStudent, nil)
				mockUserRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(user, nil)
				mockStudentEntryExitRecordsRepo.On("LockAdvisoryByStudentID", ctx, mockDb, mock.Anything).Once().Return(false, nil)
			},
		},
		{
			name:        "entryexit_queue tx closed",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.Internal, "tx is closed"),
			req: &eepb.ScanRequest{
				QrcodeContent: validQrCodeContentV2,
				TouchTime:     timestamppb.Now(),
			},
			setup: func(ctx context.Context) {
				mockCrypt.On("Decrypt", mock.Anything, mock.Anything).Once().Return(studentID, nil)
				mockStudentQRRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(nil, nil)
				mockStudentRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(mockValidStudent, nil)
				mockUserRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(user, nil)
				mockStudentEntryExitRecordsRepo.On("LockAdvisoryByStudentID", ctx, mockDb, mock.Anything).Once().Return(true, nil)
				mockStudentEntryExitRecordsRepo.On("UnLockAdvisoryByStudentID", ctx, mockDb, mock.Anything).Return(nil)
				mockStudentEntryExitRecordsRepo.On("GetLatestRecordByID", ctx, mockDb, mock.Anything).Once().Return(nil, nil)
				mockEntryExitQueueRepo.On("Create", ctx, mockDb, mock.Anything).Once().Return(pgx.ErrTxClosed)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			resp, err := s.Scan(testCase.ctx, testCase.req.(*eepb.ScanRequest))
			if err != nil {
				fmt.Println(err)
			}

			if testCase.expectedErr != nil {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
				assert.Nil(t, resp)
			} else {
				assert.Equal(t, testCase.expectedErr, err)
				assert.NotNil(t, resp)
			}

			if testCase.expectedResp != nil {
				assert.Equal(t, testCase.expectedResp.(*eepb.ScanResponse).Successful, resp.Successful)
				assert.Equal(t, testCase.expectedResp.(*eepb.ScanResponse).ParentNotified, resp.ParentNotified)
			}

			mock.AssertExpectationsForObjects(t, mockDb, mockStudentEntryExitRecordsRepo, mockEntryExitQueueRepo, mockStudentRepo, mockStudentParentRepo, mockUserRepo)
		})
	}
}

func Test_getLocation(t *testing.T) {

	expectedlocationJP, _ := time.LoadLocation("Asia/Tokyo")
	expectedlocationVN, _ := time.LoadLocation("Asia/Ho_Chi_Minh")
	expectedlocationUTC, _ := time.LoadLocation("UTC")

	type args struct {
		country cpb.Country
	}
	tests := []struct {
		name    string
		args    args
		want    time.Location
		wantErr bool
	}{
		{
			name: "JP Country uses asia/tokyo",
			args: args{
				country: cpb.Country_COUNTRY_JP,
			},
			want:    *expectedlocationJP,
			wantErr: false,
		},
		{
			name: "VN Country uses Ho chi minh",
			args: args{
				country: cpb.Country_COUNTRY_VN,
			},
			want:    *expectedlocationVN,
			wantErr: false,
		},
		{
			name: "SG uses UTC instead",
			args: args{
				country: cpb.Country_COUNTRY_SG,
			},
			want:    *expectedlocationUTC,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getLocation(tt.args.country)
			if (err != nil) != tt.wantErr {
				t.Errorf("getLocation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

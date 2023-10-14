package services

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/entryexitmgmt/constant"
	"github.com/manabie-com/backend/internal/entryexitmgmt/entities"
	"github.com/manabie-com/backend/internal/entryexitmgmt/services/uploader"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/logger"
	mock_repositories "github.com/manabie-com/backend/mock/entryexitmgmt/repositories"
	mock_services "github.com/manabie-com/backend/mock/entryexitmgmt/services"
	mock_uploader "github.com/manabie-com/backend/mock/entryexitmgmt/services/uploader"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	eepb "github.com/manabie-com/backend/pkg/manabuf/entryexitmgmt/v1"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestEntryExitModifierService_RetrieveStudentQRCodeTest(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mockDB := new(mock_database.Ext)
	mockTx := new(mock_database.Tx)
	mockStudentQRRepo := new(mock_repositories.MockStudentQRRepo)
	mockStudentRepo := new(mock_repositories.MockStudentRepo)
	mockJsm := new(mock_nats.JetStreamManagement)
	mockCrypt := new(mock_services.MockCrypt)
	mockSDKUploaderService := new(mock_uploader.Uploader)
	mockCurlUploaderService := new(mock_uploader.Uploader)
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)

	s := &EntryExitModifierService{
		DB:            mockDB,
		StudentQRRepo: mockStudentQRRepo,
		StudentRepo:   mockStudentRepo,
		JSM:           mockJsm,
		CryptV2:       mockCrypt,
		UploadServiceSelector: &UploadServiceSelector{
			SdkUploadService:  mockSDKUploaderService,
			CurlUploadService: mockCurlUploaderService,
			UnleashClient:     mockUnleashClient,
			Env:               "local",
		},
	}

	user := &entities.User{
		FullName: database.Text("Test User"),
		Country:  database.Text("COUNTRY_VN"),
	}
	student1 := &entities.Student{
		ID:       database.Text("student-id-1"),
		SchoolID: database.Int4(1),
		User:     *user,
	}

	studentQRCodeV1 := &entities.StudentQR{
		ID:        database.Int4(1),
		StudentID: student1.ID,
		QRURL:     database.Text("http://sample.com/qrcodeV1"),
		Version:   database.Text(""),
	}

	studentQRCodeV2 := &entities.StudentQR{
		ID:        database.Int4(1),
		StudentID: student1.ID,
		QRURL:     database.Text("http://sample.com/qrcode"),
		Version:   database.Text("v2"),
	}

	encryptStudentV2 := "eyJRckNvZGUiOiJBcm5SVW5iVEllOEhpNnpEWEQtUUZ3TG9ZcXI3MUwwOUl2TjZCcGN6WTVDNzByak04ZFVWMlE9PSIsIlZlcnNpb24iOiJ2MiJ9"

	testcases := []TestCase{
		{
			name: "happy case student qr code",
			ctx:  ctx,
			req: &eepb.RetrieveStudentQRCodeRequest{
				StudentId: student1.ID.String,
			},
			expectedErr: nil,
			expectedResp: &eepb.RetrieveStudentQRCodeResponse{
				QrUrl: studentQRCodeV2.QRURL.String,
			},
			setup: func(ctx context.Context) {
				mockStudentRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(&entities.Student{}, nil)
				mockStudentQRRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(studentQRCodeV2, nil)
			},
		},
		{
			name: "Create new QR Code when there's no record",
			ctx:  ctx,
			req: &eepb.RetrieveStudentQRCodeRequest{
				StudentId: student1.ID.String,
			},
			expectedErr: nil,
			expectedResp: &eepb.RetrieveStudentQRCodeResponse{
				QrUrl: studentQRCodeV2.QRURL.String,
			},
			setup: func(ctx context.Context) {
				mockStudentRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(&entities.Student{}, nil)
				mockStudentQRRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
				mockCrypt.On("Encrypt", mock.Anything, mock.Anything).Once().Return(encryptStudentV2, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(true, nil)
				mockSDKUploaderService.On("InitUploader", ctx, mock.Anything).Return(&uploader.UploadInfo{
					DownloadURL: studentQRCodeV2.QRURL.String,
					DoUploadFromFile: func(ctx context.Context, filePathName string) error {
						return nil
					},
				}, nil)
				mockStudentQRRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "Create new QR Code when version is empty or version 1",
			ctx:  ctx,
			req: &eepb.RetrieveStudentQRCodeRequest{
				StudentId: student1.ID.String,
			},
			expectedErr: nil,
			expectedResp: &eepb.RetrieveStudentQRCodeResponse{
				QrUrl: studentQRCodeV2.QRURL.String,
			},
			setup: func(ctx context.Context) {
				mockStudentRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(&entities.Student{}, nil)
				mockStudentQRRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(studentQRCodeV1, pgx.ErrNoRows)
				mockCrypt.On("Encrypt", mock.Anything, mock.Anything).Once().Return(encryptStudentV2, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(true, nil)
				mockSDKUploaderService.On("InitUploader", ctx, mock.Anything).Return(&uploader.UploadInfo{
					DownloadURL: studentQRCodeV2.QRURL.String,
					DoUploadFromFile: func(ctx context.Context, filePathName string) error {
						return nil
					},
				}, nil)
				mockStudentQRRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "failed to get student err closed pool",
			ctx:  ctx,
			req: &eepb.RetrieveStudentQRCodeRequest{
				StudentId: student1.ID.String,
			},
			expectedErr: status.Error(codes.Internal, "closed pool"),
			setup: func(ctx context.Context) {
				mockStudentRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(&entities.Student{}, puddle.ErrClosedPool)
			},
		},
		{
			name: "failed to get student err no rows",
			ctx:  ctx,
			req: &eepb.RetrieveStudentQRCodeRequest{
				StudentId: student1.ID.String,
			},
			expectedErr: status.Error(codes.Internal, "no rows in result set"),
			setup: func(ctx context.Context) {
				mockStudentRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(&entities.Student{}, pgx.ErrNoRows)
			},
		},
		{
			name: "failed to get student qr err closed pool",
			ctx:  ctx,
			req: &eepb.RetrieveStudentQRCodeRequest{
				StudentId: student1.ID.String,
			},
			expectedErr: status.Error(codes.Internal, "err StudentQRRepo.FindByID closed pool"),
			setup: func(ctx context.Context) {
				mockStudentRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(&entities.Student{}, nil)
				mockStudentQRRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(studentQRCodeV2, puddle.ErrClosedPool)
			},
		},
		{
			name: "invalid request empty student id",
			ctx:  ctx,
			req: &eepb.RetrieveStudentQRCodeRequest{
				StudentId: "",
			},
			expectedErr: status.Error(codes.InvalidArgument, "student id cannot be empty"),
			setup: func(ctx context.Context) {
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			resp, err := s.RetrieveStudentQRCode(testCase.ctx, testCase.req.(*eepb.RetrieveStudentQRCodeRequest))
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
				assert.Equal(t, testCase.expectedResp.(*eepb.RetrieveStudentQRCodeResponse).QrUrl, resp.QrUrl)
			}
			mock.AssertExpectationsForObjects(t, mockDB, mockStudentQRRepo, mockStudentRepo)
		})
	}
}

func Test_getStudentIDFromQr(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db := new(mock_database.Ext)
	mockStudentQRRepo := new(mock_repositories.MockStudentQRRepo)
	mockCrypt := new(mock_services.MockCrypt)

	const (
		studentID = "sample-student-id"
	)

	s := &EntryExitModifierService{
		DB:                 db,
		StudentQRRepo:      mockStudentQRRepo,
		CryptV2:            mockCrypt,
		encryptSecretKeyV2: "72d48c2c91e62ce3d0bf5b4bed09afb5",
	}

	withInvalidCryptKey := &EntryExitModifierService{
		DB:                 db,
		encryptSecretKeyV2: "invalid-secret-key",
	}

	type args struct {
		ctx     context.Context
		content string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		srv     *EntryExitModifierService
		setup   func(ctx context.Context)
	}{
		{
			name: "successfully fetched studentID from content",
			args: args{
				ctx:     ctx,
				content: "eyJRckNvZGUiOiJBcm5SVW5iVEllOEhpNnpEWEQtUUZ3TG9ZcXI3MUwwOUl2TjZCcGN6WTVDNzByak04ZFVWMlE9PSIsIlZlcnNpb24iOiJ2MiJ9",
			},
			srv:     s,
			wantErr: false,
			setup: func(ctx context.Context) {
				mockCrypt.On("Decrypt", mock.Anything, mock.Anything).Once().Return("student-id-1", nil)
			},
		},
		{
			name: "non base-64 content",
			args: args{
				ctx:     ctx,
				content: "invalid-content",
			},
			srv:     s,
			wantErr: true,
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "with invalid secret key",
			args: args{
				ctx:     ctx,
				content: "eyJRckNvZGUiOiJBcm5SVW5iVEllOEhpNnpEWEQtUUZ3TG9ZcXI3MUwwOUl2TjZCcGN6WTVDNzByak04ZFVWMlE9PSIsIlZlcnNpb24iOiJ2MiJ9",
			},
			srv:     withInvalidCryptKey,
			wantErr: true,
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "with decrypt error",
			args: args{
				ctx:     ctx,
				content: "eyJRckNvZGUiOiJBcm5SVW5iVEllOEhpNnpEWEQtUUZ3TG9ZcXI3MUwwOUl2TjZCcGN6WTVDNzByak04ZFVWMlE9PSIsIlZlcnNpb24iOiJ2MiJ9",
			},
			srv:     s,
			wantErr: true,
			setup: func(ctx context.Context) {
				mockCrypt.On("Decrypt", mock.Anything, mock.Anything).Once().Return("", fmt.Errorf("error Decrypt"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(tt.args.ctx)

			url, err := tt.srv.getStudentIDFromQr(tt.args.ctx, tt.args.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("EntryExitModifierService.Generate() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr == false {
				assert.NotEmpty(t, url)
			}
		})
	}
}

func Test_getStudentIDFromQrSynersia(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	ctx = golibs.ResourcePathToCtx(ctx, constant.SynersiaResourcePath)

	db := new(mock_database.Ext)
	mockStudentQRRepo := new(mock_repositories.MockStudentQRRepo)

	const (
		studentID = "sample-student-id"
	)

	encryptionKey := "fce808d465f4152a4eacbd36147e5e53"
	encryptionKeyTokyo := "2ce4b35822456ab08c9a7ae0df53dea2"
	encryptionKeySynersia := "a06fe86bae9ef10e2f21f6d7dbc4cc39"

	zapLogger := logger.NewZapLogger("debug", true)

	s := &EntryExitModifierService{
		DB:                         db,
		StudentQRRepo:              mockStudentQRRepo,
		CryptV2:                    &CryptV2{},
		encryptSecretKeyV2:         encryptionKey,
		encryptSecretKeyTokyoV2:    encryptionKeyTokyo,
		encryptSecretKeySynersiaV2: encryptionKeySynersia,
		logger:                     *zapLogger.Sugar(),
	}

	serviceWithEmptyTokyoKey := &EntryExitModifierService{
		DB:                         db,
		StudentQRRepo:              mockStudentQRRepo,
		CryptV2:                    &CryptV2{},
		encryptSecretKeyV2:         encryptionKey,
		encryptSecretKeyTokyoV2:    "",
		encryptSecretKeySynersiaV2: encryptionKeySynersia,
		logger:                     *zapLogger.Sugar(),
	}

	serviceWithEmptySynersiaKey := &EntryExitModifierService{
		DB:                         db,
		StudentQRRepo:              mockStudentQRRepo,
		CryptV2:                    &CryptV2{},
		encryptSecretKeyV2:         encryptionKey,
		encryptSecretKeyTokyoV2:    encryptionKeyTokyo,
		encryptSecretKeySynersiaV2: "",
		logger:                     *zapLogger.Sugar(),
	}

	serviceWithEmptySynersiaTokyoKey := &EntryExitModifierService{
		DB:                         db,
		StudentQRRepo:              mockStudentQRRepo,
		CryptV2:                    &CryptV2{},
		encryptSecretKeyV2:         encryptionKey,
		encryptSecretKeyTokyoV2:    "",
		encryptSecretKeySynersiaV2: "",
		logger:                     *zapLogger.Sugar(),
	}

	c := &CryptV2{}

	secret, _ := hex.DecodeString(encryptionKey)
	student1QR, _ := c.Encrypt("student-id-1", secret)
	student1ContentByte, _ := json.Marshal(QrCodeContent{QrCode: student1QR, Version: constant.V2})
	student1Content := base64.URLEncoding.EncodeToString(student1ContentByte)

	tokyoSecret, _ := hex.DecodeString(encryptionKeyTokyo)
	student1QRTokyo, _ := c.Encrypt("student-id-1-tokyo", tokyoSecret)
	student1ContentByteTokyo, _ := json.Marshal(QrCodeContent{QrCode: student1QRTokyo, Version: constant.V2})
	student1ContentTokyo := base64.URLEncoding.EncodeToString(student1ContentByteTokyo)

	synersiaSecret, _ := hex.DecodeString(encryptionKeySynersia)
	student1QRSynersia, _ := c.Encrypt("student-id-1-synersia", synersiaSecret)
	student1ContentByteSynersia, _ := json.Marshal(QrCodeContent{QrCode: student1QRSynersia, Version: constant.V2})
	student1ContentSynersia := base64.URLEncoding.EncodeToString(student1ContentByteSynersia)

	invalidContentByte, _ := json.Marshal(QrCodeContent{QrCode: "vPgmZpGq9QRVDRkPFMnutZ57mXYBCUcnxSu5xrrx0Soy_cHoEZwc7Q==", Version: constant.V2})
	invalidContent := base64.URLEncoding.EncodeToString(invalidContentByte)

	type args struct {
		ctx     context.Context
		content string
	}
	tests := []struct {
		name              string
		args              args
		wantErr           bool
		expectedStudentID string
		srv               *EntryExitModifierService
		setup             func(ctx context.Context)
	}{
		{
			name: "successfully fetched studentID using encryptSecretKeyV2",
			args: args{
				ctx:     ctx,
				content: student1Content,
			},
			srv:               s,
			wantErr:           false,
			expectedStudentID: "student-id-1",
			setup: func(ctx context.Context) {

			},
		},
		{
			name: "successfully fetched studentID using encryptSecretKeyTokyo",
			args: args{
				ctx:     ctx,
				content: student1ContentTokyo,
			},
			srv:               s,
			wantErr:           false,
			expectedStudentID: "student-id-1-tokyo",
			setup: func(ctx context.Context) {

			},
		},
		{
			name: "successfully fetched studentID using encryptSecretKeySynersia",
			args: args{
				ctx:     ctx,
				content: student1ContentSynersia,
			},
			srv:               s,
			wantErr:           false,
			expectedStudentID: "student-id-1-synersia",
			setup: func(ctx context.Context) {

			},
		},
		{
			name: "successfully fetched studentID using encryptSecretKeyTokyo using service without synersia encryption key",
			args: args{
				ctx:     ctx,
				content: student1ContentTokyo,
			},
			srv:               serviceWithEmptySynersiaKey,
			wantErr:           false,
			expectedStudentID: "student-id-1-tokyo",
			setup: func(ctx context.Context) {

			},
		},
		{
			name: "successfully fetched studentID using encryptSecretKeySynersia using service without tokyo encryption key",
			args: args{
				ctx:     ctx,
				content: student1ContentSynersia,
			},
			srv:               serviceWithEmptyTokyoKey,
			wantErr:           false,
			expectedStudentID: "student-id-1-synersia",
			setup: func(ctx context.Context) {

			},
		},
		{
			name: "successfully fetched studentID using encryptSecretKey using service without tokyo and synersia encryption key",
			args: args{
				ctx:     ctx,
				content: student1Content,
			},
			srv:               serviceWithEmptySynersiaTokyoKey,
			wantErr:           false,
			expectedStudentID: "student-id-1",
			setup: func(ctx context.Context) {

			},
		},
		{
			name: "non base-64 content",
			args: args{
				ctx:     ctx,
				content: "invalid-content",
			},
			srv:     s,
			wantErr: true,
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "with invalid secret key",
			args: args{
				ctx:     ctx,
				content: invalidContent,
			},
			srv:     s,
			wantErr: true,
			setup: func(ctx context.Context) {
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(tt.args.ctx)

			studentID, err := tt.srv.getStudentIDFromQr(tt.args.ctx, tt.args.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("EntryExitModifierService.Generate() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr == false {
				assert.NotEmpty(t, studentID)
				assert.Equal(t, tt.expectedStudentID, studentID)
			}
		})
	}
}

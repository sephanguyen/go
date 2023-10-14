package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/entryexitmgmt/entities"
	"github.com/manabie-com/backend/internal/entryexitmgmt/services/uploader"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mock_repositories "github.com/manabie-com/backend/mock/entryexitmgmt/repositories"
	mock_services "github.com/manabie-com/backend/mock/entryexitmgmt/services"
	mock_uploader "github.com/manabie-com/backend/mock/entryexitmgmt/services/uploader"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	eepb "github.com/manabie-com/backend/pkg/manabuf/entryexitmgmt/v1"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type mockPgConnError struct {
	code   string
	errMsg string
}

func (e *mockPgConnError) Error() string {
	return e.errMsg
}

func (e *mockPgConnError) Unwrap() error {
	return &pgconn.PgError{
		Code:    e.code,
		Message: e.errMsg,
	}
}

func TestEntryExitModifierService_Generate(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db := new(mock_database.Ext)
	mockTx := new(mock_database.Tx)
	mockStudentQRRepo := new(mock_repositories.MockStudentQRRepo)
	mockCrypt := new(mock_services.MockCrypt)
	mockSDKUploaderService := new(mock_uploader.Uploader)
	mockCurlUploaderService := new(mock_uploader.Uploader)
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)

	const (
		resumableUploadURL = "http://sample.com"
		v1Url              = "http://sample.com/qrcodeV1"
		v2Url              = "http://sample.com/qrcodeV2"
		encryptStudentV2   = "eyJRckNvZGUiOiJBcm5SVW5iVEllOEhpNnpEWEQtUUZ3TG9ZcXI3MUwwOUl2TjZCcGN6WTVDNzByak04ZFVWMlE9PSIsIlZlcnNpb24iOiJ2MiJ9"
		studentID          = "student-id-1"
	)

	mockV1Qr := &entities.StudentQR{
		StudentID: database.Text(studentID),
		QRURL:     database.Text(v1Url),
		Version:   database.Text(""),
	}

	mockV2Qr := &entities.StudentQR{
		StudentID: database.Text(studentID),
		QRURL:     database.Text(v2Url),
		Version:   database.Text("v2"),
	}

	s := &EntryExitModifierService{
		DB:            db,
		StudentQRRepo: mockStudentQRRepo,
		CryptV2:       mockCrypt,
		UploadServiceSelector: &UploadServiceSelector{
			SdkUploadService:  mockSDKUploaderService,
			CurlUploadService: mockCurlUploaderService,
			UnleashClient:     mockUnleashClient,
			Env:               "local",
		},
	}

	type args struct {
		ctx       context.Context
		studentID string
	}
	tests := []struct {
		name     string
		args     args
		wantErr  bool
		emptyURL bool
		setup    func(ctx context.Context)
	}{
		{
			name: "successful generation of qrcode",
			args: args{
				ctx:       ctx,
				studentID: studentID,
			},
			wantErr: false,
			setup: func(ctx context.Context) {
				mockStudentQRRepo.On("FindByID", ctx, db, mock.Anything).Once().Return(nil, nil)
				mockCrypt.On("Encrypt", mock.Anything, mock.Anything).Once().Return(encryptStudentV2, nil)
				db.On("Begin", ctx).Once().Return(mockTx, nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(true, nil)
				mockSDKUploaderService.On("InitUploader", ctx, mock.Anything).Once().Return(&uploader.UploadInfo{
					DownloadURL: v2Url,
					DoUploadFromFile: func(ctx context.Context, filePathName string) error {
						return nil
					},
				}, nil)
				mockStudentQRRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "successful return of already existing qr with updated version",
			args: args{
				ctx:       ctx,
				studentID: studentID,
			},
			wantErr: false,
			setup: func(ctx context.Context) {
				mockStudentQRRepo.On("FindByID", ctx, db, mock.Anything).Once().Return(mockV2Qr, nil)
			},
		},
		{
			name: "successful generation of new version when student QR is outdated",
			args: args{
				ctx:       ctx,
				studentID: studentID,
			},
			wantErr: false,
			setup: func(ctx context.Context) {
				mockStudentQRRepo.On("FindByID", ctx, db, mock.Anything).Once().Return(mockV1Qr, nil)
				mockCrypt.On("Encrypt", mock.Anything, mock.Anything).Once().Return(encryptStudentV2, nil)
				db.On("Begin", ctx).Once().Return(mockTx, nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(true, nil)
				mockSDKUploaderService.On("InitUploader", ctx, mock.Anything).Once().Return(&uploader.UploadInfo{
					DownloadURL: v2Url,
					DoUploadFromFile: func(ctx context.Context, filePathName string) error {
						return nil
					},
				}, nil)
				mockStudentQRRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "qrcode content encryption fails",
			args: args{
				ctx:       ctx,
				studentID: studentID,
			},
			wantErr: true,
			setup: func(ctx context.Context) {
				mockStudentQRRepo.On("FindByID", ctx, db, mock.Anything).Once().Return(nil, nil)
				mockCrypt.On("Encrypt", mock.Anything, mock.Anything).Once().Return("", fmt.Errorf("error Encrypt"))
			},
		},
		{
			name: "initialize QR URL fails",
			args: args{
				ctx:       ctx,
				studentID: studentID,
			},
			wantErr: true,
			setup: func(ctx context.Context) {
				mockStudentQRRepo.On("FindByID", ctx, db, mock.Anything).Once().Return(nil, nil)
				mockCrypt.On("Encrypt", mock.Anything, mock.Anything).Once().Return(encryptStudentV2, nil)
				db.On("Begin", ctx).Once().Return(mockTx, nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(true, nil)
				mockSDKUploaderService.On("InitUploader", ctx, mock.Anything).Once().Return(&uploader.UploadInfo{
					DownloadURL: v2Url,
					DoUploadFromFile: func(ctx context.Context, filePathName string) error {
						return nil
					},
				}, fmt.Errorf("error Generating Download URL"))
			},
		},
		{
			name: "uploading failed",
			args: args{
				ctx:       ctx,
				studentID: studentID,
			},
			wantErr: true,
			setup: func(ctx context.Context) {
				mockStudentQRRepo.On("FindByID", ctx, db, mock.Anything).Once().Return(nil, nil)
				mockCrypt.On("Encrypt", mock.Anything, mock.Anything).Once().Return(encryptStudentV2, nil)
				db.On("Begin", ctx).Once().Return(mockTx, nil)
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(true, nil)
				mockSDKUploaderService.On("InitUploader", ctx, mock.Anything).Once().Return(&uploader.UploadInfo{
					DownloadURL: v2Url,
					DoUploadFromFile: func(ctx context.Context, filePathName string) error {
						return fmt.Errorf("error on uploading")
					},
				}, nil)
				mockStudentQRRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "saving failed with error pgconn 23505",
			args: args{
				ctx:       ctx,
				studentID: studentID,
			},
			wantErr:  false,
			emptyURL: true,
			setup: func(ctx context.Context) {
				mockStudentQRRepo.On("FindByID", ctx, db, mock.Anything).Once().Return(nil, nil)
				mockCrypt.On("Encrypt", mock.Anything, mock.Anything).Once().Return(encryptStudentV2, nil)
				db.On("Begin", ctx).Once().Return(mockTx, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(true, nil)
				mockSDKUploaderService.On("InitUploader", ctx, mock.Anything).Once().Return(&uploader.UploadInfo{
					DownloadURL: v2Url,
					DoUploadFromFile: func(ctx context.Context, filePathName string) error {
						return nil
					},
				}, nil)
				mockStudentQRRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(&mockPgConnError{code: "23505", errMsg: "mock pgconn error"})
			},
		},
		{
			name: "saving failed with error pgconn 23503",
			args: args{
				ctx:       ctx,
				studentID: studentID,
			},
			wantErr:  true,
			emptyURL: true,
			setup: func(ctx context.Context) {
				mockStudentQRRepo.On("FindByID", ctx, db, mock.Anything).Once().Return(nil, nil)
				mockCrypt.On("Encrypt", mock.Anything, mock.Anything).Once().Return(encryptStudentV2, nil)

				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(true, nil)
				mockSDKUploaderService.On("InitUploader", ctx, mock.Anything).Once().Return(&uploader.UploadInfo{
					DownloadURL: v2Url,
					DoUploadFromFile: func(ctx context.Context, filePathName string) error {
						return nil
					},
				}, nil)

				for i := 0; i < 10; i++ {
					db.On("Begin", ctx).Once().Return(mockTx, nil)
					mockTx.On("Rollback", ctx).Once().Return(nil)
					mockStudentQRRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(&mockPgConnError{code: "23503", errMsg: "Foreign Key Violation error"})
				}
			},
		},
		{
			name: "saving failed with student QR RLS error",
			args: args{
				ctx:       ctx,
				studentID: studentID,
			},
			wantErr:  true,
			emptyURL: true,
			setup: func(ctx context.Context) {
				mockStudentQRRepo.On("FindByID", ctx, db, mock.Anything).Once().Return(nil, nil)
				mockCrypt.On("Encrypt", mock.Anything, mock.Anything).Once().Return(encryptStudentV2, nil)

				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(true, nil)
				mockSDKUploaderService.On("InitUploader", ctx, mock.Anything).Once().Return(&uploader.UploadInfo{
					DownloadURL: v2Url,
					DoUploadFromFile: func(ctx context.Context, filePathName string) error {
						return nil
					},
				}, nil)

				for i := 0; i < 10; i++ {
					db.On("Begin", ctx).Once().Return(mockTx, nil)
					mockTx.On("Rollback", ctx).Once().Return(nil)
					mockStudentQRRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(&mockPgConnError{code: "42501", errMsg: "new row violates row-level security policy for table \"student_qr\" (SQLSTATE 42501)"})
				}
			},
		},
		{
			name: "saving succeed with student QR RLS error after retry",
			args: args{
				ctx:       ctx,
				studentID: studentID,
			},
			wantErr: false,
			setup: func(ctx context.Context) {
				mockStudentQRRepo.On("FindByID", ctx, db, mock.Anything).Once().Return(nil, nil)
				mockCrypt.On("Encrypt", mock.Anything, mock.Anything).Once().Return(encryptStudentV2, nil)

				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(true, nil)
				mockSDKUploaderService.On("InitUploader", ctx, mock.Anything).Once().Return(&uploader.UploadInfo{
					DownloadURL: v2Url,
					DoUploadFromFile: func(ctx context.Context, filePathName string) error {
						return nil
					},
				}, nil)

				// Retry 3 times
				for i := 0; i < 3; i++ {
					db.On("Begin", ctx).Once().Return(mockTx, nil)
					mockTx.On("Rollback", ctx).Once().Return(nil)
					mockStudentQRRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(&mockPgConnError{code: "42501", errMsg: "new row violates row-level security policy for table \"student_qr\" (SQLSTATE 42501)"})
				}

				// Upsert succeed after student QR repo did not return error
				db.On("Begin", ctx).Once().Return(mockTx, nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
				mockStudentQRRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "saving failed with error pgconn 23503 but succeed in retry",
			args: args{
				ctx:       ctx,
				studentID: studentID,
			},
			wantErr: false,
			setup: func(ctx context.Context) {
				mockStudentQRRepo.On("FindByID", ctx, db, mock.Anything).Once().Return(nil, nil)
				mockCrypt.On("Encrypt", mock.Anything, mock.Anything).Once().Return(encryptStudentV2, nil)

				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(true, nil)
				mockSDKUploaderService.On("InitUploader", ctx, mock.Anything).Once().Return(&uploader.UploadInfo{
					DownloadURL: v2Url,
					DoUploadFromFile: func(ctx context.Context, filePathName string) error {
						return nil
					},
				}, nil)

				for i := 0; i < 2; i++ {
					db.On("Begin", ctx).Once().Return(mockTx, nil)
					mockTx.On("Rollback", ctx).Once().Return(nil)
					mockStudentQRRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(&mockPgConnError{code: "23503", errMsg: "Foreign Key Violation error"})
				}

				db.On("Begin", ctx).Once().Return(mockTx, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)
				mockStudentQRRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "storing info to database fails",
			args: args{
				ctx:       ctx,
				studentID: studentID,
			},
			wantErr: true,
			setup: func(ctx context.Context) {
				mockStudentQRRepo.On("FindByID", ctx, db, mock.Anything).Once().Return(nil, nil)
				mockCrypt.On("Encrypt", mock.Anything, mock.Anything).Once().Return(encryptStudentV2, nil)
				db.On("Begin", ctx).Once().Return(mockTx, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(true, nil)
				mockSDKUploaderService.On("InitUploader", ctx, mock.Anything).Once().Return(&uploader.UploadInfo{
					DownloadURL: v2Url,
					DoUploadFromFile: func(ctx context.Context, filePathName string) error {
						return nil
					},
				}, nil)
				mockStudentQRRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(pgx.ErrTxClosed)
			},
		},
		{
			name: "successful generation and upload of qrcode using curl file uploader",
			args: args{
				ctx:       ctx,
				studentID: studentID,
			},
			wantErr: false,
			setup: func(ctx context.Context) {
				mockStudentQRRepo.On("FindByID", ctx, db, mock.Anything).Once().Return(nil, nil)
				mockCrypt.On("Encrypt", mock.Anything, mock.Anything).Once().Return(encryptStudentV2, nil)
				db.On("Begin", ctx).Once().Return(mockTx, nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(false, nil)
				mockCurlUploaderService.On("InitUploader", ctx, mock.Anything).Once().Return(&uploader.UploadInfo{
					DownloadURL: v2Url,
					DoUploadFromFile: func(ctx context.Context, filePathName string) error {
						return nil
					},
				}, nil)
				mockStudentQRRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "unleash client return error",
			args: args{
				ctx:       ctx,
				studentID: studentID,
			},
			wantErr: true,
			setup: func(ctx context.Context) {
				mockStudentQRRepo.On("FindByID", ctx, db, mock.Anything).Once().Return(nil, nil)
				mockCrypt.On("Encrypt", mock.Anything, mock.Anything).Once().Return(encryptStudentV2, nil)
				db.On("Begin", ctx).Once().Return(mockTx, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(false, fmt.Errorf("unleash error"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(tt.args.ctx)

			url, err := s.Generate(tt.args.ctx, tt.args.studentID)
			if (err != nil) != tt.wantErr {
				t.Errorf("EntryExitModifierService.Generate() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.emptyURL {
				assert.Empty(t, url)
			}

			if tt.wantErr == false && !tt.emptyURL {
				assert.NotEmpty(t, url)
			}
		})
	}
}

func TestEntryExitModifierService_GenerateBatchQRCodes(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mockDB := new(mock_database.Ext)
	mockTx := new(mock_database.Tx)
	mockStudentQRRepo := new(mock_repositories.MockStudentQRRepo)
	mockCrypt := new(mock_services.MockCrypt)
	mockSDKUploaderService := new(mock_uploader.Uploader)
	mockCurlUploaderService := new(mock_uploader.Uploader)
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)

	const (
		resumableUploadURL = "http://sample.com"
		v1Url              = "http://sample.com/qrcodeV1"
		v2Url              = "http://sample.com/qrcodeV2"
		encryptStudentV2   = "eyJRckNvZGUiOiJSUU1CcmFXWDc4WjE1WmpNLXQ0NUxpTlJnQXB6UTVQUkVSZG9HN1ZDLXNnLVh1SmhYbGxZVGlEY3g2cEYiLCJWZXJzaW9uIjoidjIifQ=="
		studentID          = "sample-student-id"
	)

	var mockValidStudentIDs []string
	var mockValidGeneratedQRCodesURLs []*eepb.GenerateBatchQRCodesResponse_GeneratedQRCodesURL

	for i := 0; i < 4; i++ {
		studentID := uuid.NewString()
		mockValidStudentIDs = append(mockValidStudentIDs, studentID)
		mockValidGeneratedQRCodesURLs = append(mockValidGeneratedQRCodesURLs, &eepb.GenerateBatchQRCodesResponse_GeneratedQRCodesURL{
			StudentId: studentID,
			Url:       mock.Anything,
		})
	}

	mockInvalidStudentIDs := []string{uuid.NewString()}

	s := &EntryExitModifierService{
		DB:            mockDB,
		StudentQRRepo: mockStudentQRRepo,
		CryptV2:       mockCrypt,
		UploadServiceSelector: &UploadServiceSelector{
			SdkUploadService:  mockSDKUploaderService,
			CurlUploadService: mockCurlUploaderService,
			UnleashClient:     mockUnleashClient,
			Env:               "local",
		},
	}

	testcases := []TestCase{
		{
			name:        "invalid student ids",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "student ids cannot be empty"),
			req: &eepb.GenerateBatchQRCodesRequest{
				StudentIds: nil,
			},
			expectedResp: nil,
			setup: func(ctx context.Context) {
			},
		},
		{
			name:        "happy case",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: nil,
			req: &eepb.GenerateBatchQRCodesRequest{
				StudentIds: mockValidStudentIDs,
			},
			expectedResp: &eepb.GenerateBatchQRCodesResponse{
				QrCodes: mockValidGeneratedQRCodesURLs,
				Errors:  []*eepb.GenerateBatchQRCodesResponse_GenerateBatchQRCodesError{},
			},
			setup: func(ctx context.Context) {
				for i := 0; i < len(mockValidStudentIDs); i++ {
					mockStudentQRRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(nil, nil)
					mockCrypt.On("Encrypt", mock.Anything, mock.Anything).Once().Return(encryptStudentV2, nil)
					mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
					mockTx.On("Commit", ctx).Once().Return(nil)
					mockStudentQRRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
					mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(true, nil)
					mockSDKUploaderService.On("InitUploader", ctx, mock.Anything).Return(&uploader.UploadInfo{
						DownloadURL: v2Url,
						DoUploadFromFile: func(ctx context.Context, filePathName string) error {
							return nil
						},
					}, nil)
				}
			},
		},
		{
			name: "failed to saved",
			ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			req: &eepb.GenerateBatchQRCodesRequest{
				StudentIds: mockInvalidStudentIDs,
			},
			expectedResp: &eepb.GenerateBatchQRCodesResponse{
				Errors: []*eepb.GenerateBatchQRCodesResponse_GenerateBatchQRCodesError{
					{
						StudentId: mockInvalidStudentIDs[0],
						Error:     "s.StudentQRRepo.Upsert: tx is closed",
					},
				},
			},
			setup: func(ctx context.Context) {
				mockStudentQRRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(nil, nil)
				mockCrypt.On("Encrypt", mock.Anything, mock.Anything).Once().Return(encryptStudentV2, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(true, nil)
				mockSDKUploaderService.On("InitUploader", ctx, mock.Anything).Return(&uploader.UploadInfo{
					DownloadURL: v2Url,
					DoUploadFromFile: func(ctx context.Context, filePathName string) error {
						return nil
					},
				}, nil)
				mockStudentQRRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(pgx.ErrTxClosed)
			},
		},
		{
			name: "successful generation and upload of qrcode using curl file uploader",
			ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			req: &eepb.GenerateBatchQRCodesRequest{
				StudentIds: mockInvalidStudentIDs,
			},
			expectedResp: &eepb.GenerateBatchQRCodesResponse{
				Errors: []*eepb.GenerateBatchQRCodesResponse_GenerateBatchQRCodesError{
					{
						StudentId: mockInvalidStudentIDs[0],
						Error:     "s.StudentQRRepo.Upsert: tx is closed",
					},
				},
			},
			setup: func(ctx context.Context) {
				mockStudentQRRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(nil, nil)
				mockCrypt.On("Encrypt", mock.Anything, mock.Anything).Once().Return(encryptStudentV2, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(false, nil)
				mockCurlUploaderService.On("InitUploader", ctx, mock.Anything).Return(&uploader.UploadInfo{
					DownloadURL: v2Url,
					DoUploadFromFile: func(ctx context.Context, filePathName string) error {
						return nil
					},
				}, nil)
				mockStudentQRRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(pgx.ErrTxClosed)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			resp, err := s.GenerateBatchQRCodes(testCase.ctx, testCase.req.(*eepb.GenerateBatchQRCodesRequest))
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

			switch testCase.expectedErr {
			case nil:
				assert.Equal(t, testCase.expectedErr, err)
				assert.NotNil(t, resp)

				if testCase.expectedResp == nil {
					break
				}

				qrCodesLength := len(testCase.expectedResp.(*eepb.GenerateBatchQRCodesResponse).QrCodes)
				assert.Equal(t, testCase.expectedResp.(*eepb.GenerateBatchQRCodesResponse).Errors, resp.Errors)
				assert.Equal(t, qrCodesLength, len(resp.QrCodes))

				if qrCodesLength > 0 {
					for _, qrCode := range testCase.expectedResp.(*eepb.GenerateBatchQRCodesResponse).QrCodes {
						assert.NotEmpty(t, qrCode.StudentId)
						assert.NotEmpty(t, qrCode.Url)
					}
				}

			default:
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
				assert.Nil(t, resp)
			}

			mock.AssertExpectationsForObjects(t, mockDB, mockStudentQRRepo)
		})
	}
}

func Test_encryptWithInvalidKey(t *testing.T) {
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
		encryptSecretKeyV2: "invalid-secret-key",
	}

	type args struct {
		ctx       context.Context
		studentID string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		setup   func(ctx context.Context)
	}{
		{
			name: "invalid secret key error",
			args: args{
				ctx:       ctx,
				studentID: studentID,
			},
			wantErr: true,
			setup: func(ctx context.Context) {
				mockStudentQRRepo.On("FindByID", ctx, db, mock.Anything).Once().Return(nil, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(tt.args.ctx)

			url, err := s.Generate(tt.args.ctx, tt.args.studentID)
			if (err != nil) != tt.wantErr {
				t.Errorf("EntryExitModifierService.Generate() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr == false {
				assert.NotEmpty(t, url)
			}
		})
	}
}

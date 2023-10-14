package file_service

import (
	"context"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/utils"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"
	mockServices "github.com/manabie-com/backend/mock/payment/services/file_service"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestFileServiceUploadFile(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	var (
		db                 *mockDb.Ext
		tx                 *mockDb.Tx
		fileStorageService *mockServices.IFileStorageService
	)

	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when upload file to file storage service",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				fileStorageService.On("UploadFile", ctx,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return("", constant.ErrDefault)
				tx.On("Rollback", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
		{
			Name:        "Happy case:",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				fileStorageService.On("UploadFile", ctx,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return("123", nil)
				tx.On("Commit", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			tx = new(mockDb.Tx)
			fileStorageService = new(mockServices.IFileStorageService)
			testCase.Setup(testCase.Ctx)
			s := &FileService{
				DB:          db,
				FileStorage: fileStorageService,
			}

			resp, err := s.UploadFile(testCase.Ctx, &pb.UploadFileRequest{})
			if err != nil {
				fmt.Println(err)
			}

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.NotNil(t, resp)
			}

			mock.AssertExpectationsForObjects(
				t,
				db,
				fileStorageService,
			)
		})
	}
}

func TestFileServiceGetUploadLink(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	var (
		db                 *mockDb.Ext
		fileStorageService *mockServices.IFileStorageService
	)

	testcases := []utils.TestCase{
		{
			Name:        "Happy case:",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				fileStorageService.On("GetDownloadFileByName", ctx,
					mock.Anything,
					mock.Anything,
				).Return("123", nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			fileStorageService = new(mockServices.IFileStorageService)
			testCase.Setup(testCase.Ctx)
			s := &FileService{
				DB:          db,
				FileStorage: fileStorageService,
			}

			resp, err := s.GetEnrollmentFile(testCase.Ctx, &pb.GetEnrollmentFileRequest{})
			if err != nil {
				fmt.Println(err)
			}

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.NotNil(t, resp)
			}

			mock.AssertExpectationsForObjects(
				t,
				db,
				fileStorageService,
			)
		})
	}
}

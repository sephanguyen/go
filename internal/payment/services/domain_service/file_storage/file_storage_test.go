package file_storage

import (
	"bytes"
	"context"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"
	mockRepositories "github.com/manabie-com/backend/mock/payment/repositories"
	mockUtilsServices "github.com/manabie-com/backend/mock/payment/utils"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestFileStorageServiceUploadFile(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db       *mockDb.Ext
		fileRepo *mockRepositories.MockFileRepo
		store    *mockUtilsServices.IStorage
	)
	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Upload file to store",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				fileRepo.On("GetByFileName", ctx, mock.Anything, mock.Anything).Return(entities.File{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: create file in to database",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				fileRepo.On("GetByFileName", ctx, mock.Anything, mock.Anything).Return(entities.File{}, nil)
				store.On("UploadFromFile", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("", nil)
				fileRepo.On("Create", ctx, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name:        "Happy case",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				fileRepo.On("GetByFileName", ctx, mock.Anything, mock.Anything).Return(entities.File{
					FileID: pgtype.Text{Status: pgtype.Present, String: "1"},
				}, nil)
				store.On("UploadFromFile", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("", nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			fileRepo = new(mockRepositories.MockFileRepo)
			store = new(mockUtilsServices.IStorage)
			testCase.Setup(testCase.Ctx)
			s := &FileStorageService{
				FileRepo: fileRepo,
				Store:    store,
			}
			contentReader := bytes.NewReader([]byte("test"))
			url, err := s.UploadFile(testCase.Ctx, contentReader, db, "", "", 12)

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, url)
			}

			mock.AssertExpectationsForObjects(t, db, fileRepo, store)
		})
	}
}

func TestFileStorageServiceGetDownloadFileByName(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db       *mockDb.Ext
		fileRepo *mockRepositories.MockFileRepo
		store    *mockUtilsServices.IStorage
	)
	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Get file from repo",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				fileRepo.On("GetByFileName", ctx, mock.Anything, mock.Anything).Return(entities.File{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Happy case",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				fileRepo.On("GetByFileName", ctx, mock.Anything, mock.Anything).Return(entities.File{}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			fileRepo = new(mockRepositories.MockFileRepo)
			store = new(mockUtilsServices.IStorage)
			testCase.Setup(testCase.Ctx)
			s := &FileStorageService{
				FileRepo: fileRepo,
				Store:    store,
			}
			url, err := s.GetDownloadFileByName(testCase.Ctx, db, "")

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, url)
			}

			mock.AssertExpectationsForObjects(t, db, fileRepo, store)
		})
	}
}

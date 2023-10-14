package uploader

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/entryexitmgmt/constant"
	"github.com/manabie-com/backend/internal/golibs/configs"
	mock_filestore "github.com/manabie-com/backend/mock/entryexitmgmt/services/filestore"

	"github.com/stretchr/testify/mock"
)

func TestUploaderSDKService(t *testing.T) {

	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockFileStore := &mock_filestore.FileStore{}
	config := &configs.StorageConfig{
		Endpoint:             "http://example.com",
		Bucket:               "manabie-test",
		FileUploadFolderPath: "entryexitmgmt-upload",
	}

	uploader := &SDKUploaderService{
		Cfg:       config,
		FileStore: mockFileStore,
	}

	testObjectName := "test-object"
	testObjectPath := "/test-dir"
	fileExtension := constant.PNG
	expectedDownloadURL := fmt.Sprintf("%s/%s/%s/%s.%s", config.Endpoint, config.Bucket, config.FileUploadFolderPath, testObjectName, fileExtension)

	testCases := []struct {
		name     string
		ctx      context.Context
		req      *UploadRequest
		pathName string
		resp     *UploadInfo

		hasError bool
		setup    func(ctx context.Context)
	}{
		{
			name: "Init Upload Successfully",
			ctx:  ctx,
			req: &UploadRequest{
				ObjectName:    testObjectName,
				FileExtension: fileExtension,
			},
			pathName: testObjectPath,
			resp: &UploadInfo{
				DownloadURL: expectedDownloadURL,
			},
			setup: func(ctx context.Context) {
				mockFileStore.On("GetDownloadURL", mock.Anything).Once().Return(expectedDownloadURL)
				mockFileStore.On("UploadFromFile", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "Init Upload With Error",
			ctx:  ctx,
			req: &UploadRequest{
				ObjectName:    testObjectName,
				FileExtension: fileExtension,
			},
			pathName: testObjectPath,
			hasError: true,
			resp: &UploadInfo{
				DownloadURL: expectedDownloadURL,
			},
			setup: func(ctx context.Context) {
				mockFileStore.On("GetDownloadURL", mock.Anything).Once().Return(expectedDownloadURL)
				mockFileStore.On("UploadFromFile", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(errors.New("mock error"))
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			uploader, _ := uploader.InitUploader(testCase.ctx, testCase.req)

			err := uploader.DoUploadFromFile(testCase.ctx, testCase.pathName)
			if testCase.hasError {
				if err == nil {
					t.Errorf("Expecting an error got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expecting a nil error got %v", err)
				}
			}

			if uploader.DownloadURL != testCase.resp.DownloadURL {
				t.Errorf("Expecting download URL %s got %s", testCase.resp.DownloadURL, uploader.DownloadURL)
			}

		})
	}
}

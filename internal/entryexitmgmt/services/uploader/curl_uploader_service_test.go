package uploader

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/entryexitmgmt/constant"
	"github.com/manabie-com/backend/internal/golibs/configs"
	mock_uploads "github.com/manabie-com/backend/mock/bob/services/uploads"
	mock_curl "github.com/manabie-com/backend/mock/golibs/curl"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"
)

func TestCurlUploaderService(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockUploadService := new(mock_uploads.MockUploadReaderService)
	mockGolibHTTP := new(mock_curl.IHTTP)

	config := &configs.StorageConfig{
		Endpoint:             "http://example.com",
		Bucket:               "manabie-bob-test",
		FileUploadFolderPath: "entryexitmgmt-upload",
	}

	uploader := &CurlUploaderService{
		UploadReaderService: mockUploadService,
		HTTP:                mockGolibHTTP,
	}

	studentID := "student-id"
	tempDir, _ := ioutil.TempDir("", "qrcode-test")
	correctPath := fmt.Sprintf("%v/%v.png", tempDir, "student-id")

	fileExtension := constant.PNG
	testObjectName := "test-bob-object"

	expectedDownloadURL := fmt.Sprintf("%s/%s/%s/%s.%s", config.Endpoint, config.Bucket, config.FileUploadFolderPath, testObjectName, fileExtension)
	resumableUploadURL := "http://example-resumable.com"

	testCases := []struct {
		name     string
		ctx      context.Context
		req      *UploadRequest
		resp     *UploadInfo
		pathName string

		hasErrorInURL    bool
		hasErrorInUpload bool
		setup            func(ctx context.Context)
	}{
		{
			name: "Init Upload Successfully",
			ctx:  ctx,
			req: &UploadRequest{
				ObjectName:    testObjectName,
				FileExtension: fileExtension,
			},
			pathName: correctPath,
			resp: &UploadInfo{
				DownloadURL: expectedDownloadURL,
			},
			setup: func(ctx context.Context) {
				qrc, _ := qrcode.New(studentID)
				objectWriter, _ := standard.New(correctPath)
				qrc.Save(objectWriter)

				mockUploadService.On("GenerateResumableUploadURL", mock.Anything, mock.Anything).Once().Return(&bpb.ResumableUploadURLResponse{
					ResumableUploadUrl: resumableUploadURL,
					DownloadUrl:        expectedDownloadURL,
				}, nil)
				mockGolibHTTP.On("Request", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "generate upload url fails",
			ctx:  ctx,
			req: &UploadRequest{
				ObjectName: testObjectName,

				FileExtension: fileExtension,
			},
			pathName:      correctPath,
			hasErrorInURL: true,
			setup: func(ctx context.Context) {
				mockUploadService.On("GenerateResumableUploadURL", mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("error GenerateResumableUploadURL"))
			},
		},
		{
			name: "file type not supported",
			ctx:  ctx,
			req: &UploadRequest{
				ObjectName:    testObjectName,
				FileExtension: "txt",
			},
			pathName:         correctPath,
			hasErrorInUpload: true,
			setup: func(ctx context.Context) {
				mockUploadService.On("GenerateResumableUploadURL", mock.Anything, mock.Anything).Once().Return(&bpb.ResumableUploadURLResponse{
					ResumableUploadUrl: resumableUploadURL,
					DownloadUrl:        expectedDownloadURL,
				}, nil)
			},
		},
		{
			name: "Error in download URL",
			ctx:  ctx,
			req: &UploadRequest{
				ObjectName:    testObjectName,
				FileExtension: fileExtension,
			},
			pathName:      correctPath,
			hasErrorInURL: true,
			setup: func(ctx context.Context) {
				mockUploadService.On("GenerateResumableUploadURL", mock.Anything, mock.Anything).Once().Return(nil, errors.New("mock error"))
			},
		},
		{
			name: "non existing file",
			ctx:  ctx,
			req: &UploadRequest{
				ObjectName:    testObjectName,
				FileExtension: fileExtension,
			},
			pathName:         "Invalid-Path",
			hasErrorInUpload: true,
			setup: func(ctx context.Context) {
				mockUploadService.On("GenerateResumableUploadURL", mock.Anything, mock.Anything).Once().Return(&bpb.ResumableUploadURLResponse{
					ResumableUploadUrl: resumableUploadURL,
					DownloadUrl:        expectedDownloadURL,
				}, nil)
			},
		},
		{
			name: "error in curl",
			ctx:  ctx,
			req: &UploadRequest{
				ObjectName:    testObjectName,
				FileExtension: fileExtension,
			},
			pathName:         correctPath,
			hasErrorInUpload: true,
			setup: func(ctx context.Context) {
				qrc, _ := qrcode.New(studentID)
				objectWriter, _ := standard.New(correctPath)
				qrc.Save(objectWriter)

				mockUploadService.On("GenerateResumableUploadURL", mock.Anything, mock.Anything).Once().Return(&bpb.ResumableUploadURLResponse{
					ResumableUploadUrl: resumableUploadURL,
					DownloadUrl:        expectedDownloadURL,
				}, nil)
				mockGolibHTTP.On("Request", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(fmt.Errorf("curl error"))
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			uploader, err := uploader.InitUploader(testCase.ctx, testCase.req)
			if err != nil {
				if !testCase.hasErrorInURL {
					t.Errorf("Expecting a nil error got %v", err)
				}
				return
			} else {
				if testCase.hasErrorInURL {
					t.Errorf("Expecting %v error got nil", err)
				}
			}

			err = uploader.DoUploadFromFile(testCase.ctx, testCase.pathName)
			log.Println(uploader, err)
			if err != nil {
				if !testCase.hasErrorInUpload {
					t.Errorf("Expecting a nil error got %v", err)
				}
				return
			} else {
				if testCase.hasErrorInURL {
					t.Errorf("Expecting %v error got nil", err)
				}
			}

			if uploader.DownloadURL != testCase.resp.DownloadURL {
				t.Errorf("Expecting download URL %s got %s", testCase.resp.DownloadURL, uploader.DownloadURL)
			}

		})
	}
}

func Test_generateImagePngByte(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	studentID := "student-id"
	incorrectPath := "/unknown-dir-qr-123abc"
	tempDir, _ := ioutil.TempDir("", "qrcode-test")
	correctPath := fmt.Sprintf("%v/%v.png", tempDir, "student-id")

	type args struct {
		ctx        context.Context
		objectPath string
	}
	tests := []struct {
		name        string
		args        args
		setup       func(ctx context.Context)
		expectedErr error
	}{
		{
			name: "Happy Case",
			args: args{
				ctx:        ctx,
				objectPath: correctPath,
			},
			setup: func(ctx context.Context) {
				qrc, _ := qrcode.New(studentID)
				objectWriter, _ := standard.New(correctPath)
				qrc.Save(objectWriter)
			},
			expectedErr: nil,
		},
		{
			name: "No such file or directory",
			args: args{
				ctx:        ctx,
				objectPath: incorrectPath,
			},
			setup:       func(ctx context.Context) {},
			expectedErr: fmt.Errorf("err os.Open: open %s: no such file or directory", incorrectPath),
		},
		{
			name: "failed invalid format",
			args: args{
				ctx:        ctx,
				objectPath: tempDir,
			},
			setup:       func(ctx context.Context) {},
			expectedErr: fmt.Errorf("err image.Decode image: unknown format"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(tt.args.ctx)

			_, err := generateImagePngByte(tt.args.objectPath)
			if tt.expectedErr != nil {
				assert.EqualError(t, err, tt.expectedErr.Error())
			}
		})

		defer os.RemoveAll(tt.args.objectPath)
	}
}

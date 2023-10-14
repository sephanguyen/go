package filestorage

import (
	"context"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/configs"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func TestMinIOStorageService(t *testing.T) {

	minIOEndpoint := makeEndpointWithoutScheme("http://example.com")
	minIOClient, err := minio.New(minIOEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4("test-access-key", "test-secret-key", ""),
		Secure: false,
	})

	if err != nil {
		panic(err)
	}

	s := MinIOStorageService{client: minIOClient, storageConfig: &configs.StorageConfig{Bucket: "test-bucket"}}
	ctx := context.Background()

	testCases := []struct {
		name        string
		ctx         context.Context
		objectName  string
		objectPath  string
		contentType ContentType
		hasError    bool
	}{
		{
			name:        "Invalid file path of csv file",
			ctx:         ctx,
			objectName:  "test-csv-file.csv",
			objectPath:  "/invalid-path@132",
			contentType: ContentTypeCSV,
			hasError:    true,
		},
		{
			name:        "Invalid file path of txt file",
			ctx:         ctx,
			objectName:  "test-txt-file.txt",
			objectPath:  "/invalid-path@132",
			contentType: ContentTypeTXT,
			hasError:    true,
		},
	}

	for _, tc := range testCases {
		fileUploadInfo := FileToUploadInfo{
			ObjectName:  tc.objectName,
			PathName:    tc.objectPath,
			ContentType: tc.contentType,
		}

		err := s.UploadFile(tc.ctx, fileUploadInfo)
		if err != nil {
			if !tc.hasError {
				t.Errorf("Expecting nil error got %v", err)
			}
			continue
		} else {
			if tc.hasError {
				t.Errorf("Expecting an error got nil")
			}
		}
	}

}

func TestMinIOStorageService_Download(t *testing.T) {

	minIOEndpoint := makeEndpointWithoutScheme("http://example.com")
	minIOClient, err := minio.New(minIOEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4("test-access-key", "test-secret-key", ""),
		Secure: false,
	})

	if err != nil {
		panic(err)
	}

	s := MinIOStorageService{client: minIOClient, storageConfig: &configs.StorageConfig{Bucket: "test-bucket"}}
	ctx := context.Background()

	testCases := []struct {
		name        string
		ctx         context.Context
		objectName  string
		objectPath  string
		contentType ContentType
		hasError    bool
	}{
		{
			name:        "Invalid file path of csv file",
			ctx:         ctx,
			objectName:  "test-csv-file.csv",
			objectPath:  "/invalid-path@133",
			contentType: ContentTypeCSV,
			hasError:    true,
		},
		{
			name:        "Invalid file path of txt file",
			ctx:         ctx,
			objectName:  "test-txt-file.txt",
			objectPath:  "/invalid-path@55",
			contentType: ContentTypeTXT,
			hasError:    true,
		},
	}

	for _, tc := range testCases {
		fileDownloadInfo := FileToDownloadInfo{
			ObjectName:          tc.objectName,
			DestinationPathName: tc.objectPath,
			ContentType:         tc.contentType,
		}

		err := s.DownloadFile(tc.ctx, fileDownloadInfo)
		if err != nil {
			if !tc.hasError {
				t.Errorf("Expecting nil error got %v", err)
			}
			continue
		} else {
			if tc.hasError {
				t.Errorf("Expecting an error got nil")
			}
		}
	}

}

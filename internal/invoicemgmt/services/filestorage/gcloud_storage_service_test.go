package filestorage

import (
	"context"
	"testing"
)

func TestGCloudStorageService(t *testing.T) {

	s := GCloudStorageService{}

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
			objectName:  "test-text-file.txt",
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

func TestGCloudStorageService_Download(t *testing.T) {

	s := GCloudStorageService{}

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
			objectPath:  "/invalid-path@155",
			contentType: ContentTypeCSV,
			hasError:    true,
		},
		{
			name:        "Invalid file path of txt file",
			ctx:         ctx,
			objectName:  "test-text-file.txt",
			objectPath:  "/invalid-path@33",
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

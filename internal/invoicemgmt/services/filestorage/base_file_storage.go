package filestorage

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"

	"github.com/manabie-com/backend/internal/golibs/configs"
)

type ContentType string

const (
	GoogleCloudStorageService string = "GCS"
	MinIOService              string = "MinIO"

	ContentTypeCSV ContentType = "application/csv"
	ContentTypeTXT ContentType = "application/text"
)

type FileToUploadInfo struct {
	ObjectName  string
	PathName    string
	ContentType ContentType
}

type FileToDownloadInfo struct {
	ObjectName          string
	DestinationPathName string
	ContentType         ContentType
}

type FileStorage interface {
	GetDownloadURL(objectName string) string
	UploadFile(ctx context.Context, fileUploadInfo FileToUploadInfo) error
	FormatObjectName(string) string
	Close() error
	DownloadFile(ctx context.Context, fileDownloadInfo FileToDownloadInfo) error
}

func GetFileStorage(fileStoreageName string, storageConfig *configs.StorageConfig) (FileStorage, error) {
	if fileStoreageName == GoogleCloudStorageService {
		return NewGCloudStorageService(storageConfig)
	}

	if fileStoreageName == MinIOService {
		return NewMinIOStorageServiceService(storageConfig)
	}

	return nil, fmt.Errorf("file store %v is not yet supported", fileStoreageName)
}

type BaseFileStoreService struct {
	storageConfig *configs.StorageConfig
}

func (s *BaseFileStoreService) GetDownloadURL(objectName string) string {
	dir, filename := filepath.Dir(objectName), filepath.Base(objectName)
	if dir == "." {
		filename = url.PathEscape(filename)
	} else {
		filename = fmt.Sprintf("%s/%s", dir, url.PathEscape(filename))
	}
	return fmt.Sprintf("%s/%s/%s", s.storageConfig.Endpoint, s.storageConfig.Bucket, filename)
}

func (s *BaseFileStoreService) FormatObjectName(objectName string) string {
	// Check for file folder path
	if len(s.storageConfig.FileUploadFolderPath) != 0 {
		objectName = s.storageConfig.FileUploadFolderPath + "/" + objectName
	}

	return objectName
}

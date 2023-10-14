package filestore

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"

	"github.com/manabie-com/backend/internal/entryexitmgmt/constant"
	"github.com/manabie-com/backend/internal/golibs/configs"
)

type FileStore interface {
	UploadFromFile(ctx context.Context, objectName, pathName, contentType string) error
	GetDownloadURL(objectName string) string
	Close() error
}

func GetFileStore(fileStoreName constant.FileStoreName, conf *configs.StorageConfig) (FileStore, error) {
	if fileStoreName == constant.GoogleCloudStorageService {
		return NewGCloudStoreService(conf)
	}

	if fileStoreName == constant.MinIOService {
		return NewMinIOStoreService(conf)
	}

	return nil, fmt.Errorf("Filestore %v is not yet supported", fileStoreName)
}

type BaseFileStoreService struct {
	conf *configs.StorageConfig
}

func (s *BaseFileStoreService) GetDownloadURL(objectName string) string {
	dir, filename := filepath.Dir(objectName), filepath.Base(objectName)
	if dir == "." {
		filename = url.PathEscape(filename)
	} else {
		filename = fmt.Sprintf("%s/%s", dir, url.PathEscape(filename))
	}
	return fmt.Sprintf("%s/%s/%s", s.conf.Endpoint, s.conf.Bucket, filename)
}

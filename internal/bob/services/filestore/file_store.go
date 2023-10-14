package filestore

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/manabie-com/backend/internal/golibs/configs"
)

type ServiceName string

const (
	GoogleCloudStorageService ServiceName = "gcs"
	S3Service                 ServiceName = "s3"
	MinIOService              ServiceName = "minio"
)

type FileStore interface {
	GeneratePresignedPutObjectURL(ctx context.Context, objectName string, expiry time.Duration) (*url.URL, error)
	GenerateResumableObjectURL(ctx context.Context, objectName string, expiry time.Duration, allowOrigin, contentType string) (*url.URL, error)
	GeneratePublicObjectURL(objectName string) string
	GenerateGetObjectURL(ctx context.Context, objectName string, fileName string, expiry time.Duration) (*url.URL, error)
	GetObjectInfo(ctx context.Context, bucketName, objectName string) (*StorageObject, error)
	GetObjectsWithPrefix(ctx context.Context, bucketName, prefix, delim string) ([]*StorageObject, error)
	DeleteObject(ctx context.Context, bucketName, objectName string) error
	MoveObject(ctx context.Context, srcObjectName, dstObjectName string) error
}

// NewFileStore will return file store service, set 'serviceName'
// to appoint storage service be used.
func NewFileStore(serviceName ServiceName, serviceAccountEmail string, conf *configs.StorageConfig) (FileStore, error) {
	if len(conf.Bucket) == 0 {
		return nil, fmt.Errorf("buck name could not be empty")
	}

	switch serviceName {
	case GoogleCloudStorageService:
		return NewGoogleCloudStorage(serviceAccountEmail, conf)
	case S3Service:
		return nil, fmt.Errorf("not yet implement %s service", serviceName)
	case MinIOService:
		return NewMinIO(conf)
	default:
		return nil, fmt.Errorf("invalid file storage service %s", serviceName)
	}
}

type StorageObject struct {
	Size int64
	Name string
}

type ErrorCode int32

const (
	UnknownError      ErrorCode = 0
	FileNotFoundError ErrorCode = 1
)

type Error struct {
	Err       error
	ErrorCode ErrorCode
}

func (e *Error) Error() string {
	return e.Err.Error()
}

package filestore

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/configs"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var _ FileStore = new(MinIO)

type MinIO struct {
	conf   *configs.StorageConfig
	Client *minio.Client
}

func makeEndpointWithoutScheme(endpoint string) string {
	u, err := url.Parse(endpoint)
	if err != nil {
		return endpoint
	}
	return strings.TrimPrefix(endpoint, u.Scheme+"://")
}

func NewMinIO(c *configs.StorageConfig) (*MinIO, error) {
	minIOEndpoint := makeEndpointWithoutScheme(c.Endpoint)

	minIOClient, err := minio.New(minIOEndpoint, &minio.Options{
		Creds:     credentials.NewStaticV4(c.AccessKey, c.SecretKey, ""),
		Secure:    c.Secure,
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: c.InsecureSkipVerify}}, //nolint:gosec
	})
	if err != nil {
		return nil, fmt.Errorf("could not create minio client: %w", err)
	}

	s := &MinIO{
		conf:   c,
		Client: minIOClient,
	}
	return s, nil
}

func (m *MinIO) GeneratePresignedPutObjectURL(ctx context.Context, objectName string, expiry time.Duration) (*url.URL, error) {
	presignedURL, err := m.Client.PresignedPutObject(ctx, m.conf.Bucket, objectName, expiry)
	if err != nil {
		return nil, fmt.Errorf("could not sign url to put bucket %s, object %s: %w", m.conf.Bucket, objectName, err)
	}
	return presignedURL, nil
}

// minio doesn't support resumable upload,temporarily use presign url for this function
func (m *MinIO) GenerateResumableObjectURL(ctx context.Context, objectName string, expiry time.Duration, _, _ string) (*url.URL, error) {
	return m.GeneratePresignedPutObjectURL(ctx, objectName, expiry)
}

func (m *MinIO) GeneratePublicObjectURL(objectName string) string {
	dir, filename := filepath.Dir(objectName), filepath.Base(objectName)
	if dir == "." {
		filename = url.PathEscape(filename)
	} else {
		filename = fmt.Sprintf("%s/%s", dir, url.PathEscape(filename))
	}
	return fmt.Sprintf("%s/%s/%s", m.conf.Endpoint, m.conf.Bucket, filename)
}

func (m *MinIO) GenerateGetObjectURL(ctx context.Context, objectName string, fileName string, expiry time.Duration) (*url.URL, error) {
	// Set request parameters for content-disposition.
	reqParams := make(url.Values)
	reqParams.Set("response-content-disposition", "attachment; filename=\""+fileName+"\"")

	// Generates a presigned url which expires
	presignedURL, err := m.Client.PresignedGetObject(ctx, m.conf.Bucket, objectName, expiry, reqParams)
	if err != nil {
		return nil, fmt.Errorf("could not sign url to get object: %w", err)
	}

	return presignedURL, nil
}

func (m *MinIO) GetObjectInfo(ctx context.Context, bucketName, objectName string) (*StorageObject, error) {
	minObject, err := m.Client.GetObjectACL(ctx, bucketName, objectName)
	if err != nil {
		return nil, err
	}
	storageObject := &StorageObject{
		Size: minObject.Size,
	}
	return storageObject, nil
}

func (m *MinIO) DeleteObject(ctx context.Context, bucketName, objectName string) error {
	return m.Client.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{})
}

func (m *MinIO) GetObjectsWithPrefix(ctx context.Context, bucketName, prefix, delim string) ([]*StorageObject, error) {
	minObjectsChan := m.Client.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
		Prefix: prefix,
	})
	sb := []*StorageObject{}

	for object := range minObjectsChan {
		if object.Err != nil {
			return []*StorageObject{}, object.Err
		}
		storageObject := &StorageObject{
			Name: object.Key,
		}
		sb = append(sb, storageObject)
	}

	return sb, nil
}

func (m *MinIO) MoveObject(ctx context.Context, srcObjectName, destObjetName string) error {
	// Check object is exist
	_, err := m.Client.StatObject(ctx, m.conf.Bucket, srcObjectName, minio.StatObjectOptions{})
	if err != nil {
		return &Error{
			ErrorCode: FileNotFoundError,
			Err:       err,
		}
	}

	_, err = m.Client.CopyObject(
		ctx,
		minio.CopyDestOptions{
			Bucket: m.conf.Bucket,
			Object: destObjetName,
		},
		minio.CopySrcOptions{
			Bucket: m.conf.Bucket,
			Object: srcObjectName,
		},
	)

	if err != nil {
		return &Error{
			ErrorCode: UnknownError,
			Err:       err,
		}
	}

	err = m.Client.RemoveObject(ctx, m.conf.Bucket, srcObjectName, minio.RemoveObjectOptions{})
	if err != nil {
		return &Error{
			ErrorCode: UnknownError,
			Err:       err,
		}
	}

	return nil
}

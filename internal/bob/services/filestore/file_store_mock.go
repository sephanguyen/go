package filestore

import (
	"context"
	"net/url"
	"time"
)

var _ FileStore = new(Mock)

type Mock struct {
	GeneratePresignedPutObjectMock func(ctx context.Context, objectName string, expiry time.Duration) (*url.URL, error)
	GenerateResumableObjectURLMock func(ctx context.Context, objectName string, expiry time.Duration, allowOrigin, contentType string) (*url.URL, error)
	GeneratePublicObjectURLMock    func(objectName string) string
	GenerateGetObjectURLMock       func(ctx context.Context, objectName string, fileName string, expiry time.Duration) (*url.URL, error)
	GetObjectInfoMock              func(ctx context.Context, bucketName, objectName string) (*StorageObject, error)
	DeleteObjectMock               func(ctx context.Context, bucketName, objectName string) error
	GetObjectsWithPrefixMock       func(ctx context.Context, bucketName, prefix, delim string) ([]*StorageObject, error)
	MoveObjectMock                 func(ctx context.Context, objectName, newObjectName string) error
}

func (fs *Mock) GeneratePresignedPutObjectURL(ctx context.Context, objectName string, expiry time.Duration) (*url.URL, error) {
	return fs.GeneratePresignedPutObjectMock(ctx, objectName, expiry)
}

func (fs *Mock) GenerateResumableObjectURL(ctx context.Context, objectName string, expiry time.Duration, allowOrigin, contentType string) (*url.URL, error) {
	return fs.GenerateResumableObjectURLMock(ctx, objectName, expiry, allowOrigin, contentType)
}

func (fs *Mock) GeneratePublicObjectURL(objectName string) string {
	return fs.GeneratePublicObjectURLMock(objectName)
}

func (fs *Mock) GenerateGetObjectURL(ctx context.Context, objectName string, fileName string, expiry time.Duration) (*url.URL, error) {
	return fs.GenerateGetObjectURLMock(ctx, objectName, fileName, expiry)
}

func (fs *Mock) GetObjectInfo(ctx context.Context, bucketName, objectName string) (*StorageObject, error) {
	return fs.GetObjectInfoMock(ctx, bucketName, objectName)
}

func (fs *Mock) DeleteObject(ctx context.Context, bucketName, objectName string) error {
	return fs.DeleteObjectMock(ctx, bucketName, objectName)
}

func (fs *Mock) GetObjectsWithPrefix(ctx context.Context, bucketName, prefix, delim string) ([]*StorageObject, error) {
	return fs.GetObjectsWithPrefixMock(ctx, bucketName, prefix, delim)
}

func (fs *Mock) MoveObject(ctx context.Context, objectName, newObjectName string) error {
	return fs.MoveObjectMock(ctx, objectName, newObjectName)
}

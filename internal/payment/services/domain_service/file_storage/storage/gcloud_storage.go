package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/manabie-com/backend/internal/golibs/configs"

	"cloud.google.com/go/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GCloudStorage struct {
	client *storage.Client
	conf   *configs.StorageConfig
}

func (s *GCloudStorage) UploadFromFile(ctx context.Context, reader io.Reader, fileID string, contentType string, fileSize int64) (downloadUrl string, err error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	o := s.client.Bucket(s.conf.Bucket).Object(fileID)

	wc := o.NewWriter(ctx)
	if _, err = io.Copy(wc, reader); err != nil {
		err = status.Errorf(codes.Internal, "io.Copy: %v", err.Error())
		return
	}

	if err = wc.Close(); err != nil {
		err = status.Errorf(codes.Internal, "Writer.Close: %v", err)
		return
	}
	downloadUrl = fmt.Sprintf("%s/%s/%s", s.conf.Endpoint, s.conf.Bucket, fileID)
	return
}

func NewGCloudStoreService(conf *configs.StorageConfig) (*GCloudStorage, error) {
	ctx := context.Background()

	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	return &GCloudStorage{
		conf:   conf,
		client: client,
	}, nil
}

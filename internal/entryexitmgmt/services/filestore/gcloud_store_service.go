package filestore

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/manabie-com/backend/internal/golibs/configs"

	"cloud.google.com/go/storage"
)

type GCloudStoreService struct {
	*BaseFileStoreService
	client *storage.Client
	conf   *configs.StorageConfig
}

func NewGCloudStoreService(conf *configs.StorageConfig) (*GCloudStoreService, error) {
	ctx := context.Background()

	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	return &GCloudStoreService{
		BaseFileStoreService: &BaseFileStoreService{
			conf: conf,
		},
		conf:   conf,
		client: client,
	}, nil
}

func (s *GCloudStoreService) UploadFromFile(ctx context.Context, objectName, pathName, contentType string) error {
	f, err := os.Open(pathName)
	if err != nil {
		return fmt.Errorf("os.Open: %v", err)
	}
	defer f.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	o := s.client.Bucket(s.conf.Bucket).Object(objectName)

	wc := o.NewWriter(ctx)
	if _, err = io.Copy(wc, f); err != nil {
		return fmt.Errorf("io.Copy: %v", err)
	}

	if err := wc.Close(); err != nil {
		return fmt.Errorf("Writer.Close: %v", err)
	}

	return nil
}

func (s *GCloudStoreService) Close() error {
	return s.client.Close()
}

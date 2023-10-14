package filestorage

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/manabie-com/backend/internal/golibs/configs"

	"cloud.google.com/go/storage"
)

type GCloudStorageService struct {
	*BaseFileStoreService
	client        *storage.Client
	storageConfig *configs.StorageConfig
}

func NewGCloudStorageService(storageConfig *configs.StorageConfig) (*GCloudStorageService, error) {
	ctx := context.Background()

	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	return &GCloudStorageService{
		BaseFileStoreService: &BaseFileStoreService{
			storageConfig: storageConfig,
		},
		storageConfig: storageConfig,
		client:        client,
	}, nil
}

func (s *GCloudStorageService) UploadFile(ctx context.Context, fileUploadInfo FileToUploadInfo) error {
	f, err := os.Open(fileUploadInfo.PathName)
	if err != nil {
		return fmt.Errorf("os.Open: %v", err)
	}
	defer f.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	o := s.client.Bucket(s.storageConfig.Bucket).Object(fileUploadInfo.ObjectName)

	wc := o.NewWriter(ctx)
	if _, err = io.Copy(wc, f); err != nil {
		return fmt.Errorf("io.Copy: %v", err)
	}

	if err := wc.Close(); err != nil {
		return fmt.Errorf("Writer.Close: %v", err)
	}

	return nil
}

func (s *GCloudStorageService) Close() error {
	return s.client.Close()
}

func (s *GCloudStorageService) DownloadFile(ctx context.Context, fileDownloadInfo FileToDownloadInfo) error {
	destinationFile, err := os.OpenFile(fileDownloadInfo.DestinationPathName, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return fmt.Errorf("os.OpenFile: %v", err)
	}

	defer destinationFile.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	rc, err := s.client.Bucket(s.storageConfig.Bucket).Object(fileDownloadInfo.ObjectName).NewReader(ctx)

	if err != nil {
		return fmt.Errorf("Object(%q).NewReader: %v", fileDownloadInfo.ObjectName, err)
	}
	defer rc.Close()
	// destinationFile is where it will write the contents of the file from gcloud
	if _, err = io.Copy(destinationFile, rc); err != nil {
		return fmt.Errorf("io.Copy: %v", err)
	}

	if err = destinationFile.Close(); err != nil {
		return fmt.Errorf("destinationFile.Close: %v", err)
	}

	return nil
}

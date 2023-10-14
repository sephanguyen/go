package filestorage

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/configs"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIOStorageService struct {
	*BaseFileStoreService
	client        *minio.Client
	storageConfig *configs.StorageConfig
}

func makeEndpointWithoutScheme(endpoint string) string {
	u, err := url.Parse(endpoint)
	if err != nil {
		return endpoint
	}
	return strings.TrimPrefix(endpoint, u.Scheme+"://")
}

func NewMinIOStorageServiceService(storageConfig *configs.StorageConfig) (*MinIOStorageService, error) {
	minIOEndpoint := makeEndpointWithoutScheme(storageConfig.Endpoint)
	minIOClient, err := minio.New(minIOEndpoint, &minio.Options{
		Creds:     credentials.NewStaticV4(storageConfig.AccessKey, storageConfig.SecretKey, ""),
		Secure:    storageConfig.Secure,
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: storageConfig.InsecureSkipVerify}}, //nolint:gosec
	})

	if err != nil {
		return nil, err
	}

	return &MinIOStorageService{
		BaseFileStoreService: &BaseFileStoreService{
			storageConfig: storageConfig,
		},
		client:        minIOClient,
		storageConfig: storageConfig,
	}, nil
}

func (s *MinIOStorageService) UploadFile(ctx context.Context, fileUploadInfo FileToUploadInfo) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	_, err := s.client.FPutObject(ctx, s.storageConfig.Bucket, fileUploadInfo.ObjectName, fileUploadInfo.PathName, minio.PutObjectOptions{
		ContentType: string(fileUploadInfo.ContentType),
	})

	if err != nil {
		return fmt.Errorf("FPutObject: %w", err)
	}

	return nil
}

func (s *MinIOStorageService) Close() error {
	return nil
}

func (s *MinIOStorageService) DownloadFile(ctx context.Context, fileDownloadInfo FileToDownloadInfo) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	err := s.client.FGetObject(ctx, s.storageConfig.Bucket, fileDownloadInfo.ObjectName, fileDownloadInfo.DestinationPathName, minio.GetObjectOptions{})

	if err != nil {
		return fmt.Errorf("FGetObject: %w", err)
	}

	return nil
}

func (s *MinIOStorageService) IsObjectExists(ctx context.Context, objectName string) (bool, error) {
	_, err := s.client.StatObject(ctx, s.storageConfig.Bucket, objectName, minio.StatObjectOptions{})
	if err != nil {
		errResponse := minio.ToErrorResponse(err)
		if errResponse.Code == "NoSuchKey" {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

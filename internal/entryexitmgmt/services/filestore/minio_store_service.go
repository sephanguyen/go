package filestore

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

type MinIOStoreService struct {
	*BaseFileStoreService
	client *minio.Client
	conf   *configs.StorageConfig
}

func makeEndpointWithoutScheme(endpoint string) string {
	u, err := url.Parse(endpoint)
	if err != nil {
		return endpoint
	}
	return strings.TrimPrefix(endpoint, u.Scheme+"://")
}

func NewMinIOStoreService(conf *configs.StorageConfig) (*MinIOStoreService, error) {
	minIOEndpoint := makeEndpointWithoutScheme(conf.Endpoint)
	minIOClient, err := minio.New(minIOEndpoint, &minio.Options{
		Creds:     credentials.NewStaticV4(conf.AccessKey, conf.SecretKey, ""),
		Secure:    conf.Secure,
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: conf.InsecureSkipVerify}}, //nolint:gosec
	})

	if err != nil {
		return nil, err
	}

	return &MinIOStoreService{
		BaseFileStoreService: &BaseFileStoreService{
			conf: conf,
		},
		client: minIOClient,
		conf:   conf,
	}, nil
}

func (s *MinIOStoreService) UploadFromFile(ctx context.Context, objectName, pathName, contentType string) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	_, err := s.client.FPutObject(ctx, s.conf.Bucket, objectName, pathName, minio.PutObjectOptions{
		ContentType: contentType,
	})

	if err != nil {
		return fmt.Errorf("FPutObject: %w", err)
	}

	return nil
}

func (s *MinIOStoreService) Close() error {
	return nil
}

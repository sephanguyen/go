package storage

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/configs"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MinIOStorage struct {
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

func (s *MinIOStorage) UploadFromFile(ctx context.Context, reader io.Reader, fileID string, contentType string, fileSize int64) (downloadUrl string, err error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()
	if err != nil {
		err = status.Errorf(codes.Internal, "getting size from reader have err %v", err.Error())
		return
	}
	_, err = s.client.PutObject(ctx, s.conf.Bucket, fileID, reader, fileSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		err = status.Errorf(codes.Internal, "putting file to minio service have err %v", err)
		return
	}

	downloadUrl = fmt.Sprintf("%s/%s/%s", s.conf.Endpoint, s.conf.Bucket, fileID)
	return
}

func NewMinIOStoreService(conf *configs.StorageConfig) (*MinIOStorage, error) {
	minIOEndpoint := makeEndpointWithoutScheme(conf.Endpoint)
	minIOClient, err := minio.New(minIOEndpoint, &minio.Options{
		Creds:     credentials.NewStaticV4(conf.AccessKey, conf.SecretKey, ""),
		Secure:    conf.Secure,
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: conf.InsecureSkipVerify}}, //nolint:gosec
	})

	if err != nil {
		return nil, err
	}

	return &MinIOStorage{
		client: minIOClient,
		conf:   conf,
	}, nil
}

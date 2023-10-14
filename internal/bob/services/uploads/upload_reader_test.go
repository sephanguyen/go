package uploads_test

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/bob/services/filestore"
	"github.com/manabie-com/backend/internal/bob/services/uploads"
	"github.com/manabie-com/backend/internal/golibs/configs"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestGeneratePresignedPutObjectURL(t *testing.T) {
	t.Parallel()
	tx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockErr := fmt.Errorf("mock error")
	tcs := []struct {
		name            string
		fs              *filestore.Mock
		cfg             *configs.StorageConfig
		req             *bpb.PresignedPutObjectRequest
		expectedExpiry  time.Duration
		expectedBaseUrl string
		err             error
	}{
		{
			"happy case",
			&filestore.Mock{
				GeneratePresignedPutObjectMock: func(ctx context.Context, objectName string, expiry time.Duration) (*url.URL, error) {
					return url.Parse("https://example.com/manabie/" + objectName + "?key=1234567")
				},
				GeneratePublicObjectURLMock: func(objectName string) string {
					return "https://example.com/manabie/" + objectName
				},
			},
			&configs.StorageConfig{
				Endpoint:                 "https://example.com",
				MaximumURLExpiryDuration: time.Second * 10,
				MinimumURLExpiryDuration: time.Second * 1,
				DefaultURLExpiryDuration: time.Second * 5,
			},
			&bpb.PresignedPutObjectRequest{PrefixName: "car", Expiry: durationpb.New(time.Second * 2)},
			time.Second * 2,
			"https://example.com/manabie/car",
			nil,
		},
		{
			"expiry < min",
			&filestore.Mock{
				GeneratePresignedPutObjectMock: func(ctx context.Context, objectName string, expiry time.Duration) (*url.URL, error) {
					return url.Parse("https://example.com/manabie/" + objectName + "?key=1234567")
				},
				GeneratePublicObjectURLMock: func(objectName string) string {
					return "https://example.com/manabie/" + objectName
				},
			},
			&configs.StorageConfig{
				Endpoint:                 "https://example.com",
				MaximumURLExpiryDuration: time.Second * 10,
				MinimumURLExpiryDuration: time.Second * 1,
				DefaultURLExpiryDuration: time.Second * 5,
			},
			&bpb.PresignedPutObjectRequest{PrefixName: "car", Expiry: durationpb.New(time.Second * -1)},
			time.Second * 5,
			"https://example.com/manabie/car",
			nil,
		},
		{
			"expiry > max",
			&filestore.Mock{
				GeneratePresignedPutObjectMock: func(ctx context.Context, objectName string, expiry time.Duration) (*url.URL, error) {
					return url.Parse("https://example.com/manabie/" + objectName + "?key=1234567")
				},
				GeneratePublicObjectURLMock: func(objectName string) string {
					return "https://example.com/manabie/" + objectName
				},
			},
			&configs.StorageConfig{
				Endpoint:                 "https://example.com",
				MaximumURLExpiryDuration: time.Second * 10,
				MinimumURLExpiryDuration: time.Second * 1,
				DefaultURLExpiryDuration: time.Second * 5,
			},
			&bpb.PresignedPutObjectRequest{PrefixName: "car", Expiry: durationpb.New(time.Second * 20)},
			time.Second * 5,
			"https://example.com/manabie/car",
			nil,
		},
		{
			"give a error",
			&filestore.Mock{
				GeneratePresignedPutObjectMock: func(ctx context.Context, objectName string, expiry time.Duration) (*url.URL, error) {
					return nil, mockErr
				},
				GeneratePublicObjectURLMock: func(objectName string) string {
					return "https://example.com/manabie/" + objectName
				},
			},
			&configs.StorageConfig{
				Endpoint:                 "https://example.com",
				MaximumURLExpiryDuration: time.Second * 10,
				MinimumURLExpiryDuration: time.Second * 1,
				DefaultURLExpiryDuration: time.Second * 5,
			},
			&bpb.PresignedPutObjectRequest{PrefixName: "car", Expiry: durationpb.New(time.Second * 20)},
			time.Second * 0,
			"https://example.com/manabie/car",
			status.Error(codes.Internal, mockErr.Error()),
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(tt *testing.T) {
			s := &uploads.UploadReaderService{
				FileStore: tc.fs,
				Cfg:       *tc.cfg,
			}

			actual, err := s.GeneratePresignedPutObjectURL(tx, tc.req)
			if tc.err != nil {
				assert.Error(tt, err)
				assert.True(tt, errors.Is(err, tc.err))
			} else {
				assert.True(tt, strings.HasPrefix(actual.Name, "car"))
				assert.Equal(tt, tc.expectedExpiry, actual.Expiry.AsDuration(), tc.name)
				assert.True(tt, strings.HasPrefix(actual.PresignedUrl, "https://example.com/manabie/car"))
				// check token in presigned url
				u, err := url.Parse(actual.PresignedUrl)
				require.NoError(tt, err)
				assert.True(tt, len(u.RawQuery) != 0)

				// check random name
				urlWithoutParams := strings.Replace(actual.PresignedUrl, u.RawQuery, "", 0)
				assert.True(tt, len(urlWithoutParams) > len("https://example.com/manabie/car"))

				// check download url
				assert.True(tt, strings.HasPrefix(actual.DownloadUrl, "https://example.com/manabie/car"))

				// check token in download url
				u, err = url.Parse(actual.DownloadUrl)
				require.NoError(tt, err)
				assert.True(tt, len(u.RawQuery) == 0)
			}
		})
	}
}

func TestGenerateResumableUploadURL(t *testing.T) {
	t.Parallel()
	tx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	fileMock := &filestore.Mock{
		GenerateResumableObjectURLMock: func(ctx context.Context, objectName string, expiry time.Duration, allowOrigin, contentType string) (*url.URL, error) {
			return url.Parse("https://example.com/manabie/" + objectName + "?key=1234567" + "&contentType=" + contentType)
		},
		GeneratePublicObjectURLMock: func(objectName string) string {
			return "https://example.com/manabie/" + objectName
		},
	}

	mockErr := fmt.Errorf("mock error")
	tcs := []struct {
		name            string
		fs              *filestore.Mock
		cfg             *configs.StorageConfig
		req             *bpb.ResumableUploadURLRequest
		expectedExpiry  time.Duration
		expectedBaseUrl string
		err             error
	}{
		{
			"happy case",
			fileMock,
			&configs.StorageConfig{
				Endpoint:                 "https://example.com",
				MaximumURLExpiryDuration: time.Second * 10,
				MinimumURLExpiryDuration: time.Second * 1,
				DefaultURLExpiryDuration: time.Second * 5,
			},
			&bpb.ResumableUploadURLRequest{PrefixName: "car", Expiry: durationpb.New(time.Second * 2), AllowOrigin: "*", ContentType: ""},
			time.Second * 2,
			"https://example.com/manabie/car",
			nil,
		},
		{
			"happy case with content type",
			fileMock,
			&configs.StorageConfig{
				Endpoint:                 "https://example.com",
				MaximumURLExpiryDuration: time.Second * 10,
				MinimumURLExpiryDuration: time.Second * 1,
				DefaultURLExpiryDuration: time.Second * 5,
			},
			&bpb.ResumableUploadURLRequest{PrefixName: "car", Expiry: durationpb.New(time.Second * 2), AllowOrigin: "*", ContentType: "application/pdf"},
			time.Second * 2,
			"https://example.com/manabie/car",
			nil,
		},
		{
			"expiry < min",
			fileMock,
			&configs.StorageConfig{
				Endpoint:                 "https://example.com",
				MaximumURLExpiryDuration: time.Second * 10,
				MinimumURLExpiryDuration: time.Second * 1,
				DefaultURLExpiryDuration: time.Second * 5,
			},
			&bpb.ResumableUploadURLRequest{PrefixName: "car", Expiry: durationpb.New(time.Second * -1), AllowOrigin: "*", ContentType: ""},
			time.Second * 5,
			"https://example.com/manabie/car",
			nil,
		},
		{
			"expiry > max",
			fileMock,
			&configs.StorageConfig{
				Endpoint:                 "https://example.com",
				MaximumURLExpiryDuration: time.Second * 10,
				MinimumURLExpiryDuration: time.Second * 1,
				DefaultURLExpiryDuration: time.Second * 5,
			},
			&bpb.ResumableUploadURLRequest{PrefixName: "car", Expiry: durationpb.New(time.Second * 20), AllowOrigin: "*", ContentType: ""},
			time.Second * 5,
			"https://example.com/manabie/car",
			nil,
		},
		{
			"give a error",
			&filestore.Mock{
				GenerateResumableObjectURLMock: func(ctx context.Context, objectName string, expiry time.Duration, allowOrigin, contentType string) (*url.URL, error) {
					return nil, mockErr
				},
				GeneratePublicObjectURLMock: func(objectName string) string {
					return "https://example.com/manabie/" + objectName
				},
			},
			&configs.StorageConfig{
				Endpoint:                 "https://example.com",
				MaximumURLExpiryDuration: time.Second * 10,
				MinimumURLExpiryDuration: time.Second * 1,
				DefaultURLExpiryDuration: time.Second * 5,
			},
			&bpb.ResumableUploadURLRequest{PrefixName: "car", Expiry: durationpb.New(time.Second * 20), AllowOrigin: "*", ContentType: ""},
			time.Second * 0,
			"https://example.com/manabie/car",
			status.Error(codes.Internal, mockErr.Error()),
		},
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(tt *testing.T) {
			tt.Parallel()
			s := &uploads.UploadReaderService{
				FileStore: tc.fs,
				Cfg:       *tc.cfg,
			}

			actual, err := s.GenerateResumableUploadURL(tx, tc.req)
			if tc.err != nil {
				assert.Error(tt, err)
				assert.True(tt, errors.Is(err, tc.err))
			} else {
				assert.True(tt, strings.HasPrefix(actual.FileName, "car"))
				assert.Equal(tt, tc.expectedExpiry, actual.Expiry.AsDuration(), tc.name)
				assert.True(tt, strings.HasPrefix(actual.ResumableUploadUrl, "https://example.com/manabie/car"))
				// check token in presigned url
				u, err := url.Parse(actual.ResumableUploadUrl)
				require.NoError(tt, err)
				assert.True(tt, len(u.RawQuery) != 0)

				// check random name
				urlWithoutParams := strings.Replace(actual.ResumableUploadUrl, u.RawQuery, "", 0)
				assert.True(tt, len(urlWithoutParams) > len("https://example.com/manabie/car"))

				// check download url
				assert.True(tt, strings.HasPrefix(actual.DownloadUrl, "https://example.com/manabie/car"))

				// check token in download url
				u, err = url.Parse(actual.DownloadUrl)
				require.NoError(tt, err)
				assert.True(tt, len(u.RawQuery) == 0)

				//check content type
				assert.True(tt, strings.Contains(actual.ResumableUploadUrl, "contentType="+tc.req.ContentType), actual.ResumableUploadUrl)
			}
		})
	}
}

package filestore

import (
	"context"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/configs"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func TestMinIOStoreService(t *testing.T) {

	minIOEndpoint := makeEndpointWithoutScheme("http://example.com")
	minIOClient, err := minio.New(minIOEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4("test-access-key", "test-secret-key", ""),
		Secure: false,
	})

	if err != nil {
		panic(err)
	}

	s := MinIOStoreService{client: minIOClient, conf: &configs.StorageConfig{Bucket: "test-bucket"}}
	ctx := context.Background()

	testCases := []struct {
		name        string
		ctx         context.Context
		objectName  string
		objectPath  string
		contentType string
		hasError    bool
	}{
		{
			name:        "Invalid file path",
			ctx:         ctx,
			objectName:  "test-object.png",
			objectPath:  "/invalid-path@132",
			contentType: "image/png",
			hasError:    true,
		},
	}

	for _, tc := range testCases {

		err := s.UploadFromFile(tc.ctx, tc.objectName, tc.objectPath, tc.contentType)
		if err != nil {
			if !tc.hasError {
				t.Errorf("Expecting nil error got %v", err)
			}
			continue
		} else {
			if tc.hasError {
				t.Errorf("Expecting an error got nil")
			}
		}
	}

}

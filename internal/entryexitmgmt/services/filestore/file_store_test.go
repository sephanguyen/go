package filestore

import (
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/configs"
)

func TestGetDownloadURL(t *testing.T) {
	endpoint := "https://example.com"
	bucket := "backend"

	conf := &configs.StorageConfig{
		Endpoint: endpoint,
		Bucket:   bucket,
	}

	fs := &BaseFileStoreService{conf: conf}

	testCases := []struct {
		objectName string
		want       string
	}{
		{
			objectName: "test-file.png",
			want:       fmt.Sprintf("%s/%s/%s", endpoint, bucket, "test-file.png"),
		},
		{
			objectName: "upload-folder/test-file.png",
			want:       fmt.Sprintf("%s/%s/%s", endpoint, bucket, "upload-folder/test-file.png"),
		},
		{
			objectName: "upload-folder/my test file.png",
			want:       fmt.Sprintf("%s/%s/%s", endpoint, bucket, "upload-folder/my%20test%20file.png"),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.objectName, func(t *testing.T) {
			got := fs.GetDownloadURL(tc.objectName)
			if got != tc.want {
				t.Errorf("Expecting download URL %s got %s", tc.want, got)
			}
		})
	}

}

package filestorage

import (
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/configs"
)

func TestGetDownloadURL(t *testing.T) {
	endpoint := "https://example.com"
	bucket := "backend"

	storageConfig := &configs.StorageConfig{
		Endpoint: endpoint,
		Bucket:   bucket,
	}

	fs := &BaseFileStoreService{storageConfig: storageConfig}

	testCases := []struct {
		objectName string
		want       string
	}{
		{
			objectName: "my-csv-file.csv",
			want:       fmt.Sprintf("%s/%s/%s", endpoint, bucket, "my-csv-file.csv"),
		},
		{
			objectName: "upload-folder/my-csv-file.csv",
			want:       fmt.Sprintf("%s/%s/%s", endpoint, bucket, "upload-folder/my-csv-file.csv"),
		},
		{
			objectName: "upload-folder/はい.csv",
			want:       fmt.Sprintf("%s/%s/%s", endpoint, bucket, "upload-folder/%E3%81%AF%E3%81%84.csv"),
		},
		{
			objectName: "my-txt-file.txt",
			want:       fmt.Sprintf("%s/%s/%s", endpoint, bucket, "my-txt-file.txt"),
		},
		{
			objectName: "upload-folder/my-txt-file.txt",
			want:       fmt.Sprintf("%s/%s/%s", endpoint, bucket, "upload-folder/my-txt-file.txt"),
		},
		{
			objectName: "upload-folder/きれい.csv",
			want:       fmt.Sprintf("%s/%s/%s", endpoint, bucket, "upload-folder/%E3%81%8D%E3%82%8C%E3%81%84.csv"),
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

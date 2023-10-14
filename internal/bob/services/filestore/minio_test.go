package filestore

import (
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/configs"
)

func TestMinIOGeneratePublicObjectURL(t *testing.T) {
	endpoint := "https://example.com"
	bucket := "backend"

	minIO := &MinIO{
		conf: &configs.StorageConfig{
			Endpoint: endpoint,
			Bucket:   bucket,
		},
	}

	testCases := []struct {
		name string
		want string
	}{
		{
			name: "name-only.pdf",
			want: fmt.Sprintf("%s/%s/%s", endpoint, bucket, "name-only.pdf"),
		},
		{
			name: "2114002010_中1理科_いろいろな生物とその共通点_身のまわりの生物の観察_4月.pdf",
			want: fmt.Sprintf("%s/%s/%s", endpoint, bucket, "2114002010_%E4%B8%AD1%E7%90%86%E7%A7%91_%E3%81%84%E3%82%8D%E3%81%84%E3%82%8D%E3%81%AA%E7%94%9F%E7%89%A9%E3%81%A8%E3%81%9D%E3%81%AE%E5%85%B1%E9%80%9A%E7%82%B9_%E8%BA%AB%E3%81%AE%E3%81%BE%E3%82%8F%E3%82%8A%E3%81%AE%E7%94%9F%E7%89%A9%E3%81%AE%E8%A6%B3%E5%AF%9F_4%E6%9C%88.pdf"),
		},
		{
			name: "user-upload/test#01.pdf",
			want: fmt.Sprintf("%s/%s/%s", endpoint, bucket, "user-upload/test%2301.pdf"),
		},
		{
			name: "user-upload/test%v.pdf",
			want: fmt.Sprintf("%s/%s/%s", endpoint, bucket, "user-upload/test%25v.pdf"),
		},
		{
			name: "user-upload/test[1].pdf",
			want: fmt.Sprintf("%s/%s/%s", endpoint, bucket, "user-upload/test%5B1%5D.pdf"),
		},
		{
			name: "user-upload/space sep.pdf",
			want: fmt.Sprintf("%s/%s/%s", endpoint, bucket, "user-upload/space%20sep.pdf"),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := minIO.GeneratePublicObjectURL(tc.name)
			if got != tc.want {
				t.Errorf("minIO.GeneratePublicObjectURL(%q) = %q, expected %q", tc.name, got, tc.want)
			}
		})
	}
}

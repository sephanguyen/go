package filestore

import (
	"context"
	"testing"
)

func TestGcloudStoreService(t *testing.T) {

	s := GCloudStoreService{}

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

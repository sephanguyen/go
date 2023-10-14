package services

import (
	"context"
	"testing"

	ypb "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestBrightCoveService_CreateBrightCoveUploadUrl_InvalidArgument(t *testing.T) {
	t.Parallel()
	s := &BrightcoveService{}
	type testInput struct {
		request        *ypb.CreateBrightCoveUploadUrlRequest
		expectedCode   codes.Code
		expectedErrMsg string
	}

	testCases := map[string]testInput{
		"missing name": {
			request: &ypb.CreateBrightCoveUploadUrlRequest{
				Name: "",
			},
			expectedCode:   codes.InvalidArgument,
			expectedErrMsg: "rpc error: code = InvalidArgument desc = missing name",
		},
	}

	// test run
	for caseName, testCase := range testCases {
		caseName := caseName
		testCase := testCase
		t.Run(caseName, func(t *testing.T) {
			t.Parallel()
			resp, err := s.CreateBrightCoveUploadUrl(context.Background(), testCase.request)
			assert.Equal(t, testCase.expectedCode, status.Code(err), "%s - expecting InvalidArgument", caseName)
			assert.Equal(t, testCase.expectedErrMsg, err.Error(), "%s - expecting same error message", caseName)
			assert.Nil(t, resp, "%s - expecting nil response", caseName)
		})
	}
}

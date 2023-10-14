package services

import (
	"context"
	"testing"

	brightcove_service "github.com/manabie-com/backend/internal/golibs/brightcove"
	yasuo_service "github.com/manabie-com/backend/internal/yasuo/services"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCreateBrightCoveUploadUrl_InvalidArgument(t *testing.T) {
	t.Parallel()
	s := &CourseService{
		BrightCoveService: &yasuo_service.CourseService{
			BrightCoveProfile: "configurations.BrightCoveProfile",
			BrightcoveExtService: brightcove_service.NewBrightcoveService("configurations.BrightcoveClientID",
				"configurations.BrightcoveSecret",
				"configurations.BrightcoveAccountID",
				"configurations.BrightcovePolicyKey",
				"configurations.BrightcovePolicyKeyWithSearch",
				"configurations.BrightcoveProfile",
			),
		},
	}
	type testInput struct {
		request        *pb.CreateBrightCoveUploadUrlRequest
		expectedCode   codes.Code
		expectedErrMsg string
	}

	testCases := map[string]testInput{
		"missing name": {
			request: &pb.CreateBrightCoveUploadUrlRequest{
				Name: "",
			},
			expectedCode:   codes.InvalidArgument,
			expectedErrMsg: "rpc error: code = InvalidArgument desc = missing name",
		},
	}

	// test run
	for caseName, testCase := range testCases {
		t.Run(caseName, func(t *testing.T) {
			resp, err := s.CreateBrightCoveUploadUrl(context.Background(), testCase.request)
			assert.Equal(t, testCase.expectedCode, status.Code(err), "%s - expecting InvalidArgument", caseName)
			assert.Equal(t, testCase.expectedErrMsg, err.Error(), "%s - expecting same error message", caseName)
			assert.Nil(t, resp, "%s - expecting nil response", caseName)
		})
	}
}

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

func TestFinishUploadBrightCove_InvalidArgument(t *testing.T) {
	t.Parallel()
	s := &CourseService{
		BrightCoveService: &yasuo_service.CourseService{
			BrightCoveProfile: "configurations.BrightCoveProfile",
			BrightcoveExtService: brightcove_service.NewBrightcoveService(
				"configurations.BrightcoveClientID",
				"configurations.BrightcoveSecret",
				"configurations.BrightcoveAccountID",
				"configurations.BrightCovePolicyKey",
				"configurations.BrightCovePolicyKeyWithSearch",
				"configurations.BrightCoveProfile",
			),
		},
	}
	type testInput struct {
		request        *pb.FinishUploadBrightCoveRequest
		expectedCode   codes.Code
		expectedErrMsg string
	}

	testCases := map[string]testInput{
		"missing apiRequestUrl": {
			request: &pb.FinishUploadBrightCoveRequest{
				ApiRequestUrl: "",
				VideoId:       "some_id",
			},
			expectedCode:   codes.InvalidArgument,
			expectedErrMsg: "rpc error: code = InvalidArgument desc = missing apiRequestUrl",
		},
		"missing videoId": {
			request: &pb.FinishUploadBrightCoveRequest{
				ApiRequestUrl: "api-request-url",
				VideoId:       "",
			},
			expectedCode:   codes.InvalidArgument,
			expectedErrMsg: "rpc error: code = InvalidArgument desc = missing videoId",
		},
	}

	// test run
	for caseName, testCase := range testCases {
		caseName := caseName
		testCase := testCase
		t.Run(caseName, func(t *testing.T) {
			t.Parallel()
			resp, err := s.FinishUploadBrightCove(context.Background(), testCase.request)
			assert.Equal(t, testCase.expectedCode, status.Code(err), "%s - expecting InvalidArgument", caseName)
			assert.Equal(t, testCase.expectedErrMsg, err.Error(), "%s - expecting same error message", caseName)
			assert.Nil(t, resp, "%s - expecting nil response", caseName)
		})
	}
}

package services

import (
	"context"
	"testing"

	bobpb "github.com/manabie-com/backend/pkg/genproto/bob"
	pb "github.com/manabie-com/backend/pkg/genproto/yasuo"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func UpdateSchoolConfig_InvalidArgument(t *testing.T) {
	schoolService := &SchoolService{}
	type testInput struct {
		request        *pb.UpdateSchoolConfigRequest
		expectedCode   codes.Code
		expectedErrMsg string
	}
	testCases := map[string]testInput{
		"cannot find school config": {
			request: &pb.UpdateSchoolConfigRequest{
				SchoolConfig: &pb.SchoolConfig{
					SchoolId:     1,
					PlanId:       "School",
					Country:      bobpb.COUNTRY_VN,
					PlanDuration: 30,
					Privileges:   []pb.PlanPrivilege{pb.CAN_ACCESS_ALL_LOS, pb.CAN_ACCESS_LEARNING_TOPICS, pb.CAN_ACCESS_MOCK_EXAMS},
				},
			},
			expectedCode:   codes.InvalidArgument,
			expectedErrMsg: "rpc error: code = InvalidArgument desc = cannot find school config",
		},
		"missing school id": {
			request: &pb.UpdateSchoolConfigRequest{
				SchoolConfig: &pb.SchoolConfig{
					SchoolId:     0,
					PlanId:       "School",
					Country:      bobpb.COUNTRY_VN,
					PlanDuration: 30,
					Privileges:   []pb.PlanPrivilege{pb.CAN_ACCESS_ALL_LOS, pb.CAN_ACCESS_LEARNING_TOPICS, pb.CAN_ACCESS_MOCK_EXAMS},
				},
			},
			expectedCode:   codes.InvalidArgument,
			expectedErrMsg: "rpc error: code = InvalidArgument desc = missing school id",
		},
		"missing plan id": {
			request: &pb.UpdateSchoolConfigRequest{
				SchoolConfig: &pb.SchoolConfig{
					SchoolId:     1,
					PlanId:       "",
					Country:      bobpb.COUNTRY_VN,
					PlanDuration: 30,
					Privileges:   []pb.PlanPrivilege{pb.CAN_ACCESS_ALL_LOS, pb.CAN_ACCESS_LEARNING_TOPICS, pb.CAN_ACCESS_MOCK_EXAMS},
				},
			},
			expectedCode:   codes.InvalidArgument,
			expectedErrMsg: "rpc error: code = InvalidArgument desc = missing plan id",
		},
		"missing country": {
			request: &pb.UpdateSchoolConfigRequest{
				SchoolConfig: &pb.SchoolConfig{
					SchoolId:     1,
					PlanId:       "School",
					Country:      bobpb.COUNTRY_NONE,
					PlanDuration: 30,
					Privileges:   []pb.PlanPrivilege{pb.CAN_ACCESS_ALL_LOS, pb.CAN_ACCESS_LEARNING_TOPICS, pb.CAN_ACCESS_MOCK_EXAMS},
				},
			},
			expectedCode:   codes.InvalidArgument,
			expectedErrMsg: "rpc error: code = InvalidArgument desc = missing country",
		},
		"missing plan time": {
			request: &pb.UpdateSchoolConfigRequest{
				SchoolConfig: &pb.SchoolConfig{
					SchoolId:     1,
					PlanId:       "School",
					Country:      bobpb.COUNTRY_VN,
					PlanDuration: 0,
					Privileges:   []pb.PlanPrivilege{pb.CAN_ACCESS_ALL_LOS, pb.CAN_ACCESS_LEARNING_TOPICS, pb.CAN_ACCESS_MOCK_EXAMS},
				},
			},
			expectedCode:   codes.InvalidArgument,
			expectedErrMsg: "rpc error: code = InvalidArgument desc = missing plan time",
		},
	}

	// test run
	for caseName, testCase := range testCases {
		t.Run(caseName, func(t *testing.T) {
			resp, err := schoolService.UpdateSchoolConfig(context.Background(), testCase.request)
			assert.Equal(t, testCase.expectedCode, status.Code(err), "%s - expecting InvalidArgument", caseName)
			assert.Equal(t, testCase.expectedErrMsg, err.Error(), "%s - expecting same error message", caseName)
			assert.Nil(t, resp, "%s - expecting nil response", caseName)
		})
	}
}

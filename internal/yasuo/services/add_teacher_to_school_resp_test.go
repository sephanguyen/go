package services

import (
	"context"
	"testing"

	pb "github.com/manabie-com/backend/pkg/genproto/yasuo"
	"github.com/stretchr/testify/assert"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func AddTeacher_InvalidArgument(t *testing.T) {
	ctx := context.Background()
	schoolService := &SchoolService{}

	expectedCode := codes.InvalidArgument
	testCases := map[string]TestCase{
		"missing school id": {
			ctx: ctx,
			req: &pb.AddTeacherRequest{
				SchoolId:  0,
				TeacherId: "teacher-id",
			},
			expectedCode:   codes.InvalidArgument,
			expectedErrMsg: "missing school id",
		},
		"missing teacher id": {
			ctx: ctx,
			req: &pb.AddTeacherRequest{
				SchoolId:  1,
				TeacherId: "teacher-id",
			},
			expectedCode:   codes.InvalidArgument,
			expectedErrMsg: "missing teacher id",
		},
		"this is not a teacher": {
			ctx: ctx,
			req: &pb.AddTeacherRequest{
				SchoolId:  1,
				TeacherId: "student-id",
			},
			expectedCode:   codes.InvalidArgument,
			expectedErrMsg: "this is not a teacher",
		},
		"the teacher already is part of this school": {
			ctx: ctx,
			req: &pb.AddTeacherRequest{
				SchoolId:  2,
				TeacherId: "teacher-id",
			},
			expectedCode:   codes.InvalidArgument,
			expectedErrMsg: "the teacher already is part of this school",
		},
	}

	for caseName, testCase := range testCases {
		t.Run(caseName, func(t *testing.T) {
			rsp, err := schoolService.AddTeacher(context.Background(), testCase.req.(*pb.AddTeacherRequest))
			assert.Equal(t, testCase.expectedCode, status.Code(err), "%s - expecting InvalidArgument", caseName)
			assert.Equal(t, expectedCode, status.Code(err))
			assert.Nil(t, rsp, "expecting nil response")
		})
	}
}

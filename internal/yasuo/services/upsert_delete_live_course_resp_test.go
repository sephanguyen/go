package services

import (
	"context"
	"testing"
	"time"

	"github.com/gogo/protobuf/types"
	bobpb "github.com/manabie-com/backend/pkg/genproto/bob"
	pb "github.com/manabie-com/backend/pkg/genproto/yasuo"
	"github.com/stretchr/testify/assert"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func UpsertLiveCourse_InvalidArgument(t *testing.T) {
	ctx := context.Background()
	courseService := &CourseService{}

	expectedCode := codes.InvalidArgument
	testCases := map[string]TestCase{
		"missing school id": {
			ctx: ctx,
			req: &pb.UpsertLiveCourseRequest{
				Id:         "",
				Name:       "live-course-name",
				Grade:      "G12",
				Subject:    bobpb.SUBJECT_BIOLOGY,
				ClassIds:   []int32{1},
				TeacherIds: []string{"teacher-id"},
				SchoolId:   0,
				Country:    bobpb.COUNTRY_SG,
				StartDate:  &types.Timestamp{Seconds: time.Now().Unix()},
				EndDate:    &types.Timestamp{Seconds: time.Now().Add(time.Hour * 24 * 30).Unix()},
			},
			expectedCode:   codes.InvalidArgument,
			expectedErrMsg: "school id cannot be empty",
		},
		"missing country": {
			ctx: ctx,
			req: &pb.UpsertLiveCourseRequest{
				Id:         "",
				Name:       "live-course-name",
				Grade:      "G12",
				Subject:    bobpb.SUBJECT_BIOLOGY,
				ClassIds:   []int32{1},
				TeacherIds: []string{"teacher-id"},
				SchoolId:   1,
				Country:    bobpb.COUNTRY_NONE,
				StartDate:  &types.Timestamp{Seconds: time.Now().Unix()},
				EndDate:    &types.Timestamp{Seconds: time.Now().Add(time.Hour * 24 * 30).Unix()},
			},
			expectedCode:   codes.InvalidArgument,
			expectedErrMsg: "country cannot be empty",
		},
		"missing name": {
			ctx: ctx,
			req: &pb.UpsertLiveCourseRequest{
				Id:         "",
				Name:       "",
				Grade:      "G12",
				Subject:    bobpb.SUBJECT_BIOLOGY,
				ClassIds:   []int32{1},
				TeacherIds: []string{"teacher-id"},
				SchoolId:   1,
				Country:    bobpb.COUNTRY_SG,
				StartDate:  &types.Timestamp{Seconds: time.Now().Unix()},
				EndDate:    &types.Timestamp{Seconds: time.Now().Add(time.Hour * 24 * 30).Unix()},
			},
			expectedCode:   codes.InvalidArgument,
			expectedErrMsg: "name cannot be empty",
		},
		"missing grade": {
			ctx: ctx,
			req: &pb.UpsertLiveCourseRequest{
				Id:         "",
				Name:       "live-course-name",
				Grade:      "",
				Subject:    bobpb.SUBJECT_BIOLOGY,
				ClassIds:   []int32{1},
				TeacherIds: []string{"teacher-id"},
				SchoolId:   1,
				Country:    bobpb.COUNTRY_SG,
				StartDate:  &types.Timestamp{Seconds: time.Now().Unix()},
				EndDate:    &types.Timestamp{Seconds: time.Now().Add(time.Hour * 24 * 30).Unix()},
			},
			expectedCode:   codes.InvalidArgument,
			expectedErrMsg: "grade cannot be empty",
		},
		"missing subject": {
			ctx: ctx,
			req: &pb.UpsertLiveCourseRequest{
				Id:         "",
				Name:       "live-course-name",
				Grade:      "G12",
				Subject:    bobpb.SUBJECT_NONE,
				ClassIds:   []int32{1},
				TeacherIds: []string{"teacher-id"},
				SchoolId:   1,
				Country:    bobpb.COUNTRY_SG,
				StartDate:  &types.Timestamp{Seconds: time.Now().Unix()},
				EndDate:    &types.Timestamp{Seconds: time.Now().Add(time.Hour * 24 * 30).Unix()},
			},
			expectedCode:   codes.InvalidArgument,
			expectedErrMsg: "subject cannot be empty",
		},
		"wrong start and end date": {
			ctx: ctx,
			req: &pb.UpsertLiveCourseRequest{
				Id:         "",
				Name:       "live-course-name",
				Grade:      "G12",
				Subject:    bobpb.SUBJECT_BIOLOGY,
				ClassIds:   []int32{1},
				TeacherIds: []string{"teacher-id"},
				SchoolId:   1,
				Country:    bobpb.COUNTRY_SG,
				EndDate:    &types.Timestamp{Seconds: time.Now().Unix()},
				StartDate:  &types.Timestamp{Seconds: time.Now().Add(time.Hour * 24 * 30).Unix()},
			},
			expectedCode:   codes.InvalidArgument,
			expectedErrMsg: "start date must before end date",
		},
		"cannot add teacher of another school": {
			ctx: ctx,
			req: &pb.UpsertLiveCourseRequest{
				Id:         "",
				Name:       "live-course-name",
				Grade:      "G12",
				Subject:    bobpb.SUBJECT_BIOLOGY,
				ClassIds:   []int32{1},
				TeacherIds: []string{"teacher-id-of-another-school"},
				SchoolId:   1,
				Country:    bobpb.COUNTRY_SG,
				StartDate:  &types.Timestamp{Seconds: time.Now().Unix()},
				EndDate:    &types.Timestamp{Seconds: time.Now().Add(time.Hour * 24 * 30).Unix()},
			},
			expectedCode:   codes.InvalidArgument,
			expectedErrMsg: "cannot add teacher of another school",
		},
		"cannot add class of another school": {
			ctx: ctx,
			req: &pb.UpsertLiveCourseRequest{
				Id:         "",
				Name:       "live-course-name",
				Grade:      "G12",
				Subject:    bobpb.SUBJECT_BIOLOGY,
				ClassIds:   []int32{100}, //class id of another school
				TeacherIds: []string{"teacher-id"},
				SchoolId:   1,
				Country:    bobpb.COUNTRY_SG,
				StartDate:  &types.Timestamp{Seconds: time.Now().Unix()},
				EndDate:    &types.Timestamp{Seconds: time.Now().Add(time.Hour * 24 * 30).Unix()},
			},
			expectedCode:   codes.InvalidArgument,
			expectedErrMsg: "cannot add class of another school",
		},
	}

	for caseName, testCase := range testCases {
		t.Run(caseName, func(t *testing.T) {
			rsp, err := courseService.UpsertLiveCourse(context.Background(), testCase.req.(*pb.UpsertLiveCourseRequest))
			assert.Equal(t, testCase.expectedCode, status.Code(err), "%s - expecting InvalidArgument", caseName)
			assert.Equal(t, expectedCode, status.Code(err))
			assert.Nil(t, rsp, "expecting nil response")
		})
	}
}

func DeleteLiveCourse_InvalidArgument(t *testing.T) {
	ctx := context.Background()
	courseService := &CourseService{}

	expectedCode := codes.InvalidArgument
	testCases := map[string]TestCase{
		"missing course ids": {
			ctx: ctx,
			req: &pb.DeleteLiveCourseRequest{
				CourseIds: []string{},
			},
			expectedCode:   codes.InvalidArgument,
			expectedErrMsg: "course ids cannot empty",
		},
		"cannot find course": {
			ctx: ctx,
			req: &pb.DeleteLiveCourseRequest{
				CourseIds: []string{"not-found-course"},
			},
			expectedCode:   codes.InvalidArgument,
			expectedErrMsg: "cannot find course",
		},
	}

	for caseName, testCase := range testCases {
		t.Run(caseName, func(t *testing.T) {
			rsp, err := courseService.DeleteLiveCourse(context.Background(), testCase.req.(*pb.DeleteLiveCourseRequest))
			assert.Equal(t, testCase.expectedCode, status.Code(err), "%s - expecting InvalidArgument", caseName)
			assert.Equal(t, expectedCode, status.Code(err))
			assert.Nil(t, rsp, "expecting nil response")
		})
	}
}

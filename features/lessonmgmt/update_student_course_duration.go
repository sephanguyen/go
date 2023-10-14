package lessonmgmt

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/helper"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Suite) updateStudentCourseDuration(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := stepState.Request1.(*upb.UpsertStudentCoursePackageRequest)
	req.StudentPackageProfiles[0].StartTime = timestamppb.New(time.Now().AddDate(0, 0, 1))
	_, err := upb.NewUserModifierServiceClient(s.UserMgmtConn).UpsertStudentCoursePackage(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.ResponseErr = err
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) addCourseToStudent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	courseID := stepState.CourseIDs[0]
	locationId := stepState.CenterIDs[len(stepState.CenterIDs)-1]
	studentID := stepState.StudentIds[0]
	now := time.Now()
	profiles := []*upb.UpsertStudentCoursePackageRequest_StudentPackageProfile{
		{
			Id: &upb.UpsertStudentCoursePackageRequest_StudentPackageProfile_CourseId{
				CourseId: courseID,
			},
			StartTime: timestamppb.New(now.AddDate(0, 0, -1)),
			EndTime:   timestamppb.New(now.Add(24 * 7 * time.Hour)),
			StudentPackageExtra: []*upb.StudentPackageExtra{
				{
					LocationId: locationId,
				},
			},
		},
	}
	req := &upb.UpsertStudentCoursePackageRequest{
		StudentId:              studentID,
		StudentPackageProfiles: profiles,
	}
	stepState.StudentIDWithCourseID = []string{studentID, courseID}
	_, err := upb.NewUserModifierServiceClient(s.UserMgmtConn).
		UpsertStudentCoursePackage(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.Request1 = req
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) checkInactiveStudent(ctx context.Context) (context.Context, error) {
	time.Sleep(500 * time.Millisecond)
	stepState := StepStateFromContext(ctx)
	studentID := stepState.StudentIds[0]
	stmt := "SELECT deleted_at from lesson_members WHERE lesson_id = $1 and user_id = $2"
	var deletedAt pgtype.Timestamptz
	err := s.BobDBTrace.DB.QueryRow(ctx, stmt, stepState.CurrentLessonID, studentID).Scan(&deletedAt)
	if err != nil {
		return nil, err
	}
	if deletedAt.Status == pgtype.Null {
		return StepStateToContext(ctx, stepState), fmt.Errorf("lesson member is not correct")
	}
	return StepStateToContext(ctx, stepState), nil
}

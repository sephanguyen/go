package mastermgmt

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/try"
	fpb "github.com/manabie-com/backend/pkg/manabuf/fatima/v1"

	"github.com/segmentio/ksuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) userAddCourseWithStudentPackageExtraForAStudent(ctx context.Context) (context.Context, error) {
	time.Sleep(5 * time.Second)
	stepState := StepStateFromContext(ctx)
	now := time.Now()
	startAt := timestamppb.Now()
	stepState.StudentID = ksuid.New().String()
	endAt := timestamppb.New(now.Add(7 * 24 * time.Hour))
	studentPackageExtras := []*fpb.AddStudentPackageCourseRequest_AddStudentPackageExtra{
		{
			CourseId:   stepState.CourseIDs[0],
			LocationId: stepState.CenterIDs[0],
			ClassId:    stepState.CurrentClassId,
		},
	}

	req := &fpb.AddStudentPackageCourseRequest{
		CourseIds:           []string{stepState.CourseIDs[0]},
		StudentId:           stepState.StudentID,
		StartAt:             startAt,
		EndAt:               endAt,
		StudentPackageExtra: studentPackageExtras,
	}
	stepState.Request = req
	stepState.Response, stepState.ResponseErr = fpb.NewSubscriptionModifierServiceClient(s.FatimaConn).AddStudentPackageCourse(contextWithToken(s, ctx), req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) classMemberStoredInDB(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var count int
	query := "SELECT count(*) FROM class_member WHERE class_id = $1 AND user_id = $2 AND deleted_at is null and start_date = $3::timestamptz and end_date = $4::timestamptz"
	startDate, endDate := database.Timestamptz(stepState.Request.(*fpb.AddStudentPackageCourseRequest).StartAt.AsTime()), database.Timestamptz(stepState.Request.(*fpb.AddStudentPackageCourseRequest).EndAt.AsTime())
	err := try.Do(func(attempt int) (bool, error) {
		err := s.BobDBTrace.QueryRow(ctx, query, stepState.CurrentClassId, stepState.StudentID, startDate, endDate).Scan(&count)
		if err == nil && count > 0 {
			return false, nil
		}
		retry := attempt < 5
		if retry {
			time.Sleep(10 * time.Second)
			return true, fmt.Errorf("error querying count student subscriptions: %w", err)
		}
		return false, err
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if count == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expect have class_member, but got empty")
	}
	return StepStateToContext(ctx, stepState), nil
}

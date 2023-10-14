package lessonmgmt

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/try"
	ppb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const (
	SleepDuration = 15 * time.Second
	RetryCount    = 5

	StudentSlotInfoCountQuery = `SELECT count(*) 
		FROM lesson_student_subscriptions lss
		INNER JOIN lesson_student_subscription_access_path lssap
		ON lss.student_subscription_id = lssap.student_subscription_id
		WHERE lss.course_id = ANY($1) 
		AND lss.student_id = ANY($2) 
		AND lssap.location_id = ANY($3)
		AND lss.package_type IS NOT NULL	
		AND lss.deleted_at IS NULL
		AND lssap.deleted_at IS NULL`
)

func (s *Suite) aMessageIsPublishedToStudentCourseEventSync(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	now := time.Now()

	studentCourseSync := []*ppb.EventSyncStudentPackageCourse{
		{
			StudentId:        stepState.StudentIds[0],
			LocationId:       stepState.LocationIDs[0],
			CourseId:         stepState.CourseIDs[0],
			StudentPackageId: "bdd-student-package-id-1",
			StudentStartDate: timestamppb.New(now),
			StudentEndDate:   timestamppb.New(now.Add(5 * 24 * time.Hour)),
			CourseSlot:       wrapperspb.Int32(21),
			PackageType:      ppb.PackageType_PACKAGE_TYPE_SLOT_BASED,
		},
		{
			StudentId:         stepState.StudentIds[1],
			LocationId:        stepState.LocationIDs[0],
			CourseId:          stepState.CourseIDs[0],
			StudentPackageId:  "bdd-student-package-id-2",
			StudentStartDate:  timestamppb.New(now),
			StudentEndDate:    timestamppb.New(now.Add(28 * 24 * time.Hour)),
			CourseSlotPerWeek: wrapperspb.Int32(2),
			PackageType:       ppb.PackageType_PACKAGE_TYPE_FREQUENCY,
		},
	}
	stepState.StudentSlotInfoCount = len(studentCourseSync)

	return s.publishToStudentCourseEventSync(StepStateToContext(ctx, stepState), studentCourseSync)
}

func (s *Suite) publishToStudentCourseEventSync(ctx context.Context, studentCourseSync []*ppb.EventSyncStudentPackageCourse) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	data, err := json.Marshal(studentCourseSync)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to marshal data: %w", err)
	}

	msgID, err := s.JSM.PublishAsyncContext(ctx, constants.SubjectStudentCourseEventSync, data)
	if err != nil {
		return StepStateToContext(ctx, stepState),
			nats.HandlePushMsgFail(ctx, fmt.Errorf("SubjectStudentCourseEventSync JSM.PublishAsyncContext failed, msgID: %s, %w", msgID, err))
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) receiveStudentCourseSlotInfoSuccessfully(ctx context.Context) (context.Context, error) {
	// sleep to make sure NATS sync data successfully
	time.Sleep(SleepDuration)

	stepState := StepStateFromContext(ctx)
	expectedCount := stepState.StudentSlotInfoCount
	actualCount := 0

	err := try.Do(func(attempt int) (bool, error) {
		err := s.BobDB.QueryRow(ctx, StudentSlotInfoCountQuery, stepState.CourseIDs, stepState.StudentIds, stepState.LocationIDs).Scan(&actualCount)
		if err == nil && actualCount > 0 {
			return false, nil
		}

		retry := attempt < RetryCount
		if retry {
			time.Sleep(SleepDuration)
			return true, fmt.Errorf("error querying count student course slots: %w", err)
		}
		return false, fmt.Errorf("error querying count student course slots: %w", err)
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if actualCount != expectedCount {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected record count does not match\n expected: %d, actual: %d", expectedCount, actualCount)
	}

	return StepStateToContext(ctx, stepState), nil
}

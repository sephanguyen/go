package eureka

import (
	"context"
	"fmt"
	"time"

	bob_entities "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/try"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) enrollToTheCourse(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = contextWithToken(s, ctx)
	stmt := `SELECT email FROM users WHERE user_id = $1`
	var studentEmail string
	err := s.BobDB.QueryRow(ctx, stmt, stepState.StudentID).Scan(&studentEmail)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	locationID := idutil.ULIDNow()
	e := &bob_entities.Location{}
	database.AllNullEntity(e)
	if err := multierr.Combine(
		e.LocationID.Set(locationID),
		e.Name.Set(fmt.Sprintf("location-%s", locationID)),
		e.IsArchived.Set(false),
		e.CreatedAt.Set(time.Now()),
		e.UpdatedAt.Set(time.Now()),
	); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if _, err := database.Insert(ctx, e, s.BobDB.Exec); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	_, err = upb.NewUserModifierServiceClient(s.UsermgmtConn).UpdateStudent(
		ctx,
		&upb.UpdateStudentRequest{
			StudentProfile: &upb.UpdateStudentRequest_StudentProfile{
				Id:               stepState.StudentID,
				Name:             "test-name",
				Grade:            5,
				EnrollmentStatus: upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
				Email:            studentEmail,
				LocationIds:      []string{locationID},
			},

			SchoolId: stepState.SchoolIDInt,
		},
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to update student: %w", err)
	}

	if _, err := upb.NewUserModifierServiceClient(s.UsermgmtConn).UpsertStudentCoursePackage(ctx, &upb.UpsertStudentCoursePackageRequest{
		StudentId: stepState.StudentID,
		StudentPackageProfiles: []*upb.UpsertStudentCoursePackageRequest_StudentPackageProfile{{
			Id: &upb.UpsertStudentCoursePackageRequest_StudentPackageProfile_CourseId{
				CourseId: stepState.CourseID,
			},
			StartTime: timestamppb.New(time.Now().Add(time.Hour * -20)),
			EndTime:   timestamppb.New(time.Now().Add(time.Hour * 24 * 10)),
		}},
	}); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to update student course package: %w", err)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createStudyPlanFromTheBook(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.AuthToken = stepState.SchoolAdminToken
	ctx = contextWithToken(s, ctx)
	req := &epb.UpsertStudyPlanRequest{
		Name:                fmt.Sprintf("studyplan-%s", stepState.StudyPlanID),
		SchoolId:            constants.ManabieSchool,
		TrackSchoolProgress: true,
		Grades:              []int32{3, 4},
		Status:              epb.StudyPlanStatus_STUDY_PLAN_STATUS_ACTIVE,
		BookId:              stepState.BookID,
		CourseId:            stepState.CourseID,
	}

	resp, err := epb.NewStudyPlanModifierServiceClient(s.Conn).UpsertStudyPlan(ctx, req)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to upsert study plan: %w", err)
	}
	stepState.Request = req
	stepState.StudyPlanID = resp.StudyPlanId
	var counter pgtype.Int8

	err = try.Do(func(attempt int) (bool, error) {
		stmt := `SELECT count(*) FROM study_plan_items WHERE content_structure->>'lo_id' = ANY($1) and deleted_at IS NOT NULL`

		err = s.DB.QueryRow(ctx, stmt, database.TextArray(stepState.LoIDs)).Scan(&counter)
		if err != nil {
			return false, fmt.Errorf("unable to get data from study_plan_items: %w", err)
		}
		if counter.Int == 0 {
			time.Sleep(700 * time.Millisecond)
			return attempt < 10, fmt.Errorf("unable to get data from study_plan_items: %w", err)
		}
		return false, nil
	})
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) deleteLoStudyPlanItem(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	reqCtx := context.Background()
	if arg1 == "valid" {
		reqCtx = contextWithToken(s, ctx)
	}
	req := &epb.DeleteLOStudyPlanItemsRequest{
		LoIds: stepState.LoIDs,
	}
	_, err := epb.NewInternalModifierServiceClient(s.Conn).DeleteLOStudyPlanItems(reqCtx, req)

	stepState.ResponseErr = err
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) ourSystemHasToDeleteLoStudyPlanItemsCorrectly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	courseStudentStmt := `SELECT count(*) FROM course_students WHERE course_id = $1 and deleted_at IS NULL`
	var courseStudentCounter pgtype.Int8
	err := s.DB.QueryRow(ctx, courseStudentStmt, database.Text(stepState.CourseID)).Scan(&courseStudentCounter)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to get course_students: %w", err)
	}

	var counter pgtype.Int8
	err = try.Do(func(attempt int) (bool, error) {
		stmt := `SELECT count(*) FROM study_plan_items WHERE content_structure->>'lo_id' = ANY($1) and deleted_at IS NOT NULL`

		err = s.DB.QueryRow(ctx, stmt, database.TextArray(stepState.LoIDs)).Scan(&counter)
		if err != nil {
			return false, fmt.Errorf("unable to get data from study_plan_items: %w", err)
		}
		if counter.Int == 0 {
			time.Sleep(500 * time.Millisecond)
			return attempt < 8, fmt.Errorf("unable to get data from study_plan_items: %w", err)
		}
		return false, nil
	})

	if int(counter.Int) != len(stepState.LoIDs)*(int(courseStudentCounter.Int)+1) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("wrong delete lo study plan items")
	}
	return StepStateToContext(ctx, stepState), nil
}

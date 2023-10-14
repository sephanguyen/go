package eureka

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/try"
	common "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
)

func (s *suite) someStudentsJoinInAClass(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.ClassID = idutil.ULIDNow()
	for _, id := range stepState.StudentIDs {
		if ctx, err := s.upsertClassStudent(ctx, stepState.ClassID, id); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("classStudentRepo.Upsert: %w", err)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) someStudentsSubmitTheirAssignments(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if len(stepState.StudentIDs) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("no student exist")
	}
	if len(stepState.StudentIDs) == 1 {
		stepState.NumOfCompletedStudent = 1
	} else if len(stepState.StudentIDs) > 1 {
		if n := rand.Intn(len(stepState.StudentIDs)-1) + 1; n > 0 {
			stepState.NumOfCompletedStudent = n // else = 0
		}
	}
	for i := 0; i < stepState.NumOfCompletedStudent; i++ {
		if ctx, err := s.givenStudentSubmitTheirAssignmentInCurrentStudyPlanItem(ctx, stepState.StudentIDs[i], "existed", "single"); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unable to submit assignment: %w", err)
		}
	}
	cmd := `SELECT COUNT(*) FROM study_plan_items WHERE copy_study_plan_item_id = $1 AND completed_at IS NOT NULL`
	if err := try.Do(func(attempt int) (retry bool, err error) {
		var counter int
		if err2 := s.DB.QueryRow(ctx, cmd, database.Text(stepState.AssignmentStudyPlanItemID)).Scan(&counter); err2 != nil {
			return false, err2
		}
		if counter < 1 {
			time.Sleep(3 * time.Second)
			return attempt < 30, fmt.Errorf("unable to submit assignment: timeout")
		}
		return false, nil
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) upsertClassStudent(ctx context.Context, classID, studentID string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	classStudentRepo := &repositories.ClassStudentRepo{}

	e := &entities.ClassStudent{
		ClassID:   database.Text(classID),
		StudentID: database.Text(studentID),
	}
	e.BaseEntity.Now()
	return StepStateToContext(ctx, stepState), classStudentRepo.Upsert(ctx, s.DB, e)
}

func (s *suite) theTeacherRetrieveStatisticAssignmentClass(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Response, stepState.ResponseErr = epb.NewAssignmentReaderServiceClient(s.Conn).RetrieveStatisticAssignmentClass(contextWithToken(s, ctx), &epb.RetrieveStatisticAssignmentClassRequest{
		ClassId:         stepState.ClassID,
		StudyPlanItemId: stepState.AssignmentStudyPlanItemID,
	})

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) ourSystemHaveToReturnStatisticAssignmentClassCorrectly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp := stepState.Response.(*epb.RetrieveStatisticAssignmentClassResponse)
	if resp.StatisticItem.CompletedStudent != int32(stepState.NumOfCompletedStudent) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected number of completed student, want: %d, actual: %d", stepState.NumOfCompletedStudent, resp.StatisticItem.CompletedStudent)
	}
	if int(resp.StatisticItem.TotalAssignedStudent) != len(stepState.StudentIDs) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected number of total assigned student, want: %d, actual: %d", len(stepState.StudentIDs), resp.StatisticItem.TotalAssignedStudent)
	}

	if resp.StatisticItem.Item.StudyPlanItem.StudyPlanItemId != stepState.AssignmentStudyPlanItemID {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected study plan item, want %s, actual %s", stepState.AssignmentStudyPlanItemID, resp.StatisticItem.Item.StudyPlanItem.StudyPlanItemId)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) getAllStudentAssignedAssignments(ctx context.Context) (context.Context, []*epb.Content, error) {
	stepState := StepStateFromContext(ctx)

	resp, err := epb.NewAssignmentReaderServiceClient(s.Conn).
		ListStudentAvailableContents(contextWithToken(s, ctx), &epb.ListStudentAvailableContentsRequest{
			StudyPlanId: []string{},
		})

	if err != nil {
		return StepStateToContext(ctx, stepState), nil, err
	}

	if len(resp.Contents) == 0 {
		return StepStateToContext(ctx, stepState), nil, fmt.Errorf("cannot find any sample assignments")
	}
	return StepStateToContext(ctx, stepState), resp.Contents, nil
}

func (s *suite) getAStudentAssignedAssignment(ctx context.Context) (context.Context, []*epb.Content, error) {
	stepState := StepStateFromContext(ctx)

	ctx, contents, err := s.getAllStudentAssignedAssignments(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), nil, fmt.Errorf("unable to get all student assigned assignment: %w", err)
	}
	if len(contents) == 0 {
		return StepStateToContext(ctx, stepState), nil, fmt.Errorf("content have to not empty")
	}
	studyPlanItemRepo := &repositories.StudyPlanItemRepo{}
	if stepState.AssignmentStudyPlanItemID == "" {
		stepState.StudyPlanID = contents[0].StudyPlanItem.StudyPlanId
		studyPlanItems, err := studyPlanItemRepo.FindByIDs(ctx, db, database.TextArray([]string{contents[0].StudyPlanItem.StudyPlanItemId}))
		if err != nil {
			return StepStateToContext(ctx, stepState), nil, fmt.Errorf("unable to retrieve study plan item: %w", err)
		}
		stepState.AssignmentStudyPlanItemID = studyPlanItems[0].CopyStudyPlanItemID.String
		return StepStateToContext(ctx, stepState), contents[:1], nil
	}
	for _, c := range contents {
		studyPlanItems, err := studyPlanItemRepo.FindByIDs(ctx, db, database.TextArray([]string{c.StudyPlanItem.StudyPlanItemId}))
		if err != nil {
			return StepStateToContext(ctx, stepState), nil, fmt.Errorf("unable to retrieve study plan item: %w", err)
		}

		if studyPlanItems[0].ID.String == c.StudyPlanItem.StudyPlanItemId && studyPlanItems[0].CopyStudyPlanItemID.String == stepState.AssignmentStudyPlanItemID {
			return StepStateToContext(ctx, stepState), []*epb.Content{c}, nil
		}
	}
	return StepStateToContext(ctx, stepState), []*epb.Content{}, fmt.Errorf("unable to get student assigned assignment correctly")
}

func (s *suite) givenStudentSubmitTheirAssignmentInCurrentStudyPlanItem(ctx context.Context, studentID, contentStatus, times string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.CurrentStudentID = studentID
	token, err := generateValidAuthenticationToken(studentID, common.UserGroup_USER_GROUP_STUDENT.String())
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.AuthToken = token
	ctx, assignments, err := s.getAStudentAssignedAssignment(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if len(assignments) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("assignment have to not empty")
	}
	if ctx, err := s.ensureStudentIsCreated(ctx, studentID); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	ctx, err = s.submitAssignment(ctx, contentStatus, times, assignments)
	return StepStateToContext(ctx, stepState), err
}

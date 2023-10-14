package assignment

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"github.com/spf13/cast"
	"go.uber.org/multierr"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *Suite) aCourseAndStudyPlanWithStudent(ctx context.Context, option string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	courseID, err := utils.GenerateCourse(s.AuthHelper.SignedCtx(ctx, stepState.AdminToken), s.YasuoConn)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("utils.GenerateCourse: %w", err)
	}
	stepState.CourseID = courseID
	if err := utils.GenerateCourseBooks(s.AuthHelper.SignedCtx(ctx, stepState.AdminToken), courseID, []string{stepState.BookID}, s.EurekaConn); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("utils.GenerateCourseBooks: %w", err)
	}
	studyPlanResult, err := utils.GenerateStudyPlanV2(s.AuthHelper.SignedCtx(ctx, stepState.AdminToken), s.EurekaConn, courseID, stepState.BookID)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("utils.GenerateStudyPlanV2: %w", err)
	}
	stepState.StudyPlanID = studyPlanResult.StudyPlanID
	repo := &repositories.StudyPlanItemRepo{}
	// Find master study plan Items
	masterStudyPlanItems, err := repo.FindByStudyPlanID(ctx, s.EurekaDB, database.Text(stepState.StudyPlanID))
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to retrieve master study plan items: %w", err)
	}
	studyPlanItemIDs := make([]string, len(masterStudyPlanItems))
	for i, masterStudyPlanItem := range masterStudyPlanItems {
		studyPlanItemIDs[i] = masterStudyPlanItem.ID.String
		err = multierr.Combine(
			masterStudyPlanItem.AvailableFrom.Set(time.Now().Add(-24*time.Hour)),
			masterStudyPlanItem.AvailableTo.Set(time.Now().AddDate(0, 0, 10)),
			masterStudyPlanItem.StartDate.Set(time.Now().Add(-23*time.Hour)),
			masterStudyPlanItem.EndDate.Set(time.Now().AddDate(0, 0, 1)),
			masterStudyPlanItem.UpdatedAt.Set(time.Now()),
		)
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("multierr.Combine: %w", err)
		}
	}
	stepState.StudyPlanItemIDs = studyPlanItemIDs
	if err := repo.BulkInsert(ctx, s.EurekaDB, masterStudyPlanItems); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to update master study plan items: %w", err)
	}
	if option == "current" {
		courseStudents, err := utils.AValidCourseWithIDs(ctx, s.EurekaDB, []string{stepState.StudentID}, stepState.CourseID)
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("AValidCourseWithIDs: %w", err)
		}
		studentStudyPlan := &entities.StudentStudyPlan{}
		database.AllNullEntity(studentStudyPlan)
		err = multierr.Combine(
			studentStudyPlan.StudentID.Set(stepState.StudentID),
			studentStudyPlan.StudyPlanID.Set(stepState.StudyPlanID),
			studentStudyPlan.CreatedAt.Set(time.Now()),
			studentStudyPlan.UpdatedAt.Set(time.Now()),
		)
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("multierr.Combine: %w", err)
		}
		studentStudyPlanRepo := &repositories.StudentStudyPlanRepo{}
		err = studentStudyPlanRepo.BulkUpsert(ctx, s.EurekaDB, []*entities.StudentStudyPlan{
			studentStudyPlan,
		})
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to insert student study plan: %w", err)
		}
		m := utils.CryptRand(3) + 1
		locationIDs := make([]string, 0, m)
		for i := int32(1); i <= m; i++ {
			id := idutil.ULIDNow()
			locationIDs = append(locationIDs, id)
		}
		for _, courseStudent := range courseStudents {
			for _, locationID := range locationIDs {
				now := time.Now()
				e := &entities.CourseStudentsAccessPath{}
				database.AllNullEntity(e)
				if err := multierr.Combine(
					e.CourseStudentID.Set(courseStudent.ID.String),
					e.CourseID.Set(courseStudent.CourseID.String),
					e.StudentID.Set(courseStudent.StudentID.String),
					e.LocationID.Set(locationID),
					e.CreatedAt.Set(now),
					e.UpdatedAt.Set(now),
				); err != nil {
					return utils.StepStateToContext(ctx, stepState), fmt.Errorf("multierr.Combine: %w", err)
				}
				if _, err := database.Insert(ctx, e, s.EurekaDB.Exec); err != nil {
					return utils.StepStateToContext(ctx, stepState), fmt.Errorf("database.Insert: %w", err)
				}
			}
		}
		// stepState.CourseStudents = courseStudents
		// stepState.LocationIDs = locationIDs
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) studentSubmitUnrelatedAssignment(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	req := s.getSomeAssignmentSubmission(ctx, "single")[0]
	stepState.Request = req
	stepState.Response, stepState.ResponseErr = sspb.NewAssignmentClient(s.EurekaConn).
		SubmitAssignment(s.AuthHelper.SignedCtx(ctx, stepState.Token), req)
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemMustRecordsAllTheSubmissionsFromStudent(ctx context.Context, times string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	repo := repositories.StudentSubmissionRepo{}
	for _, v := range stepState.Submissions {
		e, err := repo.Get(ctx, s.EurekaDB, database.Text(v.Submission.SubmissionId))
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), err
		}

		if v.Submission.StudyPlanItemIdentity.LearningMaterialId != e.LearningMaterialID.String {
			return utils.StepStateToContext(ctx, stepState),
				fmt.Errorf("expecting '%s' assignment-id, got '%s'", v.Submission.StudyPlanItemIdentity.LearningMaterialId, e.LearningMaterialID.String)
		}

		actualContents := []*sspb.SubmissionContent{}
		if err := e.SubmissionContent.AssignTo(&actualContents); err != nil {
			return utils.StepStateToContext(ctx, stepState), err
		}

		for k, c := range v.Submission.SubmissionContent {
			if !proto.Equal(c, actualContents[k]) {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("expecting '%v' , got '%v'", c, actualContents[k])
			}
		}
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userSubmitTheirAssignmentTimes(ctx context.Context, times string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	requests := s.getSomeAssignmentSubmission(ctx, times)
	for _, req := range requests {
		resp, err := sspb.NewAssignmentClient(s.EurekaConn).
			SubmitAssignment(s.AuthHelper.SignedCtx(ctx, stepState.Token), req)
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to submit assignment: %w", err)
		}

		if resp.SubmissionId == "" {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("expecting SubmissionId")
		}
		req.Submission.SubmissionId = resp.SubmissionId
		stepState.Submissions = append(stepState.Submissions, req)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) getSomeAssignmentSubmission(ctx context.Context, times string) []*sspb.SubmitAssignmentRequest {
	stepState := utils.StepStateFromContext[StepState](ctx)
	requests := make([]*sspb.SubmitAssignmentRequest, 0)

	n := 1
	if times != "single" {
		// random "multiple" times
		//
		n = int(utils.CryptRand(5)) + 3
	}

	/* #nosec */
	for i := 0; i < n; i++ {
		req := &sspb.SubmitAssignmentRequest{
			Submission: &sspb.StudentSubmission{
				StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
					StudyPlanId:        stepState.StudyPlanID,
					LearningMaterialId: stepState.LearningMaterialID,
					StudentId:          wrapperspb.String(stepState.StudentID),
				},
				SubmissionContent: []*sspb.SubmissionContent{
					{
						SubmitMediaId:     fmt.Sprintf("%s media 1 %d", stepState.StudyPlanID, i),
						AttachmentMediaId: fmt.Sprintf("%s attachment 1 %d", stepState.StudyPlanID, i),
					},
					{
						SubmitMediaId:     fmt.Sprintf("%s media 2 %d", stepState.StudyPlanID, i),
						AttachmentMediaId: fmt.Sprintf("%s attachment 2 %d", stepState.StudyPlanID, i),
					},
				},
				Note:               fmt.Sprintf("random note: %s", stepState.StudyPlanID),
				CompleteDate:       timestamppb.New(time.Now().Add(time.Duration(utils.CryptRand(23)+1) * time.Hour)),
				Duration:           utils.CryptRand(99) + 1,
				CorrectScore:       wrapperspb.Float(rand.Float32() * 10),
				TotalScore:         wrapperspb.Float(rand.Float32() * 100),
				UnderstandingLevel: sspb.SubmissionUnderstandingLevel(rand.Intn(len(sspb.SubmissionUnderstandingLevel_value))),
				Status:             sspb.SubmissionStatus_SUBMISSION_STATUS_RETURNED,
			},
		}
		requests = append(requests, req)
	}

	return requests
}

func (s *Suite) ourSystemMustRecordsHighestGradeFromAssignment(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	query := `SELECT count(*) 
		FROM max_score_submission
		WHERE study_plan_id = $1 AND learning_material_id = $2 AND student_id = $3 AND max_score = $4`
	var count int
	if err := s.EurekaDB.QueryRow(ctx, query,
		database.Text(stepState.StudyPlanID),
		database.Text(stepState.LearningMaterialID),
		database.Text(stepState.StudentID),
		database.Int4(stepState.HighestScore)).Scan(&count); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to query max_score_submission: %w", err)
	}
	if count != 1 {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("highest score not recorded")
	}
	return ctx, nil
}

func (s *Suite) teacherGradeTheAssignments(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	// for min highest score to be 10
	highestScore := int(utils.CryptRand(90) + 10)
	n := int(utils.CryptRand(5) + 3)
	for i := 0; i < n; i++ {
		// first grade is highest score
		grade := highestScore - i
		_, err := epb.NewStudentAssignmentWriteServiceClient(s.EurekaConn).GradeStudentSubmission(s.AuthHelper.SignedCtx(ctx, stepState.Token), &epb.GradeStudentSubmissionRequest{
			Grade: &epb.SubmissionGrade{
				Note:         "good job",
				SubmissionId: stepState.SubmissionID,
				Grade:        cast.ToFloat64(grade),
			},
			Status: epb.SubmissionStatus_SUBMISSION_STATUS_RETURNED,
		})
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to grade student submission: %w", err)
		}
	}
	stepState.HighestScore = int32(highestScore)
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userSubmitTheirAssignmentWithOldSubmissionEndpoint(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	submission := &epb.StudentSubmission{
		AssignmentId:    stepState.AssignmentIDs[0],
		StudyPlanItemId: stepState.StudyPlanItemIDs[0],
		StudentId:       stepState.UserID,
		CourseId:        stepState.CourseID,
		Note:            "submit",
	}
	req := &epb.SubmitAssignmentRequest{
		Submission: submission,
	}
	resp, err := epb.NewStudentAssignmentWriteServiceClient(s.EurekaConn).
		SubmitAssignment(s.AuthHelper.SignedCtx(ctx, stepState.Token), req)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to submit assignment: %w", err)
	}

	if resp.SubmissionId == "" {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("expecting SubmissionId")
	}
	stepState.SubmissionID = resp.SubmissionId

	return utils.StepStateToContext(ctx, stepState), nil
}

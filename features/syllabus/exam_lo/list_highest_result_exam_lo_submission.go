package exam_lo

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/bob/constants"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *Suite) userListHighestResultExamLOSubmission(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	stepState.Response, stepState.ResponseErr = sspb.NewExamLOClient(s.EurekaConn).ListHighestResultExamLOSubmission(s.AuthHelper.SignedCtx(ctx, stepState.Token), &sspb.ListHighestResultExamLOSubmissionRequest{
		StudyPlanItemIdentities: stepState.StudyPlanItemIdentities,
	})

	return utils.StepStateToContext(ctx, stepState), nil
}

func getKeyFromStudyPlanItemIdentity(args *sspb.StudyPlanItemIdentity) string {
	return fmt.Sprintf("%s-%s-%s", args.LearningMaterialId, args.StudentId.Value, args.StudyPlanId)
}

func (s *Suite) thereAreExamLOSubmissionsExisted(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	now := time.Now()
	topicID := idutil.ULIDNow()

	// insert topic
	topic := &entities.Topic{}
	database.AllNullEntity(topic)

	err := multierr.Combine(
		topic.ID.Set(topicID),
		topic.Name.Set("topic-1"),
		topic.Country.Set(cpb.Country_COUNTRY_VN.String()),
		topic.Grade.Set(1),
		topic.Subject.Set(cpb.Subject_SUBJECT_BIOLOGY.String()),
		topic.TopicType.Set(cpb.TopicType_TOPIC_TYPE_NONE.String()),
		topic.TotalLOs.Set(0),
		topic.SchoolID.Set(constants.ManabieSchool),
		topic.CopiedTopicID.Set("copied-topic-id"),
		topic.EssayRequired.Set(false),
		topic.CreatedAt.Set(now),
		topic.UpdatedAt.Set(now))
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("multierr.Combine: %w", err)
	}

	if _, err := database.Insert(ctx, topic, s.EurekaDB.Exec); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("can't insert topic: %w", err)
	}

	insertExamLOFn := func(learningMaterialID, topicID, name string) error {
		examLO := &entities.ExamLO{}
		database.AllNullEntity(examLO)
		err := multierr.Combine(
			examLO.ID.Set(learningMaterialID),
			examLO.TopicID.Set(topicID),
			examLO.Name.Set(name),
			examLO.CreatedAt.Set(now),
			examLO.UpdatedAt.Set(now),
			examLO.ApproveGrading.Set(false),
			examLO.GradeCapping.Set(true),
			examLO.ReviewOption.Set(sspb.ExamLOReviewOption_EXAM_LO_REVIEW_OPTION_IMMEDIATELY.String()),
			examLO.SetDefaultVendorType(),
			examLO.IsPublished.Set(false),
		)
		if err != nil {
			return fmt.Errorf("multierr.Combine: %w", err)
		}

		if _, err := database.Insert(ctx, examLO, s.EurekaDB.Exec); err != nil {
			return fmt.Errorf("can't insert exam_lo: %w", err)
		}

		return nil
	}

	insertExamLOSubmissionsFn := func(studentID, studyPlanID, learningMaterialID string, results []string) error {
		for _, result := range results {
			shuffledQuizSetID := idutil.ULIDNow()
			examLOSubmission := &entities.ExamLOSubmission{}
			database.AllNullEntity(examLOSubmission)
			err := multierr.Combine(
				examLOSubmission.SubmissionID.Set(idutil.ULIDNow()),
				examLOSubmission.StudentID.Set(studentID),
				examLOSubmission.StudyPlanID.Set(studyPlanID),
				examLOSubmission.LearningMaterialID.Set(learningMaterialID),
				examLOSubmission.ShuffledQuizSetID.Set(shuffledQuizSetID),
				examLOSubmission.Result.Set(result),
				examLOSubmission.CreatedAt.Set(now),
				examLOSubmission.UpdatedAt.Set(now),
				examLOSubmission.LastAction.Set(sspb.ApproveGradingAction_APPROVE_ACTION_NONE.String()),
			)
			if err != nil {
				return fmt.Errorf("multierr.Combine: %w", err)
			}

			if _, err := database.Insert(ctx, examLOSubmission, s.EurekaDB.Exec); err != nil {
				return fmt.Errorf("can't insert exam_lo_submission: %w", err)
			}

			stepState.ExamLOSubmissionEnts = append(stepState.ExamLOSubmissionEnts, examLOSubmission)

			shuffledQuizSet := &entities.ShuffledQuizSet{
				ID:                       database.Text(shuffledQuizSetID),
				OriginalQuizSetID:        pgtype.Text{Status: pgtype.Null},
				QuizExternalIDs:          pgtype.TextArray{Status: pgtype.Null},
				Status:                   pgtype.Text{Status: pgtype.Null},
				RandomSeed:               pgtype.Text{Status: pgtype.Null},
				CreatedAt:                database.Timestamptz(now),
				UpdatedAt:                database.Timestamptz(now),
				DeletedAt:                pgtype.Timestamptz{Status: pgtype.Null},
				StudentID:                database.Text(studentID),
				StudyPlanItemID:          pgtype.Text{Status: pgtype.Null},
				TotalCorrectness:         database.Int4(rand.Int31()),
				SubmissionHistory:        database.JSONB("{}"),
				SessionID:                database.Text(idutil.ULIDNow()),
				OriginalShuffleQuizSetID: pgtype.Text{Status: pgtype.Null},
				StudyPlanID:              database.Text(studyPlanID),
				LearningMaterialID:       database.Text(learningMaterialID),
				QuestionHierarchy:        pgtype.JSONBArray{Status: pgtype.Null},
			}
			if _, err := database.Insert(ctx, shuffledQuizSet, s.EurekaDB.Exec); err != nil {
				return fmt.Errorf("can't insert shuffled_quiz_set: %w", err)
			}
		}
		return nil
	}

	mapStudyPlanItemIdentityMaxResult := make(map[string]string)
	var studyPlanItemIdentities []*sspb.StudyPlanItemIdentity

	// case results: NONE, PASSED, FAILED
	{
		studyPlanID := idutil.ULIDNow()
		learningMaterialID := idutil.ULIDNow()
		studentID := idutil.ULIDNow()

		if err := insertExamLOFn(learningMaterialID, topicID, "exam-lo-1"); err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("insertExamLOFn: %w", err)
		}

		if err := insertExamLOSubmissionsFn(studentID, studyPlanID, learningMaterialID, []string{
			sspb.ExamLOSubmissionResult_EXAM_LO_SUBMISSION_WAITING_FOR_GRADE.String(),
			sspb.ExamLOSubmissionResult_EXAM_LO_SUBMISSION_PASSED.String(),
			sspb.ExamLOSubmissionResult_EXAM_LO_SUBMISSION_FAILED.String(),
		}); err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("insertExamLOSubmissionsFn: %w", err)
		}

		studyPlanItemIdentity := &sspb.StudyPlanItemIdentity{
			StudyPlanId:        studyPlanID,
			LearningMaterialId: learningMaterialID,
			StudentId:          wrapperspb.String(studentID),
		}
		studyPlanItemIdentities = append(studyPlanItemIdentities, studyPlanItemIdentity)
		mapStudyPlanItemIdentityMaxResult[getKeyFromStudyPlanItemIdentity(studyPlanItemIdentity)] = sspb.ExamLOSubmissionResult_EXAM_LO_SUBMISSION_PASSED.String()
	}

	// case results: NONE, PASSED, FAILED, COMPLETED
	{
		studyPlanID := idutil.ULIDNow()
		learningMaterialID := idutil.ULIDNow()
		studentID := idutil.ULIDNow()

		if err := insertExamLOFn(learningMaterialID, topicID, "exam-lo-2"); err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("insertExamLOFn: %w", err)
		}

		if err := insertExamLOSubmissionsFn(studentID, studyPlanID, learningMaterialID, []string{
			sspb.ExamLOSubmissionResult_EXAM_LO_SUBMISSION_WAITING_FOR_GRADE.String(),
			sspb.ExamLOSubmissionResult_EXAM_LO_SUBMISSION_PASSED.String(),
			sspb.ExamLOSubmissionResult_EXAM_LO_SUBMISSION_FAILED.String(),
			sspb.ExamLOSubmissionResult_EXAM_LO_SUBMISSION_COMPLETED.String(),
		}); err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("insertExamLOSubmissionsFn: %w", err)
		}

		studyPlanItemIdentity := &sspb.StudyPlanItemIdentity{
			StudyPlanId:        studyPlanID,
			LearningMaterialId: learningMaterialID,
			StudentId:          wrapperspb.String(studentID),
		}
		studyPlanItemIdentities = append(studyPlanItemIdentities, studyPlanItemIdentity)
		mapStudyPlanItemIdentityMaxResult[getKeyFromStudyPlanItemIdentity(studyPlanItemIdentity)] = sspb.ExamLOSubmissionResult_EXAM_LO_SUBMISSION_PASSED.String()
	}

	// case results: NONE, FAILED
	{
		studyPlanID := idutil.ULIDNow()
		learningMaterialID := idutil.ULIDNow()
		studentID := idutil.ULIDNow()

		if err := insertExamLOFn(learningMaterialID, topicID, "exam-lo-3"); err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("insertExamLOFn: %w", err)
		}

		if err := insertExamLOSubmissionsFn(studentID, studyPlanID, learningMaterialID, []string{
			sspb.ExamLOSubmissionResult_EXAM_LO_SUBMISSION_WAITING_FOR_GRADE.String(),
			sspb.ExamLOSubmissionResult_EXAM_LO_SUBMISSION_FAILED.String(),
		}); err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("insertExamLOSubmissionsFn: %w", err)
		}

		studyPlanItemIdentity := &sspb.StudyPlanItemIdentity{
			StudyPlanId:        studyPlanID,
			LearningMaterialId: learningMaterialID,
			StudentId:          wrapperspb.String(studentID),
		}
		studyPlanItemIdentities = append(studyPlanItemIdentities, studyPlanItemIdentity)
		mapStudyPlanItemIdentityMaxResult[getKeyFromStudyPlanItemIdentity(studyPlanItemIdentity)] = sspb.ExamLOSubmissionResult_EXAM_LO_SUBMISSION_FAILED.String()
	}

	// case results: NONE, PASSED
	{
		studyPlanID := idutil.ULIDNow()
		learningMaterialID := idutil.ULIDNow()
		studentID := idutil.ULIDNow()

		if err := insertExamLOFn(learningMaterialID, topicID, "exam-lo-4"); err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("insertExamLOFn: %w", err)
		}

		if err := insertExamLOSubmissionsFn(studentID, studyPlanID, learningMaterialID, []string{
			sspb.ExamLOSubmissionResult_EXAM_LO_SUBMISSION_WAITING_FOR_GRADE.String(),
			sspb.ExamLOSubmissionResult_EXAM_LO_SUBMISSION_PASSED.String(),
		}); err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("insertExamLOSubmissionsFn: %w", err)
		}

		studyPlanItemIdentity := &sspb.StudyPlanItemIdentity{
			StudyPlanId:        studyPlanID,
			LearningMaterialId: learningMaterialID,
			StudentId:          wrapperspb.String(studentID),
		}
		studyPlanItemIdentities = append(studyPlanItemIdentities, studyPlanItemIdentity)
		mapStudyPlanItemIdentityMaxResult[getKeyFromStudyPlanItemIdentity(studyPlanItemIdentity)] = sspb.ExamLOSubmissionResult_EXAM_LO_SUBMISSION_PASSED.String()
	}

	// case results: NONE, COMPLETED
	{
		studyPlanID := idutil.ULIDNow()
		learningMaterialID := idutil.ULIDNow()
		studentID := idutil.ULIDNow()

		if err := insertExamLOFn(learningMaterialID, topicID, "exam-lo-5"); err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("insertExamLOFn: %w", err)
		}

		if err := insertExamLOSubmissionsFn(studentID, studyPlanID, learningMaterialID, []string{
			sspb.ExamLOSubmissionResult_EXAM_LO_SUBMISSION_WAITING_FOR_GRADE.String(),
			sspb.ExamLOSubmissionResult_EXAM_LO_SUBMISSION_COMPLETED.String(),
		}); err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("insertExamLOSubmissionsFn: %w", err)
		}

		studyPlanItemIdentity := &sspb.StudyPlanItemIdentity{
			StudyPlanId:        studyPlanID,
			LearningMaterialId: learningMaterialID,
			StudentId:          wrapperspb.String(studentID),
		}
		studyPlanItemIdentities = append(studyPlanItemIdentities, studyPlanItemIdentity)
		mapStudyPlanItemIdentityMaxResult[getKeyFromStudyPlanItemIdentity(studyPlanItemIdentity)] = sspb.ExamLOSubmissionResult_EXAM_LO_SUBMISSION_WAITING_FOR_GRADE.String()
	}

	// case results: don't insert anything
	{
		studyPlanID := idutil.ULIDNow()
		learningMaterialID := idutil.ULIDNow()
		studentID := idutil.ULIDNow()

		studyPlanItemIdentity := &sspb.StudyPlanItemIdentity{
			StudyPlanId:        studyPlanID,
			LearningMaterialId: learningMaterialID,
			StudentId:          wrapperspb.String(studentID),
		}
		studyPlanItemIdentities = append(studyPlanItemIdentities, studyPlanItemIdentity)
		mapStudyPlanItemIdentityMaxResult[getKeyFromStudyPlanItemIdentity(studyPlanItemIdentity)] = sspb.ExamLOSubmissionResult_EXAM_LO_SUBMISSION_WAITING_FOR_GRADE.String()
	}

	stepState.MapStudyPlanItemIdentityMaxResult = mapStudyPlanItemIdentityMaxResult
	stepState.StudyPlanItemIdentities = studyPlanItemIdentities

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemMustReturnHighestResultExamLOSubmissionsCorrectly(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	response := stepState.Response.(*sspb.ListHighestResultExamLOSubmissionResponse)

	for _, result := range response.StudyPlanItemResults {
		maxResult := stepState.MapStudyPlanItemIdentityMaxResult[getKeyFromStudyPlanItemIdentity(result.StudyPlanItemIdentity)]
		if maxResult != result.LatestExamLoSubmissionResult.String() {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("expected highest result %s, got %s", maxResult, result.LatestExamLoSubmissionResult.String())
		}
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

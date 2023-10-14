package student_progress

import (
	"context"
	"fmt"
	"strconv"
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
)

// nolint

func (s *Suite) hasCreatedABookWithEachLearningMaterialType(ctx context.Context, role string, numLos, numFlashcards, numAsses, numTaskAsses, numExamLos, numTopics, numChapters, numQuizzes int) (context.Context, error) {
	var err error
	stepState := utils.StepStateFromContext[StepState](ctx)
	stepState.Token = stepState.SchoolAdmin.Token

	stepState.NumTopics = numTopics * numChapters
	stepState.NumChapter = numChapters
	stepState.NumQuizzes = numQuizzes

	stepState.BookID = idutil.ULIDNow()
	if _, err := epb.NewBookModifierServiceClient(s.EurekaConn).UpsertBooks(s.AuthHelper.SignedCtx(ctx, stepState.Token), &epb.UpsertBooksRequest{
		Books: []*epb.UpsertBooksRequest_Book{
			{
				BookId: stepState.BookID,
				Name:   fmt.Sprintf("book-name+%s", stepState.BookID),
			},
		},
	}); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("school admin unable to create a book: %v", err)
	}

	if ctx, err = s.schoolAdminCreateContentBookForStudentProgressWithNewLM(ctx, numLos, numFlashcards, numAsses, numTaskAsses, numExamLos, numTopics, numChapters, numQuizzes); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable create a content book: %v", err)
	}

	stepState.CourseID, err = utils.GenerateCourse(s.AuthHelper.SignedCtx(ctx, stepState.Token), s.YasuoConn)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to generate course: %v", err)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) schoolAdminCreateContentBookForStudentProgressWithNewLM(ctx context.Context, numLos, numFlashcard, numAsses, numTaskAsses, numExamLos, numTopics, numChapters, numQuizzes int) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	stepState.Token = stepState.SchoolAdmin.Token
	authCtx := s.AuthHelper.SignedCtx(ctx, stepState.Token)
	chaptersRsp, err := utils.GenerateChaptersV2(authCtx, stepState.BookID, numChapters, nil, s.EurekaConn)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("GenerateChaptersV2: %w", err)
	}
	stepState.ChapterIDs = append(stepState.ChapterIDs, chaptersRsp.ChapterIDs...)
	for _, chapterID := range chaptersRsp.ChapterIDs {
		topicsRsp, err := utils.GenerateTopicsV2(authCtx, chapterID, numTopics, nil, s.EurekaConn)
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("GenerateTopicsV2: %w", err)
		}
		stepState.TopicIDs = append(stepState.TopicIDs, topicsRsp.TopicIDs...)

		for _, topicID := range topicsRsp.TopicIDs {
			loIDs, err := s.generateLearningObjectivesWithQuizzes(authCtx, topicID, numLos, numQuizzes)
			if err != nil {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("generateLearningObjectivesWithQuizzes: %w", err)
			}
			stepState.LoIDs = append(stepState.LoIDs, loIDs...)
			stepState.LearningMaterialIDs = append(stepState.LearningMaterialIDs, loIDs...)

			flashCardIDs, err := s.generateFlashCardWithQuizzes(ctx, authCtx, topicID, numFlashcard, numQuizzes)
			if err != nil {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("generateFlashCardWithQuizzes: %w", err)
			}
			stepState.FlashCardIDs = append(stepState.FlashCardIDs, flashCardIDs...)
			stepState.LearningMaterialIDs = append(stepState.LearningMaterialIDs, flashCardIDs...)

			loFcIDs := make([]string, 0, numLos+numFlashcard)
			loFcIDs = append(loFcIDs, loIDs...)
			loFcIDs = append(loFcIDs, flashCardIDs...)

			assRsp, err := utils.GenerateAssignment(authCtx, topicID, numAsses, loFcIDs, s.EurekaConn, nil)
			if err != nil {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("GenerateAssignment: %w", err)
			}
			stepState.AssignmentIDs = append(stepState.AssignmentIDs, assRsp.AssignmentIDs...)
			stepState.LearningMaterialIDs = append(stepState.LearningMaterialIDs, assRsp.AssignmentIDs...)

			taskAssignmentIDs, err := s.generateTaskAssignment(authCtx, topicID, numTaskAsses)
			if err != nil {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("generateTaskAssignment: %w", err)
			}
			stepState.TaskAssignmentIDs = append(stepState.TaskAssignmentIDs, taskAssignmentIDs...)
			stepState.LearningMaterialIDs = append(stepState.LearningMaterialIDs, taskAssignmentIDs...)

			examLOIDs, err := s.generateExamLO(authCtx, topicID, numExamLos)
			if err != nil {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("generateExamLO: %w", err)
			}
			stepState.ExamLoIDs = append(stepState.ExamLoIDs, examLOIDs...)
			stepState.LearningMaterialIDs = append(stepState.LearningMaterialIDs, examLOIDs...)
		}
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) doTestAndDoneLosFlashcardsWithCorrectlyAndAssignmentsTaskAssignmentsWithPointAndSkipTopics(ctx context.Context, arg1 string, numWorkonLO, numWorkonFc, numCorrectQuiz, numWorkonAss, numWorkonTaskAss, assPoint, numWorkonExamLo, examLoPoint, skippedTopic string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	lnumWorkonLO, _ := strconv.Atoi((numWorkonLO))
	lnumWorkonFC, _ := strconv.Atoi((numWorkonFc))
	lnumCorrectQuizInput, _ := strconv.Atoi((numCorrectQuiz))
	lnumWorkonAss, _ := strconv.Atoi((numWorkonAss))
	lnumWorkonTaskAss, _ := strconv.Atoi((numWorkonTaskAss))
	lnumWorkonExamLo, _ := strconv.Atoi(numWorkonExamLo)
	lSkippedTopic, _ := strconv.Atoi((skippedTopic))

	if ctx, err := s.waitForImportStudentStudyPlanCompleted(ctx); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to fetch student study plan: %w", err)
	}

	learningMaterialRepo := repositories.LearningMaterialRepo{}
	lmInfos, err := learningMaterialRepo.FindInfoByStudyPlanItemIdentity(ctx, s.EurekaDB, database.Text(stepState.Student.ID), database.Text(stepState.StudyPlanID), database.Text(""))
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to fetch learning material info: %w", err)
	}

	topicLOs := make(map[string]int)
	topicFCs := make(map[string]int)
	topicAss := make(map[string]int)
	topicTaskAss := make(map[string]int)
	topicExamLo := make(map[string]int)
	chapters := make(map[string][]string)

	for _, each := range lmInfos {
		chapters[each.ChapterID.String] = append(chapters[each.ChapterID.String], each.TopicID.String)
	}
	stepState.SkippedTopics = nil
	for _, each := range chapters {
		for i := 0; i < lSkippedTopic; i++ {
			stepState.SkippedTopics = append(stepState.SkippedTopics, each[i])
		}
	}
	for _, each := range lmInfos {
		if containsStr(stepState.SkippedTopics, each.TopicID.String) {
			continue
		}

		if each.Type.String == LEARNING_OBJECTIVE_TYPE || each.Type.String == FLASH_CARD_TYPE {
			if each.Type.String == LEARNING_OBJECTIVE_TYPE && topicLOs[each.TopicID.String] == lnumWorkonLO {
				continue
			}

			if each.Type.String == FLASH_CARD_TYPE && topicFCs[each.TopicID.String] == lnumWorkonFC {
				continue
			}

			ctx, err = s.doLosOrFlashcard(utils.StepStateToContext(ctx, stepState), lnumCorrectQuizInput, each.LearningMaterialID.String)
			if err != nil {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to do los: %w", err)
			}
			if each.Type.String == LEARNING_OBJECTIVE_TYPE {
				topicLOs[each.TopicID.String]++
			}
			if each.Type.String == FLASH_CARD_TYPE {
				topicFCs[each.TopicID.String]++
			}

			err = s.markIsCompleted(ctx, stepState.Student.ID, stepState.StudyPlanID, each.LearningMaterialID.String)
			if err != nil {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to mark learning material as completed: %w", err)
			}

			stepState.CompletedLmIDs = append(stepState.CompletedLmIDs, each.LearningMaterialID.String)
		}

		if each.Type.String == ASSIGNMENT_TYPE || each.Type.String == TASK_ASSIGNMENT_TYPE {
			if each.Type.String == ASSIGNMENT_TYPE && topicAss[each.TopicID.String] == lnumWorkonAss {
				continue
			}

			if each.Type.String == TASK_ASSIGNMENT_TYPE && topicTaskAss[each.TopicID.String] == lnumWorkonTaskAss {
				continue
			}

			ctx, err = s.doAssignmentOrTaskAssignment(utils.StepStateToContext(ctx, stepState), each.LearningMaterialID.String, cast.ToFloat32(assPoint))
			if err != nil {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to do assignment: %w", err)
			}
			stepState.CompletedLmIDs = append(stepState.CompletedLmIDs, each.LearningMaterialID.String)

			if each.Type.String == ASSIGNMENT_TYPE {
				topicAss[each.TopicID.String]++
			}
			if each.Type.String == TASK_ASSIGNMENT_TYPE {
				topicTaskAss[each.TopicID.String]++
			}
		}

		if each.Type.String == EXAM_LO_TYPE {
			if topicExamLo[each.TopicID.String] == lnumWorkonExamLo {
				continue
			}

			ctx, err = s.doExamLO(utils.StepStateToContext(ctx, stepState), each.LearningMaterialID.String, examLoPoint)
			if err != nil {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to do exam lo: %w", err)
			}

			err = s.markIsCompleted(ctx, stepState.Student.ID, stepState.StudyPlanID, each.LearningMaterialID.String)
			if err != nil {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to mark learning material as completed: %w", err)
			}
			stepState.CompletedLmIDs = append(stepState.CompletedLmIDs, each.LearningMaterialID.String)

			topicExamLo[each.TopicID.String]++
		}
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) correctLoCompletedWithFullLm(ctx context.Context, doneLOs, doneFCs, doneAssignments, doneTaskAssignments, doneExamLos string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	lDoneLo, _ := strconv.ParseInt(doneLOs, 10, 32)
	lDoneFC, _ := strconv.ParseInt(doneFCs, 10, 32)
	lDoneAssignments, _ := strconv.ParseInt(doneAssignments, 10, 32)
	lDoneTaskAssignments, _ := strconv.ParseInt(doneTaskAssignments, 10, 32)
	lDoneExamLos, _ := strconv.ParseInt(doneExamLos, 10, 32)

	resp := stepState.Response.(*sspb.GetStudentProgressResponse)

	for _, each := range resp.StudentStudyPlanProgresses[0].TopicProgress {
		if each.CompletedStudyPlanItem.Value != (int32(lDoneAssignments) + int32(lDoneLo) + int32(lDoneFC) + int32(lDoneTaskAssignments) + int32(lDoneExamLos)) {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("mismatch completed study plan items")
		}
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) doExamLO(ctx context.Context, learningMaterialID, examLoPoint string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	stepState.Token = stepState.Student.Token

	submissionID := idutil.ULIDNow()
	now := time.Now()
	examLOSubmission := &entities.ExamLOSubmission{}
	database.AllNullEntity(examLOSubmission)
	if err := multierr.Combine(
		examLOSubmission.SubmissionID.Set(submissionID),
		examLOSubmission.LearningMaterialID.Set(learningMaterialID),
		examLOSubmission.ShuffledQuizSetID.Set(fmt.Sprintf("shuffled %s", submissionID)),
		examLOSubmission.StudyPlanID.Set(stepState.StudyPlanID),
		examLOSubmission.StudentID.Set(stepState.Student.ID),
		examLOSubmission.TeacherFeedback.Set("teacher-feedback"),
		examLOSubmission.Status.Set(sspb.SubmissionStatus_SUBMISSION_STATUS_RETURNED.String()),
		examLOSubmission.Result.Set(sspb.ExamLOSubmissionResult_EXAM_LO_SUBMISSION_COMPLETED.String()),
		examLOSubmission.CreatedAt.Set(now),
		examLOSubmission.UpdatedAt.Set(now),
		examLOSubmission.TotalPoint.Set(10),
		examLOSubmission.LastAction.Set(sspb.ApproveGradingAction_APPROVE_ACTION_NONE.String()),
	); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("multierr.Combine: %w", err)
	}

	if _, err := database.Insert(ctx, examLOSubmission, s.EurekaDB.Exec); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("can't insert exam_lo_submission: %w", err)
	}

	examLOSubmissionAnswer := &entities.ExamLOSubmissionAnswer{}
	database.AllNullEntity(examLOSubmissionAnswer)
	err := multierr.Combine(
		examLOSubmissionAnswer.SubmissionID.Set(submissionID),
		examLOSubmissionAnswer.ShuffledQuizSetID.Set(fmt.Sprintf("shuffled %s", submissionID)),
		examLOSubmissionAnswer.QuizID.Set(fmt.Sprintf("quiz %s", submissionID)),
		examLOSubmissionAnswer.StudentID.Set(stepState.Student.ID),
		examLOSubmissionAnswer.LearningMaterialID.Set(learningMaterialID),
		examLOSubmissionAnswer.StudyPlanID.Set(stepState.StudyPlanID),
		examLOSubmissionAnswer.CorrectIndexAnswer.Set([]int32{1, 2, 3}),
		examLOSubmissionAnswer.CorrectTextAnswer.Set([]string{"1", "2", "3"}),
		examLOSubmissionAnswer.StudentIndexAnswer.Set([]int32{1, 2, 3}),
		examLOSubmissionAnswer.StudentTextAnswer.Set([]string{"1", "2", "3"}),
		examLOSubmissionAnswer.IsCorrect.Set(true),
		examLOSubmissionAnswer.IsAccepted.Set(false),
		examLOSubmissionAnswer.Point.Set(examLoPoint),
		examLOSubmissionAnswer.CreatedAt.Set(now),
		examLOSubmissionAnswer.UpdatedAt.Set(now),
	)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("multierr.Combine: %w", err)
	}

	if _, err := database.Insert(ctx, examLOSubmissionAnswer, s.EurekaDB.Exec); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("can't insert exam_lo_submission_answer: %w", err)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

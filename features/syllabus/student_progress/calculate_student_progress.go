package student_progress

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/try"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"github.com/jackc/pgx/v4"
	"github.com/spf13/cast"
	"go.uber.org/multierr"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// nolint
const (
	COURSE_ID         = "course_id"
	STUDENT_ID        = "student_id"
	BOOK_ID           = "book_id"
	STUDY_PLAN_ID     = "study_plan_id"
	NumberOfStudyPlan = 1

	LEARNING_OBJECTIVE_TYPE = "LEARNING_MATERIAL_LEARNING_OBJECTIVE"
	FLASH_CARD_TYPE         = "LEARNING_MATERIAL_FLASH_CARD"
	ASSIGNMENT_TYPE         = "LEARNING_MATERIAL_GENERAL_ASSIGNMENT"
	TASK_ASSIGNMENT_TYPE    = "LEARNING_MATERIAL_TASK_ASSIGNMENT"
	EXAM_LO_TYPE            = "LEARNING_MATERIAL_EXAM_LO"
)

func (s *Suite) userAssignStudyPlanToAStudent(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	err := studentJoinCourse(ctx, s.EurekaDB, stepState.Student.ID, stepState.CourseID)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}

	stepState.Token = stepState.Student.Token
	stepState.NumberOfStudyPlan = NumberOfStudyPlan
	studyPlanIDs, err := utils.GenerateStudyPlans(s.AuthHelper.SignedCtx(ctx, stepState.SchoolAdmin.Token), s.EurekaConn, stepState.CourseID, stepState.BookID, stepState.NumberOfStudyPlan)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("can't generate study plan: %v", err)
	}
	stepState.StudyPlanID = studyPlanIDs[0]

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) adminInsertIndividualStudyPlan(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	for _, learningMaterialID := range stepState.LearningMaterialIDs {
		req := generateIndividualStudyPlanRequest(stepState.StudyPlanID, learningMaterialID, stepState.Student.ID)
		_, stepState.ResponseErr = sspb.NewStudyPlanClient(s.EurekaConn).UpsertIndividual(s.AuthHelper.SignedCtx(ctx, stepState.SchoolAdmin.Token), req)
		if stepState.ResponseErr != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("can't insert individual study plan to student:%s ", stepState.ResponseErr.Error())
		}
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func generateIndividualStudyPlanRequest(spID, lmID, studentID string) *sspb.UpsertIndividualInfoRequest {
	req := &sspb.UpsertIndividualInfoRequest{
		IndividualItems: []*sspb.StudyPlanItem{
			{
				StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
					StudyPlanId:        spID,
					LearningMaterialId: lmID,
					StudentId: &wrapperspb.StringValue{
						Value: studentID,
					},
				},
				AvailableFrom: timestamppb.New(time.Now().Add(-24 * time.Hour)),
				AvailableTo:   timestamppb.New(time.Now().AddDate(0, 0, 10)),
				StartDate:     timestamppb.New(time.Now().Add(-23 * time.Hour)),
				EndDate:       timestamppb.New(time.Now().AddDate(0, 0, 1)),
				Status:        sspb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_ACTIVE,
			},
		},
	}

	return req
}

func (s *Suite) schoolAdminTeacherAndStudentLogin(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	ctx, err := s.aSignedIn(ctx, "school admin")
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	ctx, err = s.aSignedIn(ctx, "teacher")
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	ctx, err = s.aSignedIn(ctx, "student")
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) schoolAdminCreateContentBookForStudentProgress(ctx context.Context, numLos, numAsses, numTopics, numChapters, numQuizzes int) (context.Context, error) {
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

			assRsp, err := utils.GenerateAssignment(authCtx, topicID, numAsses, loIDs, s.EurekaConn, nil)
			if err != nil {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("GenerateAssignment: %w", err)
			}
			stepState.AssignmentIDs = append(stepState.AssignmentIDs, assRsp.AssignmentIDs...)
			stepState.LearningMaterialIDs = append(stepState.LearningMaterialIDs, assRsp.AssignmentIDs...)
			// time.Sleep(200 * time.Millisecond) // Work around no healthy upstream error
		}
		// time.Sleep(200 * time.Millisecond) // Work around no healthy upstream error
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) hasCreatedABookWithEachLosAssignmentsTopicsChaptersQuizzes(ctx context.Context, role string, numLos, numAsses, numTopics, numChapters, numQuizzes int) (context.Context, error) {
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

	if ctx, err = s.schoolAdminCreateContentBookForStudentProgress(ctx, numLos, numAsses, numTopics, numChapters, numQuizzes); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable create a content book: %v", err)
	}

	stepState.CourseID, err = utils.GenerateCourse(s.AuthHelper.SignedCtx(ctx, stepState.Token), s.YasuoConn)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to generate course: %v", err)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) waitForImportStudentStudyPlanCompleted(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	studentStudyPlanRepo := &repositories.StudentStudyPlanRepo{}

	err := try.Do(func(attempt int) (bool, error) {
		findByStudentIDsResp, err := studentStudyPlanRepo.FindByStudentIDs(ctx, s.EurekaDB, database.TextArray([]string{stepState.Student.ID}))
		if err != nil {
			if err == pgx.ErrNoRows {
				time.Sleep(time.Second)
				return attempt < 10, fmt.Errorf("no row found")
			}
			return false, err
		}

		if len(findByStudentIDsResp) != 0 {
			return false, nil
		}

		time.Sleep(time.Second)
		return attempt < 10, fmt.Errorf("timeout sync import student study plan")
	})

	return utils.StepStateToContext(ctx, stepState), err
}

func (s *Suite) calculateStudentProgress(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	stepState.Response, stepState.ResponseErr = sspb.NewStatisticsClient(s.EurekaConn).GetStudentProgress(s.AuthHelper.SignedCtx((ctx), stepState.Token), &sspb.GetStudentProgressRequest{
		StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
			StudentId: wrapperspb.String(stepState.Student.ID),
		},
		CourseId: stepState.CourseID,
	})

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) studentCalculateStudentProgress(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	stepState.Token = stepState.Student.Token

	stepState.Response, stepState.ResponseErr = sspb.NewStatisticsClient(s.EurekaConn).GetStudentProgress(s.AuthHelper.SignedCtx((ctx), stepState.Token), &sspb.GetStudentProgressRequest{
		StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
			StudyPlanId: stepState.StudyPlanID,
			StudentId:   wrapperspb.String(stepState.Student.ID),
		},
		CourseId: stepState.CourseID,
	})

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) doTestAndDoneLosWithCorrectlyAndAssignmentsWithPointAndSkipTopics(ctx context.Context, arg1 string, numWorkonLO, numCorrectQuiz, numWorkonAss, assPoint, skippedTopic string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	lnumWorkonLO, _ := strconv.Atoi((numWorkonLO))
	lnumCorrectQuizInput, _ := strconv.Atoi((numCorrectQuiz))
	lnumWorkonAss, _ := strconv.Atoi((numWorkonAss))
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
	topicAss := make(map[string]int)
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

		if each.Type.String == LEARNING_OBJECTIVE_TYPE {
			if topicLOs[each.TopicID.String] == lnumWorkonLO {
				continue
			}
			ctx, err = s.doLosOrFlashcard(utils.StepStateToContext(ctx, stepState), lnumCorrectQuizInput, each.LearningMaterialID.String)
			if err != nil {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to do los: %w", err)
			}
			topicLOs[each.TopicID.String]++

			err = s.markIsCompleted(ctx, stepState.Student.ID, stepState.StudyPlanID, each.LearningMaterialID.String)
			if err != nil {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to mark learning material as completed: %w", err)
			}
			stepState.CompletedLmIDs = append(stepState.CompletedLmIDs, each.LearningMaterialID.String)
		}

		if each.Type.String == ASSIGNMENT_TYPE {
			if topicAss[each.TopicID.String] == lnumWorkonAss {
				continue
			}

			ctx, err = s.doAssignmentOrTaskAssignment(utils.StepStateToContext(ctx, stepState), each.LearningMaterialID.String, cast.ToFloat32(assPoint))
			if err != nil {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to do assignment: %w", err)
			}

			stepState.CompletedLmIDs = append(stepState.CompletedLmIDs, each.LearningMaterialID.String)

			topicAss[each.TopicID.String]++
		}
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) doTestAndDoneLosWithCorrectlyAndAssignmentsWithPointInFourTopics(ctx context.Context, arg1 string, numWorkonLO1, numCorrectQuiz1, numWorkonAss1, assPoint1, numWorkonLO2, numCorrectQuiz2, numWorkonAss2, assPoint2 string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	if ctx, err := s.waitForImportStudentStudyPlanCompleted(ctx); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to fetch student study plan: %w", err)
	}

	learningMaterialRepo := repositories.LearningMaterialRepo{}
	lmInfos, err := learningMaterialRepo.FindInfoByStudyPlanItemIdentity(ctx, s.EurekaDB, database.Text(stepState.Student.ID), database.Text(stepState.StudyPlanID), database.Text(""))
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to fetch learning material info: %w", err)
	}

	topicLOs := make(map[string]int)
	topicAss := make(map[string]int)
	chapters := make(map[string][]string)
	lnumWorkonLO := 0
	lnumCorrectQuizInput := 0
	lnumWorkonAss := 0
	lassPoint := 0
	firstTopicPair := []string{stepState.TopicIDs[0], stepState.TopicIDs[1], stepState.TopicIDs[4], stepState.TopicIDs[5], stepState.TopicIDs[8], stepState.TopicIDs[9]}

	for _, each := range lmInfos {
		chapters[each.ChapterID.String] = append(chapters[each.ChapterID.String], each.TopicID.String)
	}

	for _, each := range lmInfos {
		if containsStr(firstTopicPair, each.TopicID.String) {
			lnumWorkonLO, _ = strconv.Atoi((numWorkonLO1))
			lnumCorrectQuizInput, _ = strconv.Atoi((numCorrectQuiz1))
			lnumWorkonAss, _ = strconv.Atoi((numWorkonAss1))
			lassPoint, _ = strconv.Atoi((assPoint1))
		} else {
			lnumWorkonLO, _ = strconv.Atoi((numWorkonLO2))
			lnumCorrectQuizInput, _ = strconv.Atoi((numCorrectQuiz2))
			lnumWorkonAss, _ = strconv.Atoi((numWorkonAss2))
			lassPoint, _ = strconv.Atoi((assPoint2))
		}

		if each.Type.String == LEARNING_OBJECTIVE_TYPE {
			if topicLOs[each.TopicID.String] == lnumWorkonLO {
				continue
			}
			ctx, err = s.doLosOrFlashcard(utils.StepStateToContext(ctx, stepState), lnumCorrectQuizInput, each.LearningMaterialID.String)
			if err != nil {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to do los: %w", err)
			}
			topicLOs[each.TopicID.String]++

			err = s.markIsCompleted(ctx, stepState.Student.ID, stepState.StudyPlanID, each.LearningMaterialID.String)
			if err != nil {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to mark learning material as completed: %w", err)
			}
		}

		if each.Type.String == ASSIGNMENT_TYPE {
			if topicAss[each.TopicID.String] == lnumWorkonAss {
				continue
			}
			ctx, err = s.doAssignmentOrTaskAssignment(utils.StepStateToContext(ctx, stepState), each.LearningMaterialID.String, cast.ToFloat32(lassPoint))
			if err != nil {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to do assignment: %w", err)
			}

			topicAss[each.TopicID.String]++
		}
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) topicScoreIsAndChapterScoreIs(ctx context.Context, topicScore, chapterScore int) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	res := stepState.Response.(*sspb.GetStudentProgressResponse)
	studyPlanProgress := res.StudentStudyPlanProgresses[0]

	if len(studyPlanProgress.TopicProgress) != stepState.NumTopics || len(studyPlanProgress.ChapterProgress) != stepState.NumChapter {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("incorrect result for number of chapters and topics: actual topics: %v, actual chapters: %v", len(studyPlanProgress.TopicProgress), len(studyPlanProgress.ChapterProgress))
	}

	for _, each := range studyPlanProgress.TopicProgress {
		if !containsStr(stepState.SkippedTopics, each.TopicId) && each.AverageScore != nil && each.AverageScore.Value != int32(topicScore) {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("incorrect result for topic average got = %v, want = %v", each.AverageScore.Value, topicScore)
		}

		if !containsStr(stepState.SkippedTopics, each.TopicId) && each.AverageScore == nil && int32(topicScore) != 0 {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("incorrect result for topic average got = %v, want = %v", 0, topicScore)
		}
	}

	for _, each := range studyPlanProgress.ChapterProgress {
		if each.AverageScore != nil && each.AverageScore.Value != int32(chapterScore) {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("incorrect result for chapter average got = %v, want = %v", each.AverageScore.Value, chapterScore)
		}

		if each.AverageScore == nil && int32(chapterScore) != 0 {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("incorrect result for chapter average got = %v, want = %v", 0, chapterScore)
		}
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) firstPairTopicScoreIsAndSecondPairTopicScoreIsAndChapterScoreIs(ctx context.Context, firstPairTopicScore, secondPairTopicScore, chapterScore int) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	res := stepState.Response.(*sspb.GetStudentProgressResponse)
	studyPlanProgress := res.StudentStudyPlanProgresses[0]
	firstPairTopicIds := []string{stepState.TopicIDs[0], stepState.TopicIDs[1], stepState.TopicIDs[4], stepState.TopicIDs[5], stepState.TopicIDs[8], stepState.TopicIDs[9]}
	secondPairTopicIds := []string{stepState.TopicIDs[2], stepState.TopicIDs[3], stepState.TopicIDs[6], stepState.TopicIDs[7], stepState.TopicIDs[10], stepState.TopicIDs[11]}

	if len(studyPlanProgress.TopicProgress) != stepState.NumTopics || len(studyPlanProgress.ChapterProgress) != stepState.NumChapter {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("incorrect result for number of chapters and topics")
	}

	for _, each := range studyPlanProgress.TopicProgress {
		if containsStr(firstPairTopicIds, each.TopicId) && each.AverageScore != nil && each.AverageScore.Value != int32(firstPairTopicScore) {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("incorrect result for first topic pair average")
		}
		if containsStr(secondPairTopicIds, each.TopicId) && each.AverageScore != nil && each.AverageScore.Value != int32(secondPairTopicScore) {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("incorrect result for second topic pair average")
		}
	}

	for _, each := range studyPlanProgress.ChapterProgress {
		if each.AverageScore != nil && each.AverageScore.Value != int32(chapterScore) {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("incorrect result for chapter average")
		}
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) correctLoCompletedWithAnd(ctx context.Context, doneLOs, doneAssignments string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	lDoneLo, _ := strconv.Atoi((doneLOs))
	lDoneAssignments, _ := strconv.Atoi((doneAssignments))

	resp := stepState.Response.(*sspb.GetStudentProgressResponse)

	for _, each := range resp.StudentStudyPlanProgresses[0].TopicProgress {
		if each.CompletedStudyPlanItem.Value != (int32(lDoneAssignments) + int32(lDoneLo)) {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("mismatch completed study plan items")
		}
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) schoolAdminDeleteTopics(ctx context.Context, indexList string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	stepState.Token = stepState.SchoolAdmin.Token
	topicIds := []string{}
	for _, idx := range strings.Split(indexList, ",") {
		i, err := strconv.Atoi(strings.TrimSpace(idx))
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), err
		}
		topicIds = append(topicIds, stepState.TopicIDs[i])
	}

	if _, err := epb.NewTopicModifierServiceClient(s.EurekaConn).DeleteTopics(s.AuthHelper.SignedCtx(ctx, stepState.Token), &epb.DeleteTopicsRequest{
		TopicIds: topicIds,
	}); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to delete topics: %w", err)
	}

	stepState.NumTopics -= len(topicIds)

	return utils.StepStateToContext(ctx, stepState), nil
}

func containsStr(s []string, target string) bool {
	for _, val := range s {
		if target == val {
			return true
		}
	}
	return false
}

func (s *Suite) ourSystemMustReturnLearningMaterialResultAndBookTreeCorrectly(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	response := stepState.Response.(*sspb.GetStudentProgressResponse)
	studyPlanProgress := response.StudentStudyPlanProgresses[0]

	if len(studyPlanProgress.LearningMaterialResults) != len(stepState.LearningMaterialIDs) {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("incorrect number of learning material results")
	}

	if len(studyPlanProgress.StudyPlanTrees) != len(stepState.LearningMaterialIDs) {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("incorrect number of book tree")
	}

	for _, lmResult := range studyPlanProgress.LearningMaterialResults {
		if !golibs.InArrayString(lmResult.LearningMaterial.LearningMaterialId, stepState.LearningMaterialIDs) {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected learning material id, expected: %s not exist in LearningMaterialIDs %s", lmResult.LearningMaterial.LearningMaterialId, stepState.LearningMaterialIDs)
		}
		if !golibs.InArrayString(lmResult.LearningMaterial.TopicId, stepState.TopicIDs) {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected chapter_id, expected: %s exist in ChapterIDs %s", lmResult.LearningMaterial.TopicId, stepState.TopicIDs)
		}

		if golibs.InArrayString(lmResult.LearningMaterial.LearningMaterialId, stepState.CompletedLmIDs) != lmResult.IsCompleted {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected learning material completed status, %s is completed should be %t", lmResult.LearningMaterial.LearningMaterialId, lmResult.IsCompleted)
		}
	}

	for _, bookTree := range studyPlanProgress.StudyPlanTrees {
		if !golibs.InArrayString(bookTree.BookTree.LearningMaterialId, stepState.LearningMaterialIDs) {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected learning material id, expected: %s not exist in LearningMaterialIDs %s", bookTree.BookTree.LearningMaterialId, stepState.LearningMaterialIDs)
		}
		if stepState.BookID != bookTree.BookTree.BookId {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected book_id, expected: %s actual: %s", stepState.BookID, bookTree.BookTree.BookId)
		}
		if !golibs.InArrayString(bookTree.BookTree.ChapterId, stepState.ChapterIDs) {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected chapter_id, expected: %s exist in ChapterIDs %s", bookTree.BookTree.ChapterId, stepState.ChapterIDs)
		}
		if !golibs.InArrayString(bookTree.BookTree.TopicId, stepState.TopicIDs) {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected topic_id, expected: %s exist in TopicIDs %s", bookTree.BookTree.TopicId, stepState.TopicIDs)
		}
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) markIsCompleted(ctx context.Context, studentID, studyPlanID, learningMaterialID string) error {
	query := `
			INSERT INTO student_event_logs(student_id, study_plan_id, learning_material_id, event_type, created_at, resource_path, payload)
			VALUES($1::TEXT, $2::TEXT, $3::TEXT, 'quiz_answer_selected_test', now(), $4, '{"event": "completed"}');
			`
	if _, err := s.EurekaDB.Exec(ctx, query, studentID, studyPlanID, learningMaterialID, fmt.Sprintf("%d", constants.ManabieSchool)); err != nil {
		return fmt.Errorf("unable to insert student event logs: %v", err)
	}

	return nil
}

func (s *Suite) generateLearningObjectivesWithQuizzes(authCtx context.Context, topicID string, numLos, numQuizzes int) ([]string, error) {
	losRsp, err := utils.GenerateLearningObjectivesV2(authCtx, topicID, numLos, cpb.LearningObjectiveType_LEARNING_OBJECTIVE_TYPE_LEARNING, nil, s.EurekaConn)
	if err != nil {
		return nil, fmt.Errorf("GenerateLearningObjectivesV2: %w", err)
	}

	for _, loID := range losRsp.LoIDs {
		err := utils.GenerateQuizzes(authCtx, loID, numQuizzes, nil, s.EurekaConn)
		if err != nil {
			return nil, fmt.Errorf("GenerateQuizzes: %w", err)
		}
	}
	return losRsp.LoIDs, nil
}

func (s *Suite) generateFlashCardWithQuizzes(ctx, authCtx context.Context, topicID string, numFlashcard, numQuizzes int) ([]string, error) {
	flashCardIDs := make([]string, 0, numFlashcard)
	for i := 0; i < numFlashcard; i++ {
		fcRsp, err := utils.GenerateFlashcard(authCtx, s.EurekaConn, topicID)
		if err != nil {
			return nil, fmt.Errorf("GenerateFlashcard: %w", err)
		}

		// For check quiz correctness
		query := `
				insert into learning_objectives(lo_id, name, updated_at, created_at, resource_path)
				values($1, 'flashcard', now(), now(), $2)`

		if _, err := s.EurekaDB.Exec(ctx, query, fcRsp[0], fmt.Sprintf("%d", constants.ManabieSchool)); err != nil {
			return nil, fmt.Errorf("unable to insert learning_objectives: %v", err)
		}
		flashCardIDs = append(flashCardIDs, fcRsp...)
	}

	for _, flashCardID := range flashCardIDs {
		err := utils.GenerateQuizzes(authCtx, flashCardID, numQuizzes, nil, s.EurekaConn)
		if err != nil {
			return nil, fmt.Errorf("GenerateQuizzes: %w", err)
		}
	}

	return flashCardIDs, nil
}

func (s *Suite) generateTaskAssignment(authCtx context.Context, topicID string, numTaskAsses int) ([]string, error) {
	taskAssignmentIDs := make([]string, 0)
	for i := 0; i < numTaskAsses; i++ {
		taRsp, err := sspb.NewTaskAssignmentClient(s.EurekaConn).InsertTaskAssignment(authCtx, &sspb.InsertTaskAssignmentRequest{
			TaskAssignment: &sspb.TaskAssignmentBase{
				Base: &sspb.LearningMaterialBase{
					TopicId: topicID,
					Name:    fmt.Sprintf("task-assignment-name-%s", idutil.ULIDNow()),
				},
				Attachments: []string{"attachment-1", "attachment-2"},
				Instruction: "instruction",
			},
		})
		if err != nil {
			return nil, fmt.Errorf("unable to insert task assignment: %w", err)
		}
		taskAssignmentIDs = append(taskAssignmentIDs, taRsp.LearningMaterialId)
	}

	return taskAssignmentIDs, nil
}

func (s *Suite) generateExamLO(authCtx context.Context, topicID string, numExamLos int) ([]string, error) {
	examLOIDs := make([]string, 0)
	for i := 0; i < numExamLos; i++ {
		resp, err := sspb.NewExamLOClient(s.EurekaConn).InsertExamLO(authCtx, &sspb.InsertExamLORequest{
			ExamLo: &sspb.ExamLOBase{
				Base: &sspb.LearningMaterialBase{
					TopicId: topicID,
					Name:    "exam-lo-name",
				},
				Instruction:   "instruction",
				GradeToPass:   wrapperspb.Int32(10),
				ManualGrading: false,
				TimeLimit:     wrapperspb.Int32(100),
			},
		})
		if err != nil {
			return nil, fmt.Errorf("unable to insert exam lo: %w", err)
		}
		examLOIDs = append(examLOIDs, resp.GetLearningMaterialId())
	}

	return examLOIDs, nil
}

func (s *Suite) doLosOrFlashcard(ctx context.Context, lnumCorrectQuizInput int, learningMaterialID string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	stepState.Token = stepState.Student.Token
	resp, err := sspb.NewQuizClient(s.EurekaConn).CreateQuizTestV2(s.AuthHelper.SignedCtx((ctx), stepState.Token), &sspb.CreateQuizTestV2Request{
		StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
			StudyPlanId:        stepState.StudyPlanID,
			StudentId:          wrapperspb.String(stepState.Student.ID),
			LearningMaterialId: learningMaterialID,
		},
		KeepOrder: false,
		Paging: &cpb.Paging{
			Limit: uint32(100),
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: int64(1),
			},
		},
	})
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to create quiz test: %w", err)
	}

	lNumCorrectQuiz := lnumCorrectQuizInput
	for _, quiz := range resp.Quizzes {
		if lNumCorrectQuiz > 0 {
			if _, err := epb.NewQuizModifierServiceClient(s.EurekaConn).CheckQuizCorrectness(s.AuthHelper.SignedCtx((ctx), stepState.Token), &epb.CheckQuizCorrectnessRequest{
				SetId:  resp.ShuffleQuizSetId,
				QuizId: quiz.Core.ExternalId,
				Answer: []*epb.Answer{
					{Format: &epb.Answer_FilledText{FilledText: "A"}},
				},
			}); err != nil {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to check quiz correctness: %w", err)
			}
		} else {
			if _, err := epb.NewQuizModifierServiceClient(s.EurekaConn).CheckQuizCorrectness(s.AuthHelper.SignedCtx((ctx), stepState.Token), &epb.CheckQuizCorrectnessRequest{
				SetId:  resp.ShuffleQuizSetId,
				QuizId: quiz.Core.ExternalId,
				Answer: []*epb.Answer{
					{Format: &epb.Answer_FilledText{FilledText: "B"}},
				},
			}); err != nil {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to check quiz correctness: %w", err)
			}
			stepState.WrongQuizExternalIDs = append(stepState.WrongQuizExternalIDs, quiz.Core.ExternalId)
		}
		lNumCorrectQuiz--
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) doAssignmentOrTaskAssignment(ctx context.Context, learningMaterialID string, assPoint float32) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	stepState.Token = stepState.Student.Token

	submitAssRes, err := sspb.NewAssignmentClient(s.EurekaConn).SubmitAssignment(s.AuthHelper.SignedCtx((ctx), stepState.Token), &sspb.SubmitAssignmentRequest{
		Submission: &sspb.StudentSubmission{
			StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
				StudyPlanId:        stepState.StudyPlanID,
				StudentId:          wrapperspb.String(stepState.Student.ID),
				LearningMaterialId: learningMaterialID,
			},
			CourseId:     stepState.CourseID,
			Note:         "submit",
			CompleteDate: timestamppb.Now(),
			CorrectScore: wrapperspb.Float(assPoint),
			TotalScore:   wrapperspb.Float(10),
		},
	})
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to submit assignment: %w", err)
	}
	stepState.SubmissionIDs = append(stepState.SubmissionIDs, submitAssRes.SubmissionId)

	return utils.StepStateToContext(ctx, stepState), nil
}

func studentJoinCourse(ctx context.Context, db database.QueryExecer, studentID, courseID string) error {
	courseStudent := &entities.CourseStudent{}
	database.AllNullEntity(courseStudent)
	err := multierr.Combine(
		courseStudent.StudentID.Set(studentID),
		courseStudent.CourseID.Set(courseID),
		courseStudent.ID.Set(idutil.ULIDNow()),
		courseStudent.StartAt.Set(time.Now().Add(-24*time.Hour)),
		courseStudent.EndAt.Set(time.Now().AddDate(0, 0, 10)),
		courseStudent.BaseEntity.CreatedAt.Set(time.Now()),
		courseStudent.BaseEntity.UpdatedAt.Set(time.Now()),
	)
	if err != nil {
		return fmt.Errorf("unable to set course student: %w", err)
	}

	courseStudentRepo := repositories.CourseStudentRepo{}
	if err := courseStudentRepo.BulkUpsertV2(ctx, db, []*entities.CourseStudent{
		courseStudent,
	}); err != nil {
		return fmt.Errorf("unable to create course student: %w", err)
	}

	return nil
}

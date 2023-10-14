package exam_lo

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
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *Suite) addExamLOToTopic(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	lo, err := utils.GenerateExamLO(s.AuthHelper.SignedCtx(ctx, stepState.Token), stepState.TopicID, nil, wrapperspb.Int32(7), true, true, s.EurekaConn)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("utils.GenerateExamLO: %w", err)
	}
	stepState.LoID = lo
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) createSomeTags(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	questionTagTypeID := idutil.ULIDNow()
	err := utils.GenerateQuestionTagType(s.AuthHelper.SignedCtx(ctx, stepState.Token), s.EurekaDB, questionTagTypeID)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("utils.GenerateQuestionTagType: %s", err.Error())
	}
	questionTagIDs := make([]string, 0)
	uuid := idutil.ULIDNow()
	for i := 0; i < 4; i++ {
		questionTagIDs = append(questionTagIDs, uuid+strconv.Itoa(i))
	}
	err = utils.GenerateQuestionTags(ctx, s.EurekaDB, questionTagIDs, questionTagTypeID)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("utils.GenerateQuestionTags: %s", err.Error())
	}
	stepState.QuestionTagIDs = questionTagIDs
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) addSomeQuizzesToExamLOWithTags(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	tags := [][]string{
		{stepState.QuestionTagIDs[0], stepState.QuestionTagIDs[1], stepState.QuestionTagIDs[2]}, // true
		{stepState.QuestionTagIDs[0], stepState.QuestionTagIDs[1]},                              // true
		{stepState.QuestionTagIDs[3], stepState.QuestionTagIDs[1]},                              // false
	}
	quizIDs, err := utils.GenerateQuizzesWithTag(s.AuthHelper.SignedCtx(ctx, stepState.Token), stepState.LoID, 3, tags, s.EurekaConn)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("utils.GenerateQuizzesWithTag: %w", err)
	}
	stepState.QuizIDs = quizIDs
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) createAStudyPlanWithBook(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	courseID, err := utils.GenerateCourse(s.AuthHelper.SignedCtx(ctx, stepState.Token), s.YasuoConn)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("utils.GenerateCourse: %w", err)
	}
	stepState.CourseID = courseID
	studyPlanID, err := utils.GenerateStudyPlan(s.AuthHelper.SignedCtx(ctx, stepState.Token), s.EurekaConn, courseID, stepState.BookID)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("utils.GenerateStudyPlan: %w", err)
	}
	stepState.StudyPlanID = studyPlanID
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) aStudentJoinCourse(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	studentID := idutil.ULIDNow()
	stepState.StudentID = studentID
	err := utils.InsertUserIntoBob(s.AuthHelper.SignedCtx(ctx, stepState.Token), s.BobDB, studentID)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("utils.InsertUserIntoBob: %w", err)
	}

	// time.Sleep(2 * time.Second)
	// err = utils.UserAssignCourseToAStudent(s.AuthHelper.SignedCtx(ctx, stepState.Token), s.UserMgmtConn, studentID, stepState.CourseID)
	// if err != nil {
	//	return utils.StepStateToContext(ctx, stepState), fmt.Errorf("utils.UserAssignCourseToAStudent: %w", err)
	//}

	courseStudents, err := utils.AValidCourseWithIDs(s.AuthHelper.SignedCtx(ctx, stepState.Token), s.EurekaDB, []string{studentID}, stepState.CourseID)
	if err != nil {
		stepState.ResponseErr = err
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("AValidCourseWithIDs: %w", err)
	}
	stepState.CourseStudents = courseStudents

	// Update start_date, end_date, available_to, available_from of master_study_plan_item to student available to do ExamLO
	masterStudyPlanItems, err := (&repositories.StudyPlanItemRepo{}).FindByStudyPlanID(s.AuthHelper.SignedCtx(ctx, stepState.Token), s.EurekaDB, database.Text(stepState.StudyPlanID))
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to retrieve master study plan items: %w", err)
	}

	// Find child study plan item
	for _, masterStudyPlanItem := range masterStudyPlanItems {
		childStudyPlanItems, err := (&repositories.StudyPlanItemRepo{}).RetrieveChildStudyPlanItem(s.AuthHelper.SignedCtx(ctx, stepState.Token), s.EurekaDB, masterStudyPlanItem.ID, database.TextArray([]string{stepState.StudentID}))
		if err != nil {
			stepState.ResponseErr = err
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("error retrieve child study plan item")
		}
		stepState.StudyPlanItems = append(stepState.StudyPlanItems, masterStudyPlanItem)
		stepState.StudyPlanItemsIDs = append(stepState.StudyPlanItemsIDs, masterStudyPlanItem.ID.String)
		for _, childStudyPlanItem := range childStudyPlanItems {
			stepState.StudyPlanItems = append(stepState.StudyPlanItems, childStudyPlanItem)
			stepState.StudyPlanItemsIDs = append(stepState.StudyPlanItemsIDs, childStudyPlanItem.ID.String)
		}
	}

	for _, masterStudyPlanItem := range masterStudyPlanItems {
		stepState.StudyPlanItems = append(stepState.StudyPlanItems, masterStudyPlanItem)
		stepState.StudyPlanItemsIDs = append(stepState.StudyPlanItemsIDs, masterStudyPlanItem.ID.String)
	}

	upsertSpiReq := &epb.UpsertStudyPlanItemV2Request{}
	for i := 0; i < len(stepState.LoIDs); i++ {
		for _, item := range stepState.StudyPlanItems {
			cse := &entities.ContentStructure{}
			err := item.ContentStructure.AssignTo(cse)
			if err != nil {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("error unmarshal ContentStructure: %w", err)
			}

			cs := &epb.ContentStructure{}
			err = item.ContentStructure.AssignTo(cs)
			if err != nil {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("error unmarshal ContentStructure: %w", err)
			}

			if len(cse.LoID) != 0 {
				cs.ItemId = &epb.ContentStructure_LoId{LoId: wrapperspb.String(cse.LoID)}
			} else if len(cse.AssignmentID) != 0 {
				stepState.AssignedStudentIDs = append(stepState.AssignedStudentIDs, cse.AssignmentID)
				cs.ItemId = &epb.ContentStructure_AssignmentId{AssignmentId: wrapperspb.String(cse.AssignmentID)}
			}

			upsertSpiReq.StudyPlanItems = append(upsertSpiReq.StudyPlanItems, &epb.StudyPlanItem{
				StudyPlanId:             item.StudyPlanID.String,
				StudyPlanItemId:         item.ID.String,
				AvailableFrom:           timestamppb.New(time.Now().Add(-24 * time.Hour)),
				AvailableTo:             timestamppb.New(time.Now().AddDate(0, 0, 10)),
				StartDate:               timestamppb.New(time.Now().Add(-23 * time.Hour)),
				EndDate:                 timestamppb.New(time.Now().AddDate(0, 0, 1)),
				ContentStructure:        cs,
				ContentStructureFlatten: item.ContentStructureFlatten.String,
				Status:                  epb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_ACTIVE,
			})
		}
	}
	_, err = epb.NewStudyPlanModifierServiceClient(s.EurekaConn).UpsertStudyPlanItemV2(s.AuthHelper.SignedCtx(ctx, stepState.Token), upsertSpiReq)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to upsert study plan item: %w", err)
	}

	stepState.StudyPlanItemID = stepState.StudyPlanItemsIDs[0]

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) aStudentDoExamLO(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	quizTest, err := epb.NewQuizModifierServiceClient(s.EurekaConn).CreateQuizTest(s.AuthHelper.SignedCtx(ctx, stepState.Token), &epb.CreateQuizTestRequest{
		StudyPlanItemId: stepState.StudyPlanItemID,
		LoId:            stepState.LoID,
		StudentId:       stepState.StudentID,
		KeepOrder:       true,
		Paging: &cpb.Paging{
			Limit: uint32(5),
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: 1,
			},
		},
	})
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("CreateQuizTest: %w", err)
	}

	_, err = epb.NewCourseModifierServiceClient(s.EurekaConn).SubmitQuizAnswers(s.AuthHelper.SignedCtx(ctx, stepState.Token), &epb.SubmitQuizAnswersRequest{
		SetId: quizTest.QuizzesId,
		QuizAnswer: []*epb.QuizAnswer{
			{
				QuizId: stepState.QuizIDs[0],
				Answer: []*epb.Answer{
					{
						Format: &epb.Answer_SelectedIndex{
							SelectedIndex: 1,
						},
					},
				},
			},
			{
				QuizId: stepState.QuizIDs[1],
				Answer: []*epb.Answer{
					{
						Format: &epb.Answer_SelectedIndex{
							SelectedIndex: 1,
						},
					},
				},
			},
		},
	})
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("SubmitQuizAnswers: %w", err)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

type point struct {
	gradePoint uint32
	totalPoint uint32
}

func (s *Suite) userRetrieveMetadataTaggingResult(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	var examLOSubmissionID string
	query := `SELECT submission_id FROM exam_lo_submission WHERE student_id = $1 AND learning_material_id = $2`
	if err := s.EurekaDB.QueryRow(s.AuthHelper.SignedCtx(ctx, stepState.Token), query, stepState.StudentID, stepState.LoID).Scan(&examLOSubmissionID); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable get exam lo submission: %v", err)
	}

	res, err := sspb.NewExamLOClient(s.EurekaConn).RetrieveMetadataTaggingResult(s.AuthHelper.SignedCtx(ctx, stepState.Token), &sspb.RetrieveMetadataTaggingResultRequest{
		SubmissionId: examLOSubmissionID,
	})
	stepState.Response = res
	stepState.ResponseErr = err

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) metadataTaggingResultIsCorrect(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	res := stepState.Response.(*sspb.RetrieveMetadataTaggingResultResponse)
	responseTaggingResult := make(map[string]point)
	for _, taggingResult := range res.TaggingResults {
		responseTaggingResult[taggingResult.TagId] = point{
			gradePoint: taggingResult.GradedPoint,
			totalPoint: taggingResult.TotalPoint,
		}
	}

	if len(responseTaggingResult) != 4 ||
		responseTaggingResult[stepState.QuestionTagIDs[0]].gradePoint != 2 ||
		responseTaggingResult[stepState.QuestionTagIDs[1]].gradePoint != 2 ||
		responseTaggingResult[stepState.QuestionTagIDs[2]].gradePoint != 1 ||
		responseTaggingResult[stepState.QuestionTagIDs[3]].gradePoint != 0 ||
		responseTaggingResult[stepState.QuestionTagIDs[0]].totalPoint != 2 ||
		responseTaggingResult[stepState.QuestionTagIDs[1]].totalPoint != 3 ||
		responseTaggingResult[stepState.QuestionTagIDs[2]].totalPoint != 1 ||
		responseTaggingResult[stepState.QuestionTagIDs[3]].totalPoint != 1 {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("return wrong result")
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

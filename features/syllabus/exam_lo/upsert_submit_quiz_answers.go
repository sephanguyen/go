package exam_lo

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"go.uber.org/multierr"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *Suite) userCreatesAValidBookContent(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	bookResult, err := utils.GenerateBooksV2(s.AuthHelper.SignedCtx(ctx, stepState.Token), 1, nil, s.EurekaConn)
	if err != nil {
		stepState.ResponseErr = err
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("GenerateBooksV2: %w", err)
	}
	stepState.BookID = bookResult.BookIDs[0]

	chapterResult, err := utils.GenerateChaptersV2(s.AuthHelper.SignedCtx(ctx, stepState.Token), stepState.BookID, 1, nil, s.EurekaConn)
	if err != nil {
		stepState.ResponseErr = err
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("GenerateChaptersV2: %w", err)
	}
	stepState.ChapterID = chapterResult.ChapterIDs[0]

	topicResult, err := utils.GenerateTopicsV2(s.AuthHelper.SignedCtx(ctx, stepState.Token), stepState.ChapterID, 1, nil, s.EurekaConn)
	if err != nil {
		stepState.ResponseErr = err
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("GenerateTopicsV2: %w", err)
	}
	stepState.TopicID = topicResult.TopicIDs[0]

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userCreatesACourseAndAddStudentsIntoTheCourse(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	courseID, err := utils.GenerateCourse(s.AuthHelper.SignedCtx(ctx, stepState.Token), s.YasuoConn)
	if err != nil {
		stepState.ResponseErr = err
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("GenerateCourse: %w", err)
	}
	stepState.CourseID = courseID

	studentIDs, err := utils.InsertMultiUserIntoBob(s.AuthHelper.SignedCtx(ctx, stepState.Token), s.BobDB, 1)
	if err != nil {
		stepState.ResponseErr = err
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("InsertMultiUserIntoBob: %w", err)
	}
	stepState.StudentIDs = studentIDs

	courseStudents, err := utils.AValidCourseWithIDs(s.AuthHelper.SignedCtx(ctx, stepState.Token), s.EurekaDB, stepState.StudentIDs, stepState.CourseID)
	if err != nil {
		stepState.ResponseErr = err
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("AValidCourseWithIDs: %w", err)
	}
	stepState.CourseStudents = courseStudents

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userAddsAMasterStudyPlanWithTheCreatedBook(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	studyPlanID, err := utils.GenerateStudyPlan(s.AuthHelper.SignedCtx(ctx, stepState.Token), s.EurekaConn, stepState.CourseID, stepState.BookID)
	if err != nil {
		stepState.ResponseErr = err
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("GenerateStudyPlan: %w", err)
	}
	stepState.StudyPlanID = studyPlanID

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userCreatesAnExamLoWithManualGradingIsGradeToPassIsApproveGradingIs(ctx context.Context, manualGrading string, gradeToPass string, approveGrading string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	vManualGrading, _ := strconv.ParseBool(manualGrading)
	var vGradeToPass *wrapperspb.Int32Value
	if gradeToPass != "" {
		pGradeToPass, _ := strconv.ParseInt(gradeToPass, 10, 32)
		vGradeToPass = wrapperspb.Int32(int32(pGradeToPass))
	}
	vApproveGrading, _ := strconv.ParseBool(approveGrading)

	examLOID, err := utils.GenerateExamLO(s.AuthHelper.SignedCtx(ctx, stepState.Token), stepState.TopicID, nil, vGradeToPass, vManualGrading, vApproveGrading, s.EurekaConn)
	if err != nil {
		stepState.ResponseErr = err
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("GenerateExamLO: %w", err)
	}
	stepState.LearningMaterialID = examLOID
	stepState.LoIDs = append(stepState.LoIDs, examLOID)

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userAddsQuizzesInTypeAndSetsPointForEachQuiz(ctx context.Context, numOfQuizzes int, typeOfQuizzes string, point int) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	var vTypeOfQuizzes cpb.QuizType
	switch typeOfQuizzes {
	case "multiple choice":
		vTypeOfQuizzes = cpb.QuizType_QUIZ_TYPE_MCQ
	case "fill in blank":
		vTypeOfQuizzes = cpb.QuizType_QUIZ_TYPE_FIB
	case "multiple answer":
		vTypeOfQuizzes = cpb.QuizType_QUIZ_TYPE_MAQ
	default:
		vTypeOfQuizzes = cpb.QuizType_QUIZ_TYPE_MCQ
	}

	upsertSingleQuizResult, err := utils.GenerateUpsertSingleQuiz(s.AuthHelper.SignedCtx(ctx, stepState.Token), stepState.LearningMaterialID, vTypeOfQuizzes, numOfQuizzes, int32(point), s.EurekaConn)
	if err != nil {
		stepState.ResponseErr = err
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("GenerateUpsertSingleQuiz: %w", err)
	}
	stepState.ExternalIDs = upsertSingleQuizResult.ExternalIDs
	stepState.TotalPoint = upsertSingleQuizResult.TotalPoint

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userUpdatesStudyPlanForTheExamLO(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	// Find master study plan Items
	masterStudyPlanItems, err := (&repositories.StudyPlanItemRepo{}).FindByStudyPlanID(s.AuthHelper.SignedCtx(ctx, stepState.Token), s.EurekaDB, database.Text(stepState.StudyPlanID))
	if err != nil {
		stepState.ResponseErr = err
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to retrieve master study plan items: %w", err)
	}

	// Find child study plan item
	for _, masterStudyPlanItem := range masterStudyPlanItems {
		childStudyPlanItems, err := (&repositories.StudyPlanItemRepo{}).RetrieveChildStudyPlanItem(s.AuthHelper.SignedCtx(ctx, stepState.Token), s.EurekaDB, masterStudyPlanItem.ID, database.TextArray(stepState.StudentIDs))
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
	resp, err := epb.NewStudyPlanModifierServiceClient(s.EurekaConn).UpsertStudyPlanItemV2(s.AuthHelper.SignedCtx(ctx, stepState.Token), upsertSpiReq)
	if err != nil {
		stepState.ResponseErr = err
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("NewStudyPlanModifierServiceClient.UpsertStudyPlanItemV2: %w", err)
	}

	stepState.StudyPlanItemID = resp.StudyPlanItemIds[0]

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userStartsAndSubmitsAnswersInMultipleChoiceTypeAndExit(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	stepState.Offset = 1
	stepState.NextPage = nil
	stepState.SetID = ""
	stepState.SessionID = idutil.ULIDNow()
	stepState.CurrentStudentID = stepState.StudentIDs[0]
	request := &epb.CreateQuizTestRequest{
		LoId:            stepState.LearningMaterialID,
		StudentId:       stepState.CurrentStudentID,
		StudyPlanItemId: stepState.StudyPlanItemID,
		SessionId:       stepState.SessionID,
		Paging: &cpb.Paging{
			Limit: uint32(5),
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: 1,
			},
		},
	}

	resp, err := epb.NewQuizModifierServiceClient(s.EurekaConn).CreateQuizTest(s.AuthHelper.SignedCtx(ctx, stepState.Token), request)
	if err != nil {
		stepState.ResponseErr = err
		return utils.StepStateToContext(ctx, stepState), nil
	}
	stepState.ShuffledQuizSetID = resp.QuizzesId
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("cannot select exam lo submission, err: %w", err)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) loProgressionAndLoProgressionAnswersHasBeenCreated(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	now := time.Now()
	progressionID := idutil.ULIDNow()
	var loP entities.LOProgression

	database.AllNullEntity(&loP)
	if err := multierr.Combine(
		loP.CreatedAt.Set(now),
		loP.UpdatedAt.Set(now),
		loP.ProgressionID.Set(progressionID),
		loP.LastIndex.Set(database.Int4(int32(len(stepState.ExternalIDs)/2+1))),
		loP.ShuffledQuizSetID.Set(stepState.ShuffledQuizSetID),
		loP.QuizExternalIDs.Set(stepState.ExternalIDs),
		loP.StudentID.Set(stepState.CurrentStudentID),
		loP.StudyPlanID.Set(stepState.StudyPlanID),
		loP.LearningMaterialID.Set(stepState.LearningMaterialID),
	); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("cannot set value for lo progression, err: %w", err)
	}

	var loPAs []*entities.LOProgressionAnswer

	lenExternalIDs := len(stepState.ExternalIDs)/2 + 1 // because do half the answer and exit
	for i := 0; i < lenExternalIDs; i++ {
		var loPA entities.LOProgressionAnswer
		database.AllNullEntity(&loPA)
		now = time.Now()
		if err := multierr.Combine(
			loPA.CreatedAt.Set(now),
			loPA.UpdatedAt.Set(now),
			loPA.ProgressionAnswerID.Set(idutil.ULIDNow()),
			loPA.ProgressionID.Set(progressionID),
			loPA.ShuffledQuizSetID.Set(stepState.ShuffledQuizSetID),
			loPA.StudentID.Set(stepState.CurrentStudentID),
			loPA.StudyPlanID.Set(stepState.StudyPlanID),
			loPA.LearningMaterialID.Set(stepState.LearningMaterialID),
			loPA.QuizExternalID.Set(stepState.ExternalIDs[i]),
			loPA.StudentIndexAnswers.Set([]int32{1}),
		); err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("cannot set value for lo progression answer, err: %w", err)
		}
		loPAs = append(loPAs, &loPA)
	}
	// will update later when have insert api
	fieldNames := database.GetFieldNames(&loP)
	placeHolders := database.GeneratePlaceholders(len(fieldNames))
	query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s)`, loP.TableName(), strings.Join(fieldNames, ","), placeHolders)
	if _, err := s.EurekaDB.Exec(ctx, query, database.GetScanFields(&loP, fieldNames)...); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("cannot insert lo progression, error: %s", err.Error())
	}

	var LoARepo repositories.LOProgressionAnswerRepo
	if err := LoARepo.BulkUpsert(ctx, s.EurekaDB, loPAs); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("cannot upsert lo progression answers, error: %s", err.Error())
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) loProgressionAndLoProgressionAnswersHasBeenDeletedCorrectly(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	var total int
	query := `SELECT COUNT(*) FROM lo_progression WHERE study_plan_id = $1::TEXT AND learning_material_id = $2::TEXT AND student_id = $3::TEXT AND deleted_at IS NOT NULL`
	if err := s.EurekaDB.QueryRow(ctx, query, stepState.StudyPlanID, stepState.LearningMaterialID, stepState.CurrentStudentID).Scan(&total); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("cannot count lo progression table, err: %w", err)
	}
	if total != 1 {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("delete lo progression wrong expected %d get %d", 1, total)
	}

	total = 0
	query = `SELECT COUNT(*) FROM lo_progression_answer WHERE study_plan_id = $1::TEXT AND learning_material_id = $2::TEXT AND student_id = $3::TEXT AND deleted_at IS NOT NULL`
	if err := s.EurekaDB.QueryRow(ctx, query, stepState.StudyPlanID, stepState.LearningMaterialID, stepState.CurrentStudentID).Scan(&total); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("cannot count lo progression answer table, err: %w", err)
	}
	if total != len(stepState.ExternalIDs)/2+1 {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("delete lo progression answer wrong expected %d get %d", len(stepState.ExternalIDs)/2+1, total)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userStartsAndSubmitsAnswersInMultipleChoiceType(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	stepState.Offset = 1
	stepState.NextPage = nil
	stepState.SetID = ""
	stepState.SessionID = strconv.Itoa(rand.Int())
	stepState.CurrentStudentID = stepState.StudentIDs[0]
	request := &epb.CreateQuizTestRequest{
		LoId:            stepState.LearningMaterialID,
		StudentId:       stepState.CurrentStudentID,
		StudyPlanItemId: stepState.StudyPlanItemID,
		SessionId:       stepState.SessionID,
		Paging: &cpb.Paging{
			Limit: uint32(5),
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: 1,
			},
		},
	}

	resp, err := epb.NewQuizModifierServiceClient(s.EurekaConn).CreateQuizTest(s.AuthHelper.SignedCtx(ctx, stepState.Token), request)
	if err != nil {
		stepState.ResponseErr = err
		return utils.StepStateToContext(ctx, stepState), nil
	}

	stepState.ShuffledQuizSetID = resp.QuizzesId

	answers := make([]*epb.QuizAnswer, 0, len(stepState.ExternalIDs))
	for _, externalID := range stepState.ExternalIDs {
		answer := &epb.QuizAnswer{
			QuizId: externalID,
			Answer: []*epb.Answer{
				{
					Format: &epb.Answer_SelectedIndex{
						SelectedIndex: 1,
					},
				},
			},
		}
		answers = append(answers, answer)
	}

	stepState.Response, err = epb.NewCourseModifierServiceClient(s.EurekaConn).SubmitQuizAnswers(s.AuthHelper.SignedCtx(ctx, stepState.Token), &epb.SubmitQuizAnswersRequest{
		SetId:      stepState.ShuffledQuizSetID,
		QuizAnswer: answers,
	})
	if err != nil {
		stepState.ResponseErr = err
		return utils.StepStateToContext(ctx, stepState), nil
	}

	var submission entities.ExamLOSubmission
	stmt := `SELECT submission_id, student_id, study_plan_id, learning_material_id FROM exam_lo_submission WHERE shuffled_quiz_set_id = $1::TEXT`
	if err := database.Select(ctx, s.EurekaDB, stmt, stepState.ShuffledQuizSetID).ScanOne(&submission); err != nil {
		stepState.ResponseErr = err
		return utils.StepStateToContext(ctx, stepState), nil
	}

	stepState.SubmissionID = submission.SubmissionID.String
	stepState.StudentID = submission.StudentID.String
	stepState.StudyPlanID = submission.StudyPlanID.String
	stepState.LearningMaterialID = submission.LearningMaterialID.String

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemMustReturnSubmitResultCorrectly(ctx context.Context, result string, numOfquizzes int, point int) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	submissionResult := stepState.Response.(*epb.SubmitQuizAnswersResponse).SubmissionResult
	if result != submissionResult.String() {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("wrong SubmissionResult expected: %v, got: %v", result, submissionResult.String())
	}

	totalPoint := stepState.Response.(*epb.SubmitQuizAnswersResponse).TotalPoint
	expectedPoint := stepState.TotalPoint
	if uint32(expectedPoint) != totalPoint.Value {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("wrong TotalPoint expected: %v, got: %v", expectedPoint, totalPoint)
	}

	totalQuestion := stepState.Response.(*epb.SubmitQuizAnswersResponse).TotalQuestion
	if numOfquizzes != int(totalQuestion) {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("wrong TotalQuestion expected: %v, got: %v", numOfquizzes, totalQuestion)
	}

	// TotalCorrectAnswer
	totalCorrectAnswer := stepState.Response.(*epb.SubmitQuizAnswersResponse).TotalCorrectAnswer
	totalGrainedPoint := stepState.Response.(*epb.SubmitQuizAnswersResponse).TotalGradedPoint
	if uint32(totalCorrectAnswer*int32(point)) != totalGrainedPoint.Value {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("wrong totalCorrectAnswer, totalGrainedPoint expected: %v, got: %v", totalCorrectAnswer*int32(point), totalGrainedPoint.Value)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

package eureka

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/s3"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *suite) schoolAdminAddStudentToACourseHaveAStudyPlan(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.AuthToken = stepState.SchoolAdminToken
	ctx, err := s.userCreateACourseWithAStudyPlan(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.userCreateACourseWithAStudyPlan: %w", err)
	}
	ctx, err = s.addStudentToCourse(ctx, entities.UserGroupStudent)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.addStudentToCourse: %w", err)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userCreateQuizTest(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx = contextWithToken(s, ctx)
	stepState.TopicIDs = append(stepState.TopicIDs, stepState.TopicID)

	stepState.Response, stepState.ResponseErr = pb.NewQuizModifierServiceClient(s.Conn).CreateQuizTest(ctx, &pb.CreateQuizTestRequest{
		LoId:      stepState.LoID,
		StudentId: stepState.StudentID,
		Paging: &cpb.Paging{
			Limit: 10,
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: 1,
			},
		},
	})
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentCreateQuizTest(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.AuthToken = stepState.StudentToken
	stepState.TopicIDs = append(stepState.TopicIDs, stepState.TopicID)
	ctx, err := s.userGetListTodoItemsByTopics(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.userGetListTodoItemsByTopics: %w", err)
	}
	resp := stepState.Response.(*pb.ListToDoItemsByTopicsResponse)
	res, err := pb.NewQuizModifierServiceClient(s.Conn).CreateQuizTest(s.signedCtx(ctx), &pb.CreateQuizTestRequest{
		StudyPlanItemId: resp.GetItems()[0].GetTodoItems()[0].GetStudyPlanItem().StudyPlanItemId,
		LoId:            stepState.LoID,
		StudentId:       stepState.StudentID,
		Paging: &cpb.Paging{
			Limit: 10,
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: 1,
			},
		},
		KeepOrder: true,
	})
	stepState.Response, stepState.ResponseErr = res, err
	if res != nil {
		stepState.ShuffledQuizSetID = res.QuizzesId
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) ourSystemMustReturnsQuizzesCorrectly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp := stepState.Response.(*pb.CreateQuizTestResponse)
	items := resp.GetItems()
	for i, item := range items {
		if stepState.QuizLOList[i].Quiz.GetPoint() != nil {
			if item.Core.Point.Value != stepState.QuizLOList[i].Quiz.Point.Value {
				return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected quiz point: want %v, got %v", stepState.QuizLOList[i].Quiz.Point.Value, item.Core.Point.Value)
			}
		} else {
			var defaultPoint int32 = 1
			if item.Core.Point.Value != defaultPoint {
				return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected quiz point: want %v, got %v", defaultPoint, item.Core.Point.Value)
			}
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) quizTestHaveQuestionHierarchy(ctx context.Context) (context.Context, error) {
	// check data in question_hierarchy column of quiz_sets table
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return ctx, nil
	}

	query := fmt.Sprintf(`
		SELECT question_hierarchy FROM shuffled_quiz_sets
		WHERE shuffled_quiz_set_id = $1
			AND deleted_at IS NULL`)
	questionHierarchy := &pgtype.JSONBArray{}
	err := s.DB.QueryRow(ctx, query, stepState.ShuffledQuizSetID).Scan(questionHierarchy)
	if err != nil {
		return ctx, fmt.Errorf("db.QueryRow: %w", err)
	}
	actualQuestionHierarchy := make([]*entities.QuestionHierarchyObj, 0)
	questionHierarchy.AssignTo(&actualQuestionHierarchy)

	// compare between question hierarchy of expected state and actual in db
	if length := len(actualQuestionHierarchy); length != len(stepState.ExistingQuestionHierarchy) {
		return ctx, fmt.Errorf("expected %d item in question_hierarchy column of shuffled_quiz_sets but got %d", len(stepState.ExistingQuestionHierarchy), length)
	}

	for i := range actualQuestionHierarchy {
		actual := actualQuestionHierarchy[i]
		expected := stepState.ExistingQuestionHierarchy[i]

		if actual.Type != expected.Type {
			return ctx, fmt.Errorf("expected type of item question_hierarchy %d is %s but got %s", i, expected.Type, actual.Type)
		}
		if actual.ID != expected.ID {
			return ctx, fmt.Errorf("expected id of item question_hierarchy %d is %s but got %s", i, expected.ID, actual.ID)
		}
		if len(actual.ChildrenIDs) != len(expected.ChildrenIDs) {
			return ctx, fmt.Errorf("expected children of item question_hierarchy %d is %d but got %d", i, len(expected.ChildrenIDs), len(actual.ChildrenIDs))
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) quizBelongToQuestionGroup(ctx context.Context, numItems int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	token := stepState.AuthToken
	ctx, err := s.aSignedIn(ctx, "school admin")
	if err != nil {
		return ctx, err
	}

	questionGrIDs := stepState.ExistingQuestionHierarchy.GetQuestionGroupIDs()
	quizLO := utils.GenerateQuizLOProtobufMessage(numItems*len(questionGrIDs), stepState.LoID)
	for i, id := range questionGrIDs {
		quizLO[i*2].Quiz.QuestionGroupId, quizLO[i*2+1].Quiz.QuestionGroupId = wrapperspb.String(id), wrapperspb.String(id)
		if err = stepState.ExistingQuestionHierarchy.AddChildrenIDsForQuestionGroup(id, quizLO[i*2].Quiz.ExternalId, quizLO[i*2+1].Quiz.ExternalId); err != nil {
			return ctx, fmt.Errorf("ExistingQuestionHierarchy.AddChildrenIDsForQuestionGroup: %w", err)
		}
	}

	ctx = s.signedCtx(ctx)
	if _, err := utils.UpsertQuizzes(ctx, s.Conn, quizLO); err != nil {
		return ctx, fmt.Errorf("UpsertQuizzes: %w", err)
	}
	stepState.QuizLOList = append(stepState.QuizLOList, quizLO...)
	stepState.AuthToken = token
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetQuizTestResponse(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	type questionWithGroup struct {
		ID              string
		QuestionGroupID string
	}
	expectedQuiz := make([]*questionWithGroup, 0)
	expectedQuestionGroup := make([]*cpb.QuestionGroup, 0)
	url, _ := s3.GenerateUploadURL("", "", "rendered rich text")
	for _, v := range stepState.ExistingQuestionHierarchy {
		if v.Type == entities.QuestionHierarchyQuestion {
			expectedQuiz = append(expectedQuiz, &questionWithGroup{
				ID: v.ID,
			})
		} else {
			var totalPoints int32
			for _, id := range v.ChildrenIDs {
				expectedQuiz = append(expectedQuiz, &questionWithGroup{
					ID:              id,
					QuestionGroupID: v.ID,
				})
				quiz := s.getStepStateQuizLOByID(ctx, id)
				if quiz == nil {
					return ctx, fmt.Errorf("could not found quiz id %s in step state", id)
				}
				totalPoints += quiz.Quiz.Point.GetValue()
			}
			expectedQuestionGroup = append(expectedQuestionGroup, &cpb.QuestionGroup{
				QuestionGroupId:    v.ID,
				LearningMaterialId: stepState.LoID,
				TotalChildren:      int32(len(v.ChildrenIDs)),
				TotalPoints:        totalPoints,
				RichDescription: &cpb.RichText{
					Raw:      "raw rich text",
					Rendered: url,
				},
			})
		}
	}

	// compare between response's data and expected data
	var (
		items          []*cpb.Quiz
		questionGroups []*cpb.QuestionGroup
	)

	// nolint
	switch resp := stepState.Response.(type) {
	case *pb.CreateQuizTestResponse:
		items = resp.GetItems()
		questionGroups = resp.GetQuestionGroups()
	case *pb.CreateRetryQuizTestResponse:
		items = resp.GetItems()
		questionGroups = resp.GetQuestionGroups()
	default:
		return StepStateToContext(ctx, stepState), fmt.Errorf("invalid response type")
	}

	for _, item := range items {
		if len(stepState.GroupedQuizzes) != 0 && !sliceutils.Contains(stepState.GroupedQuizzes, item.Core.ExternalId) {
			return StepStateToContext(ctx, stepState), nil
		}
	}

	if len(expectedQuiz) != len(items) {
		return ctx, fmt.Errorf("expected %d quiz in response but got %d", len(expectedQuiz), len(items))
	}

	for i := range items {
		if expectedQuiz[i].ID != items[i].Core.ExternalId {
			return ctx, fmt.Errorf("expected in position %d in quiz field response is id %s but got %s", i, expectedQuiz[i].ID, items[i].Core.ExternalId)
		}

		actGrID := ""
		if items[i].Core.GetQuestionGroupId() != nil {
			actGrID = items[i].Core.GetQuestionGroupId().Value
		}
		if expectedQuiz[i].QuestionGroupID != actGrID {
			return ctx, fmt.Errorf("expected quiz %s have question group id is %s but got %s", expectedQuiz[i].ID, expectedQuiz[i].QuestionGroupID, actGrID)
		}
	}

	if len(expectedQuestionGroup) != len(questionGroups) {
		return ctx, fmt.Errorf("expected %d questions in response but got %d", len(expectedQuestionGroup), len(questionGroups))
	}
	for i := range questionGroups {
		if expectedQuestionGroup[i].QuestionGroupId != questionGroups[i].QuestionGroupId {
			return ctx, fmt.Errorf("expected in position %d in question group field response is id %s but got %s", i, expectedQuestionGroup[i].QuestionGroupId, questionGroups[i].QuestionGroupId)
		}

		if expectedQuestionGroup[i].LearningMaterialId != questionGroups[i].LearningMaterialId {
			return ctx, fmt.Errorf("expected question group %s have learning material id is %s but got %s", expectedQuestionGroup[i].QuestionGroupId, expectedQuestionGroup[i].LearningMaterialId, questionGroups[i].LearningMaterialId)
		}

		if expectedQuestionGroup[i].TotalChildren != questionGroups[i].TotalChildren {
			return ctx, fmt.Errorf("expected question group %s have total children is %d but got %d", expectedQuestionGroup[i].QuestionGroupId, expectedQuestionGroup[i].TotalChildren, questionGroups[i].TotalChildren)
		}

		if expectedQuestionGroup[i].TotalPoints != questionGroups[i].TotalPoints {
			return ctx, fmt.Errorf("expected question group %s have total ponts is %d but got %d", expectedQuestionGroup[i].QuestionGroupId, expectedQuestionGroup[i].TotalPoints, questionGroups[i].TotalPoints)
		}

		if expectedQuestionGroup[i].RichDescription.Raw != questionGroups[i].RichDescription.Raw {
			return ctx, fmt.Errorf("expected question group %s have total raw rich description is %s but got %s", expectedQuestionGroup[i].QuestionGroupId, expectedQuestionGroup[i].RichDescription.Raw, questionGroups[i].RichDescription.Raw)
		}

		if !strings.Contains(questionGroups[i].RichDescription.Rendered, strings.ReplaceAll(expectedQuestionGroup[i].RichDescription.Rendered, "//", "")) {
			return ctx, fmt.Errorf("expected question group %s have total rendered rich description is %s but got %s", expectedQuestionGroup[i].QuestionGroupId, expectedQuestionGroup[i].RichDescription.Rendered, questionGroups[i].RichDescription.Rendered)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) getStepStateQuizLOByID(ctx context.Context, id string) *pb.QuizLO {
	stepState := StepStateFromContext(ctx)
	for _, quiz := range stepState.QuizLOList {
		if quiz.GetQuiz().ExternalId == id {
			return quiz
		}
	}

	return nil
}

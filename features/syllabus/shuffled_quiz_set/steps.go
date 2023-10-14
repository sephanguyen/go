package shuffled_quiz_set

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/syllabus/entity"
	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/eureka/entities"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type StepState struct {
	StartDate              *timestamppb.Timestamp
	Token                  string
	Response               interface{}
	Request                interface{}
	ResponseErr            error
	BookID                 string
	StudyPlanID            string
	CourseID               string
	TopicIDs               []string
	TopicID                string
	LoID                   string
	ChapterID              string
	SchoolIDInt            int32
	QuizID                 string
	ShuffledQuizSetID      string
	RetryShuffledQuizSetID string
	QuizLOList             []*epb.QuizLO
	SchoolAdmin            entity.SchoolAdmin
	Student                entity.Student
	Teacher                entity.Teacher
	Parent                 entity.Parent
	HQStaff                entity.HQStaff
	QuestionGroupID        string

	ExistingQuestionHierarchy entities.QuestionHierarchy

	StudyPlanItemIdentities []*sspb.StudyPlanItemIdentity
}

func InitStep(s *Suite) map[string]interface{} {
	steps := map[string]interface{}{
		`^<shuffled_quiz_set> a valid book content$`:               s.aValidBookContent,
		`^school admin add student to a course have a study plan$`: s.schoolAdminAddStudentToACourseHaveAStudyPlan,
		`^user create a quiz using v(\d+)$`:                        s.userCreateAQuizUsingV2,
		`^user create quiz test v(\d+)$`:                           s.userCreateQuizTestV2,
		`^<shuffled_quiz_set> a signed in "([^"]*)"$`:              s.aSignedIn,
		`^<shuffled_quiz_set> returns "([^"]*)" status code$`:      s.returnsStatusCode,
		`^shuffled quiz test have been stored$`:                    s.shuffledQuizTestHaveBeenStored,
		`^user create retry quiz test v(\d+)$`:                     s.userCreateRetryQuizTestV,
		`^retry shuffled quiz test have been stored$`:              s.retryShuffledQuizTestHaveBeenStored,
		`^<(\d+)> existing question group$`:                        s.existingQuestionGroups,
		`^<(\d+)> existing questions$`:                             s.existingQuestions,
		`^user got quiz test response$`:                            s.userGetQuizTestResponse,
		`^<(\d+)> quiz belong to question group$`:                  s.quizBelongToQuestionGroup,

		`^a quizset with "([^"]*)" quizzes using v(\d+)$`:          s.validQuizSet,
		`^"([^"]*)" students do test of a study plan$`:             s.studentDoQuizTestSuccess,
		`^teacher get quiz test of a study plan$`:                  s.teacherGetQuizTestByStudyPlanItemIdentity,
		`^teacher get quiz test without study plan item identity$`: s.teacherGetQuizTestWithoutStudyPlanItemIdentity,
		`^get quiz test of a study plan by "([^"]*)"$`:             s.retrieveQuizTestsV2ByRole,
		`^"([^"]*)" quiz tests infor$`:                             s.quizTestsInfo,
	}
	return steps
}

func (s *Suite) aSignedIn(ctx context.Context, arg string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	// reset token
	stepState.Token = ""
	userID, authToken, err := s.AuthHelper.AUserSignedInAsRole(ctx, arg)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	//TODO: no need if you're not use it. Just an example.
	switch arg {
	case "student":
		stepState.Student.Token = authToken
		stepState.Student.ID = userID
	case "school admin", "admin":
		stepState.SchoolAdmin.Token = authToken
		stepState.SchoolAdmin.ID = userID
	case "teacher", "current teacher":
		stepState.Teacher.Token = authToken
		stepState.Teacher.ID = userID
	case "parent":
		stepState.Parent.Token = authToken
		stepState.Parent.ID = userID
	case "hq staff":
		stepState.HQStaff.Token = authToken
		stepState.HQStaff.ID = userID
	default:
		stepState.Parent.Token = authToken
		stepState.Parent.ID = userID
	}
	stepState.Token = authToken
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) returnsStatusCode(ctx context.Context, arg string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	return utils.StepStateToContext(ctx, stepState), utils.ValidateStatusCode(stepState.ResponseErr, arg)
}

func (s *Suite) aValidBookContent(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	bookResp, err := epb.NewBookModifierServiceClient(s.EurekaConn).UpsertBooks(s.AuthHelper.SignedCtx(ctx, stepState.Token), &epb.UpsertBooksRequest{
		Books: utils.GenerateBooks(1, nil),
	})
	if err != nil {
		if err.Error() == "rpc error: code = PermissionDenied desc = auth: not allowed" {
			stepState.ResponseErr = err
			return utils.StepStateToContext(ctx, stepState), nil
		}
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to create book: %w", err)
	}
	stepState.BookID = bookResp.BookIds[0]
	chapterResp, err := epb.NewChapterModifierServiceClient(s.EurekaConn).UpsertChapters(s.AuthHelper.SignedCtx(ctx, stepState.Token), &epb.UpsertChaptersRequest{
		Chapters: utils.GenerateChapters(stepState.BookID, 1, nil),
		BookId:   stepState.BookID,
	})
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to create chapter: %w", err)
	}
	stepState.ChapterID = chapterResp.ChapterIds[0]
	topicResp, err := epb.NewTopicModifierServiceClient(s.EurekaConn).Upsert(s.AuthHelper.SignedCtx(ctx, stepState.Token), &epb.UpsertTopicsRequest{
		Topics: utils.GenerateTopics(stepState.ChapterID, 1, nil),
	})
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to create topic: %w", err)
	}
	stepState.TopicID = topicResp.TopicIds[0]
	stepState.TopicIDs = append(stepState.TopicIDs, stepState.TopicID)
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) existingQuestionGroups(ctx context.Context, numItems int) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	token := stepState.Token
	ctx, err := s.aSignedIn(ctx, "school admin")
	if err != nil {
		return ctx, err
	}
	ctx = s.AuthHelper.SignedCtx(ctx, stepState.Token)

	for i := 0; i < numItems; i++ {
		res, err := utils.InsertANewQuestionGroup(ctx, s.EurekaConn, stepState.LoID)
		if err != nil {
			return ctx, fmt.Errorf("InsertANewQuestionGroup: %w", err)
		}
		stepState.ExistingQuestionHierarchy.AddQuestionGroupID(res.QuestionGroupId)
	}

	stepState.Token = token
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) existingQuestions(ctx context.Context, numItems int) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	token := stepState.Token
	ctx, err := s.aSignedIn(ctx, "school admin")
	if err != nil {
		return ctx, err
	}
	ctx = s.AuthHelper.SignedCtx(ctx, stepState.Token)

	quizLO := utils.GenerateQuizLOProtobufMessage(numItems, stepState.LoID)
	for i := 0; i < numItems; i++ {
		if _, err := utils.UpsertQuizzes(ctx, s.EurekaConn, quizLO); err != nil {
			return ctx, fmt.Errorf("UpsertQuizzes: %w", err)
		}
		stepState.ExistingQuestionHierarchy.AddQuestionID(quizLO[i].GetQuiz().ExternalId)
	}

	stepState.QuizLOList = append(stepState.QuizLOList, quizLO...)
	stepState.Token = token
	return utils.StepStateToContext(ctx, stepState), nil
}

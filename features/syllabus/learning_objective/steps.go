package learning_objective

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/syllabus/entity"
	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
)

type StepState struct {
	Token                      string
	Response                   interface{}
	Request                    interface{}
	ResponseErr                error
	BookID                     string
	TopicIDs                   []string
	ChapterIDs                 []string
	SchoolAdmin                entity.SchoolAdmin
	StudentIDs                 []string
	LearningMaterialID         string
	MapLearningMaterial        map[string]*entity.LearningMaterialPb
	LearningMaterialIDs        []string
	CurrentLM                  *entity.LearningMaterialPb
	UserId                     string
	SchoolAdminToken           string
	StudentToken               string
	LearningObjectiveID        string
	TopicLODisplayOrderCounter int32
	QuizID                     string
	ExistingQuestionHierarchy  entities.QuestionHierarchy
	QuestionGroupID            string
	ExternalIDs                []string
	CourseID                   string
	CourseStudents             []*entities.CourseStudent
	StudyPlanID                string
	QuizLOList                 []*epb.QuizLO
	ShuffledQuizSetID          string
	SessionID                  string
}

func InitStep(s *Suite) map[string]interface{} {
	steps := map[string]interface{}{
		`^<learning_objective>a signed in "([^"]*)"$`:         s.aSignedIn,
		`^<learning_objective>a valid book content$`:          s.aValidBookContent,
		`^<learning_objective>returns "([^"]*)" status code$`: s.returnsStatusCode,
		// insert learning_objective
		`^our system generates a correct display order for learning objective$`: s.ourSystemGeneratesACorrectDisplayOrderForLearningObjective,
		`^there are learning objectives existed in topic$`:                      s.thereAreLearningObjectivesExistedInTopic,
		`^user inserts a learning objective$`:                                   s.userInsertALearningObjective,
		`^user inserts a learning objective with "([^"]*)"`:                     s.userInsertALearningObjectiveWithField,
		`^our system must create LO with "([^"]*)"`:                             s.ourSystemMustCreateLOWithValue,

		// update learning_objective
		`^our system updates the learning objective correctly$`:                            s.ourSystemUpdatesTheLearningObjectiveCorrectly,
		`^user updates a learning objective$`:                                              s.userUpdatesALearningObjective,
		`^our system updates topic display order counter of learning objective correctly$`: s.ourSystemUpdatesTopicDisplayOrderCounterOfLearningObjectiveCorrectly,

		// list learning_objective
		`^user list learning objectives$`:                                    s.userListLearningObjectives,
		`^our system must return learning objectives correctly$`:             s.ourSystemMustReturnLearningObjectivesCorrectly,
		`^a valid learning objective with quizzes by learning material ids$`: s.validLearningObjectivesQuizzesByLoIDs,
		`^our system must return learning objective has a total question$`:   s.ourSystemMustReturnLOHasTotalQuestion,

		// question group
		`^insert a new question group$`:                                           s.insertANewQuestionGroup,
		`^new question group was added at the end of list$`:                       s.newQuestionGroupWasAddedAtTheEndList,
		`^new question group was added$`:                                          s.newQuestionGroupWasAddedAtTheEndList,
		`^question group was updated`:                                             s.questionGroupWasUpdated,
		`^a learning object$`:                                                     s.userInsertALearningObjective,
		`^a "([^"]*)" quiz$`:                                                      s.existingQuiz,
		`^existing question group$`:                                               s.existingQuestionGroup,
		`^update a question group$`:                                               s.updateAQuestionGroup,
		`^update display order of index (\d+) and index (\d+)$`:                   s.updateDisplayOrder,
		`^update display order of index (\d+) and index (\d+) in question group$`: s.updateDisplayOrderInQuestionGroup,
		`^new display order is updated$`:                                          s.newDisplayOrderIsUpdated,
		`^delete existing question group$`:                                        s.deleteExistingQuestionGroup,
		`^question group is deleted$`:                                             s.questionGroupDeleted,

		// upsert learning objective progression
		`^student upsert lo progression with (\d+) answers$`: s.upsertLOProgressionWithAnswers,

		// retrieve lo progression
		`^<learning_objective>user creates a course and add students into the course$`: s.userCreatesACourseAndAddStudentsIntoTheCourse,
		`^<learning_objective>user adds a master study plan with the created book$`:    s.userAddsAMasterStudyPlanWithTheCreatedBook,
		`^<learning_objective>user create quizzes$`:                                    s.createQuizzes,
		`^<learning_objective>user create quiz test v2$`:                               s.userCreateQuizTestV2,
		`^there is exam LO existed in topic$`:                                          s.thereIsExamLOExistedInTopic,
		`^student list lo progression$`:                                                s.listLOProgression,
		`^there are (\d+) answers in the response$`:                                    s.thereAreAnswers,
	}

	return steps
}

func (s *Suite) aSignedIn(ctx context.Context, arg string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	// reset token
	stepState.Token = ""
	_, authToken, err := s.AuthHelper.AUserSignedInAsRole(ctx, arg)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
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
	bookID, chapterIDs, topicIDs, err := utils.AValidBookContent(s.AuthHelper.SignedCtx(ctx, stepState.Token), s.EurekaConn, s.EurekaDB, constant.ManabieSchool)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("AValidBookContent: %w", err)
	}
	stepState.BookID = bookID
	stepState.ChapterIDs = chapterIDs
	stepState.TopicIDs = topicIDs
	return utils.StepStateToContext(ctx, stepState), nil
}

package flashcard

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/manabie-com/backend/features/syllabus/entity"
	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	"github.com/manabie-com/backend/internal/yasuo/entities"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"
)

type StepState struct {
	Token                                  string
	AdminToken                             string
	StudentToken                           string
	TeacherToken                           string
	UserID                                 string
	StudentID                              string
	Response                               interface{}
	Request                                interface{}
	ResponseErr                            error
	BookID                                 string
	TopicIDs                               []string
	ChapterIDs                             []string
	CourseID                               string
	StudyPlanID                            string
	SchoolAdmin                            entity.SchoolAdmin
	Student                                entity.Student
	LearningMaterialID                     string
	FlashcardID                            string
	StudySetID                             string
	TopicLODisplayOrderCounter             int32
	LearningMaterialIDs                    []string
	LatestFlashcardStudyProgressStudySetId string
	QuizID                                 string
	QuizFlashcardList                      []*cpb.QuizCore
	OldSpeeches                            []*entities.Speeches
}

func InitStep(s *Suite) map[string]interface{} {
	steps := map[string]interface{}{
		`^<flashcard>a signed in "([^"]*)"$`:                          s.aSignedIn,
		`^<flashcard>an exists "([^"]*)" signed in$`:                  s.anExistsSignedIn,
		`^<flashcard>a valid book content$`:                           s.aValidBookContent,
		`^<flashcard>returns "([^"]*)" status code$`:                  s.returnsStatusCode,
		`^a valid flashcard with quizzes$`:                            s.aValidFlashcardWithQuizzes,
		`^<flashcard>a course and study plan with "([^"]*)" student$`: s.aCourseAndStudyPlanWithStudent,

		// insert flashcard
		`^our system generates a correct display order for flashcard$`: s.ourSystemGeneratesACorrectDisplayOrderForFlashcard,
		`^our system updates topic LODisplayOrderCounter correctly$`:   s.ourSystemUpdatesTopicLODisplayOrderCounterCorrectly,
		`^user inserts a flashcard$`:                                   s.userInsertsAFlashcard,
		`^there are flashcards existed in topic$`:                      s.thereAreFlashcardsExistedInTopic,
		`^user inserts a lmsv2 flashcard$`:                             s.userInsertsALmsv2Flashcard,
		`^the lmsv2 flashcard is created with correct data$`:           s.theLmsv2FlashcardIsCreatedWithCorrectData,
		// update flashcard
		`^user updates a flashcard$$`:                  s.userUpdateAFlashcard,
		`^our system updates the flashcard correctly$`: s.ourSystemUpdatesTheFlashcardCorrectly,
		// list flashcard
		`^user list flashcard$`:                                     s.userListFlashcard,
		`^our system must return flashcards correctly$`:             s.ourSystemMustReturnFlashcardsCorrectly,
		`^a valid flashcard with quizzes by learning material ids$`: s.validFCQuizzesByLoIDs,
		`^our system must return flashcards has a total question$`:  s.ourSystemMustReturnFCHasTotalQuestion,

		// create flashcard study
		`^user create flashcard study$`:                        s.userCreateFlashcardStudy,
		`^our system creates flashcard progression correctly$`: s.ourSystemCreatesFlashcardProgressionCorrectly,

		`^returns latest flashcard study progress correctly$`: s.returnsLatestFlashcardStudyProgressCorrectly,
		`^user create some flashcard studies$`:                s.userCreateSomeFlashcardStudies,
		`^user get latest flashcard study progress$`:          s.userGetLatestFlashcardStudyProgress,

		// finish flashcard study
		`^user finish flashcard study with "([^"]*)"$`:                  s.userFinishFlashcardStudyWith,
		`^our system updates flashcard study correctly with "([^"]*)"$`: s.ourSystemUpdatesFlashcardStudyCorrectlyWith,
		`^user create a flashcard content$`:                             s.userCreateAFlashcardContent,
		`^user create a flashcard content with "([^"]*)"$`:              s.userCreateAFlashcardContentWith,
		`^options and speeches updated correctly$`:                      s.optionsAndSpeechesUpdatedCorrectly,
		`^regenerate speeches audio link$`:                              s.regenerateSpeechesAudioLink,
	}

	return steps
}

func (s *Suite) aSignedIn(ctx context.Context, role string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	// reset token
	stepState.Token = ""
	userID, authToken, err := s.AuthHelper.AUserSignedInAsRole(ctx, role)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	stepState.UserID = userID
	switch role {
	case "school admin":
		stepState.AdminToken = authToken
	case "student":
		stepState.StudentToken = authToken
		stepState.StudentID = userID
	case "teacher":
		stepState.TeacherToken = authToken
	}
	stepState.Token = authToken
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) anExistsSignedIn(ctx context.Context, role string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	switch role {
	case "student":
		if stepState.StudentToken == "" {
			return s.aSignedIn(ctx, role)
		}
		stepState.Token = stepState.StudentToken
		stepState.UserID = stepState.StudentID

	case "school admin":
		if stepState.AdminToken == "" {
			return s.aSignedIn(ctx, role)
		}
		stepState.Token = stepState.AdminToken

	case "teacher":
		if stepState.TeacherToken == "" {
			return s.aSignedIn(ctx, role)
		}
		stepState.Token = stepState.TeacherToken

	default:
		return s.aSignedIn(ctx, role)
	}

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

func (s *Suite) aValidFlashcardWithQuizzes(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	// create flashcard
	flashcardReq := &sspb.InsertFlashcardRequest{
		Flashcard: &sspb.FlashcardBase{
			Base: &sspb.LearningMaterialBase{
				TopicId: stepState.TopicIDs[0],
				Name:    fmt.Sprintf("flashcard-name"),
			},
		},
	}
	resp, err := sspb.NewFlashcardClient(s.EurekaConn).InsertFlashcard(s.AuthHelper.SignedCtx(ctx, stepState.Token), flashcardReq)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("sspb.NewFlashcardClient(s.EurekaConn).InsertFlashcard: %w", err)
	}
	stepState.FlashcardID = resp.LearningMaterialId

	// create quizzes
	if err := utils.GenerateQuizzes(s.AuthHelper.SignedCtx(ctx, stepState.Token), stepState.FlashcardID, rand.Intn(7)+3, nil, s.EurekaConn); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("utils.GenerateQuizzes: %w", err)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

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
	switch option {
	case "current":
		if _, err := epb.NewAssignmentModifierServiceClient(s.EurekaConn).AssignStudyPlan(s.AuthHelper.SignedCtx(ctx, stepState.AdminToken), &epb.AssignStudyPlanRequest{
			StudyPlanId: stepState.StudyPlanID,
			Data: &epb.AssignStudyPlanRequest_StudentId{
				StudentId: stepState.StudentID,
			},
		}); err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("epb.NewAssignmentModifierServiceClient(s.EurekaConn).AssignStudyPlan: %w", err)
		}
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

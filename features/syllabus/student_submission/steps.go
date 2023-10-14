package student_submission

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/manabie-com/backend/features/syllabus/entity"
	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/constants"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type StepState struct {
	Token                      string
	Response                   interface{}
	Request                    interface{}
	ResponseErr                error
	BookID                     string
	StudentID                  string
	StudyPlanID                string
	CourseID                   string
	TopicID                    string
	TopicIDs                   []string
	ChapterID                  string
	ChapterIDs                 []string
	SchoolAdmin                entity.SchoolAdmin
	Student                    entity.Student
	Teacher                    entity.Teacher
	Parent                     entity.Parent
	HQStaff                    entity.HQStaff
	LearningMaterialID         string
	FlashcardID                string
	TopicLODisplayOrderCounter int32
	LearningMaterialIDs        []string
	StudyPlanItemIDs           []string
	AssignmentIDs              []string
	LoIDs                      []string
	StudentIDs                 []string
	CourseStudents             []*entities.CourseStudent
	LocationIDs                []string
	SubmissionIDs              []string
	AssignmentID               string
	StudyPlanIdentities        []*sspb.StudyPlanItemIdentity
	ShuffleQuizSetID           string
}

func InitStep(s *Suite) map[string]interface{} {
	steps := map[string]interface{}{
		`^<student_submission> a signed in "([^"]*)"$`:                                 s.aSignedIn,
		`^<student_submission> a valid book content$`:                                  s.aValidBookContent,
		`^<student_submission> a valid book content with "([^"]*)"$`:                   s.aValidBookContentWith,
		`^<student_submission> returns "([^"]*)" status code$`:                         s.returnsStatusCode,
		`^students submit their assignments$`:                                          s.studentsSubmitTheirAssignments,
		`^user using list submissions v(\d+)$`:                                         s.userUsingListSubmissionsV,
		`^returns list student submissions correctly$`:                                 s.returnsListStudentSubmissionsCorrectly,
		`^<student_submission> some students added to course in some valid locations$`: s.someStudentsAddedToCourseInSomeValidLocations,
		`^create a study plan for that course$`:                                        s.createAStudyPlanForThatCourse,

		`^user retrieve submissions$`:              s.userRetrieveSubmissions,
		`^retrieve student submissions correctly$`: s.retrieveStudentSubmissionsCorrectly,

		`^student do test of "([^"]*)"$`:                          s.studentDoTestOf,
		`^user retrieve student submission history$`:              s.userRetrieveStudentSubmissionHistory,
		`^our system returns correct student submission history$`: s.ourSystemReturnsCorrectStudentSubmissionHistory,

		`^user using list submissions v(\d+) without student name$`:        s.userUsingListSubmissionsV4,
		`^user using list submissions v(\d+) with "([^"]*)" student name$`: s.userUsingListSubmissionsVWithStudentName,
		`^returns list student submissions correctly with "([^"]*)"$`:      s.returnsListStudentSubmissionsCorrectlyWith,

		`^user using list submissions ver2 with "([^"]*)" student name$`: s.userUsingListSubmissionsV2WithStudentName,
		`^returns list student submissions v2 correctly with "([^"]*)"$`: s.returnsListStudentSubmissionsV2CorrectlyWith,
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
	bookID, chapterIDs, topicIDs, err := utils.AValidBookContent(s.AuthHelper.SignedCtx(ctx, stepState.Token), s.EurekaConn, s.EurekaDB, constants.ManabieSchool)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("AValidBookContent: %w", err)
	}
	stepState.BookID = bookID
	stepState.ChapterIDs = chapterIDs
	stepState.ChapterID = chapterIDs[0]
	stepState.TopicIDs = topicIDs
	stepState.TopicID = topicIDs[0]

	numOfLos := rand.Intn(5) + 1
	los := make([]*cpb.LearningObjective, 0, numOfLos)
	for i := 0; i < numOfLos; i++ {
		lo := utils.GenerateLearningObjective(stepState.TopicIDs[0])
		stepState.LoIDs = append(stepState.LoIDs, lo.Info.Id)
		los = append(los, lo)
	}

	if _, err := epb.NewLearningObjectiveModifierServiceClient(s.EurekaConn).UpsertLOs(s.AuthHelper.SignedCtx(ctx, stepState.Token),
		&epb.UpsertLOsRequest{
			LearningObjectives: los,
		}); err != nil {
		if e, ok := status.FromError(err); ok && e.Code() == codes.PermissionDenied {
			stepState.ResponseErr = err
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("AValidBookContent: %w", err)
		}
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable create los: %w", err)
	}

	resp, err := sspb.NewTaskAssignmentClient(s.EurekaConn).InsertTaskAssignment(s.AuthHelper.SignedCtx(ctx, stepState.Token), &sspb.InsertTaskAssignmentRequest{
		TaskAssignment: &sspb.TaskAssignmentBase{
			Base: &sspb.LearningMaterialBase{
				TopicId: stepState.TopicID,
				Name:    "name",
			},
		},
	})
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable insert task assignment: %w", err)
	}
	stepState.AssignmentID = resp.LearningMaterialId
	stepState.LearningMaterialIDs = append(stepState.LearningMaterialIDs, stepState.AssignmentID)

	courseID, err := utils.GenerateCourse(s.AuthHelper.SignedCtx(ctx, stepState.Token), s.YasuoConn)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	stepState.CourseID = courseID

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) aValidBookContentWith(ctx context.Context, lmType string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	ctx = s.AuthHelper.SignedCtx(ctx, stepState.Token)
	// create book, topic
	bookID, chapterIDs, topicIDs, err := utils.AValidBookContent(ctx, s.EurekaConn, s.EurekaDB, constants.ManabieSchool)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("AValidBookContent: %w", err)
	}
	stepState.BookID = bookID
	stepState.ChapterIDs = chapterIDs
	stepState.ChapterID = chapterIDs[0]
	stepState.TopicIDs = topicIDs
	stepState.TopicID = topicIDs[0]

	loType := cpb.LearningObjectiveType_LEARNING_OBJECTIVE_TYPE_NONE
	switch lmType {
	case "learning objective":
		loType = cpb.LearningObjectiveType_LEARNING_OBJECTIVE_TYPE_LEARNING
	case "flashcard":
		loType = cpb.LearningObjectiveType_LEARNING_OBJECTIVE_TYPE_FLASH_CARD
	}

	res, err := utils.GenerateLearningObjectivesV2(ctx, stepState.TopicID, 1, loType, nil, s.EurekaConn)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("utils.GenerateLearningObjectivesV2: %w", err)
	}

	stepState.LearningMaterialIDs = append(stepState.LearningMaterialIDs, res.LoIDs...)
	stepState.LearningMaterialID = res.LoIDs[0]

	switch lmType {
	case "learning objective":
		if _, err := utils.GenerateUpsertSingleQuiz(ctx, stepState.LearningMaterialID, cpb.QuizType_QUIZ_TYPE_MCQ, 1, 1, s.EurekaConn); err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("utils.GenerateUpsertSingleQuiz: %w", err)
		}
	case "flashcard":
		if _, err := utils.GenerateUpsertFlashcardContent(ctx, stepState.LearningMaterialID, 1, s.EurekaConn); err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("utils.GenerateUpsertFlashcardContent: %w", err)
		}
	}

	courseID, err := utils.GenerateCourse(ctx, s.YasuoConn)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	stepState.CourseID = courseID

	return utils.StepStateToContext(ctx, stepState), nil
}

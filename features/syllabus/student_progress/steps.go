package student_progress

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/syllabus/entity"
	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
)

type StepState struct {
	Response    interface{}
	Request     interface{}
	ResponseErr error
	BookID      string
	TopicIDs    []string
	ChapterIDs  []string
	Token       string
	SchoolAdmin entity.SchoolAdmin
	Student     entity.Student
	Teacher     entity.Teacher
	Parent      entity.Parent
	HQStaff     entity.HQStaff

	SchoolIDInt            int32
	Grade                  int32
	CourseID               string
	AssignedStudentIDs     []string
	ChapterID              string
	TopicID                string
	SkippedTopics          []string
	LoID                   string
	LoIDs                  []string
	FlashCardIDs           []string
	AssignmentIDs          []string
	TaskAssignmentIDs      []string
	ExamLoIDs              []string
	QuizIDs                []string
	SessionID              string
	StudyPlanID            string
	AvailableStudyPlanIDs  []string
	StudyPlanItemIDs       []string
	WrongQuizExternalIDs   []string
	NumTopics              int
	Assignments            []*pb.Assignment
	NumChapter             int
	NumQuizzes             int
	OldStudyPlanItemStatus pb.StudyPlanItemStatus

	NumberOfStudyPlan   int
	LearningMaterialIDs []string
	SubmissionIDs       []string
	CompletedLmIDs      []string
}

func InitStep(s *Suite) map[string]interface{} {
	steps := map[string]interface{}{
		`^<student_progress>a signed in "([^"]*)"$`:                   s.aSignedIn,
		`^<student_progress>returns "([^"]*)" status code$`:           s.returnsStatusCode,
		`^<student_progress>school admin, teacher and student login$`: s.schoolAdminTeacherAndStudentLogin,

		`^"([^"]*)" has created a book with each (\d+) los, (\d+) assignments, (\d+) topics, (\d+) chapters, (\d+) quizzes$`: s.hasCreatedABookWithEachLosAssignmentsTopicsChaptersQuizzes,

		`^<student_progress>individual study plan created$`: s.adminInsertIndividualStudyPlan,
		`^<student_progress>study plan assign to student$`:  s.userAssignStudyPlanToAStudent,

		`^"([^"]*)" do test and done "([^"]*)" los with "([^"]*)" correctly and "([^"]*)" assignments with "([^"]*)" point in the first two topics and done "([^"]*)" los with "([^"]*)" correctly and "([^"]*)" assignments with "([^"]*)" point in the other$`: s.doTestAndDoneLosWithCorrectlyAndAssignmentsWithPointInFourTopics,
		`^"([^"]*)" do test and done "([^"]*)" los with "([^"]*)" correctly and "([^"]*)" assignments with "([^"]*)" point and skip "([^"]*)" topics$`:                                                                                                           s.doTestAndDoneLosWithCorrectlyAndAssignmentsWithPointAndSkipTopics,
		`^topic score is "([^"]*)" and chapter score is "([^"]*)"$`:                                                     s.topicScoreIsAndChapterScoreIs,
		`^first pair topic score is "([^"]*)" and second pair topic score is "([^"]*)" and chapter score is "([^"]*)"$`: s.firstPairTopicScoreIsAndSecondPairTopicScoreIsAndChapterScoreIs,
		`^student calculate student progress$`:                                                                          s.studentCalculateStudentProgress,
		`^calculate student progress$`:                                                                                  s.calculateStudentProgress,
		`^correct lo completed with "([^"]*)" and "([^"]*)"$`:                                                           s.correctLoCompletedWithAnd,
		`^school admin delete "([^"]*)" topics$`:                                                                        s.schoolAdminDeleteTopics,
		`^our system must return learning material result and book tree correctly$`:                                     s.ourSystemMustReturnLearningMaterialResultAndBookTreeCorrectly,

		`^"([^"]*)" has created a book with each (\d+) los, (\d+) flashcard, (\d+) assignment, (\d+) task assignment, (\d+) exam los, (\d+) topics, (\d+) chapters, (\d+) quizzes$`:                                                         s.hasCreatedABookWithEachLearningMaterialType,
		`^"([^"]*)" do test and done "([^"]*)" los and "([^"]*)" flashcards with "([^"]*)" correctly, "([^"]*)" assignments and "([^"]*)" task assignments with "([^"]*)" point, "([^"]*)" with "([^"]*)" point and skip "([^"]*)" topics$`: s.doTestAndDoneLosFlashcardsWithCorrectlyAndAssignmentsTaskAssignmentsWithPointAndSkipTopics,
		`^correct lo completed with "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)" and "([^"]*)"$`:                                                                                                                                              s.correctLoCompletedWithFullLm,

		`^parent calculate student progress$`:                                 s.parentCalculateStudentProgress,
		`^<student_progress>school admin, parent, teacher and student login$`: s.schoolAdminParentTeacherAndStudentLogin,
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
		stepState.Student.Token = authToken
		stepState.Student.ID = userID
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

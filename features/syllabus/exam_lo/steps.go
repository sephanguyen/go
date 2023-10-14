package exam_lo

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/syllabus/entity"
	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"
)

type StepState struct {
	Token                      string
	Response                   interface{}
	Request                    interface{}
	ResponseErr                error
	LocationID                 string
	StudentIDs                 []string
	CourseID                   string
	CourseIDs                  []string
	BookID                     string
	BookIDs                    []string
	TopicID                    string
	TopicIDs                   []string
	Topics                     []*epb.Topic
	ChapterID                  string
	ChapterIDs                 []string
	CurrentChapterIDs          []string
	LearningMaterialIDs        []string
	LoID                       string
	LOIDs                      []string
	TopicLODisplayOrderCounter int32
	CorrectorID                string

	StudyPlanItemIdentities           []*sspb.StudyPlanItemIdentity
	MapStudyPlanItemIdentityMaxResult map[string]string

	LocationIDs    []string
	CourseStudents []*entities.CourseStudent

	CurrentUserID    string
	CurrentStudentID string
	SchoolIDInt      int

	UserFillInTheBlankOld bool

	ExamLOBase                    *sspb.ExamLOBase
	QuizItems                     []*cpb.Quiz
	Quizzes                       entities.Quizzes
	QuizSet                       entities.QuizSet
	SubmissionID                  string
	ShuffledQuizSetID             string
	SetID                         string
	SelectedIndex                 map[string]map[string][]*epb.Answer
	SelectedQuiz                  []int
	QuizAnswers                   []*epb.QuizAnswer
	CheckQuizCorrectnessResponses []*epb.CheckQuizCorrectnessResponse
	FilledText                    map[string]map[string][]*epb.Answer

	StudyPlanItemID string

	ExamLOSubmissionScoreEnts  []*entities.ExamLOSubmissionScore
	ExamLOSubmissionAnswerEnts []*entities.ExamLOSubmissionAnswer
	ExamLOSubmissionEnts       []*entities.ExamLOSubmission

	Limit     int
	Offset    int
	NextPage  *cpb.Paging
	SessionID string

	SchoolAdminToken         string
	Grade                    int
	AssIDs                   []string
	StudentToken             string
	TeacherToken             string
	TeacherID                string
	NumTopics                int
	NumChapter               int
	NumQuizzes               int
	ClassID                  string
	ClassIDs                 []string
	QuizIDs                  []string
	StudentID                string
	StudyPlanIDs             []string
	ShuffledQuizSetIDs       []string
	MasterStudyPlanItems     []string
	AvailableStudyPlanIDs    []string
	WrongQuizExternalIDs     []string
	ArchivedStudyPlanItemIDs []string
	Students                 []entity.Student
	SkippedTopics            []string
	TestResp                 interface{}
	StudyPlanID              string
	LearningMaterialID       string
	StudyPlanItems           []*entities.StudyPlanItem
	StudyPlanItemsIDs        []string
	LoIDs                    []string
	AssignedStudentIDs       []string
	ExternalIDs              []string
	TotalPoint               int32
	TotalTeacherExamGrades   int32
	QuestionTagIDs           []string
	UserID                   string
	QuestionGroupID          string
}

func InitStep(s *Suite) map[string]interface{} {
	steps := map[string]interface{}{
		`^<exam_lo>a signed in "([^"]*)"$`:                        s.aSignedIn,
		`^<exam_lo>returns "([^"]*)" status code$`:                s.returnsStatusCode,
		`^<exam_lo>a valid book content$`:                         s.aValidBookContent,
		`^there are exam LOs existed in topic$`:                   s.thereAreExamLOsExistedInTopic,
		`^some students added to course in some valid locations$`: s.someStudentsAddedToCourseInSomeValidLocations,

		// insert exam lo
		`^user insert a valid exam LO$`:                                               s.userInsertAValidExamLO,
		`^our system generates a correct display order for exam LO$`:                  s.ourSystemGeneratesACorrectDisplayOrderForExamLO,
		`^our system updates topic LODisplayOrderCounter correctly with new exam LO$`: s.ourSystemUpdatesTopicLODisplayOrderCounterCorrectlyWithNewExamLo,
		`^user insert a exam LO without "([^"]*)"$`:                                   s.userInsertAExamLOWithoutField,
		`^our system must create exam LO with "([^"]*)" as default value$`:            s.ourSystemMustCreateExamLOWithDefaultValue,

		// update exam lo
		`^our system update exam LO correctly$`:                     s.ourSystemUpdateExamLOCorrectly,
		`^user update a valid exam LO$`:                             s.userUpdateAValidExamLO,
		`^user update a exam LO with "([^"]*)"$`:                    s.userUpdateAExamLOWithField,
		`^our system must update exam LO with "([^"]*)" correctly$`: s.ourSystemMustUpdateExamLOWithUpdatedFieldCorrectly,

		// list exam lo
		`^user list exam LOs$`:                                   s.userListExamLOs,
		`^our system must return exam LOs correctly$`:            s.ourSystemMustReturnExamLOsCorrectly,
		`^a valid quiz set for exam LO$`:                         s.validQuizSetForExamLO,
		`^our system must return exam LOs has a total question$`: s.ourSystemMustReturnExamLOHasTotalQuestion,

		// list highest exam lo submission
		`^there are exam lo submissions existed$`:                               s.thereAreExamLOSubmissionsExisted,
		`^user list highest result exam LO submission$`:                         s.userListHighestResultExamLOSubmission,
		`^our system must return highest result exam lo submissions correctly$`: s.ourSystemMustReturnHighestResultExamLOSubmissionsCorrectly,

		// list exam lo submission
		`^our system must returns list exam lo submissions correctly$`:                                                             s.ourSystemMustReturnsListExamLOSubmissionsCorrectly,
		`^list exam lo submissions with valid locations$`:                                                                          s.listExamLOSubmissionsWithValidLocations,
		`^list exam lo submissions with invalid locations$`:                                                                        s.listExamLOSubmissionsWithInvalidLocations,
		`^student choose option "([^"]*)" of the quiz "([^"]*)" for submit quiz answers$`:                                          s.studentChooseOptionOfTheQuizForSubmitQuizAnswers,
		`^student submit quiz answers$`:                                                                                            s.studentSubmitQuizAnswers,
		`^create quiz tests and answers for exam lo$`:                                                                              s.createQuizTestsAndAnswersForExamLO,
		`^a quiz test include "([^"]*)" multiple choice quizzes with "([^"]*)" quizzes per page and do quiz test for exam lo$`:     s.aQuizTestIncludeMultipleChoiceQuizzesWithQuizzesPerPageAndDoQuizTestForExamLO,
		`^a quiz test include "([^"]*)" pair of word quizzes with "([^"]*)" quizzes per page and do quiz test for exam lo$`:        s.aQuizTestIncludePairOfWordQuizzesWithQuizzesPerPageAndDoQuizTestForExamLO,
		`^a quiz test include "([^"]*)" term and definition quizzes with "([^"]*)" quizzes per page and do quiz test for exam lo$`: s.aQuizTestIncludeTermAndDefinitionQuizzesWithQuizzesPerPageAndDoQuizTestForExamLO,
		`^a quiz test "([^"]*)" fill in the blank quizzes with "([^"]*)" quizzes per page and do quiz test for exam lo$`:           s.aQuizTestFillInTheBlankQuizzesWithQuizzesPerPageAndDoQuizTestForExamLO,
		`^list exam lo submissions with filter by "([^"]*)"$`:                                                                      s.listExamLOSubmissionsWithFilterBy,
		`^our system must returns list exam lo submissions with filter by "([^"]*)" correctly$`:                                    s.ourSystemMustReturnsListExamLOSubmissionsWithFilterCorrectly,
		`^all student answers and submit quizzes belong to exam$`:                                                                  s.allStudentsAnswerAndSubmitQuizzesBelongToExam,
		`^teacher manually grade exam submission$`:                                                                                 s.teacherGradesSubmission,
		// list exam lo submission score
		`^there are exam lo submission scores existed$`:                      s.thereAreExamLOSubmissionScoresExisted,
		`^user list exam lo submission scores$`:                              s.userListExamLoSubmissionScores,
		`^our system must returns list exam lo submission scores correctly$`: s.ourSystemMustReturnsListExamLOSubmissionScoresCorrectly,

		// list exam lo submission result
		`^user list exam lo submission result$`:                              s.userListExamLOSubmissionResult,
		`^our system must returns list exam lo submission result correctly$`: s.ourSystemMustReturnsListExamLoSubmissionResultCorrectly,

		// grade book
		`^<exam_lo>"(\d+)" student login$`:                                                                      s.studentLogin,
		`^<exam_lo>a teacher login$`:                                                                            s.teacherLogin,
		`^<exam_lo>a school admin login$`:                                                                       s.schoolAdminLogin,
		`^<exam_lo>"school admin" has created a studyplan for all student$`:                                     s.hasCreatedAStudyplanForStudent,
		`^<exam_lo>"school admin" has created a course with a book$`:                                            s.schoolAdminCreateACourseWithABook,
		`^<exam_lo>"([^"]*)" has updated course duration for student$`:                                          s.updateCouseDurationForStudent,
		`^<exam_lo>our system returns correct topic statistic$`:                                                 s.returnsCorrectTopicStatistic,
		`^<exam_lo>retrieve topic statistic$`:                                                                   s.retrieveTopicStatistic,
		`^<exam_lo>topic total assigned student is (\d+), completed students is (\d+), average score is (\d+)$`: s.topicTotalAssignedStudentIsCompletedStudentsIsAverageScoreIs,
		`^<exam_lo>"([^"]*)" has created a book with each "(\d+)" los with "(\d+)" point, "(\d+)" assignments, "(\d+)" topics, "(\d+)" chapters, "(\d+)" quizzes$`:                                                              s.hasCreatedABookWithEachLosAssignmentsTopicsChaptersQuizzesV2,
		`^<exam_lo>"(\d+)" students do test and each student done "(\d+)" los with "(\d+)" correctly and "(\d+)" assignments with "(\d+)" point and skip "(\d+)" topics$`:                                                       s.someStudentDoTestAndDoneLosWithCorrectlyAndAssignmentsWithPointAndSkipTopics,
		`^<exam_lo>our system returns "(\d+)" items with "(\d+)" exam los, "(\d+)" completed los, "(\d+)" grade to pass items, "(\d+)" items and "(\d+)" items with "(\d+)" point, "(\d+)" grade point, "(\d+)" total attempts`: s.ourSystemReturnsGradeBookViewCorrectly,
		`^<exam_lo>retrieve grade book with "([^"]*)"$`:                                      s.retrieveGradeBookWithSetting,
		`^<exam_lo>teacher updates "([^"]*)" of exam lo$`:                                    s.teachUpdateResultOfExamLo,
		`^<exam_lo>user upsert grade book setting$`:                                          s.userUpsertGradeBookSetting,
		`admin respectively create 3 books with 0 exam lo, 1 exam lo and 2 exam los`:         s.adminRespectivelyCreate3BooksWith0ExamLO1ExamLOAnd2ExamLOs,
		`admin create 1st course with 3 study plans using 3 books`:                           s.adminCreateCourseWith3StudyPlanUsing3Books,
		`admin create 2nd course with a study plan with book have 2 exam los`:                s.adminCreate2ndCourseWithStudyPlanHave2ExamLOs,
		`admin create 3rd course with a study plan with book have 0 exam lo`:                 s.adminCreate3rdCourseWithStudyPlanHave0ExamLO,
		`admin create 4th course with no study plan`:                                         s.adminCreate4thCourseWithNoStudyPlan,
		`admin create 1st student at grade 5 join all courses`:                               s.adminCreate1stStudentAtGrade5JoinAllCourses,
		`admin create 2nd student at grade 6 join 1st course`:                                s.adminCreate2ndStudentAtGrade6Join1stCourse,
		`admin create 3rd student at grade 7 join 3rd course`:                                s.adminCreate3rdStudentAtGrade7Join3rdCourse,
		`admin create 4th student at grade 5 join no course`:                                 s.adminCreate4thStudentAtGrade5,
		`admin get list grade book with "([^"]*)" and "([^"]*)" and "([^"]*)" and "([^"]*)"`: s.adminGetListGradeBookWith,
		`returns correct "(\d+)"`:                                                            s.ReturnsCorrectResponseStudentAndTotalItems,

		// upsert submit quiz answer
		`^user creates a valid book content$`:                                                                                  s.userCreatesAValidBookContent,
		`^user creates a course and add students into the course$`:                                                             s.userCreatesACourseAndAddStudentsIntoTheCourse,
		`^user adds a master study plan with the created book$`:                                                                s.userAddsAMasterStudyPlanWithTheCreatedBook,
		`^user creates an Exam LO with manual grading is "([^"]*)", grade to pass is "([^"]*)", approve grading is "([^"]*)"$`: s.userCreatesAnExamLoWithManualGradingIsGradeToPassIsApproveGradingIs,
		`^user adds (\d+) quizzes in "([^"]*)" type and sets (\d+) point for each quiz$`:                                       s.userAddsQuizzesInTypeAndSetsPointForEachQuiz,
		`^user updates study plan for the Exam LO$`:                                                                            s.userUpdatesStudyPlanForTheExamLO,
		`^user starts and submits answers in multiple choice type$`:                                                            s.userStartsAndSubmitsAnswersInMultipleChoiceType,
		`^user starts and submits answers in multiple choice type and exit$`:                                                   s.userStartsAndSubmitsAnswersInMultipleChoiceTypeAndExit,
		`^our system must return submit "([^"]*)", (\d+), (\d+) correctly$`:                                                    s.ourSystemMustReturnSubmitResultCorrectly,
		`^user grades a submission answers to "([^"]*)" status$`:                                                               s.userGradesASubmissionAnswersToStatus,
		`^our system must returns graded score correctly$`:                                                                     s.ourSystemMustReturnsGradedScoreCorrectly,
		`^lo progression and lo progression answers has been created$`:                                                         s.loProgressionAndLoProgressionAnswersHasBeenCreated,
		`^lo progression and lo progression answers has been deleted correctly$`:                                               s.loProgressionAndLoProgressionAnswersHasBeenDeletedCorrectly,

		// delete exam lo submission
		`^create student event logs after do quiz$`:                           s.createStudentEventLogsAfterDoQuiz,
		`^user delete exam lo submission$`:                                    s.userDeleteExamLoSubmission,
		`^exam lo submission and related tables have been deleted correctly$`: s.examLoSubmissionAndRelatedTablesHaveBeenDeletedCorrectly,

		// bulk approve reject submission
		`^user action bulk "([^"]*)" submission$`:                                   s.userActionBulkSubmission,
		`^our system must returns "([^"]*)" status and "([^"]*)" result correctly$`: s.ourSystemMustReturnsStatusCorrectly,

		// retrieve metadata tagging result
		`^<exam_lo>add a exam_lo to topic$`:                s.addExamLOToTopic,
		`^<exam_lo>create some tags$`:                      s.createSomeTags,
		`^<exam_lo>add some quizzes to exam_lo with tags$`: s.addSomeQuizzesToExamLOWithTags,
		`^<exam_lo>create study plan with book$`:           s.createAStudyPlanWithBook,
		`^<exam_lo>a student join course$`:                 s.aStudentJoinCourse,
		`^<exam_lo>a student do exam lo$`:                  s.aStudentDoExamLO,
		`^<exam_lo>user retrieve metadata tagging result$`: s.userRetrieveMetadataTaggingResult,
		`^<exam_lo>metadata tagging result is correct$`:    s.metadataTaggingResultIsCorrect,
		`^insert a question group$`:                        s.insertANewQuestionGroup,
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
	stepState.Token = authToken
	stepState.UserID = userID
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
	stepState.TopicID = topicIDs[0]
	return utils.StepStateToContext(ctx, stepState), nil
}

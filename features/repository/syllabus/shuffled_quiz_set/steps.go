package shuffled_quiz_set

import "github.com/manabie-com/backend/internal/eureka/entities"

type StepState struct {
	ShuffledQuizSetID string
	RoleIDs           []string
	StudentID         string
	CurrentStudentID  string
	StudyPlanID       string
	StudyPlanItemID   string
	TopicID           string
	LoID              string
	AssignmentID      string
	StudyPlanItemIDs  []string

	Quizzes entities.Quizzes

	TotalAnswerTrue int
	TotalPoint      int
}

func InitStep(s *Suite) map[string]interface{} {
	steps := map[string]interface{}{
		`^our system stored study plan item identity of shuffle quiz set correctly$`: s.ourSystemStoredStudyPlanItemIdentityOfShuffleQuizSetCorrectly,
		`^user create a shuffle quiz set$`:                                           s.userCreateAShuffleQuizSet,
		`^user create a study plan of assignment to database$`:                       s.userCreateAStudyPlanOfAssignmentToDatabase,

		`^user create a study plan of exam lo to database$`:                                             s.userCreateAStudyPlanOfExamLOToDatabase,
		`^user create a shuffle quiz set for exam lo with (\d+) of quizzes$`:                            s.userCreateAShuffleQuizSetForExamLOWithNumberOfQuizzes,
		`^user submitted with (\d+) of answers$`:                                                        s.userSubmittedWithNumberOfAnswers,
		`^database has a record in exam lo submission and (\d+) records in exam lo submissions answer$`: s.databaseHasARecordInExamLoSubmissionAndNumOfRecordsInExamLoSubmissionsAnswer,

		`^database has (\d+) of record in flash card submission and (\d+) answers in flash card submission answer table$`: s.databaseHasOfRecordInFlashcardSubmissionAndAnswerInFlashcardSubmissionAnswer,
		`^user create a shuffle quiz set for flash card with (\d+) of quizzes$`:                                           s.userCreateAShuffleQuizSetForFlashcardWithNumberOfQuizzes,
		`^user create a study plan of flash card to database$`:                                                            s.userCreateAStudyPlanOfFlashcardToDatabase,
		`^user submitted with (\d+) of flash card answers$`:                                                               s.userSubmittedWithNumberOfFlashcardAnswers,

		`^a valid learning objective in database$`:                                                          s.aValidLearningObjectiveInDatabase,
		`^a study plan of lo in database$`:                                                                  s.aStudyPlanOfLOInDB,
		`^student create a valid shuffle quiz set for lo$`:                                                  s.studentCreateAValidShuffleQuizSetWithQuizzesForLO,
		`^student submit with (\d+) of answers$`:                                                            s.studentSubmitWithNumberOfAnswers,
		`^database must have (\d+) records lo submission and (\d+) records lo submission answer correctly$`: s.databaseMustHaveNumOfSubmissionRecordInLOSubmissionTableCorrectly,

		`^user create a study plan of "([^"]*)" to database$`:   s.userCreateAStudyPlanWithLearningMaterialToDatabase,
		`^user create a shuffle quiz set with some of quizzes$`: s.userCreateAShuffleQuizSetForLOsWithNumberOfQuizzes,
		`^user submitted with some answers$`:                    s.userSubmittedWithSomeAnswers,
		`^user calculate highest submission score correctly$`:   s.userCalculateHighestSubmissionScore,

		`^user create study plan of exam lo to database$`:             s.userCreateAStudyPlanWithExamLOToDatabase,
		`^user calculate highest exam lo submission score correctly$`: s.userCalculateExamLOHighestSubmissionScore,

		`^valid completeness exam lo in database$`:   s.validStudentsLearningObjectivesCompleteness,
		`^user get highest exam lo score correctly$`: s.userGetExamLoScore,
	}
	return steps
}

package entities

import (
	"os"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEntity(t *testing.T) {
	t.Parallel()
	sv, err := database.NewSchemaVerifier("eureka")
	require.NoError(t, err)

	ents := []database.Entity{
		&AssignStudyPlanTask{},
		&AssignmentStudyPlanItem{},
		&Assignment{},
		&ClassStudent{},
		&ClassStudyPlan{},
		&CourseClass{},
		&CourseStudent{},
		&CourseStudyPlan{},
		&LoStudyPlanItem{},
		&StudentStudyPlan{},
		&StudentSubmissionGrade{},
		&StudentSubmission{},
		&StudyPlanItem{},
		&StudyPlan{},
		&StudentLatestSubmission{},
		&TopicsAssignments{},
		&CourseStudentsAccessPath{},
		&LearningObjective{},
		&Topic{},
		&BookChapter{},
		&TopicsLearningObjectives{},
		&Chapter{},
		&CopiedChapter{},
		&Book{},
		&CoursesBooks{},
		&Quiz{},
		&ShuffledQuizSet{},
		&QuizSet{},
		&FlashcardProgression{},
		&StudentEventLog{},
		&StudentLearningTimeDaily{},
		&StudentsLearningObjectivesCompleteness{},
		&LearningMaterial{},
		&GeneralAssignment{},
		&Flashcard{},
		&LearningObjectiveV2{},
		&ExamLO{},
		&ExamLOSubmission{},
		&ExamLOSubmissionAnswer{},
		&ExamLOSubmissionScore{},
		&IndividualStudyPlan{},
		&TaskAssignment{},
		&MasterStudyPlan{},
		&QuestionGroup{},
		&IndividualStudyPlansView{},
		&QuestionTag{},
		&QuestionTagType{},
		&GradeBookSetting{},
		&AllocateMarker{},
		&CorrectnessInfo{},
		&LOSubmissionAnswer{},
		&LOSubmissionAnswerKey{},
		&FlashCardSubmissionAnswer{},
		&FlashCardSubmissionAnswerKey{},
		&LOProgression{},
		&LOProgressionAnswer{},
		&StudentTag{},
		&ImportStudyPlanTask{},
		&AssessmentSession{},
		&Student{},
		&User{},
		&ContentBankMedia{},
		&Assessment{},
	}

	assert := assert.New(t)
	dir, err := os.Getwd()
	assert.NoError(err)

	count, err := database.CheckEntity(dir)
	assert.NoError(err)
	assert.Equalf(count, len(ents), "found %d entities in package, but only %d are being checked; please add new entities to the unit test", count, len(ents))

	for _, e := range ents {
		assert.NoError(database.CheckEntityDefinition(e))
		assert.NoError(sv.Verify(e))
	}
}

func TestEntities(t *testing.T) {
	t.Parallel()
	ents := []database.Entities{
		&AssignmentStudyPlanItems{},
		&Assignments{},
		&ClassStudents{},
		&LoStudyPlanItems{},
		&StudentStudyPlans{},
		&StudentSubmissionGrades{},
		&StudentSubmissions{},
		&StudyPlanItems{},
		&StudyPlans{},
		&StudentLatestSubmissions{},
		&CourseStudents{},
		&CourseStudentsAccessPaths{},
		&LearningObjectives{},
		&Topics{},
		&BookChapters{},
		&Chapters{},
		&CopiedChapters{},
		&Books{},
		&CoursesBookss{},
		&Quizzes{},
		&ShuffledQuizSets{},
		&QuizSets{},
		&LearningMaterials{},
		&LearningObjectiveV2s{},
		&GeneralAssignments{},
		&ExamLOs{},
		&ExamLOSubmissions{},
		&ExamLOSubmissionAnswers{},
		&ExamLOSubmissionScores{},
		&Flashcards{},
		&TaskAssignments{},
		&MasterStudyPlans{},
		&QuestionGroups{},
		&IndividualStudyPlansViews{},
		&QuestionTags{},
		&QuestionTagTypes{},
		&AllocateMarkers{},
		&IndividualStudyPlans{},
		&LOSubmissionAnswers{},
		&FlashCardSubmissionAnswers{},
		&LOProgressions{},
		&LOProgressionAnswers{},
		&StudentTags{},
		&AssessmentSessions{},
		&Students{},
		&Users{},
		&Assessments{},
	}

	assert := assert.New(t)
	dir, err := os.Getwd()
	assert.NoError(err)

	count, err := database.CheckEntities(dir)
	assert.NoError(err)
	assert.Equalf(count, len(ents), "found %d entities in package, but only %d are being checked; please add new entities to the unit test", count, len(ents))

	for _, e := range ents {
		assert.NoError(database.CheckEntitiesDefinition(e))
	}
}

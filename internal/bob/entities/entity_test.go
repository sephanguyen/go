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
	sv, err := database.NewSchemaVerifier("bob")
	require.NoError(t, err)

	ents := []database.Entity{
		&ActivityLog{},
		&AppleUser{},
		&Assignment{},
		&BookChapter{},
		&Book{},
		&Chapter{},
		&CopiedChapter{},
		&ClassMember{},
		&Class{},
		&Config{},
		&ConversionTask{},
		&CourseClass{},
		&Course{},
		&CoursesBooks{},
		&Group{},
		&JprepSyncDataLog{},
		&LearningObjective{},
		&LessonGroup{},
		&LessonMember{},
		&LessonPolling{},
		&Lesson{},
		&LessonsTeachers{},
		&Location{},
		&LocationType{},
		&CourseAccessPath{},
		&LessonReport{},
		&LessonReportDetail{},
		&Media{},
		&PresetStudyPlan{},
		&PresetStudyPlanWeekly{},
		&StudentsStudyPlansWeekly{},
		&Question{},
		&QuestionTagLo{},
		&QuestionSets{},
		&Quiz{},
		&QuizSet{},
		&ShuffledQuizSet{},
		&SchoolAdmin{},
		&SchoolConfig{},
		&City{},
		&District{},
		&School{},
		&StudentAssignment{},
		&StudentComment{},
		&StudentEventLog{},
		&StudentLearningTimeDaily{},
		&StudentOrder{},
		&StudentTopicCompleteness{},
		&StudentTopicOverdue{},
		&Student{},
		&StudentStat{},
		&StudentsLearningObjectivesCompleteness{},
		&StudentSubmission{},
		&StudentSubmissionScore{},
		&Teacher{},
		&Topic{},
		&UserGroup{},
		&User{},
		&AcademicYear{},
		&CourseAcademicYear{},
		&Parent{},
		&StudentParent{},
		&LessonMemberState{},
		&Speeches{},
		&PartnerFormConfig{},
		&PartnerDynamicFormFieldValue{},
		&TopicsLearningObjectives{},
		&FlashcardProgression{},
		&LessonReportApprovalRecord{},
		&StudentSubscription{},
		&StudentSubscriptionAccessPath{},
		&PostgresUser{},
		&PostgresNamespace{},
		&VirtualClassRoomLog{},
		&Username{},
		&StudentEnrollmentStatusHistory{},
		&SchoolInfo{},
		&SchoolHistory{},
		&ClassMemberV2{},
		&TaggedUser{},
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
		&ActivityLogs{},
		&Books{},
		&BookChapters{},
		&Chapters{},
		&CopiedChapters{},
		&Classes{},
		&ConversionTasks{},
		&Courses{},
		&CoursesBookss{},
		&LearningObjectives{},
		&LessonGroups{},
		&LessonMembers{},
		&Lessons{},
		&CourseAccessPaths{},
		&StudentSubscriptionAccessPaths{},
		&LessonReports{},
		&LessonReportDetails{},
		&Medias{},
		&PresetStudyPlansWeekly{},
		&Questions{},
		&Quizzes{},
		&QuizSets{},
		&ShuffledQuizSets{},
		&Schools{},
		&Topics{},
		&UserGroups{},
		&Users{},
		&LessonMemberStates{},
		&Citites{},
		&Districts{},
		&LessonReportApprovalRecords{},
		&PartnerFormConfigs{},
		&PartnerDynamicFormFieldValues{},
		&StudentSubscriptions{},
		&PostgresUsers{},
		&PostgresNamespaces{},
		&SchoolInfos{},
		&SchoolHistories{},
		&ClassMembers{},
		&ClassMembersV2{},
		&TaggedUsers{},
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

package course_statistical

import (
	"context"
	"math/rand"
	"time"

	"github.com/manabie-com/backend/features/syllabus/entity"
	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"

	"github.com/jackc/pgtype"
)

type StepState struct {
	Token                    string
	Response                 interface{}
	Request                  interface{}
	ResponseErr              error
	SchoolAdminToken         string
	StudentToken             string
	TeacherToken             string
	NumTopics                int
	NumChapter               int
	NumQuizzes               int
	Grade                    int32
	BookID                   string
	CourseID                 string
	ChapterID                string
	ClassID                  string
	ClassIDs                 []string
	ChapterIDs               []string
	TopicID                  string
	TopicIDs                 []string
	SchoolIDInt              int32
	LoID                     string
	LoIDs                    []string
	AssIDs                   []string
	QuizIDs                  []string
	StudentID                string
	StudyPlanID              string
	StudyPlanIDs             []string
	StudyPlanItems           []*entities.StudyPlanItem
	StudyPlanItemsIDs        []string
	MasterStudyPlanItems     []string
	AssignedStudentIDs       []string
	ShuffledQuizSetID        string
	AvailableStudyPlanIDs    []string
	WrongQuizExternalIDs     []string
	ArchivedStudyPlanItemIDs []string
	Students                 []entity.Student
	StudentIDs               []string
	SkippedTopics            []string
	TestResp                 interface{}
	debug                    int
	SchoolIDs                []string
	TagIDs                   []string
}

func InitStep(s *Suite) map[string]interface{} {
	steps := map[string]interface{}{
		`^<course_statistical>a signed in "([^"]*)"$`:              s.aSignedIn,
		`^<course_statistical>student are member of some class$`:   s.studentsAreMembersOfSomeClasses,
		`^user create a book$`:                                     s.userCreateABook,
		`^<course_statistical>returns "([^"]*)" status code$`:      s.returnsStatusCode,
		`^"(\d+)" student login$`:                                  s.studentLogin,
		`^<course_statistical>a teacher login$`:                    s.teacherLogin,
		`^<course_statistical>a school admin login$`:               s.schoolAdminLogin,
		`^"school admin" has created a studyplan for all student$`: s.hasCreatedAStudyplanForStudent,
		`^"school admin" has created a course with a book$`:        s.schoolAdminCreateACourseWithABook,
		`^"([^"]*)" has updated course duration for student$`:      s.updateCouseDurationForStudent,
		`^our system returns correct topic statistic$`:             s.returnsCorrectTopicStatistic,
		// `^our system returns correct course statistic v3$`:                                             s.returnsCorrectCourseStatisticV3,
		`retrieve course statistic with no class filter`:                                                                                                            s.retrieveCourseStatisticWithNoClassFilter,
		`retrieve course statistic v3 with no class, "([^"]*)", "([^"]*)" filter`:                                                                                   s.retrieveCourseV3StatisticWithNoClassFilter,
		`retrieve course statistic with class filter`:                                                                                                               s.retrieveCourseStatisticWithClassFilter,
		`retrieve course statistic v3 with class filter`:                                                                                                            s.retrieveCourseStatisticWithClassFilter,
		`^topic total assigned student is (\d+), completed students is (\d+), average score is (\d+)$`:                                                              s.topicTotalAssignedStudentIsCompletedStudentsIsAverageScoreIs,
		`^<course_statistical_v3>topic total assigned student is (\d+), completed students is (\d+), average score is (\d+)$`:                                       s.topicTotalAssignedStudentIsCompletedStudentsIsAverageScoreIsV3,
		`^"([^"]*)" has created a book with each "(\d+)" los, "(\d+)" assignments, "(\d+)" topics, "(\d+)" chapters, "(\d+)" quizzes$`:                              s.hasCreatedABookWithEachLosAssignmentsTopicsChaptersQuizzesV2,
		`^"(\d+)" students do test and each student done "(\d+)" los with "(\d+)" correctly and "(\d+)" assignments with "(\d+)" point and skip "(\d+)" topics$`:    s.someStudentDoTestAndDoneLosWithCorrectlyAndAssignmentsWithPointAndSkipTopics,
		`^"(\d+)" students do test and each student done "(\d+)" los with "(\d+)" correctly and "(\d+)" assignments with "(\d+)" point and skip "(\d+)" topics V2$`: s.someStudentDoTestAndDoneLosWithCorrectlyAndAssignmentsWithPointAndSkipTopicsV2,
		`<course_statistical> "(\d+)" in a "([^"]*)" class`:                                                                                                         s.AssignStudentsToClass,
		`course "([^"]*)" a class`: s.AssignClassToCourse,

		`^tag users valid exists in DB$`: s.tagUsersValid,
		// retrieve school history
		`^retrieve school history by student in course$`: s.retrieveSchoolHistoryByStudentInCourse,
		`^students exists in school history in DB$`:      s.studentsExistsInSchoolHistory,
		`^there are (\d+) school information$`:           s.thereAreNumberSchoolInfo,

		`^user get list tag by student in course$`:      s.userGetListTagByStudentInCourse,
		`^user creates tagged user$`:                    s.userCreatesTaggedUser,
		`^our system must returns list tags correctly$`: s.ourSystemMustReturnsListTagsCorrectly,
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

func (s *Suite) studentsAreMembersOfSomeClasses(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	for _, student := range stepState.Students {
		classID := idutil.ULIDNow()
		if rand.Int()&1 == 0 {
			classID = idutil.ULIDNow()
		}

		err := (&repositories.CourseClassRepo{}).BulkUpsert(ctx, s.EurekaDB, []*entities.CourseClass{
			{
				BaseEntity: entities.BaseEntity{
					CreatedAt: database.Timestamptz(time.Now()),
					UpdatedAt: database.Timestamptz(time.Now()),
					DeletedAt: pgtype.Timestamptz{Status: pgtype.Null},
				},
				ID:       database.Text(idutil.ULIDNow()),
				CourseID: database.Text(stepState.CourseID),
				ClassID:  database.Text(classID),
			},
		})
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), err
		}

		err = (&repositories.ClassStudentRepo{}).Upsert(ctx, s.EurekaDB, &entities.ClassStudent{
			BaseEntity: entities.BaseEntity{
				CreatedAt: database.Timestamptz(time.Now()),
				UpdatedAt: database.Timestamptz(time.Now()),
				DeletedAt: pgtype.Timestamptz{Status: pgtype.Null},
			},
			StudentID: database.Text(student.ID),
			ClassID:   database.Text(classID),
		})
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), err
		}
		stepState.ClassIDs = append(stepState.ClassIDs, classID)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

package study_plan

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/manabie-com/backend/features/syllabus/entity"
	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type StepState struct {
	StartDate                     *timestamppb.Timestamp
	Token                         string
	Response                      interface{}
	Request                       interface{}
	ResponseErr                   error
	LearningMaterialType          sspb.LearningMaterialType
	BookID                        string
	StudentID                     string
	StudentIDs                    []string
	StudyPlanID                   string
	CourseID                      string
	TopicID                       string
	StudyPlanItems                []*entities.StudyPlanItem
	StudyPlanItemsIDs             []string
	TopicIDs                      []string
	ChapterID                     string
	ChapterIDs                    []string
	SchoolAdmin                   entity.SchoolAdmin
	Student                       entity.Student
	Teacher                       entity.Teacher
	Parent                        entity.Parent
	HQStaff                       entity.HQStaff
	LearningMaterialID            string
	FlashcardID                   string
	TopicLODisplayOrderCounter    int32
	LearningMaterialIDs           []string
	StudyPlanItemIDs              []string
	AssignmentIDs                 []string
	LoIDs                         []string
	NumberOfStudyPlan             int
	StudyPlanIDs                  []string
	Status                        string
	StudentToken                  string
	SchoolAdminToken              string
	RequestSentAt                 time.Time
	SubmissionIDs                 []string
	PaginatedToDoItems            [][]*sspb.StudyPlanToDoItem
	PaginatedStudentStudyPlanItem [][]*sspb.StudentStudyPlanItem
	PaginatedTopicProgress        [][]*sspb.StudentTopicStudyProgress
	AllocateMarker                *entities.AllocateMarker
}

func InitStep(s *Suite) map[string]interface{} {
	steps := map[string]interface{}{
		`^<master_study_plan> a signed in "([^"]*)"`:                   s.aSignedIn,
		`^<master_study_plan>a valid book content$`:                    s.aValidBookContent,
		`^<master_study_plan>returns "([^"]*)" status code$`:           s.returnsStatusCode,
		`^<master_study_plan> list study plan with learning material$`: s.listStudyPlanWithLearningMaterial,
		`^admin insert master study plan$`:                             s.adminUpsertMasterStudyPlanInfo,

		`^our system stores master study plan correctly$`:                 s.ourSystemStoresIndividualStudyPlanCorrectly,
		`^admin update info of master study plan$`:                        s.adminUpdateInfoOfMasterStudyPlan,
		`^our system updates start date for master study plan correctly$`: s.ourSystemUpdatesStartDateForIndividualStudyPlanCorrectly,

		// retrieve study plan identity
		`^"([^"]*)" has created a study plan with the book content for student$`: s.hasCreatedAStudyPlanForStudent,
		`^our system return study plan identity correctly$`:                      s.ourSystemReturnStudyPlanIdentityCorrectly,
		`^some study plan items for the study plan$`:                             s.someStudyPlanItemsForTheStudyPlan,
		`^user retrieves study plan identity$`:                                   s.userRetrievesStudyPlanIdentity,

		`^<study_plan> a signed in "([^"]*)"`:                s.aSignedIn,
		`^<study_plan> a valid book content$`:                s.aValidBookContent,
		`^<study_plan>returns "([^"]*)" status code$`:        s.returnsStatusCode,
		`^<study_plan> school admin and student login$`:      s.schoolAdminAndStudentLogin,
		`^student is assigned some valid study plans$`:       s.studentsIsAssignedSomeValidStudyPlans,
		`^user list "([^"]*)" to do item$`:                   s.userListToDoItems,
		`^user try list to do item$`:                         s.userTryListToDoItems,
		`^our system must return list to do item correctly$`: s.ourSystemReturnToDoItemsCorrectly,

		`^valid course and study plan in database$`:                  s.dataForList,
		`^user list to do item structured book tree$`:                s.ListToDoItemStructuredBookTree,
		`^our system must return data correctly$`:                    s.ourSystemReturnCorrectlyToDoItem,
		`^a study plan in course$`:                                   s.aStudyPlanInCourse,
		`^"(\d+)" student in a course$`:                              s.studentInCourse,
		`^<study_plan>user list student study plans$`:                s.ListStudentStudyPlans,
		`^our system must return list student study plan correctly$`: s.OurSysReturnCorrectStudentStudyPlan,

		// Retrieve allocate marker
		`^an existing allocate marker$`:                              s.aValidAllocateMarker,
		`^<study_plan>user send a retrieve allocate marker request$`: s.userRetrieveAllocateMarker,
		`^our system must return allocate marker correctly$`:         s.ourSystemReturnAllocateMarkerCorrectly,

		// import study plan
		`^<study_plan>user creates a valid book content$`:                      s.userCreatesAValidBookContent,
		`^<study_plan>user creates a course and add students into the course$`: s.userCreatesACourseAndAddStudentsIntoTheCourse,
		`^<study_plan>user adds a master study plan with the created book$`:    s.userAddsAMasterStudyPlanWithTheCreatedBook,
		`^user create a learning material in "([^"]*)" type$`:                  s.userCreateALearningMaterialInType,
		`^user bulk upload csv with above study plan$`:                         s.userBulkUploadCSV,
		`^<study_plan>"school admin" has created a studyplan for all student$`: s.hasCreatedAStudyplanForStudent,
	}
	return steps
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

	courseID, err := utils.GenerateCourse(s.AuthHelper.SignedCtx(ctx, stepState.Token), s.YasuoConn)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	stepState.CourseID = courseID

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) aStudyPlanInCourse(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	ctx = s.AuthHelper.SignedCtx(ctx, stepState.Token)
	if err := utils.GenerateCourseBooks(ctx, stepState.CourseID, []string{stepState.BookID}, s.EurekaConn); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("utils.GenerateCourseBooks: %w", err)
	}
	r, err := utils.GenerateStudyPlanV2(s.AuthHelper.SignedCtx(ctx, stepState.Token), s.EurekaConn, stepState.CourseID, stepState.BookID)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	stepState.StudyPlanID = r.StudyPlanID

	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	_, err = utils.GenerateCourseStudyPlan(ctx, r.StudyPlanID, stepState.CourseID, s.EurekaDB)

	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) studentInCourse(ctx context.Context, number int) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	for i := 0; i < number; i++ {
		stepState.StudentIDs = append(stepState.StudentIDs, idutil.ULIDNow())
	}
	_, err := utils.AValidCourseWithIDs(ctx, s.EurekaDB, stepState.StudentIDs, stepState.CourseID)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

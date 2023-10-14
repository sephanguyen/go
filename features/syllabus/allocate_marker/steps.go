package allocate_marker

import (
	"context"

	"github.com/manabie-com/backend/features/syllabus/entity"
	"github.com/manabie-com/backend/features/syllabus/utils"

	"github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type StepState struct {
	StartDate                  *timestamppb.Timestamp
	Token                      string
	Response                   interface{}
	Request                    interface{}
	ResponseErr                error
	BookID                     string
	StudentID                  string
	StudyPlanID                string
	CourseID                   string
	CourseIDs                  []string
	CourseTypeIDs              []string
	StudentIDs                 []string
	StudyPlanIDs               []string
	LoIDs                      []string
	TeacherIDs                 []string
	NumberAllocatedSubmissions []int32
	SchoolDate                 *timestamp.Timestamp
	TopicIDs                   []string
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
	LocationIDs                []string
}

func InitStep(s *Suite) map[string]interface{} {
	steps := map[string]interface{}{
		`^<allocate_marker>a signed in "([^"]*)"$`:         s.aSignedIn,
		`^<allocate_marker>returns "([^"]*)" status code$`: s.returnsStatusCode,
		`^admin insert allocate marker with "([^"]*)" submissions and first teacher has "([^"]*)" submissions,second teacher has "([^"]*)" submissions$`: s.adminInsertAllocateMarker,
		`^admin insert allocate marker with "([^"]*)" submissions for first teacher$`:                                                                    s.adminInsertAllocateMarkerForFirstTeacher,
		`^our system stores allocate marker correctly$`:                                                                                                  s.ourSystemStoresAllocateMarkerCorrectly,
		`^"([^"]*)" teachers access "([^"]*)" courses by location$`:                                                                                      s.teacherAccessCoursesByLocation,
		`^admin lists allocate teacher$`:                                                                                                                 s.adminListAllocateTeacher,
		`^our system returns allocate teacher correctly$`:                                                                                                s.ourSystemReturnsAllocateTeacherCorrectly,
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

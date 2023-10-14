package template

var PackageTemplate = `package %s
import(
	"context"
)
`
var FuncTemplate = `
func (s *suite) %s(%s) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	return StepStateToContext(ctx, stepState), nil
}
`
var SyllabusBddTemplate = `package %s

import "github.com/manabie-com/backend/features/syllabus/utils"

type Suite utils.Suite[StepState]
`

var SyllabusPackageTemplate = `package %s
import(
	"context"
	"github.com/manabie-com/backend/features/syllabus/utils"
)
`
var SyllabusFuncTemplate = `
func (s *Suite) %s(%s) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	return utils.StepStateToContext(ctx, stepState), nil
}
`

var ProtoTemplate = `message %[1]sRequest {
}

message %[1]sResponse {
}
`

var ProtoFuncTemplate = ` rpc %[1]s(%[1]sRequest) returns (%[1]sResponse);
}
`

var AllProtoTemplate = `
syntax = "proto3";

package %[1]s.v1;

option go_package = "github.com/manabie-com/backend/pkg/manabuf/%[1]s/v1;%[2]s";

%[3]s

service %[4]sService {
  %[5]s
`

var ServiceFuncTemplate = `
func(s *%sService) %[2]s(ctx context.Context, req *%[3]s.%[2]sRequest) (*%[3]s.%[2]sResponse, error){
	return &%[3]s.%[2]sResponse{}.nil
}
`

var AllServiceTemplate = `package %[1]s

import (
	"context"
	%[2]s "github.com/manabie-com/backend/pkg/manabuf/%[1]s/v1"
	"github.com/manabie-com/backend/internal/golibs/database"
)

type %[4]sService struct {
	DB database.Ext
}

%[3]s
`
var SyllabusStepsTemplate = `package %s
import (
	"context"

	"github.com/manabie-com/backend/features/syllabus/entity"
	"github.com/manabie-com/backend/features/syllabus/utils"
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
}

func InitStep(s *Suite) map[string]interface{} {
	steps := map[string]interface{}{}

	return steps
}

func (s *Suite) aSignedIn(ctx context.Context, arg string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	//reset token
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
`

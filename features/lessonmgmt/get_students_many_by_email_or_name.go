package lessonmgmt

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/lessonmgmt/modules/user/infrastructure/repo"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"
)

func (s *Suite) GetStudentsManyByEmailOrName(ctx context.Context, keyword string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	userRepo := repo.UserRepo{}
	students, err := userRepo.GetStudentsManyReferenceByNameOrEmail(ctx, s.BobDB, keyword, 1, 0)
	res := &lpb.GetStudentsManyReferenceByNameOrEmailResponse{}
	res.Students = make([]*lpb.GetStudentsManyReferenceByNameOrEmailResponse_StudentInfo, 0, len(students))
	for _, value := range students {
		res.Students = append(res.Students, &lpb.GetStudentsManyReferenceByNameOrEmailResponse_StudentInfo{
			UserId: value.ID,
			Name:   value.Name,
			Email:  value.Email,
		})
	}
	stepState.Response = res
	stepState.ResponseErr = err
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) GotListStudentsByEmailOrName(ctx context.Context, keyword string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	res := stepState.Response.(*lpb.GetStudentsManyReferenceByNameOrEmailResponse)
	if len(res.Students) == 0 {
		return nil, fmt.Errorf("no result, expect found at least 1 student")
	}
	if len(res.Students) > 1 {
		return nil, fmt.Errorf("result return failed, expect get only 1 row")
	}
	name := res.Students[0].Name
	email := res.Students[0].Email
	if !strings.Contains(name, keyword) && !strings.Contains(email, keyword) {
		return nil, fmt.Errorf("name or email doesn't contain keyword")
	}
	return StepStateToContext(ctx, stepState), nil
}

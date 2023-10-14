package bob

import (
	"context"
	"fmt"
	"strconv"

	pb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
)

func (s *suite) userRetrieveClassByIds(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	classID := strconv.Itoa(int(stepState.CurrentClassID))
	stepState.Request = &pb.RetrieveClassByIDsRequest{
		ClassIds: []string{classID},
	}
	stepState.Response, stepState.ResponseErr = pb.NewClassReaderServiceClient(s.Conn).RetrieveClassByIDs(contextWithToken(s, ctx), stepState.Request.(*pb.RetrieveClassByIDsRequest))
	return StepStateToContext(ctx, stepState), nil

}
func (s *suite) bobMustReturnCorrectClassIds(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	query := "SELECT count(*) FROM classes WHERE class_id = $1 AND name = $2"
	rsp := stepState.Response.(*pb.RetrieveClassByIDsResponse)
	for _, class := range rsp.Classes {
		id, err := strconv.Atoi(class.Id)
		if err != nil {
			return StepStateToContext(ctx, stepState), err

		}
		var count int
		if err := s.DB.QueryRow(ctx, query, id, &class.Name).Scan(&count); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		if count != 1 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot find class with %s id and %s name", class.Id, class.Name)
		}
	}
	for i := 0; i < len(rsp.Classes)-1; i++ {
		if rsp.Classes[i].Name > rsp.Classes[i+1].Name {
			return StepStateToContext(ctx, stepState), fmt.Errorf("wrong response order, want: %s->%s, actual: %s->%s", rsp.Classes[i].Name, rsp.Classes[i+1].Name, rsp.Classes[i+1].Name, rsp.Classes[i].Name)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

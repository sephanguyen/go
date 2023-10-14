package bob

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/bob/repositories"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"

	"github.com/jackc/pgtype"
	"github.com/segmentio/ksuid"
)

func (s *suite) aEditClassRequestWithClassNameIs(ctx context.Context, className string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.Request = &pb.EditClassRequest{ClassName: className + ksuid.New().String()}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) AEditClassRequestWithClassNameIs(ctx context.Context, className string) (context.Context, error) {
	return s.aEditClassRequestWithClassNameIs(ctx, className)
}
func (s *suite) userEditAClass(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.RequestSentAt = time.Now()
	ctx, err := s.createClassUpsertedSubscribe(StepStateToContext(ctx, stepState))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.createClassUpsertedSubscribe: %w", err)
	}
	stepState.Response, stepState.ResponseErr = pb.NewClassClient(s.Conn).EditClass(contextWithToken(s, ctx), stepState.Request.(*pb.EditClassRequest))

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) UserEditAClass(ctx context.Context) (context.Context, error) {
	return s.userEditAClass(ctx)
}
func (s *suite) aClassIdInEditClassRequest(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if arg1 == "valid" {
		stepState.Request.(*pb.EditClassRequest).ClassId = stepState.CurrentClassID
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) AClassIdInEditClassRequest(ctx context.Context, arg1 string) (context.Context, error) {
	return s.aClassIdInEditClassRequest(ctx, arg1)
}
func (s *suite) bobMustUpdateClassInDB(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := stepState.Request.(*pb.EditClassRequest)

	classRepo := &repositories.ClassRepo{}
	class, err := classRepo.FindByID(ctx, s.DB, pgtype.Int4{Int: req.ClassId, Status: 2})
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	if class.Name.String != req.ClassName {
		return StepStateToContext(ctx, stepState), errors.New("classname does not edit")
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) BobMustUpdateClassInDB(ctx context.Context) (context.Context, error) {
	return s.bobMustUpdateClassInDB(ctx)
}

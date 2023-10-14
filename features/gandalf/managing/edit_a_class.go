package managing

import (
	"context"

	"github.com/manabie-com/backend/features/bob"

	"github.com/pkg/errors"
)

func (s *suite) aTeacherWhoIsOwnerCurrentClass(ctx context.Context) (context.Context, error) {
	if bob.StepStateFromContext(ctx).CurrentTeacherID == "" {
		return ctx, errors.New("missing CurrentTeacherID")
	}
	if bob.StepStateFromContext(ctx).CurrentClassID == 0 {
		return ctx, errors.New("missing CurrentClassID")
	}
	return ctx, nil
}

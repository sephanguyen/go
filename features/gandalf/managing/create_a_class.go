package managing

import (
	"context"

	"github.com/lestrrat-go/jwx/jwt"
	"github.com/manabie-com/backend/features/bob"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
)

func (s *suite) aOwnerIdWithSchoolIdIsInCreateClassRequest(ctx context.Context, number int, role string, schoolID int) (context.Context, error) {
	stepState := GandalfStepStateFromContext(ctx)
	if role == "" {
		return ctx, nil
	}
	authToken := bob.StepStateFromContext(ctx).AuthToken
	ownerIDs := bob.StepStateFromContext(ctx).Request.(*pb.CreateClassRequest).OwnerIds
	stepState.GandalfStateTeacherIDsMap = make(map[string]string)
	for number > 0 {
		ctx, err := s.bobSuite.ASignedInWithSchool(ctx, role, schoolID)
		if err != nil {
			return GandalfStepStateToContext(ctx, stepState), err

		}
		t, _ := jwt.ParseString(bob.StepStateFromContext(ctx).AuthToken)
		ownerIDs = append(ownerIDs, t.Subject())
		stepState.GandalfStateTeacherIDsMap[t.Subject()] = bob.StepStateFromContext(ctx).AuthToken
		number--
	}
	bobStepState := bob.StepStateFromContext(ctx)
	bobStepState.Request.(*pb.CreateClassRequest).OwnerIds = ownerIDs
	bobStepState.AuthToken = authToken
	return bob.StepStateToContext(ctx, bobStepState), nil
}

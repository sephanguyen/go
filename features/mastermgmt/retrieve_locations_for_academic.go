package mastermgmt

import (
	"context"

	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"
)

func (s *suite) retrieveLocationsForAcademic(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &mpb.RetrieveLocationsForAcademicRequest{
		AcademicYearId: stepState.AcademicYearIDs[0],
	}
	stepState.Response, stepState.ResponseErr = mpb.NewAcademicYearServiceClient(s.MasterMgmtConn).
		RetrieveLocationsForAcademic(contextWithToken(s, ctx), req)
	return StepStateToContext(ctx, stepState), nil
}

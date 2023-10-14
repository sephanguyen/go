package eureka

import (
	"context"

	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
)

func (s *suite) studentRetrieveSubmissionHistoryByLo_ids(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Response, stepState.ResponseErr = epb.NewStudentSubmissionReaderServiceClient(s.Conn).RetrieveStudentSubmissionHistoryByLoIDs(s.signedCtx(ctx), &epb.RetrieveStudentSubmissionHistoryByLoIDsRequest{
		LoIds: []string{stepState.LoID},
	})
	return StepStateToContext(ctx, stepState), nil
}

package eureka

import (
	"context"
	"strconv"
	"strings"

	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
)

func (s *suite) userGetTotalQuizOfLoWithRole(ctx context.Context, totalQuizOfLo string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = contextWithToken(s, ctx)
	loIDs := make([]int, 0)
	for _, idx := range strings.Split(totalQuizOfLo, ",") {
		i, _ := strconv.Atoi(strings.TrimSpace(idx))
		loIDs = append(loIDs, i)
	}

	stepState.LOIDsInReq = make([]string, len(loIDs))
	for i := range stepState.LOIDsInReq {
		stepState.LOIDsInReq[i] = stepState.LOIDs[loIDs[i]-1]
	}
	stepState.Response, stepState.ResponseErr = epb.NewQuizReaderServiceClient(s.Conn).RetrieveTotalQuizLOs(ctx, &epb.RetrieveTotalQuizLOsRequest{
		LoIds: stepState.LOIDsInReq,
	})

	return StepStateToContext(ctx, stepState), nil
}

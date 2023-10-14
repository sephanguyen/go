package mastermgmt

import (
	"context"
	"fmt"
	"time"

	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"
)

func (s *suite) validWorkingHoursCsvPayload(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	locationID := stepState.CenterIDs[0]
	format := "%s,%s,%s"
	r1 := fmt.Sprintf(format, "Monday", "17:00", "21:00")
	r2 := fmt.Sprintf(format, "Tuesday", "17:00", "21:00")
	r3 := fmt.Sprintf(format, "Wednesday", "17:00", "21:00")

	csv := fmt.Sprintf(`day,opening_time,closing_time
	%s
	%s
	%s`, r1, r2, r3)
	stepState.Request = &mpb.ImportWorkingHoursRequest{
		Payload:    []byte(csv),
		LocationId: locationID,
	}
	stepState.ValidCsvRows = []string{r1, r2, r3}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) importWorkingHours(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = mpb.NewWorkingHoursServiceClient(s.MasterMgmtConn).
		ImportWorkingHours(contextWithToken(s, ctx), stepState.Request.(*mpb.ImportWorkingHoursRequest))
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) invalidWorkingHoursCsvPayload(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	locationID := stepState.CenterIDs[0]
	format := "%s,%s,%s"
	r1 := fmt.Sprintf(format, "Monday", "17:00", "21:00")
	r2 := fmt.Sprintf(format, "Tuesday", "23:00", "21:00")
	r3 := fmt.Sprintf(format, "Wednesday", "17:00", "21:00")

	csv := fmt.Sprintf(`day,opening_time,closing_time
	%s
	%s
	%s`, r1, r2, r3)
	stepState.Request = &mpb.ImportWorkingHoursRequest{
		Payload:    []byte(csv),
		LocationId: locationID,
	}
	stepState.ValidCsvRows = []string{r1, r2, r3}
	return StepStateToContext(ctx, stepState), nil
}

package mastermgmt

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"
)

func (s *suite) validTimeSlotCsvPayload(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	locationID := stepState.CenterIDs[0]
	format := "%s,%s,%s"
	r1 := fmt.Sprintf(format, "Internal "+idutil.ULIDNow(), "11:00", "14:00")
	r2 := fmt.Sprintf(format, "Internal "+idutil.ULIDNow(), "13:30", "17:00")
	r3 := fmt.Sprintf(format, "Internal "+idutil.ULIDNow(), "11:00", "17:00")

	csv := fmt.Sprintf(`time_slot_internal_id,start_time,end_time
	%s
	%s
	%s`, r1, r2, r3)
	stepState.Request = &mpb.ImportTimeSlotRequest{
		Payload:    []byte(csv),
		LocationId: locationID,
	}
	stepState.ValidCsvRows = []string{r1, r2, r3}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) importTimeSlot(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = mpb.NewTimeSlotServiceClient(s.MasterMgmtConn).
		ImportTimeSlots(contextWithToken(s, ctx), stepState.Request.(*mpb.ImportTimeSlotRequest))
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) invalidTimeSlotCsvPayload(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	locationID := stepState.CenterIDs[0]
	format := "%s,%s,%s"
	r1 := fmt.Sprintf(format, "Internal "+idutil.ULIDNow(), "11:00", "14:70")
	r2 := fmt.Sprintf(format, "Internal "+idutil.ULIDNow(), "13:30", "27:00")
	r3 := fmt.Sprintf(format, "Internal "+idutil.ULIDNow(), "11:00", "14:00")

	csv := fmt.Sprintf(`time_slot_internal_id,start_time,end_time
	%s
	%s
	%s`, r1, r2, r3)
	stepState.Request = &mpb.ImportTimeSlotRequest{
		Payload:    []byte(csv),
		LocationId: locationID,
	}
	stepState.ValidCsvRows = []string{r1, r2, r3}
	return StepStateToContext(ctx, stepState), nil
}

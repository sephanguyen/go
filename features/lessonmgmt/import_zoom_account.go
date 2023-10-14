package lessonmgmt

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/lessonmgmt/modules/zoom/domain"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"
)

func (s *Suite) theZoomAccountRequestPayloadWith(ctx context.Context, rowCondition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	validRow1 := "id1,email1@gmail.com,Upsert"
	validRow2 := "id2,email2@gmail.com,Delete"
	validRow3 := "id3,email3@gmail.com,Upsert"
	validRow4 := ",email4@gmail.com,Upsert"

	invalidEmptyRow1 := ",2023-01-12 05:00:00,,1"
	invalidEmptyRow2 := ",,2023-01-10 09:00:00,1"

	stepState.ValidCsvRows = []string{}
	stepState.InvalidCsvRows = []string{}
	header := fmt.Sprintf("%s,%s,%s", domain.ZoomIDLabel, domain.ZoomUsernameLabel, domain.ZoomAccountActionLabel)
	switch rowCondition {
	case "all valid rows":
		stepState.Request = &lpb.ImportZoomAccountRequest{
			Payload: []byte(fmt.Sprintf(`%s
			%s
			%s
			%s
			%s`, header, validRow1, validRow2, validRow3, validRow4)),
		}
		stepState.ValidCsvRows = []string{validRow1, validRow2, validRow3}
	case "empty value row":
		stepState.Request = &lpb.ImportZoomAccountRequest{
			Payload: []byte(fmt.Sprintf(`%s
					%s
					%s`, header, invalidEmptyRow1, invalidEmptyRow2)),
		}
		stepState.InvalidCsvRows = []string{invalidEmptyRow1, invalidEmptyRow2}
	}
	stepState.ImportLessonPartnerInternalIDs = []string{"partner-internal-id-5-19"}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) importingZoomAccount(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.RequestSentAt = time.Now()

	stepState.Response, stepState.ResponseErr = lpb.NewZoomAccountServiceClient(s.LessonMgmtConn).
		ImportZoomAccount(contextWithToken(s, ctx), stepState.Request.(*lpb.ImportZoomAccountRequest))

	return StepStateToContext(ctx, stepState), nil
}

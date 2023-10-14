package lessonmgmt

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/classdo/domain"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"
)

func (s *Suite) haveSomeImportedClassDoAccounts(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	accountsCount := 5
	payloadString := fmt.Sprintf("%s,%s,%s,%s",
		domain.ClassDoIDLabel,
		domain.ClassDoEmailLabel,
		domain.ClassDoAPIKeyLabel,
		domain.ClassDoActionLabel,
	)

	for i := 0; i < accountsCount; i++ {
		newID := idutil.ULIDNow()
		stepState.ValidCsvRows = append(stepState.ValidCsvRows, newID)
		payloadString += fmt.Sprintf(`
			%s`, ","+newID+"@email.com,APIKEY"+newID+",Upsert")
	}

	req := &lpb.ImportClassDoAccountRequest{
		Payload: []byte(payloadString),
	}

	return s.importClassDoAccount(StepStateToContext(ctx, stepState), req)
}

func (s *Suite) exportClassDoAccounts(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &lpb.ExportClassDoAccountRequest{}

	stepState.Response, stepState.ResponseErr = lpb.NewClassDoAccountServiceClient(s.LessonMgmtConn).
		ExportClassDoAccount(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

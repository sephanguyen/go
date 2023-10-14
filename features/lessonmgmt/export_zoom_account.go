package lessonmgmt

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/exporter"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/zoom/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/zoom/infrastructure/repo"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"go.uber.org/multierr"
)

func (s *Suite) exportZoomAccounts(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &lpb.ExportZoomAccountRequest{}
	stepState.Response, stepState.ResponseErr = lpb.NewZoomAccountServiceClient(s.LessonMgmtConn).
		ExportZoomAccount(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) haveSomeZoomAccounts(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	locationIDs := stepState.LocationIDs

	for range locationIDs {
		newID := idutil.ULIDNow()
		zoomAccount := &repo.ZoomAccount{}
		database.AllNullEntity(zoomAccount)

		err := multierr.Combine(
			zoomAccount.ID.Set(newID),
			zoomAccount.Email.Set(fmt.Sprintf("classroom-%s", newID)),
			zoomAccount.UserName.Set(fmt.Sprintf("classroom-%s", newID)),
			zoomAccount.CreatedAt.Set(timeutil.Now()),
			zoomAccount.UpdatedAt.Set(timeutil.Now()),
		)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to set classroom values: %w", err)
		}

		cmdTag, err := database.Insert(ctx, zoomAccount, s.BobDB.Exec)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		if cmdTag.RowsAffected() != 1 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot insert classroom")
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) returnsZoomAccountInCsv(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return ctx, fmt.Errorf("can not export zoom account: %s", stepState.ResponseErr.Error())
	}
	resp := stepState.Response.(*lpb.ExportZoomAccountResponse)
	zoomAccountRepo := repo.ZoomAccountRepo{}
	expectedData, err := zoomAccountRepo.GetAllZoomAccount(ctx, s.BobDBTrace)
	if err != nil {
		return ctx, fmt.Errorf("can not get expected zoom account: %s", err)
	}
	exportCols := []exporter.ExportColumnMap{
		{
			DBColumn:  "zoom_id",
			CSVColumn: domain.ZoomIDLabel,
		},
		{
			DBColumn:  "email",
			CSVColumn: domain.ZoomUsernameLabel,
		},
		{
			DBColumn: "updated_at",
		},
		{
			DBColumn: "created_at",
		},
		{
			DBColumn: "deleted_at",
		},
	}
	exportable := sliceutils.Map(expectedData, func(d *repo.ZoomAccount) database.Entity {
		return d
	})

	str, err := exporter.ExportBatch(exportable, exportCols)
	if err != nil {
		return nil, fmt.Errorf("ExportBatch: %w", err)
	}

	if string(exporter.ToCSV(str)) != string(resp.GetData()) {
		return ctx, fmt.Errorf("zoom account csv is not valid:\ngot:\n%s \nexpected: \n%s", resp.Data, exporter.ToCSV(str))
	}
	return StepStateToContext(ctx, stepState), nil
}

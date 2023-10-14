package queries

import (
	"context"

	"github.com/manabie-com/backend/internal/calendar/application/queries/payloads"
	"github.com/manabie-com/backend/internal/calendar/infrastructure"
	"github.com/manabie-com/backend/internal/golibs/database"
)

type DateInfoQueryHandler struct {
	DB           database.QueryExecer
	DateInfoRepo infrastructure.DateInfoPort
}

func (c *DateInfoQueryHandler) FetchDateInfoByDateRangeAndLocationID(ctx context.Context, req *payloads.FetchDateInfoByDateRangeRequest) (*payloads.FetchDateInfoByDateRangeResponse, error) {
	dateInfos, err := c.DateInfoRepo.GetDateInfoDetailedByDateRangeAndLocationID(ctx, c.DB, req.StartDate, req.EndDate, req.LocationID, req.Timezone)
	if err != nil {
		return nil, err
	}

	return &payloads.FetchDateInfoByDateRangeResponse{
		DateInfos: dateInfos,
	}, nil
}

func (c *DateInfoQueryHandler) ExportDayInfo(ctx context.Context) (data []byte, err error) {
	data, err = c.DateInfoRepo.GetAllToExport(ctx, c.DB)
	if err != nil {
		return nil, err
	}
	return
}

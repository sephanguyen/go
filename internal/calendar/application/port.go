package application

import (
	"context"

	"github.com/manabie-com/backend/internal/calendar/application/command"
	"github.com/manabie-com/backend/internal/calendar/application/queries/payloads"
	"github.com/manabie-com/backend/internal/golibs/database"
	cpb "github.com/manabie-com/backend/pkg/manabuf/calendar/v1"
)

type QueryDateInfoPort interface {
	FetchDateInfoByDateRangeAndLocationID(ctx context.Context, req *payloads.FetchDateInfoByDateRangeRequest) (*payloads.FetchDateInfoByDateRangeResponse, error)
	ExportDayInfo(ctx context.Context) (data []byte, err error)
}

type UpsertDateInfoPort interface {
	UpsertDateInfo(ctx context.Context, req *command.UpsertDateInfoRequest) error
	DuplicateDateInfo(ctx context.Context, req *command.DuplicateDateInfoRequest) error
}

type CreateSchedulerPort interface {
	CreateScheduler(ctx context.Context, db database.QueryExecer, req *command.CreateSchedulerRequest) (*command.CreateSchedulerResponse, error)
	CreateManySchedulers(ctx context.Context, db database.QueryExecer, req *cpb.CreateManySchedulersRequest) (*cpb.CreateManySchedulersResponse, error)
}

type UpdateSchedulerPort interface {
	UpdateScheduler(ctx context.Context, db database.QueryExecer, req *command.UpdateSchedulerRequest) error
}

type GetStaffPort interface {
	GetStaffsByLocation(ctx context.Context, db database.QueryExecer, req *payloads.GetStaffRequest) (*payloads.GetStaffResponse, error)
	GetStaffsByLocationIDsAndNameOrEmail(ctx context.Context, db database.QueryExecer, req *payloads.GetStaffByLocationIDsAndNameOrEmailRequest) (*payloads.GetStaffByLocationIDsAndNameOrEmailResponse, error)
}

type QueryLessonPort interface {
	GetLessonDetail(ctx context.Context, db database.QueryExecer, req *payloads.GetLessonDetailRequest) (*payloads.GetLessonDetailResponse, error)
	GetLessonIDsForBulkStatusUpdate(ctx context.Context, db database.QueryExecer, req *payloads.GetLessonIDsForBulkStatusUpdateRequest) ([]*payloads.GetLessonIDsForBulkStatusUpdateResponse, error)
}

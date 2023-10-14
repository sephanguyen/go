package application

import (
	"context"

	"github.com/manabie-com/backend/internal/mastermgmt/modules/schedule_class/application/commands"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/schedule_class/domain"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type ReserveClassCommandHandlerPort interface {
	UpsertReserveClass(ctx context.Context, reserveClass *domain.ReserveClass) (string, *timestamppb.Timestamp, error)
	PublicReserveClassEvt(ctx context.Context, msg *mpb.EvtScheduleClass) error
	CheckWillReserveClass(ctx context.Context, req *mpb.ScheduleStudentClassRequest) (bool, string, error)
	ReserveStudentClass(ctx context.Context, req *mpb.ScheduleStudentClassRequest, currentClassID string) error
	CancelReserveClass(ctx context.Context, payload commands.CancelReserveClassCommandPayload) error
	DeleteReserveClassesByEffectiveDate(ctx context.Context, date string) error
}

type ReserveClassQueryHandlerPort interface {
	RetrieveScheduledClass(ctx context.Context, studentID string) (*mpb.RetrieveScheduledStudentClassResponse, error)
	GetReserveClassesByEffectiveDate(ctx context.Context, date string) ([]*mpb.GetReserveClassesByEffectiveDateResponse_ReserveClass, error)
}

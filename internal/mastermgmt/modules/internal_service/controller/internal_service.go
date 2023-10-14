package controller

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/nats"
	classInfra "github.com/manabie-com/backend/internal/mastermgmt/modules/class/infrastructure"
	reserveClassApplication "github.com/manabie-com/backend/internal/mastermgmt/modules/schedule_class/application"
	reserveClassCommands "github.com/manabie-com/backend/internal/mastermgmt/modules/schedule_class/application/commands"
	reserveClassQueries "github.com/manabie-com/backend/internal/mastermgmt/modules/schedule_class/application/queries"
	reserveInfra "github.com/manabie-com/backend/internal/mastermgmt/modules/schedule_class/infrastructure"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MasterInternalService struct {
	ReserveClassCommandHandler reserveClassApplication.ReserveClassCommandHandlerPort
	ReserveClassQueryHandler   reserveClassApplication.ReserveClassQueryHandlerPort
}

func NewMasterInternalService(
	bobDB database.Ext,
	reserveClassRepo reserveInfra.ReserveClassRepo,
	courseRepo reserveInfra.CourseRepo,
	classRepo reserveInfra.ClassRepo,
	studentPackageClassRepo reserveInfra.StudentPackageClassRepo,
	classMemberRepo classInfra.ClassMemberRepo,
	jsm nats.JetStreamManagement,
) *MasterInternalService {
	return &MasterInternalService{
		ReserveClassCommandHandler: &reserveClassCommands.ReserveClassCommandHandler{
			DB:               bobDB,
			ReserveClassRepo: reserveClassRepo,
			JSM:              jsm,
			ClassMemberRepo:  classMemberRepo,
		},
		ReserveClassQueryHandler: &reserveClassQueries.ReserveClassQueryHandler{
			DB:                      bobDB,
			ReserveClassRepo:        reserveClassRepo,
			CourseRepo:              courseRepo,
			ClassRepo:               classRepo,
			StudentPackageClassRepo: studentPackageClassRepo,
		},
	}
}

func (i *MasterInternalService) GetReserveClassesByEffectiveDate(ctx context.Context, req *mpb.GetReserveClassesByEffectiveDateRequest) (*mpb.GetReserveClassesByEffectiveDateResponse, error) {
	effectiveDate := req.EffectiveDate.AsTime().Format("2006/01/02")
	reserveClasses, err := i.ReserveClassQueryHandler.GetReserveClassesByEffectiveDate(ctx, effectiveDate)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("GetReserveClassesByEffectiveDate at %s fail: %w", effectiveDate, err).Error())
	}

	return &mpb.GetReserveClassesByEffectiveDateResponse{
		ReserveClasses: reserveClasses,
	}, nil
}

func (i *MasterInternalService) DeleteReserveClassesByEffectiveDate(ctx context.Context, req *mpb.DeleteReserveClassByEffectiveDateRequest) (*mpb.DeleteReserveClassByEffectiveDateResponse, error) {
	effectiveDate := req.EffectiveDate.AsTime().Format("2006/01/02")
	err := i.ReserveClassCommandHandler.DeleteReserveClassesByEffectiveDate(ctx, effectiveDate)
	if err != nil {
		return &mpb.DeleteReserveClassByEffectiveDateResponse{
			Successful: false,
		}, status.Error(codes.Internal, fmt.Errorf("DeleteReserveClassesByEffectiveDate at %s fail: %w", effectiveDate, err).Error())
	}

	return &mpb.DeleteReserveClassByEffectiveDateResponse{
		Successful: true,
	}, nil
}

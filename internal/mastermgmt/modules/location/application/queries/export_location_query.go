package queries

import (
	"context"
	"sort"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/exporter"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/infrastructure"
	location_repo "github.com/manabie-com/backend/internal/mastermgmt/modules/location/infrastructure/repo"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ExportLocationQueryHandler struct {
	DB               database.Ext
	LocationRepo     infrastructure.LocationRepo
	LocationTypeRepo infrastructure.LocationTypeRepo
}

func (e *ExportLocationQueryHandler) ExportLocation(ctx context.Context) (data []byte, err error) {
	allLocation, err := e.LocationRepo.GetAllLocations(ctx, e.DB)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	ec := []exporter.ExportColumnMap{
		{
			DBColumn: "location_id",
		},
		{
			DBColumn: "partner_internal_id",
		},
		{
			DBColumn: "name",
		},
		{
			DBColumn: "location_type",
		},
		{
			DBColumn: "partner_internal_parent_id",
		},
	}
	var toEntity = func(l *location_repo.Location) database.Entity {
		return l
	}
	var isOrg = func(l *location_repo.Location) bool {
		return l.LocationType.String == "org"
	}
	exportableLocation := sliceutils.MapSkip(allLocation, toEntity, isOrg)

	str, err := exporter.ExportBatch(exportableLocation, ec)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return exporter.ToCSV(str), nil
}

func (e *ExportLocationQueryHandler) ExportLocationType(ctx context.Context) (data []byte, err error) {
	locTypes, err := e.LocationTypeRepo.GetAllLocationTypes(ctx, e.DB)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	ec := []exporter.ExportColumnMap{
		{
			DBColumn: "location_type_id",
		},
		{
			DBColumn: "name",
		},
		{
			DBColumn: "display_name",
		},
		{
			DBColumn: "level",
		},
	}

	// sort by level ASC
	sort.Slice(locTypes, func(i, j int) bool {
		return locTypes[i].Level.Int < locTypes[j].Level.Int
	})

	var isOrg = func(l *location_repo.LocationType) bool {
		return l.Name.String == "org"
	}
	var toEntity = func(l *location_repo.LocationType) database.Entity {
		return l
	}
	exportableLocationTypes := sliceutils.MapSkip(locTypes, toEntity, isOrg)

	str, err := exporter.ExportBatch(exportableLocationTypes, ec)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return exporter.ToCSV(str), nil
}

package queries

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/exporter"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/class/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/class/infrastructure"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ExportClassesQueryHandler struct {
	DB        database.Ext
	ClassRepo infrastructure.ClassRepo
}

func (e *ExportClassesQueryHandler) ExportClasses(ctx context.Context) (data []byte, err error) {
	allClasses, err := e.ClassRepo.GetAll(ctx, e.DB)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	ec := []exporter.ExportColumnMap{
		{
			DBColumn:  "class_id",
			CSVColumn: "class_id",
		},
		{
			DBColumn:  "name",
			CSVColumn: "class_name",
		},
		{
			DBColumn:  "course_id",
			CSVColumn: "course_id",
		},
		{
			DBColumn:  "location_id",
			CSVColumn: "location_id",
		},
	}

	exportableClasses := sliceutils.Map(allClasses, func(g *domain.ExportingClass) database.Entity {
		return g
	})

	str, err := exporter.ExportBatch(exportableClasses, ec)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return exporter.ToCSV(str), nil
}

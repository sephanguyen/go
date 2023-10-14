package queries

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/exporter"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/grade/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/grade/infrastructure"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ExportGradesQueryHandler struct {
	DB        database.Ext
	GradeRepo infrastructure.GradeRepo
}

func (e *ExportGradesQueryHandler) ExportGrades(ctx context.Context) (data []byte, err error) {
	allGrades, err := e.GradeRepo.GetAll(ctx, e.DB)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	ec := []exporter.ExportColumnMap{
		{
			DBColumn:  "grade_id",
			CSVColumn: "grade_id",
		},
		{
			DBColumn:  "partner_internal_id",
			CSVColumn: "grade_partner_id",
		},
		{
			DBColumn:  "name",
			CSVColumn: "name",
		},
		{
			DBColumn: "sequence",
		},
		{
			DBColumn: "remarks",
		},
	}

	exportableGrades := sliceutils.Map(allGrades, func(g *domain.Grade) database.Entity {
		return g
	})

	str, err := exporter.ExportBatch(exportableGrades, ec)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return exporter.ToCSV(str), nil
}

package queries

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/exporter"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/subject/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/subject/infrastructure"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ExportSubjectsQueryHandler struct {
	DB          database.Ext
	SubjectRepo infrastructure.SubjectRepo
}

func (e *ExportSubjectsQueryHandler) ExportSubjects(ctx context.Context) (data []byte, err error) {
	allSubjects, err := e.SubjectRepo.GetAll(ctx, e.DB)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	ec := []exporter.ExportColumnMap{
		{
			DBColumn:  "subject_id",
			CSVColumn: "subject_id",
		},
		{
			DBColumn:  "name",
			CSVColumn: "name",
		},
	}

	exportableSubjects := sliceutils.Map(allSubjects, func(s *domain.Subject) database.Entity {
		return s
	})

	str, err := exporter.ExportBatch(exportableSubjects, ec)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return exporter.ToCSV(str), nil
}

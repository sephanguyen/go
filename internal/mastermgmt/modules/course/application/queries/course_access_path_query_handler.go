package queries

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/exporter"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/course/infrastructure"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/course/infrastructure/repo"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CourseAccessPathQueryHandler struct {
	DB                   database.Ext
	CourseAccessPathRepo infrastructure.CourseAccessPathRepo
}

func (c *CourseAccessPathQueryHandler) ExportCourseAccessPaths(ctx context.Context) (data []byte, err error) {
	allCap, err := c.CourseAccessPathRepo.GetAll(ctx, c.DB)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	ec := []exporter.ExportColumnMap{
		{
			DBColumn:  "id",
			CSVColumn: "course_access_path_id",
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

	capEntities := sliceutils.Map(allCap, func(c *repo.CourseAccessPath) database.Entity {
		return c
	})

	str, err := exporter.ExportBatch(capEntities, ec)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return exporter.ToCSV(str), nil
}

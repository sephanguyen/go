package queries

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/exporter"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/course/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/course/infrastructure"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/course/infrastructure/repo"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CourseQueryHandler struct {
	DB         database.Ext
	CourseRepo infrastructure.CourseRepo
}

func (c *CourseQueryHandler) GetCoursesByIDs(ctx context.Context, payload GetCoursesByIDsPayload) ([]*domain.Course, error) {
	if len(payload.IDs) == 0 {
		return []*domain.Course{}, nil
	}
	return c.CourseRepo.GetByIDs(ctx, c.DB, payload.IDs)
}

func (c *CourseQueryHandler) ExportCourses(ctx context.Context, enableTeachingMethod bool) (data []byte, err error) {
	allCourses, err := c.CourseRepo.GetAll(ctx, c.DB)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	ec := []exporter.ExportColumnMap{
		{
			DBColumn:  "course_id",
			CSVColumn: "course_id",
		},
		{
			DBColumn:  "name",
			CSVColumn: "course_name",
		},
		{
			CSVColumn: "course_type_id",
			DBColumn:  "course_type_id",
		},
		{
			CSVColumn: "course_partner_id",
			DBColumn:  "course_partner_id",
		},
		{
			CSVColumn: "remarks",
			DBColumn:  "remarks",
		},
	}
	if enableTeachingMethod {
		teachingMethodColumn := exporter.ExportColumnMap{
			DBColumn:  "teaching_method",
			CSVColumn: "teaching_method",
		}

		ec = append(ec[:4], append([]exporter.ExportColumnMap{teachingMethodColumn}, ec[4:]...)...)
	}

	exportableCourses := sliceutils.Map(allCourses, func(c *repo.Course) database.Entity {
		return c
	})

	str, err := exporter.ExportBatch(exportableCourses, ec)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return exporter.ToCSV(str), nil
}

package queries

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/exporter"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
)

type CourseTeachingTimeQueryHandler struct {
	WrapperConnection *support.WrapperDBConnection
	CourseRepo        infrastructure.CourseRepo
}

func (cl *CourseTeachingTimeQueryHandler) ExportCourseTeachingTime(ctx context.Context) (data []byte, err error) {
	exportCols := []exporter.ExportColumnMap{
		{
			DBColumn: "course_id",
		},
		{
			DBColumn:  "name",
			CSVColumn: "course_name",
		},
		{
			DBColumn: "preparation_time",
		},
		{
			DBColumn: "break_time",
		},
	}
	conn, err := cl.WrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	return cl.CourseRepo.ExportAllCoursesWithTeachingTimeValue(ctx, conn, exportCols)
}

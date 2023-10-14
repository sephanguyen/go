package lessonmgmt

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/golibs/exporter"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure/repo"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"
)

func (s *Suite) exportCoursesTeachingTime(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &lpb.ExportCourseTeachingTimeRequest{}
	stepState.Response, stepState.ResponseErr = lpb.NewLessonExecutorServiceClient(s.LessonMgmtConn).
		ExportCourseTeachingTime(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) returnsCourseTeachingTimeInCsv(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return ctx, fmt.Errorf("can not export course teaching time: %s", stepState.ResponseErr.Error())
	}
	resp := stepState.Response.(*lpb.ExportCourseTeachingTimeResponse)

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
	courseRepo := repo.CourseRepo{}
	expectedData, err := courseRepo.ExportAllCoursesWithTeachingTimeValue(ctx, s.BobDBTrace, exportCols)
	if err != nil {
		return ctx, fmt.Errorf("can not get expected course teaching time: %s", err)
	}

	if string(expectedData) != string(resp.GetData()) {
		return ctx, fmt.Errorf("course csv is not valid:\ngot:\n%s \nexpected: \n%s", resp.Data, expectedData)
	}
	return StepStateToContext(ctx, stepState), nil
}

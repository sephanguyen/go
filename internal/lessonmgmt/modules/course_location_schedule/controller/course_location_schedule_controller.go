package controller

import (
	"context"

	services "github.com/manabie-com/backend/internal/lessonmgmt/modules/course_location_schedule/service"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"
)

type CourseLocationScheduleController struct {
	wrapperConnection *support.WrapperDBConnection
	service           services.ICourseLocationScheduleService
}

func NewCourseLocationControllerController(wrapperConnection *support.WrapperDBConnection, service services.ICourseLocationScheduleService) *CourseLocationScheduleController {
	return &CourseLocationScheduleController{
		wrapperConnection: wrapperConnection,
		service:           service,
	}
}

func (c *CourseLocationScheduleController) ImportCourseLocationSchedule(ctx context.Context, req *lpb.ImportCourseLocationScheduleRequest) (*lpb.ImportCourseLocationScheduleResponse, error) {
	return c.service.ImportCourseLocationSchedule(ctx, req)
}

func (c *CourseLocationScheduleController) ExportCourseLocationSchedule(ctx context.Context, _ *lpb.ExportCourseLocationScheduleRequest) (*lpb.ExportCourseLocationScheduleResponse, error) {
	return c.service.ExportCourseLocationSchedule(ctx)
}

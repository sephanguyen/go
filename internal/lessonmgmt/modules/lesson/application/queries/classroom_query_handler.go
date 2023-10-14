package queries

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/exporter"
	"github.com/manabie-com/backend/internal/golibs/unleashclient"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/application/queries/payloads"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"golang.org/x/exp/slices"
)

type ClassroomQueryHandler struct {
	WrapperConnection   *support.WrapperDBConnection
	ClassroomRepo       infrastructure.ClassroomRepo
	LessonClassroomRepo infrastructure.LessonClassroomRepo
	Env                 string
	UnleashClientIns    unleashclient.ClientInstance
}

func (cl *ClassroomQueryHandler) ExportClassrooms(ctx context.Context) (data []byte, err error) {
	conn, err := cl.WrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	isUnleashEnabled, err := cl.UnleashClientIns.IsFeatureEnabled("Lesson_LessonManagement_BackOffice_UpdateClassroomFlow", cl.Env)
	if err != nil {
		return nil, fmt.Errorf("l.connectToUnleash: %w", err)
	}
	exportCols := []exporter.ExportColumnMap{
		{
			DBColumn: "location_id",
		},
		{
			DBColumn: "location_name",
		},
		{
			DBColumn: "classroom_id",
		},
		{
			DBColumn: "classroom_name",
		},
		{
			DBColumn: "remarks",
		},
	}
	if isUnleashEnabled {
		exportCols = append(exportCols, []exporter.ExportColumnMap{
			{
				DBColumn: "room_area",
			},
			{
				DBColumn: "seat_capacity",
			}}...)
	} else {
		exportCols = append(exportCols, exporter.ExportColumnMap{
			DBColumn: "is_archived",
		})
	}
	return cl.ClassroomRepo.ExportAllClassrooms(ctx, conn, exportCols)
}

func (cl *ClassroomQueryHandler) RetrieveClassroomsByLocationID(ctx context.Context, req *lpb.RetrieveClassroomsByLocationIDRequest) (classrooms []*domain.Classroom, err error) {
	conn, err := cl.WrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	payload, err := buildGetClassroomArgPayload(req)
	if err != nil {
		return nil, fmt.Errorf("cannot build classroom payload: %w", err)
	}

	classrooms, err = cl.ClassroomRepo.RetrieveClassroomsByLocationID(ctx, conn, payload)
	if err != nil {
		return nil, fmt.Errorf("cannot retrieve classroom for this location: %w", err)
	}

	if !payload.StartTime.IsZero() && !payload.EndTime.IsZero() {
		classrooms, err = cl.getClassroomStatusByLocationAndTimeRange(ctx, conn, classrooms, payload)
		if err != nil {
			return nil, fmt.Errorf("cannot get occupied classroom: %w", err)
		}
	}
	return
}

func (cl *ClassroomQueryHandler) getClassroomStatusByLocationAndTimeRange(ctx context.Context, db database.Ext, classrooms []*domain.Classroom, payload *payloads.GetClassroomListArg) ([]*domain.Classroom, error) {
	lessonClassrooms, err := cl.LessonClassroomRepo.GetOccupiedClassroomByTime(ctx, db, payload.LocationIDs, payload.LessonID, payload.StartTime, payload.EndTime, payload.Timezone)
	if err != nil {
		return nil, fmt.Errorf("cannot retrieve classroom for this location: %w", err)
	}
	lcIDs := lessonClassrooms.GetIDs()

	for _, classroom := range classrooms {
		if slices.Contains(lcIDs, classroom.ClassroomID) {
			classroom.ClassroomStatus = domain.InUsed
		}
	}

	return classrooms, nil
}

func buildGetClassroomArgPayload(req *lpb.RetrieveClassroomsByLocationIDRequest) (payload *payloads.GetClassroomListArg, err error) {
	locationID := req.GetLocationId()
	timezone := "UTC"

	if locationID == "" {
		return nil, fmt.Errorf("location_id is required: %w", err)
	}
	locationIDs := append(req.GetLocationIds(), locationID)

	payload = &payloads.GetClassroomListArg{
		LocationIDs: golibs.Uniq(locationIDs),
		KeyWord:     req.GetKeyword(),
		Limit:       20,
		Offset:      0,
	}
	if req.Paging != nil {
		if req.Paging.GetOffsetInteger() > 0 {
			payload.Offset = int32(req.Paging.GetOffsetInteger())
		}
		if req.Paging.GetLimit() > 0 {
			payload.Limit = int32(req.Paging.GetLimit())
		}
	}

	if req.TimeZone != "" {
		timezone = req.GetTimeZone()
	}

	if err != nil {
		return nil, fmt.Errorf("invalid client timezone")
	}

	payload.Timezone = timezone
	if startTime := req.GetStartTime(); startTime != nil {
		payload.StartTime = req.GetStartTime().AsTime()
	}
	if endTime := req.GetEndTime(); endTime != nil {
		payload.EndTime = req.GetEndTime().AsTime()
	}

	if req.GetLessonId() != "" {
		payload.LessonID = req.GetLessonId()
	}

	return payload, nil
}

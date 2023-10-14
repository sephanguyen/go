package controller

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/unleashclient"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/application/queries"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ClassroomReaderService struct {
	wrapperConnection     *support.WrapperDBConnection
	classroomQueryHandler queries.ClassroomQueryHandler
}

func NewClassroomReaderService(
	wrapperConnection *support.WrapperDBConnection,
	classroomRepo infrastructure.ClassroomRepo,
	lessonClassroomRepo infrastructure.LessonClassroomRepo,
	env string,
	unleashClientIns unleashclient.ClientInstance,
) *ClassroomReaderService {
	return &ClassroomReaderService{
		wrapperConnection: wrapperConnection,
		classroomQueryHandler: queries.ClassroomQueryHandler{
			WrapperConnection:   wrapperConnection,
			ClassroomRepo:       classroomRepo,
			Env:                 env,
			UnleashClientIns:    unleashClientIns,
			LessonClassroomRepo: lessonClassroomRepo,
		},
	}
}

func (c *ClassroomReaderService) RetrieveClassroomsByLocationID(ctx context.Context, req *lpb.RetrieveClassroomsByLocationIDRequest) (*lpb.RetrieveClassroomsByLocationIDResponse, error) {
	classrooms, err := c.classroomQueryHandler.RetrieveClassroomsByLocationID(ctx, req)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	items := []*lpb.Classroom{}
	for _, classroom := range classrooms {
		items = append(items, &lpb.Classroom{
			ClassroomId:   classroom.ClassroomID,
			ClassroomName: classroom.Name,
			LocationId:    classroom.LocationID,
			RoomArea:      classroom.RoomArea,
			SeatCapacity:  uint32(classroom.SeatCapacity),
			Remarks:       classroom.Remarks,
			Status:        lpb.ClassroomStatus(lpb.ClassroomStatus_value[string(classroom.ClassroomStatus)]),
		})
	}

	return &lpb.RetrieveClassroomsByLocationIDResponse{
		Items: items,
	}, nil
}

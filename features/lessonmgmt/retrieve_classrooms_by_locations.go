package lessonmgmt

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/features/helper"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"golang.org/x/exp/slices"
)

func (s *Suite) userGetClassroomsOfLocations(ctx context.Context, locationStrIDs string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	locationIDs = strings.Split(locationStrIDs, ",")
	req := &lpb.RetrieveClassroomsByLocationIDRequest{
		TimeZone: LocalTimezone,
		Paging: &cpb.Paging{
			Limit: 50,
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: 0,
			},
		},
		LocationId:  locationIDs[0],
		LocationIds: locationIDs,
	}
	stepState.Request = req
	stepState.Response, stepState.ResponseErr = lpb.NewClassroomReaderServiceClient(s.Connections.LessonMgmtConn).
		RetrieveClassroomsByLocationID(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) theListClassroomOfLocations(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := stepState.Request.(*lpb.RetrieveClassroomsByLocationIDRequest)
	classrooms := stepState.Response.(*lpb.RetrieveClassroomsByLocationIDResponse).Items

	if len(classrooms) > 0 {
		locationList := append(req.GetLocationIds(), req.GetLocationId())
		for _, classroom := range classrooms {
			if !slices.Contains(locationList, classroom.LocationId) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("this classroom(%s) not belong to these locations: %s", classroom.ClassroomId, locationList)
			}
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

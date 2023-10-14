package lessonmgmt

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"github.com/pkg/errors"
)

func (s *Suite) avalidClassroomRequestPayload(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	locationID := "location-id-2"
	locationName := "location-id-2"
	request := fmt.Sprintf(`location_id,location_name,classroom_id,classroom_name,room_area,seat_capacity,remarks
		%s,%s,,classroom-1,floor 1,20,remark
		%s,%s,,classroom-2,floor 2,15,classroom
		%s,%s,,classroom-3,floor 3,32,teacer seat`,
		locationID, locationName,
		locationID, locationName,
		locationID, locationName,
	)

	stepState.Request = &lpb.ImportClassroomRequest{
		Payload: []byte(request),
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) anInvalidClassroomRequestPayload(ctx context.Context, invalidFormat string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	switch invalidFormat {
	case "no data":
		stepState.Request = &lpb.ImportClassroomRequest{}
	case "header only":
		stepState.Request = &lpb.ImportClassroomRequest{
			Payload: []byte(`location_id,location_name,classroom_id,classroom_name,room_area,seat_capacity,remarks`),
		}
	case "mismatched number of fields in header and content":
		stepState.Request = &lpb.ImportClassroomRequest{
			Payload: []byte(`location_id,location_name,classroom_id,classroom_name,room_area,seat_capacity,remarks
			location-id-3,location-id-3
			location-id-3,location-id-3
			location-id-3,location-id-3`),
		}
	case "wrong id column name in header":
		stepState.Request = &lpb.ImportClassroomRequest{
			Payload: []byte(`location_id,location_name,classroom_name,classroom_id,room_area,seat_capacity,remarks
			location-id-2,location-id-2,classroom-1,,floor 1,20,remark
			location-id-2,location-id-2,classroom-1,,floor 1,20,remark
			location-id-2,location-id-2,classroom-1,,floor 1,20,remark`),
		}
	case "wrong name column name in header":
		stepState.Request = &lpb.ImportClassroomRequest{
			Payload: []byte(`location_id,location_name,classroom_id,name,room_area,seat_capacity,remarks
			location-id-2,location-id-2,,classroom-1,floor 1,20,remark
			location-id-2,location-id-2,,classroom-1,floor 1,20,remark
			location-id-2,location-id-2,,classroom-1,floor 1,20,remark`),
		}
	case "mismatched valid and invalid rows":
		stepState.Request = &lpb.ImportClassroomRequest{
			Payload: []byte(`location_id,location_name,classroom_id,classroom_name,room_area,seat_capacity,remarks
			location-id-2,location-id-2,,classroom-1,floor 1,20,remark
			location-id-2,location-id-2,,classroom-2,floor 1,20,remark
			location-id-2,location-id-2,classroom-1,,floor 1,20,remark`),
		}
	case "invalid location_id":
		stepState.Request = &lpb.ImportClassroomRequest{
			Payload: []byte(`location_id,location_name,classroom_id,classroom_name,room_area,seat_capacity,remarks
			locationid2,location-id-2,,classroom-1,floor 1,20,remark`),
		}
	case "missing value in madatory column":
		stepState.Request = &lpb.ImportClassroomRequest{
			Payload: []byte(`location_id,location_name,classroom_id,classroom_name,room_area,seat_capacity,remarks
			location-id-2,,,classroom-1,floor 1,20,remark
			location-id-2,location-id-2,,,floor 2,45,remark cls 3
			`),
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) importingClassrooms(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = lpb.NewLessonExecutorServiceClient(s.LessonMgmtConn).
		ImportClassroom(contextWithToken(s, ctx), stepState.Request.(*lpb.ImportClassroomRequest))

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) theValidClassroomLinesAreImportedSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	res := stepState.Response.(*lpb.ImportClassroomResponse)
	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	if len(res.Errors) > 0 {
		return ctx, fmt.Errorf("response errors: %s", res.Errors)
	}

	classrooms, err := s.selectNewClassrooms(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	for _, row := range stepState.ValidCsvRows {
		rowValues := strings.Split(row, ",")
		classroomID := rowValues[2]
		seatCap, err := strconv.ParseInt(rowValues[5], 10, 32)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("capacity must be a numberic")
		}

		found := false
		for _, classroom := range classrooms {
			if classroom.ClassroomID == classroomID && classroom.Name == rowValues[3] && classroom.LocationID == rowValues[0] && classroom.RoomArea == rowValues[4] && classroom.SeatCapacity == int32(seatCap) {
				found = true
				break
			}
		}
		if !found {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to import valid csv row: %s", rowValues)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) selectNewClassrooms(ctx context.Context) ([]*domain.Classroom, error) {
	var allEntities []*domain.Classroom
	stmt := `SELECT classroom_id, name, location_id, COALESCE(remarks, '') as remarks, COALESCE(room_area, '') as room_area, COALESCE(seat_capacity, 0) as seat_capacity
		FROM classroom
		where deleted_at is null
		order by updated_at desc limit 50`
	rows, err := s.BobDBTrace.Query(ctx, stmt)
	if err != nil {
		return nil, errors.Wrap(err, "query new classrooms")
	}
	defer rows.Close()
	for rows.Next() {
		c := &domain.Classroom{}
		err := rows.Scan(
			&c.ClassroomID,
			&c.Name,
			&c.LocationID,
			&c.Remarks,
			&c.RoomArea,
			&c.SeatCapacity,
		)
		if err != nil {
			return nil, errors.WithMessage(err, "rows.Scan new classrooms")
		}
		allEntities = append(allEntities, c)
	}
	return allEntities, nil
}

func (s *Suite) theInvalidClassroomMustReturnedError(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp := stepState.Response.(*lpb.ImportClassroomResponse)
	if len(resp.Errors) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("invalid file is not returned error list in response")
	}
	return StepStateToContext(ctx, stepState), nil
}

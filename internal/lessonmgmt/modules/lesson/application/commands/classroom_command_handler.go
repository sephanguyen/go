package commands

import (
	"bytes"
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/scanner"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"
)

type ClassroomCommandHandler struct {
	WrapperConnection *support.WrapperDBConnection
	ClassroomRepo     infrastructure.ClassroomRepo
	MasterDataPort    infrastructure.MasterDataPort
}

func (cl *ClassroomCommandHandler) ImportClassroom(ctx context.Context, req *lpb.ImportClassroomRequest) (*lpb.ImportClassroomResponse, error) {
	conn, err := cl.WrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	payloads, errorCSVs := cl.buildImportClassroomArgs(ctx, conn, req.Payload)
	res := lpb.ImportClassroomResponse{
		Errors: errorCSVs,
	}
	if len(errorCSVs) > 0 {
		return &res, nil
	}

	err = cl.ClassroomRepo.UpsertClassrooms(ctx, conn, payloads.Classrooms)
	if err != nil {
		return &res, err
	}
	return &res, nil
}

func (cl *ClassroomCommandHandler) buildImportClassroomArgs(ctx context.Context, db database.Ext, data []byte) (payloads *ImportClassroomPayload, errorCSVs []*lpb.ImportError) {
	var errors map[int]error
	sc1 := scanner.NewCSVScanner(bytes.NewReader(data))
	columnsIndex := map[string]int{
		"location_id":    0,
		"location_name":  1,
		"classroom_name": 3,
		"classroom_id":   -1,
		"room_area":      -1,
		"seat_capacity":  -1,
	}

	errors = ValidateImportFileHeader(sc1, columnsIndex)
	if len(errors) > 0 {
		errorCSVs = ConvertErrToImportCSVErr(errors)
		return nil, errorCSVs
	}

	locationIDs := make([]string, 0, len(sc1.GetRow()))
	classroomIDs := make([]string, 0, len(sc1.GetRow()))
	mapLocationName := make(map[string]string)
	for sc1.Scan() {
		if sc1.Text("location_id") != "" {
			locationID := sc1.Text("location_id")
			locationName := sc1.Text("location_name")
			locationIDs = append(locationIDs, locationID)
			lName, existed := mapLocationName[locationID]
			if existed && lName != locationName {
				errors[0] = fmt.Errorf("\"Invalid location %s: Existed name %s but got new name %s\"", locationID, lName, locationName)
				errorCSVs = ConvertErrToImportCSVErr(errors)
				return nil, errorCSVs
			}
			mapLocationName[locationID] = locationName
		}
		if sc1.Text("classroom_id") != "" {
			classroomIDs = append(classroomIDs, sc1.Text("classroom_id"))
		}
	}

	sc2 := scanner.NewCSVScanner(bytes.NewReader(data))
	payloads = NewImportClassroomPayload().
		WithScanner(sc2).
		WithLocationIDs(locationIDs).
		WithClassroomIDs(classroomIDs).
		WithMapLocationName(mapLocationName).
		WithClassroomRepo(cl.ClassroomRepo).
		WithMasterDataPort(cl.MasterDataPort)

	errors = payloads.buildImportClassroomPayload(ctx, db)
	if len(errors) > 0 {
		errorCSVs = ConvertErrToImportCSVErr(errors)
	}

	return payloads, errorCSVs
}

func ConvertErrToImportCSVErr(errors map[int]error) []*lpb.ImportError {
	errorCSVs := []*lpb.ImportError{}
	for line, err := range errors {
		errorCSVs = append(errorCSVs, &lpb.ImportError{
			RowNumber: int32(line),
			Error:     fmt.Sprintf("unable to parse this item: %s", err),
		})
	}
	return errorCSVs
}

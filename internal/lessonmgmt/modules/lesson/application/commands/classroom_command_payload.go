package commands

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/scanner"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure"
)

type ImportClassroomPayload struct {
	Classrooms      []*domain.Classroom
	Scanner         scanner.CSVScanner
	Timezone        string
	LocationIDs     []string
	ClassroomIDs    []string
	MapLocationName map[string]string

	// port
	ClassroomRepo  infrastructure.ClassroomRepo
	MasterDataPort infrastructure.MasterDataPort
}

func NewImportClassroomPayload() *ImportClassroomPayload {
	return &ImportClassroomPayload{}
}

func (p *ImportClassroomPayload) WithScanner(sc scanner.CSVScanner) *ImportClassroomPayload {
	p.Scanner = sc
	return p
}

func (p *ImportClassroomPayload) WithLocationIDs(ids []string) *ImportClassroomPayload {
	p.LocationIDs = ids
	return p
}

func (p *ImportClassroomPayload) WithMapLocationName(mapLocation map[string]string) *ImportClassroomPayload {
	p.MapLocationName = mapLocation
	return p
}

func (p *ImportClassroomPayload) WithClassroomIDs(ids []string) *ImportClassroomPayload {
	p.ClassroomIDs = ids
	return p
}

func (p *ImportClassroomPayload) WithClassroomRepo(repo infrastructure.ClassroomRepo) *ImportClassroomPayload {
	p.ClassroomRepo = repo
	return p
}

func (p *ImportClassroomPayload) WithMasterDataPort(port infrastructure.MasterDataPort) *ImportClassroomPayload {
	p.MasterDataPort = port
	return p
}

func (p *ImportClassroomPayload) IsValid(ctx context.Context, db database.Ext) map[int]error {
	errors := make(map[int]error)
	locationIDs := golibs.Uniq(p.LocationIDs)
	classroomIDs := golibs.Uniq(p.ClassroomIDs)

	if len(locationIDs) == 0 {
		errors[0] = fmt.Errorf("location_id is a required")
		return errors
	}

	err := p.MasterDataPort.CheckLocationByIDs(ctx, db, locationIDs, p.MapLocationName)
	if err != nil {
		errors[0] = fmt.Errorf("invalid location: %w", err)
		return errors
	}

	if len(classroomIDs) > 0 {
		err := p.ClassroomRepo.CheckClassroomIDs(ctx, db, classroomIDs)
		if err != nil {
			errors[0] = fmt.Errorf("the classroom_id not existed")
			return errors
		}
	}

	return errors
}

func (p *ImportClassroomPayload) buildImportClassroomPayload(ctx context.Context, db database.Ext) map[int]error {
	sc := p.Scanner
	errors := p.IsValid(ctx, db)
	if len(errors) > 0 {
		return errors
	}

	for sc.Scan() {
		locationID := sc.Text("location_id")
		classroomID := sc.Text("classroom_id")
		seatCapStr := sc.Text("seat_capacity")
		currentRow := sc.GetCurRow()

		if classroomID == "" {
			classroomID = idutil.ULIDNow()
		}

		if locationID == "" || sc.Text("location_name") == "" || sc.Text("classroom_name") == "" {
			errors[currentRow] = fmt.Errorf("missing mandatory value")
			continue
		}

		seatCap, err := strconv.Atoi(seatCapStr)
		if err != nil {
			errors[currentRow] = fmt.Errorf("capacity must be a numberic")
			continue
		}

		classroom := domain.NewClassroom(classroomID).
			WithName(sc.Text("classroom_name")).
			WithLocationID(locationID).
			WithRemark(sc.Text("remarks")).
			WithRoomArea(sc.Text("room_area")).
			WithSeatCapacity(seatCap).
			WithIsArchived(false).
			WithModificationTime(time.Now(), time.Now())

		p.Classrooms = append(p.Classrooms, classroom)
	}
	if len(errors) > 0 {
		return errors
	}
	return nil
}

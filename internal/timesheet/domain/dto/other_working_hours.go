package dto

import (
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	"github.com/manabie-com/backend/internal/timesheet/domain/constant"
	"github.com/manabie-com/backend/internal/timesheet/domain/entity"
	pbc "github.com/manabie-com/backend/pkg/genproto/bob"
	pb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"

	"github.com/jackc/pgtype"
)

type ListOtherWorkingHours []*OtherWorkingHours

type OtherWorkingHours struct {
	ID                string
	TimesheetID       string
	TimesheetConfigID string
	StartTime         time.Time
	EndTime           time.Time
	TotalHour         int16
	Remarks           string
	IsDeleted         bool
}

// ================= other working hours ==========

func NewOtherWorkingHoursFromRPCRequest(timesheetID string, req *pb.OtherWorkingHoursRequest) *OtherWorkingHours {
	timeStart := time.Time{}
	timeEnd := time.Time{}
	if req.GetStartTime() != nil {
		timeStart = req.GetStartTime().AsTime()
	}
	if req.GetEndTime() != nil {
		timeEnd = req.GetEndTime().AsTime()
	}
	return &OtherWorkingHours{
		ID:                req.GetOtherWorkingHoursId(),
		TimesheetID:       timesheetID,
		TimesheetConfigID: req.GetTimesheetConfigId(),
		StartTime:         timeStart,
		EndTime:           timeEnd,
		Remarks:           req.GetRemarks(),
	}
}

func NewOtherWorkingHoursFromEntity(owhsEntity *entity.OtherWorkingHours) *OtherWorkingHours {
	isDeleted := false
	if owhsEntity.DeletedAt.Status != pgtype.Null {
		isDeleted = true
	}

	return &OtherWorkingHours{
		ID:                owhsEntity.ID.String,
		TimesheetID:       owhsEntity.TimesheetID.String,
		TimesheetConfigID: owhsEntity.TimesheetConfigID.String,
		StartTime:         owhsEntity.StartTime.Time,
		EndTime:           owhsEntity.EndTime.Time,
		Remarks:           owhsEntity.Remarks.String,
		TotalHour:         owhsEntity.TotalHour.Int,
		IsDeleted:         isDeleted,
	}
}

func (h *OtherWorkingHours) Validate() error {
	switch {
	case h.TimesheetConfigID == "":
		return fmt.Errorf("other working type must not be empty")
	case h.StartTime.IsZero():
		return fmt.Errorf("other working hours start time must not be nil")
	case h.EndTime.IsZero():
		return fmt.Errorf("other working hours end time must not be nil")
	case h.EndTime.Equal(h.StartTime) || h.EndTime.Before(h.StartTime):
		return fmt.Errorf("other working hours end time must after start time")
	case len([]rune(h.Remarks)) > constant.KOtherWorkingHoursRemarksLimit:
		return fmt.Errorf("other working hours remarks must limit to 100 characters")
	}
	return nil
}

func (h *OtherWorkingHours) ToEntity() *entity.OtherWorkingHours {
	owhsE := &entity.OtherWorkingHours{
		ID:                database.Text(h.ID),
		TimesheetID:       database.Text(h.TimesheetID),
		TimesheetConfigID: database.Text(h.TimesheetConfigID),
		StartTime:         database.Timestamptz(h.StartTime),
		EndTime:           database.Timestamptz(h.EndTime),
		TotalHour:         database.Int2(h.TotalHour),
		Remarks:           database.Text(h.Remarks),
		CreatedAt:         pgtype.Timestamptz{Status: pgtype.Null},
		UpdatedAt:         pgtype.Timestamptz{Status: pgtype.Null},
		DeletedAt:         pgtype.Timestamptz{Status: pgtype.Null},
	}

	if h.IsDeleted {
		owhsE.DeletedAt = database.Timestamptz(time.Now())
	}
	return owhsE
}

func (h *OtherWorkingHours) NormalizedData() {
	h.StartTime = h.StartTime.In(timeutil.Timezone(pbc.COUNTRY_JP))
	h.EndTime = h.EndTime.In(timeutil.Timezone(pbc.COUNTRY_JP))

	h.StartTime = time.Date(h.StartTime.Year(), h.StartTime.Month(), h.StartTime.Day(), h.StartTime.Hour(), h.StartTime.Minute(), 0, 0, h.StartTime.Location())
	h.EndTime = time.Date(h.EndTime.Year(), h.EndTime.Month(), h.EndTime.Day(), h.EndTime.Hour(), h.EndTime.Minute(), 0, 0, h.EndTime.Location())
}

func (h *OtherWorkingHours) IsEqual(other *OtherWorkingHours) bool {
	if h == nil && other == nil {
		return true
	}

	if h == nil || other == nil {
		return false
	}

	if h.ID == other.ID &&
		h.TimesheetID == other.TimesheetID &&
		h.TimesheetConfigID == other.TimesheetConfigID &&
		h.TotalHour == other.TotalHour &&
		h.Remarks == other.Remarks &&
		h.StartTime.Equal(other.StartTime) &&
		h.EndTime.Equal(other.EndTime) {
		return true
	}
	return false
}

func NewListOtherWorkingHoursFromRPCRequest(timesheetID string, listOWHsReq []*pb.OtherWorkingHoursRequest) ListOtherWorkingHours {
	listOWHs := make(ListOtherWorkingHours, 0, len(listOWHsReq))
	for _, owhsReq := range listOWHsReq {
		owhs := NewOtherWorkingHoursFromRPCRequest(timesheetID, owhsReq)
		listOWHs = append(listOWHs, owhs)
	}
	return listOWHs
}

func NewListOtherWorkingHoursFromEntities(listOWHsEntity []*entity.OtherWorkingHours) ListOtherWorkingHours {
	listOWHs := make(ListOtherWorkingHours, 0, len(listOWHsEntity))
	for _, owhsReq := range listOWHsEntity {
		owhs := NewOtherWorkingHoursFromEntity(owhsReq)
		listOWHs = append(listOWHs, owhs)
	}
	return listOWHs
}

func (lh ListOtherWorkingHours) Validate() error {
	if len(lh) > constant.KListOtherWorkingHoursLimit {
		return fmt.Errorf("list other working hours must be limit to 5 rows")
	}

	for _, owhs := range lh {
		err := owhs.Validate()
		if err != nil {
			return err
		}
	}
	return nil
}

func (lh ListOtherWorkingHours) ToEntities() []*entity.OtherWorkingHours {
	listOWHsE := make(entity.ListOtherWorkingHours, 0, len(lh))

	for _, owhs := range lh {
		owhsE := owhs.ToEntity()
		listOWHsE = append(listOWHsE, owhsE)
	}
	return listOWHsE
}

func (lh ListOtherWorkingHours) UpdateTimesheetID(timsheetID string) {
	for i := range lh {
		lh[i].TimesheetID = timsheetID
	}
}

func (lh ListOtherWorkingHours) IsEqual(listOtherWorkingHours ListOtherWorkingHours) bool {
	if len(lh) != len(listOtherWorkingHours) {
		return false
	}
	if len(lh) == 0 && len(listOtherWorkingHours) == 0 {
		return true
	}
	mapOtherWorkingHours := make(map[string]*OtherWorkingHours, len(lh))
	for _, e := range lh {
		mapOtherWorkingHours[e.ID] = e
	}

	for _, e := range listOtherWorkingHours {
		if !e.IsEqual(mapOtherWorkingHours[e.ID]) {
			return false
		}
	}
	return true
}

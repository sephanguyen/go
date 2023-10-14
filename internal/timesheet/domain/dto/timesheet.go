package dto

import (
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	"github.com/manabie-com/backend/internal/timesheet/domain/constant"
	"github.com/manabie-com/backend/internal/timesheet/domain/entity"
	pbc "github.com/manabie-com/backend/pkg/genproto/bob"
	pb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"

	"github.com/jackc/pgtype"
)

type Timesheet struct {
	ID                         string
	StaffID                    string
	LocationID                 string
	TimesheetStatus            string
	TimesheetDate              time.Time
	Remark                     string
	ListOtherWorkingHours      ListOtherWorkingHours
	ListTimesheetLessonHours   ListTimesheetLessonHours
	ListTransportationExpenses ListTransportationExpenses
	IsCreated                  bool
	IsDeleted                  bool
}

type TimesheetLocationDto struct {
	TimesheetID string
	LocationID  string
}

func NewTimesheetFromRPCCreateRequest(req *pb.CreateTimesheetRequest) *Timesheet {
	timesheet := &Timesheet{}
	timesheet.StaffID = req.GetStaffId()
	timesheet.LocationID = req.GetLocationId()
	timesheet.Remark = req.GetRemark()

	if req.GetTimesheetDate() != nil {
		timesheet.TimesheetDate = req.GetTimesheetDate().AsTime()
	}

	if len(req.GetListOtherWorkingHours()) > 0 {
		timesheet.ListOtherWorkingHours = NewListOtherWorkingHoursFromRPCRequest(timesheet.ID, req.GetListOtherWorkingHours())
	}

	if len(req.GetListTransportationExpenses()) > 0 {
		timesheet.ListTransportationExpenses = NewListTransportExpensesFromRPCRequest(timesheet.ID, req.GetListTransportationExpenses())
	}

	return timesheet
}

func NewTimesheetFromRPCUpdateRequest(req *pb.UpdateTimesheetRequest) *Timesheet {
	timesheet := &Timesheet{}
	timesheet.ID = req.GetTimesheetId()
	timesheet.Remark = req.GetRemark()

	if len(req.GetListOtherWorkingHours()) > 0 {
		timesheet.ListOtherWorkingHours = NewListOtherWorkingHoursFromRPCRequest(timesheet.ID, req.GetListOtherWorkingHours())
	}

	if len(req.GetListTransportationExpenses()) > 0 {
		timesheet.ListTransportationExpenses = NewListTransportExpensesFromRPCRequest(timesheet.ID, req.GetListTransportationExpenses())
	}
	return timesheet
}

func NewTimesheetFromEntity(timesheetE *entity.Timesheet) *Timesheet {
	isCreated := false
	isDeleted := false
	if !timesheetE.CreatedAt.Time.IsZero() {
		isCreated = true
	}
	if !timesheetE.DeletedAt.Time.IsZero() {
		isDeleted = true
	}

	return &Timesheet{
		ID:              timesheetE.TimesheetID.String,
		StaffID:         timesheetE.StaffID.String,
		LocationID:      timesheetE.LocationID.String,
		TimesheetStatus: timesheetE.TimesheetStatus.String,
		TimesheetDate:   timesheetE.TimesheetDate.Time,
		Remark:          timesheetE.Remark.String,
		IsCreated:       isCreated,
		IsDeleted:       isDeleted,
	}
}

func (t *Timesheet) ValidateCreateInfo() error {
	switch {
	case t.StaffID == "":
		return fmt.Errorf("staff id must not be empty")
	case t.LocationID == "":
		return fmt.Errorf("location id must not be empty")
	case t.TimesheetDate.IsZero():
		return fmt.Errorf("date must not be nil")
	case t.TimesheetDate.Before(GetMinTimesheetDateValid()):
		return fmt.Errorf("date must be greater than 1st Jan 2022")
	case len([]rune(t.Remark)) > constant.KTimesheetRemarkLimit:
		return fmt.Errorf("remark must be limit to 500 characters")
	}

	if len(t.ListOtherWorkingHours) == 0 {
		return fmt.Errorf("other working hours must be not empty")
	}

	if len(t.ListOtherWorkingHours) > 0 {
		err := t.ListOtherWorkingHours.Validate()
		if err != nil {
			return err
		}
	}

	if len(t.ListTransportationExpenses) > 0 {
		return t.ListTransportationExpenses.Validate()
	}

	return nil
}

func (t *Timesheet) ValidateUpdateInfo() error {
	switch {
	case t.ID == "":
		return fmt.Errorf("timesheet id must not be empty")
	case len([]rune(t.Remark)) > constant.KTimesheetRemarkLimit:
		return fmt.Errorf("remark must be limit to 500 characters")
	}

	if len(t.ListOtherWorkingHours) > 0 {
		err := t.ListOtherWorkingHours.Validate()
		if err != nil {
			return err
		}
	}

	if len(t.ListTransportationExpenses) > 0 {
		return t.ListTransportationExpenses.Validate()
	}

	return nil
}

func (t *Timesheet) ToEntity() *entity.Timesheet {
	timesheetE := &entity.Timesheet{
		TimesheetID:     database.Text(t.ID),
		StaffID:         database.Text(t.StaffID),
		LocationID:      database.Text(t.LocationID),
		TimesheetStatus: database.Text(t.TimesheetStatus),
		TimesheetDate:   database.Timestamptz(t.TimesheetDate),
		Remark:          database.Text(t.Remark),
		CreatedAt:       pgtype.Timestamptz{Status: pgtype.Null},
		UpdatedAt:       pgtype.Timestamptz{Status: pgtype.Null},
		DeletedAt:       pgtype.Timestamptz{Status: pgtype.Null},
	}

	if len(t.Remark) == 0 {
		timesheetE.Remark = pgtype.Text{Status: pgtype.Null}
	}

	return timesheetE
}

func (t *Timesheet) MakeNewID() {
	t.ID = idutil.ULIDNow()

	if len(t.ListOtherWorkingHours) > 0 {
		t.ListOtherWorkingHours.UpdateTimesheetID(t.ID)
	}
	if len(t.ListTimesheetLessonHours) > 0 {
		t.ListTimesheetLessonHours.UpdateTimesheetID(t.ID)
	}
}

func (t *Timesheet) NormalizedData() {
	t.TimesheetDate = t.TimesheetDate.In(timeutil.Timezone(pbc.COUNTRY_JP))
	t.TimesheetDate = time.Date(t.TimesheetDate.Year(), t.TimesheetDate.Month(), t.TimesheetDate.Day(), 0, 0, 0, 0, t.TimesheetDate.Location())
}

func (t *Timesheet) GetTimesheetLessonHoursNew() []*TimesheetLessonHours {
	var listTimesheetLessonHoursNew []*TimesheetLessonHours
	for _, e := range t.ListTimesheetLessonHours {
		if !e.IsCreated {
			listTimesheetLessonHoursNew = append(listTimesheetLessonHoursNew, e)
		}
	}
	return listTimesheetLessonHoursNew
}

// validateMerge validate 2 timesheet want to merge
// 2 timesheet must have same general info,
// and if both are created or not created yet, 2 timesheet must have same timesheet status and remark
func (t *Timesheet) validateMerge(timesheet2 *Timesheet) bool {
	normalizedDate1 := timeutil.NormalizeToStartOfDay(t.TimesheetDate, pbc.COUNTRY_JP).Format("2006-01-02")
	normalizedDate2 := timeutil.NormalizeToStartOfDay(timesheet2.TimesheetDate, pbc.COUNTRY_JP).Format("2006-01-02")
	if t.StaffID != timesheet2.StaffID ||
		t.LocationID != timesheet2.LocationID ||
		!(normalizedDate1 == normalizedDate2) {
		return false
	}

	if (t.IsCreated && timesheet2.IsCreated) || (!t.IsCreated && !timesheet2.IsCreated) {
		if t.ID != timesheet2.ID ||
			t.TimesheetStatus != timesheet2.TimesheetStatus ||
			t.Remark != timesheet2.Remark {
			return false
		}
	}

	if t.ID != "" && timesheet2.ID != "" && t.ID != timesheet2.ID {
		return false
	}

	if (t.IsCreated && t.ID == "") ||
		(timesheet2.IsCreated && timesheet2.ID == "") {
		return false
	}

	return true
}

// Merge will merge 2 timesheet have the same general info as one timesheet
func (t *Timesheet) Merge(timesheet2 *Timesheet) (*Timesheet, error) {
	if !t.validateMerge(timesheet2) {
		return nil, fmt.Errorf("validateMerge failed")
	}

	var baseTimesheet *Timesheet
	if t.IsCreated {
		baseTimesheet = t
	} else if timesheet2.IsCreated {
		baseTimesheet = timesheet2
	} else {
		// if validateMerge success and t.IsCreated == timesheet2.IsCreated == false
		// then timesheet info of t and timesheet2 is the same
		// and we can use t as baseTimesheet
		baseTimesheet = t
	}

	mergedListTimesheetLessonHours := t.ListTimesheetLessonHours.Merge(timesheet2.ListTimesheetLessonHours)

	for _, e := range mergedListTimesheetLessonHours {
		if e.TimesheetID == "" {
			e.TimesheetID = baseTimesheet.ID
		}
	}

	mergedTimesheet := &Timesheet{
		ID:                       baseTimesheet.ID,
		StaffID:                  baseTimesheet.StaffID,
		LocationID:               baseTimesheet.LocationID,
		TimesheetStatus:          baseTimesheet.TimesheetStatus,
		TimesheetDate:            baseTimesheet.TimesheetDate,
		Remark:                   baseTimesheet.Remark,
		ListOtherWorkingHours:    baseTimesheet.ListOtherWorkingHours,
		ListTimesheetLessonHours: mergedListTimesheetLessonHours,
		IsCreated:                baseTimesheet.IsCreated,
	}

	return mergedTimesheet, nil
}

func (t *Timesheet) IsEqual(timesheet2 *Timesheet) bool {
	if t == nil && timesheet2 == nil {
		return true
	}

	if t == nil || timesheet2 == nil {
		return false
	}

	if !t.TimesheetDate.Equal(timesheet2.TimesheetDate) ||
		t.TimesheetStatus != timesheet2.TimesheetStatus ||
		t.StaffID != timesheet2.StaffID ||
		t.LocationID != timesheet2.LocationID ||
		t.Remark != timesheet2.Remark ||
		t.ID != timesheet2.ID ||
		t.IsCreated != timesheet2.IsCreated {
		return false
	}

	if !t.ListTimesheetLessonHours.IsEqual(timesheet2.ListTimesheetLessonHours) {
		return false
	}

	if !t.ListOtherWorkingHours.IsEqual(timesheet2.ListOtherWorkingHours) {
		return false
	}

	return true
}

func (t *Timesheet) DeleteTimesheetLessonHours(lessonID string) bool {
	for _, timesheetLessonHours := range t.ListTimesheetLessonHours {
		if timesheetLessonHours.LessonID == lessonID {
			timesheetLessonHours.IsDeleted = true
			return true
		}
	}
	return false
}

// IsTimesheetEmpty return true if  timesheet only contains general info and status
func (t *Timesheet) IsTimesheetEmpty() bool {
	for _, otherWorkingHours := range t.ListOtherWorkingHours {
		if !otherWorkingHours.IsDeleted {
			return false
		}
	}

	for _, timesheetLessonHours := range t.ListTimesheetLessonHours {
		if !timesheetLessonHours.IsDeleted {
			return false
		}
	}

	for _, transportExpense := range t.ListTransportationExpenses {
		if !transportExpense.IsDeleted {
			return false
		}
	}

	return true
}

func GetMinTimesheetDateValid() time.Time {
	return time.Date(2022, 01, 01, 00, 00, 00, 00, timeutil.Timezone(pbc.COUNTRY_JP))
}

func GetTimesheetEndOfMonthDate() time.Time {
	now := time.Now()
	firstDay := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, timeutil.Timezone(pbc.COUNTRY_JP))
	// first day of month + 1 month - 1 nanosecond = {END_OF_MONTH} 23:59:59.999999999
	return firstDay.AddDate(0, 1, 0).Add(time.Nanosecond * -1)
}

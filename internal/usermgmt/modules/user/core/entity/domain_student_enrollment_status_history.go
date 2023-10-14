package entity

import (
	"sort"
	"time"

	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/valueobj"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/utils"
)

const (
	StartDateFieldEnrollmentStatusHistory = "start_date"
	EndDateFieldEnrollmentStatusHistory   = "end_date"
)

type FieldEnrollmentStatusHistory string

const (
	FieldEnrollmentStatusHistoryEnrollmentStatus    FieldEnrollmentStatusHistory = "enrollment_status"
	FieldEnrollmentStatusHistoryStartDate           FieldEnrollmentStatusHistory = "start_date"
	FieldEnrollmentStatusHistoryEndDate             FieldEnrollmentStatusHistory = "end_date"
	FieldEnrollmentStatusHistoryOrderID             FieldEnrollmentStatusHistory = "order_id"
	FieldEnrollmentStatusHistoryOrderSequenceNumber FieldEnrollmentStatusHistory = "order_sequence_number"
)

type EnrollmentStatusHistory interface {
	EnrollmentStatus() field.String
	StartDate() field.Time
	EndDate() field.Time
	OrderID() field.String
	OrderSequenceNumber() field.Int32
}

type DomainEnrollmentStatusHistory interface {
	EnrollmentStatusHistory
	valueobj.HasUserID
	valueobj.HasLocationID
	valueobj.HasOrganizationID
	valueobj.HasCreatedAt
}

type EnrollmentStatusHistoryWillBeDelegated struct {
	EnrollmentStatusHistory
	valueobj.HasUserID
	valueobj.HasLocationID
	valueobj.HasOrganizationID
	valueobj.HasCreatedAt
}

type DefaultDomainEnrollmentStatusHistory struct{}

func (e DefaultDomainEnrollmentStatusHistory) UserID() field.String {
	return field.NewNullString()
}

func (e DefaultDomainEnrollmentStatusHistory) LocationID() field.String {
	return field.NewNullString()
}

func (e DefaultDomainEnrollmentStatusHistory) EnrollmentStatus() field.String {
	return field.NewNullString()
}

func (e DefaultDomainEnrollmentStatusHistory) StartDate() field.Time {
	return field.NewNullTime()
}

func (e DefaultDomainEnrollmentStatusHistory) EndDate() field.Time {
	return field.NewNullTime()
}

func (e DefaultDomainEnrollmentStatusHistory) OrderID() field.String {
	return field.NewNullString()
}

func (e DefaultDomainEnrollmentStatusHistory) OrderSequenceNumber() field.Int32 {
	return field.NewNullInt32()
}

func (e DefaultDomainEnrollmentStatusHistory) OrganizationID() field.String {
	return field.NewNullString()
}

func (e DefaultDomainEnrollmentStatusHistory) CreatedAt() field.Time {
	return field.NewNullTime()
}

type DomainEnrollmentStatusHistories []DomainEnrollmentStatusHistory

func (e DomainEnrollmentStatusHistories) EnrollmentStatuses() []field.String {
	enrollmentStatuses := make([]field.String, 0)
	for _, enrollmentStatusHistory := range e {
		enrollmentStatuses = append(enrollmentStatuses, enrollmentStatusHistory.EnrollmentStatus())
	}
	return enrollmentStatuses
}

func (e DomainEnrollmentStatusHistories) UserIDs() []field.String {
	userIDs := make([]field.String, 0, len(e))
	for _, enrollmentStatusHistory := range e {
		userIDs = append(userIDs, enrollmentStatusHistory.UserID())
	}
	return userIDs
}

func (e DomainEnrollmentStatusHistories) LocationIDs() []field.String {
	locationIDs := make([]field.String, 0, len(e))
	for _, enrollmentStatusHistory := range e {
		locationIDs = append(locationIDs, enrollmentStatusHistory.LocationID())
	}
	return locationIDs
}

func (e DomainEnrollmentStatusHistories) GetActivatedByRequest(userID field.String, reqEnrollmentStatusHistories DomainEnrollmentStatusHistories) DomainEnrollmentStatusHistories {
	enrollmentStatusHistories := make(DomainEnrollmentStatusHistories, 0)
	roundedCurrentDateTime := utils.TruncateTimeToStartOfDay(time.Now())
	for _, enrollmentStatusHistoryDB := range e {
		roundedStartDateTimeFromDB := utils.TruncateTimeToStartOfDay(enrollmentStatusHistoryDB.StartDate().Time())
		roundedEndDateTimeFromDB := utils.TruncateTimeToStartOfDay(enrollmentStatusHistoryDB.EndDate().Time())
		if enrollmentStatusHistoryDB.UserID().Equal(userID) {
			if roundedStartDateTimeFromDB.Before(roundedCurrentDateTime) || roundedStartDateTimeFromDB.Equal(roundedCurrentDateTime) {
				if roundedEndDateTimeFromDB.After(roundedCurrentDateTime) || enrollmentStatusHistoryDB.EndDate().Time().IsZero() {
					enrollmentStatusHistories = append(enrollmentStatusHistories, enrollmentStatusHistoryDB)
				}
			}
		}
	}

	validActivatedEnrollmentStatusHistories := make(DomainEnrollmentStatusHistories, 0)
	for _, enrollmentStatusHistoryDB := range enrollmentStatusHistories {
		for _, reqEnrollmentStatusHistory := range reqEnrollmentStatusHistories {
			if reqEnrollmentStatusHistory.EnrollmentStatus().String() != StudentEnrollmentStatusTemporary {
				validActivatedEnrollmentStatusHistories = append(validActivatedEnrollmentStatusHistories, enrollmentStatusHistoryDB)
			}
		}
	}
	return validActivatedEnrollmentStatusHistories
}

func (e DomainEnrollmentStatusHistories) GetByUserIDWithUniqLocation(userID field.String) DomainEnrollmentStatusHistories {
	enrollmentStatusHistories := make(DomainEnrollmentStatusHistories, 0)
	uniqLocation := make(map[string]bool)
	for _, enrollmentStatusHistoryDB := range e {
		if enrollmentStatusHistoryDB.UserID().Equal(userID) {
			if !uniqLocation[enrollmentStatusHistoryDB.LocationID().String()] {
				uniqLocation[enrollmentStatusHistoryDB.LocationID().String()] = true
				enrollmentStatusHistories = append(enrollmentStatusHistories, enrollmentStatusHistoryDB)
			}
		}
	}
	return enrollmentStatusHistories
}

func (e DomainEnrollmentStatusHistories) GetActivatedByUserIDLocationID(userID field.String, locationID field.String) DomainEnrollmentStatusHistory {
	roundedCurrentDateTime := utils.TruncateTimeToStartOfDay(time.Now())
	for _, enrollmentStatusHistoryDB := range e {
		roundedStartDateTimeFromDB := utils.TruncateTimeToStartOfDay(enrollmentStatusHistoryDB.StartDate().Time())
		roundedEndDateTimeFromDB := utils.TruncateTimeToStartOfDay(enrollmentStatusHistoryDB.EndDate().Time())
		if enrollmentStatusHistoryDB.UserID().Equal(userID) && enrollmentStatusHistoryDB.LocationID().Equal(locationID) {
			if roundedStartDateTimeFromDB.Before(roundedCurrentDateTime) || roundedStartDateTimeFromDB.Equal(roundedCurrentDateTime) {
				if roundedEndDateTimeFromDB.After(roundedCurrentDateTime) || enrollmentStatusHistoryDB.EndDate().Time().IsZero() {
					return enrollmentStatusHistoryDB
				}
			}
		}
	}
	return nil
}

func (e DomainEnrollmentStatusHistories) GetExactly(enrollmentStatusHistoryReq DomainEnrollmentStatusHistory) DomainEnrollmentStatusHistory {
	roundedStartDateTimeReq := utils.TruncateTimeToStartOfDay(enrollmentStatusHistoryReq.StartDate().Time())
	roundedEndDateTimeReq := utils.TruncateTimeToStartOfDay(enrollmentStatusHistoryReq.EndDate().Time())
	for _, enrollmentStatusHistoryDB := range e {
		roundedStartDateTimeFromDB := utils.TruncateTimeToStartOfDay(enrollmentStatusHistoryDB.StartDate().Time())
		roundedEndDateTimeFromDB := utils.TruncateTimeToStartOfDay(enrollmentStatusHistoryDB.EndDate().Time())
		if enrollmentStatusHistoryDB.UserID().Equal(enrollmentStatusHistoryReq.UserID()) &&
			enrollmentStatusHistoryDB.LocationID().Equal(enrollmentStatusHistoryReq.LocationID()) &&
			roundedStartDateTimeFromDB.Equal(roundedStartDateTimeReq) &&
			roundedEndDateTimeFromDB.Equal(roundedEndDateTimeReq) &&
			enrollmentStatusHistoryDB.EnrollmentStatus().Equal(enrollmentStatusHistoryReq.EnrollmentStatus()) {
			return enrollmentStatusHistoryDB
		}
	}
	return nil
}

func (e DomainEnrollmentStatusHistories) GetLatestByUserIDLocationID(userID field.String, locationID field.String) DomainEnrollmentStatusHistory {
	enrollmentStatusHistoriesDB := make(DomainEnrollmentStatusHistories, 0)
	for _, enrollmentStatusHistoryDB := range e {
		if enrollmentStatusHistoryDB.UserID().Equal(userID) &&
			enrollmentStatusHistoryDB.LocationID().Equal(locationID) {
			enrollmentStatusHistoriesDB = append(enrollmentStatusHistoriesDB, enrollmentStatusHistoryDB)
		}
	}
	sort.Slice(enrollmentStatusHistoriesDB, func(i, j int) bool {
		return enrollmentStatusHistoriesDB[i].CreatedAt().Time().After(enrollmentStatusHistoriesDB[j].CreatedAt().Time())
	})

	if len(enrollmentStatusHistoriesDB) == 0 {
		return nil
	}
	return enrollmentStatusHistoriesDB[0]
}

func (e DomainEnrollmentStatusHistories) GetAllByUserIDLocationID(userID field.String, locationID field.String) DomainEnrollmentStatusHistories {
	enrollmentStatusHistoriesDB := make(DomainEnrollmentStatusHistories, 0)
	for _, enrollmentStatusHistoryDB := range e {
		if enrollmentStatusHistoryDB.UserID().Equal(userID) &&
			enrollmentStatusHistoryDB.LocationID().Equal(locationID) {
			enrollmentStatusHistoriesDB = append(enrollmentStatusHistoriesDB, enrollmentStatusHistoryDB)
		}
	}
	return enrollmentStatusHistoriesDB
}

type enrollmentStatusHistoryWillBeDelegated struct {
	EnrollmentStatusHistoryWillBeDelegated

	startDate field.Time
}

func (e enrollmentStatusHistoryWillBeDelegated) StartDate() field.Time {
	return e.startDate
}

func NewEnrollmentStatusHistoryWithStartDate(history DomainEnrollmentStatusHistory, startDate field.Time) DomainEnrollmentStatusHistory {
	return enrollmentStatusHistoryWillBeDelegated{
		EnrollmentStatusHistoryWillBeDelegated: EnrollmentStatusHistoryWillBeDelegated{
			EnrollmentStatusHistory: history,
			HasUserID:               history,
			HasLocationID:           history,
			HasOrganizationID:       history,
			HasCreatedAt:            history,
		},
		startDate: startDate,
	}
}

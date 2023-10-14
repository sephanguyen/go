package payloads

import (
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/calendar/domain/dto"
)

type FetchDateInfoByDateRangeRequest struct {
	StartDate  time.Time
	EndDate    time.Time
	LocationID string
	Timezone   string
}

type FetchDateInfoByDateRangeResponse struct {
	DateInfos []*dto.DateInfo
}

func (r *FetchDateInfoByDateRangeRequest) Validate() error {
	if r.StartDate.IsZero() {
		return fmt.Errorf("start date cannot be empty")
	}

	if r.EndDate.IsZero() {
		return fmt.Errorf("end date cannot be empty")
	}

	if len(r.LocationID) == 0 {
		return fmt.Errorf("location ID cannot be empty")
	}

	if r.EndDate.Before(r.StartDate) {
		return fmt.Errorf("end date could not before start date")
	}

	if len(r.Timezone) == 0 {
		r.Timezone = "UTC"
	}

	return nil
}

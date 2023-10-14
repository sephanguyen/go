package dto

import (
	"time"
)

type TimesheetCountByStatusAndLocationIdsReq struct {
	Status      string
	FromDate    time.Time
	ToDate      time.Time
	LocationIds []string
}

type TimesheetCountByStatusAndLocationIdsResp struct {
	Count int64
}

package dto

import (
	pb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"
)

type CountSubmittedTimesheetsReq struct {
	LocationIds []string
}

type CountSubmittedTimesheetsResp struct {
	Count int64
}

func NewCountSubmittedTimesheetsRequest(req *pb.CountSubmittedTimesheetsRequest) *CountSubmittedTimesheetsReq {
	return &CountSubmittedTimesheetsReq{
		LocationIds: req.GetLocationIds(),
	}
}

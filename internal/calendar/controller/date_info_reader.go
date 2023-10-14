package controller

import (
	"context"

	"github.com/manabie-com/backend/internal/calendar/application"
	"github.com/manabie-com/backend/internal/calendar/application/queries"
	"github.com/manabie-com/backend/internal/calendar/application/queries/payloads"
	"github.com/manabie-com/backend/internal/calendar/domain/dto"
	"github.com/manabie-com/backend/internal/calendar/infrastructure"
	"github.com/manabie-com/backend/internal/golibs/database"
	cpb "github.com/manabie-com/backend/pkg/manabuf/calendar/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type DateInfoReaderService struct {
	dateInfoQueryHandler application.QueryDateInfoPort
}

func NewDateInfoReaderService(db database.QueryExecer, dateInfoRepo infrastructure.DateInfoPort) *DateInfoReaderService {
	return &DateInfoReaderService{
		dateInfoQueryHandler: &queries.DateInfoQueryHandler{
			DB:           db,
			DateInfoRepo: dateInfoRepo,
		},
	}
}

func (d *DateInfoReaderService) FetchDateInfo(ctx context.Context, req *cpb.FetchDateInfoRequest) (*cpb.FetchDateInfoResponse, error) {
	request := &payloads.FetchDateInfoByDateRangeRequest{
		StartDate:  req.StartDate.AsTime(),
		EndDate:    req.EndDate.AsTime(),
		LocationID: req.GetLocationId(),
		Timezone:   req.GetTimezone(),
	}

	if err := request.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	response, err := d.dateInfoQueryHandler.FetchDateInfoByDateRangeAndLocationID(ctx, request)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	dateInfoListLength := len(response.DateInfos)
	if dateInfoListLength == 0 {
		return &cpb.FetchDateInfoResponse{
			Successful: true,
			Message:    "no date info found",
		}, nil
	}

	items := make([]*cpb.DateInfoDetailed, 0, dateInfoListLength)
	for _, dateInfo := range response.DateInfos {
		items = append(items, convertToDateInfoDetailedProto(dateInfo))
	}

	return &cpb.FetchDateInfoResponse{
		Successful: true,
		Message:    "fetched date info list successfully",
		DateInfos:  items,
	}, nil
}

func convertToDateInfoDetailedProto(dateInfo *dto.DateInfo) *cpb.DateInfoDetailed {
	return &cpb.DateInfoDetailed{
		DateInfo: &cpb.DateInfo{
			Date:        timestamppb.New(dateInfo.Date),
			LocationId:  dateInfo.LocationID,
			DateTypeId:  dateInfo.DateTypeID,
			OpeningTime: dateInfo.OpeningTime,
			Status:      dateInfo.Status,
			Timezone:    dateInfo.TimeZone,
		},
		DateTypeDisplayName: dateInfo.DateTypeDisplayName,
	}
}

func (d *DateInfoReaderService) ExportDayInfo(ctx context.Context, req *cpb.ExportDayInfoRequest) (res *cpb.ExportDayInfoResponse, err error) {
	bytes, err := d.dateInfoQueryHandler.ExportDayInfo(ctx)
	if err != nil {
		return &cpb.ExportDayInfoResponse{}, status.Error(codes.Internal, err.Error())
	}
	res = &cpb.ExportDayInfoResponse{
		Data: bytes,
	}
	return res, nil
}

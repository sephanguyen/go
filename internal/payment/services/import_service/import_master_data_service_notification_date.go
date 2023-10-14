package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"log"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *ImportMasterDataService) ImportNotificationDate(ctx context.Context, req *pb.ImportNotificationDateRequest) (*pb.ImportNotificationDateResponse, error) {
	errors := make([]*pb.ImportNotificationDateResponse_ImportNotificationDateError, 0)

	r := csv.NewReader(bytes.NewReader(req.Payload))
	lines, err := r.ReadAll()
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if len(lines) < 2 {
		return nil, status.Error(codes.InvalidArgument, constant.NoDataInCsvFile)
	}

	header := lines[0]
	headerTitles := []string{
		"notification_date_id",
		"order_type",
		"notification_date",
		"is_archived",
	}
	err = utils.ValidateCsvHeader(
		len(headerTitles),
		header,
		headerTitles,
	)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("csv file invalid format - %s", err.Error()))
	}
	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) (err error) {
		// first line is header
		mapOrderType := make(map[string]bool, 0)
		for i, line := range lines[1:] {
			notificationDate, err := NotificationDateEntityFromCsv(line)
			if err != nil {
				errors = append(errors, &pb.ImportNotificationDateResponse_ImportNotificationDateError{
					RowNumber: int32(i) + 2, // i = 0 <=> line number 2 in csv file
					Error:     fmt.Sprintf(constant.UnableToParseNotificationDate, err),
				})
				continue
			}
			err = validateNotificationDate(notificationDate, mapOrderType)
			if err != nil {
				errors = append(errors, &pb.ImportNotificationDateResponse_ImportNotificationDateError{
					RowNumber: int32(i) + 2, // i = 0 <=> line number 2 in csv file
					Error:     fmt.Sprintf("invalid notification date: %s", err),
				})
				continue
			}

			err = s.NotificationDateRepo.Upsert(ctx, tx, notificationDate)
			if err != nil {
				upsertErr := fmt.Sprintf("unable to create new notification date item: %s", err)
				if notificationDate.NotificationDateID.Get() != nil {
					upsertErr = fmt.Sprintf("unable to update notification date item: %s", err)
				}

				errors = append(errors, &pb.ImportNotificationDateResponse_ImportNotificationDateError{
					RowNumber: int32(i) + 2,
					Error:     upsertErr,
				})
			}
		}
		if len(errors) > 0 {
			return fmt.Errorf(errors[0].Error)
		}
		return nil
	})
	if err != nil {
		log.Printf("Error when importing notification date: %s", err.Error())
	}
	return &pb.ImportNotificationDateResponse{
		Errors: errors,
	}, nil
}

func validateNotificationDate(notificationDate *entities.NotificationDate, mapOrderType map[string]bool) error {
	if notificationDate.NotificationDate.Int < 1 || notificationDate.NotificationDate.Int > 31 {
		return fmt.Errorf("notification_date should be greater than 30 and smaller than 0")
	}
	if _, exists := pb.OrderType_value[notificationDate.OrderType.String]; !exists {
		return fmt.Errorf("order_type is invalid")
	}
	if _, exists := mapOrderType[notificationDate.OrderType.String]; exists {
		return fmt.Errorf("duplicate order type: %s", notificationDate.OrderType.String)
	}
	mapOrderType[notificationDate.OrderType.String] = true

	return nil
}
func getMandatoryColumnsForNotificationDate(i int) string {
	mandatoryColumns := []string{
		"notification_date_id",
		"order_type",
		"notification_date",
		"is_archived",
	}
	return mandatoryColumns[i]
}

func NotificationDateEntityFromCsv(line []string) (*entities.NotificationDate, error) {
	const (
		NotificationDateID = iota
		OrderType
		NotificationDate
		IsArchived
	)

	mandatory := []int{OrderType, NotificationDate, IsArchived}

	areMandatoryDataPresent, colPosition := checkMandatoryColumnAndGetIndex(line, mandatory)
	if !areMandatoryDataPresent {
		return nil, fmt.Errorf("missing mandatory data: %v", getMandatoryColumnsForNotificationDate(colPosition))
	}

	notificationDate := &entities.NotificationDate{}

	if err := multierr.Combine(
		utils.StringToFormatString("notification_date_id", line[NotificationDateID], true, notificationDate.NotificationDateID.Set),
		utils.StringToFormatString("order_type", line[OrderType], false, notificationDate.OrderType.Set),
		utils.StringToInt("notification_date", line[NotificationDate], false, notificationDate.NotificationDate.Set),
		utils.StringToBool("is_archived", line[IsArchived], false, notificationDate.IsArchived.Set),
	); err != nil {
		return nil, err
	}

	return notificationDate, nil
}

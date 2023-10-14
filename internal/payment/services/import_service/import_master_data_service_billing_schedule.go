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

func (s *ImportMasterDataService) ImportBillingSchedule(ctx context.Context, req *pb.ImportBillingScheduleRequest) (*pb.ImportBillingScheduleResponse, error) {
	errors := []*pb.ImportBillingScheduleResponse_ImportBillingScheduleError{}

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
		"billing_schedule_id",
		"name",
		"remarks",
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
		for i, line := range lines[1:] {
			billingSchedule, err := BillingScheduleFromCsv(line, headerTitles)
			if err != nil {
				errors = append(errors, &pb.ImportBillingScheduleResponse_ImportBillingScheduleError{
					RowNumber: int32(i) + 2, // i = 0 <=> line number 2 in csv file
					Error:     fmt.Sprintf(constant.UnableToParseBillingScheduleItem, err),
				})
				continue
			}
			if billingSchedule.BillingScheduleID.Get() == nil {
				err = s.BillingScheduleRepo.Create(ctx, tx, billingSchedule)
				if err != nil {
					errors = append(errors, &pb.ImportBillingScheduleResponse_ImportBillingScheduleError{
						RowNumber: int32(i) + 2,
						Error:     fmt.Sprintf("unable to create new billing schedule item: %s", err),
					})
				}
			} else {
				err = s.BillingScheduleRepo.Update(ctx, tx, billingSchedule)
				if err != nil {
					errors = append(errors, &pb.ImportBillingScheduleResponse_ImportBillingScheduleError{
						RowNumber: int32(i) + 2,
						Error:     fmt.Sprintf("unable to update billing schedule item: %s", err),
					})
				}
			}
		}
		if len(errors) > 0 {
			return fmt.Errorf(errors[0].Error)
		}
		return nil
	})
	if err != nil {
		log.Printf("Error when importing billing schedule: %s", err.Error())
	}
	return &pb.ImportBillingScheduleResponse{
		Errors: errors,
	}, nil
}

func BillingScheduleFromCsv(line []string, columnNames []string) (*entities.BillingSchedule, error) {
	const (
		BillingScheduleID = iota
		Name
		Remarks
		IsArchived
	)

	mandatory := []int{Name, IsArchived}

	areMandatoryDataPresent, colPosition := checkMandatoryColumnAndGetIndex(line, mandatory)
	if !areMandatoryDataPresent {
		return nil, fmt.Errorf("missing mandatory data: %v", columnNames[colPosition])
	}

	billingSchedule := &entities.BillingSchedule{}

	if err := multierr.Combine(
		utils.StringToFormatString("billing_schedule_id", line[BillingScheduleID], true, billingSchedule.BillingScheduleID.Set),
		utils.StringToFormatString("name", line[Name], false, billingSchedule.Name.Set),
		utils.StringToFormatString("remarks", line[Remarks], true, billingSchedule.Remarks.Set),
		utils.StringToBool("is_archived", line[IsArchived], false, billingSchedule.IsArchived.Set),
	); err != nil {
		return nil, err
	}

	return billingSchedule, nil
}

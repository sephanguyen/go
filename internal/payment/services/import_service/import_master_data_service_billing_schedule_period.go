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

func (s *ImportMasterDataService) ImportBillingSchedulePeriod(ctx context.Context, req *pb.ImportBillingSchedulePeriodRequest) (*pb.ImportBillingSchedulePeriodResponse, error) {
	errors := []*pb.ImportBillingSchedulePeriodResponse_ImportBillingSchedulePeriodError{}

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
		"billing_schedule_period_id",
		"name",
		"billing_schedule_id",
		"start_date",
		"end_date",
		"billing_date",
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
			billingSchedulePeriod, err := ReadBillingSchedulePeriodFromCsv(line, headerTitles)
			rowLine := int32(i) + 2 // i = 0 <=> line number 2 in csv file
			if err != nil {
				errors = append(errors, &pb.ImportBillingSchedulePeriodResponse_ImportBillingSchedulePeriodError{
					RowNumber: rowLine,
					Error:     fmt.Sprintf(constant.UnableToParseBillingSchedulePeriodItem, err),
				})
				continue
			}
			if billingSchedulePeriod.BillingSchedulePeriodID.Get() == nil {
				err := s.BillingSchedulePeriodRepo.Create(ctx, tx, billingSchedulePeriod)
				if err != nil {
					errors = append(errors, &pb.ImportBillingSchedulePeriodResponse_ImportBillingSchedulePeriodError{
						RowNumber: rowLine,
						Error:     fmt.Sprintf("unable to create new billing schedule period item: %s", err),
					})
				}
				continue
			}
			err = s.BillingSchedulePeriodRepo.Update(ctx, tx, billingSchedulePeriod)
			if err != nil {
				errors = append(errors, &pb.ImportBillingSchedulePeriodResponse_ImportBillingSchedulePeriodError{
					RowNumber: rowLine,
					Error:     fmt.Sprintf("unable to update billing schedule period item: %s", err),
				})
			}
		}
		if len(errors) > 0 {
			return fmt.Errorf(errors[0].Error)
		}
		return nil
	})
	if err != nil {
		log.Printf("Error when importing billing schedule period: %s", err.Error())
	}
	return &pb.ImportBillingSchedulePeriodResponse{
		Errors: errors,
	}, nil
}

func ReadBillingSchedulePeriodFromCsv(line []string, columnNames []string) (*entities.BillingSchedulePeriod, error) {
	const (
		BillingSchedulePeriodID = iota
		Name
		BillingScheduleID
		StartDate
		EndDate
		BillingDate
		Remarks
		IsArchived
	)

	mandatory := []int{Name, BillingScheduleID, StartDate, EndDate, BillingDate, IsArchived}

	areMandatoryDataPresent, colPosition := checkMandatoryColumnAndGetIndex(line, mandatory)
	if !areMandatoryDataPresent {
		return nil, fmt.Errorf("missing mandatory data: %v", columnNames[colPosition])
	}

	billingSchedulePeriod := &entities.BillingSchedulePeriod{}

	if err := multierr.Combine(
		utils.StringToFormatString("billing_schedule_period_id", line[BillingSchedulePeriodID], true, billingSchedulePeriod.BillingSchedulePeriodID.Set),
		utils.StringToFormatString("name", line[Name], false, billingSchedulePeriod.Name.Set),
		utils.StringToFormatString("billing_schedule_id", line[BillingScheduleID], false, billingSchedulePeriod.BillingScheduleID.Set),
		utils.StringToDate("start_date", line[StartDate], false, billingSchedulePeriod.StartDate.Set),
		utils.StringToDate("end_date", line[EndDate], false, billingSchedulePeriod.EndDate.Set),
		utils.StringToDate("billing_date", line[BillingDate], false, billingSchedulePeriod.BillingDate.Set),
		utils.StringToFormatString("remarks", line[Remarks], true, billingSchedulePeriod.Remarks.Set),
		utils.StringToBool("is_archived", line[IsArchived], false, billingSchedulePeriod.IsArchived.Set),
	); err != nil {
		return nil, err
	}

	return billingSchedulePeriod, nil
}

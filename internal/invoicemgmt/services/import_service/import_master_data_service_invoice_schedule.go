package services

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/jackc/pgx/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *ImportMasterDataService) ImportInvoiceSchedule(ctx context.Context, req *invoice_pb.ImportInvoiceScheduleRequest) (*invoice_pb.ImportInvoiceScheduleResponse, error) {
	lines, err := s.validateHeaderColumnRequest(req.Payload, invoice_pb.ImportMasterAction_INVOICE_SCHEDULE.String())

	if err != nil {
		return nil, err
	}

	errors := []*invoice_pb.ImportInvoiceScheduleResponse_ImportInvoiceScheduleError{}

	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) (err error) {
		useKECFeedbackPh1, err := s.UnleashClient.IsFeatureEnabled(constant.EnableKECFeedbackPh1, s.Env)
		if err != nil {
			return status.Error(codes.Internal, fmt.Sprintf("%v unleashClient.IsFeatureEnabled err: %v", constant.EnableKECFeedbackPh1, err))
		}

		// Skip the first item, which is the CSV header
		for i, line := range lines[1:] {
			// Validate each line after the header
			invoiceSchedule, err := s.validateImportScheduleCSV(ctx, line, useKECFeedbackPh1)
			if err != nil {
				errors = append(errors, s.generateImportInvoiceScheduleError(int32(i)+2, fmt.Sprintf("unable to parse invoice schedule: %s", err)))
				continue
			}

			// Cancel a scheduled invoice date if there's any
			if err := s.InvoiceScheduleRepo.CancelScheduleIfExists(ctx, tx, invoiceSchedule.InvoiceDate.Time); err != nil {
				errors = append(errors, s.generateImportInvoiceScheduleError(int32(i)+2, fmt.Sprintf("unable to cancel schedule: %s", err)))
				continue
			}

			// Create invoice record if there's no invoice schedule ID provided; otherwise, update the existing record
			if invoiceSchedule.InvoiceScheduleID.String == "" {
				if err := s.InvoiceScheduleRepo.Create(ctx, tx, invoiceSchedule); err != nil {
					errors = append(errors, s.generateImportInvoiceScheduleError(int32(i)+2, fmt.Sprintf("unable to create invoice schedule: %s", err)))
				}
			} else {
				if err := s.InvoiceScheduleRepo.Update(ctx, tx, invoiceSchedule); err != nil {
					errors = append(errors, s.generateImportInvoiceScheduleError(int32(i)+2, fmt.Sprintf("unable to update invoice schedule: %s", err)))
				}
			}
		}

		// This causes the changes to be rolled back if there's at least one error
		if len(errors) > 0 {
			return fmt.Errorf(errors[0].Error)
		}

		return nil
	})

	if err != nil {
		log.Printf("Error when importing invoice schedule: %s", err.Error())
	}

	return &invoice_pb.ImportInvoiceScheduleResponse{
		Errors: errors,
	}, nil
}

// Validates header column names and count and any format issues from reading the file a CSV file
func (s *ImportMasterDataService) validateHeaderColumnRequest(reqPayload []byte, importAction string) ([][]string, error) {
	var lines [][]string

	r := csv.NewReader(bytes.NewReader(reqPayload))
	lines, err := r.ReadAll()
	if err != nil {
		return lines, status.Error(codes.InvalidArgument, err.Error())
	}

	if len(lines) < 2 {
		return lines, status.Error(codes.InvalidArgument, "No data in CSV file")
	}

	header := lines[0]

	headerTitles := GenerateHeaderTitles(importAction)

	if err = validateCsvHeader(
		len(headerTitles),
		header,
		headerTitles,
	); err != nil {
		return lines, status.Error(codes.InvalidArgument, fmt.Sprintf("Invalid CSV format: %s", err.Error()))
	}

	return lines, nil
}

// Validates the CSV contents according to master guidelines with additional input validations
func (s *ImportMasterDataService) validateImportScheduleCSV(ctx context.Context, line []string, useKECFeedbackPh1 bool) (*entities.InvoiceSchedule, error) {
	// represents the zero-based order in CSV
	const (
		InvoiceScheduleID = iota
		InvoiceSchedule
		IsArchived
		Remarks
	)

	var isArchived bool
	var invoiceSchedule = &entities.InvoiceSchedule{}
	var err error

	database.AllNullEntity(invoiceSchedule)

	invoiceScheduleIdTrimmed := strings.TrimSpace(line[InvoiceScheduleID])
	isArchivedTrimmed := strings.TrimSpace(line[IsArchived])
	invoiceScheduleTrimmed := strings.TrimSpace(line[InvoiceSchedule])

	// (1) if there's invoice_schedule_id, there should be is_archived and vice versa
	if (len(invoiceScheduleIdTrimmed) != 0 && len(isArchivedTrimmed) == 0) ||
		(len(invoiceScheduleIdTrimmed) == 0 && len(isArchivedTrimmed) != 0) {
		return nil, fmt.Errorf("invoice_schedule_id and is_archived can only be both present or absent")
	}

	// (2) if the invoice_schedule_id and is_archived fields are not provided, invoice_schedule should exist to create a record
	if len(invoiceScheduleIdTrimmed) == 0 && len(isArchivedTrimmed) == 0 && len(invoiceScheduleTrimmed) == 0 {
		return nil, fmt.Errorf("invoice date is required")
	}

	// (3) if the row is to archive a schedule, validating the invoice_date value isn't needed
	if len(invoiceScheduleTrimmed) > 0 {
		// time.Parse's first argument specifies YYYY/MM/DD
		invoiceScheduleDate, err := time.Parse("2006/01/02", invoiceScheduleTrimmed)
		if err != nil {
			return nil, fmt.Errorf("invalid date format")
		}

		// Use JST by default
		location, err := utils.GetTimeLocationByCountry(utils.CountryJp)
		if err != nil {
			return nil, fmt.Errorf("error getTimeLocationByCountry: %v", err)
		}

		// get current date in JST
		nowJST := time.Now().In(location)

		// set to default 00:00 for quick comparison
		currentDayJST12AM := utils.ResetTimeComponent(nowJST)
		partnerInvoiceDate12AM := time.Date(invoiceScheduleDate.Year(), invoiceScheduleDate.Month(), invoiceScheduleDate.Day(), 0, 0, 0, 0, location)

		// validate the invoice_date.
		// if feature toggle is on, only validate past dates
		// if feature toggle is off, validate past dates and current date
		if useKECFeedbackPh1 {
			if partnerInvoiceDate12AM.Before(currentDayJST12AM) {
				return nil, fmt.Errorf("invoice schedule should be a present date or future date")
			}
		} else {
			if partnerInvoiceDate12AM == currentDayJST12AM || partnerInvoiceDate12AM.Before(currentDayJST12AM) {
				return nil, fmt.Errorf("invoice schedule should be a future date")
			}
		}

		invoiceSchedule.InvoiceDate = database.Timestamptz(partnerInvoiceDate12AM)

		// Set the scheduled date to invoice_date + 1 day
		invoiceSchedule.ScheduledDate = database.Timestamptz(invoiceSchedule.InvoiceDate.Time.Add(24 * time.Hour))
	}

	// (4) is_archived value should be valid
	if len(isArchivedTrimmed) > 0 {
		isArchived, err = strconv.ParseBool(isArchivedTrimmed)
		if err != nil {
			return nil, fmt.Errorf("invalid IsArchived value")
		}
	}

	// (5) if there's invoice_schedule_id, it should exist in DB
	if len(invoiceScheduleIdTrimmed) > 0 {
		_, err := s.InvoiceScheduleRepo.RetrieveInvoiceScheduleByID(ctx, s.DB, strings.TrimSpace(line[InvoiceScheduleID]))
		if err != nil {
			return nil, fmt.Errorf("cannot find invoice_schedule_id with error '%v'", err)
		}

		invoiceSchedule.InvoiceScheduleID = database.Text(invoiceScheduleIdTrimmed)
	}

	invoiceSchedule.Remarks = database.Text(line[Remarks])
	invoiceSchedule.IsArchived = database.Bool(isArchived)
	invoiceSchedule.Status = database.Text(invoice_pb.InvoiceScheduleStatus_INVOICE_SCHEDULE_SCHEDULED.String())

	if isArchived {
		invoiceSchedule.Status = database.Text(invoice_pb.InvoiceScheduleStatus_INVOICE_SCHEDULE_CANCELLED.String())
	}

	return invoiceSchedule, nil
}

func (s *ImportMasterDataService) generateImportInvoiceScheduleError(rowNumber int32, errorMsg string) *invoice_pb.ImportInvoiceScheduleResponse_ImportInvoiceScheduleError {
	return &invoice_pb.ImportInvoiceScheduleResponse_ImportInvoiceScheduleError{
		RowNumber: rowNumber,
		Error:     errorMsg,
	}
}

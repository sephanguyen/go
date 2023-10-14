package services

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/exporter"
	export_entities "github.com/manabie-com/backend/internal/invoicemgmt/export_entities"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *ExportMasterDataService) ExportInvoiceSchedule(ctx context.Context, req *invoice_pb.ExportInvoiceScheduleRequest) (*invoice_pb.ExportInvoiceScheduleResponse, error) {
	invoiceSchedules, err := s.InvoiceScheduleRepo.FindAll(ctx, s.DB)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("s.InvoiceScheduleRepo.FindAll err: %v", err))
	}

	entities := make([]database.Entity, len(invoiceSchedules))
	for i, e := range invoiceSchedules {
		// Convert invoice date UTC to VNT timezone
		location, err := utils.GetTimeLocationByCountry(utils.CountryJp)
		if err != nil {
			return nil, fmt.Errorf("error getTimeLocationByCountry: %v", err)
		}
		invoiceDateStr := e.InvoiceDate.Time.In(location).Format("2006/01/02")

		entities[i] = &export_entities.InvoiceScheduleExportData{
			InvoiceScheduleID: e.InvoiceScheduleID.String,
			InvoiceDate:       invoiceDateStr,
			IsArchived:        e.IsArchived.Bool,
			Remarks:           e.Remarks.String,
		}
	}

	colMap := []exporter.ExportColumnMap{
		{
			CSVColumn: "invoice_schedule_id",
			DBColumn:  "invoice_schedule_id",
		},
		{
			CSVColumn: "invoice_date",
			DBColumn:  "invoice_date",
		},
		{
			CSVColumn: "is_archived",
			DBColumn:  "is_archived",
		},
		{
			CSVColumn: "remarks",
			DBColumn:  "remarks",
		},
	}

	res, err := exporter.ExportBatch(entities, colMap)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("exporter.ExportBatch err: %v", err))
	}

	csvBytes := exporter.ToCSV(res)
	return &invoice_pb.ExportInvoiceScheduleResponse{
		Data: csvBytes,
	}, nil
}

package service

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/exporter"
	export_entities "github.com/manabie-com/backend/internal/payment/export_entities"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *ExportService) ExportStudentBilling(ctx context.Context, req *pb.ExportStudentBillingRequest) (exportDataResp *pb.ExportStudentBillingResponse, err error) {
	exportData, err := s.BillItemService.GetExportStudentBilling(ctx, s.DB, req.LocationIds)
	if err != nil {
		return
	}

	entities := make([]database.Entity, len(exportData))

	for i, e := range exportData {
		entities[i] = &export_entities.StudentBillingExport{
			StudentName:     e.StudentName,
			StudentID:       e.StudentID,
			Grade:           e.Grade,
			Location:        e.Location,
			CreatedDate:     e.CreatedDate,
			Status:          e.Status,
			BillingItemName: e.BillingItemName,
			Courses:         e.Courses,
			DiscountName:    e.DiscountName,
			DiscountAmount:  e.DiscountAmount,
			TaxAmount:       e.TaxAmount,
			BillingAmount:   e.BillingAmount,
		}
	}

	colMap := []exporter.ExportColumnMap{
		{
			CSVColumn: "student_name",
			DBColumn:  "student_name",
		},
		{
			CSVColumn: "student_id",
			DBColumn:  "student_id",
		},
		{
			CSVColumn: "grade",
			DBColumn:  "grade",
		},
		{
			CSVColumn: "location",
			DBColumn:  "location",
		},
		{
			CSVColumn: "created_date",
			DBColumn:  "created_date",
		},
		{
			CSVColumn: "status",
			DBColumn:  "status",
		},
		{
			CSVColumn: "billing_item_name",
			DBColumn:  "billing_item_name",
		},
		{
			CSVColumn: "courses",
			DBColumn:  "courses",
		},
		{
			CSVColumn: "discount_name",
			DBColumn:  "discount_name",
		},
		{
			CSVColumn: "discount_amount",
			DBColumn:  "discount_amount",
		},
		{
			CSVColumn: "tax_amount",
			DBColumn:  "tax_amount",
		},
		{
			CSVColumn: "billing_amount",
			DBColumn:  "billing_amount",
		},
	}

	res, err := exporter.ExportBatch(entities, colMap)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("exporter.ExportStudentBilling err: %v", err))
	}

	csvBytes := exporter.ToCSV(res)

	return &pb.ExportStudentBillingResponse{
		Data: csvBytes,
	}, nil
}

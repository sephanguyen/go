package services

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/exporter"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *ExportMasterDataService) ExportBankMapping(ctx context.Context, req *invoice_pb.ExportBankMappingRequest) (*invoice_pb.ExportBankMappingResponse, error) {
	bankMappings, err := s.BankMappingRepo.FindAll(ctx, s.DB)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("s.BankMappingRepo.FindAll err: %v", err))
	}

	entities := make([]database.Entity, len(bankMappings))
	for i, bankMapping := range bankMappings {
		entities[i] = bankMapping
	}

	colMap := []exporter.ExportColumnMap{
		{
			CSVColumn: "bank_mapping_id",
			DBColumn:  "bank_mapping_id",
		},
		{
			CSVColumn: "bank_id",
			DBColumn:  "bank_id",
		},
		{
			CSVColumn: "partner_bank_id",
			DBColumn:  "partner_bank_id",
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
	return &invoice_pb.ExportBankMappingResponse{
		Data: csvBytes,
	}, nil
}

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

func (s *ExportMasterDataService) ExportBank(ctx context.Context, req *invoice_pb.ExportBankRequest) (*invoice_pb.ExportBankResponse, error) {
	banks, err := s.BankRepo.FindAll(ctx, s.DB)

	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("s.BankRepo.FindAll err: %v", err))
	}

	entities := make([]database.Entity, len(banks))
	for i, bank := range banks {
		entities[i] = bank
	}

	colMap := []exporter.ExportColumnMap{
		{
			CSVColumn: "bank_id",
			DBColumn:  "bank_id",
		},
		{
			CSVColumn: "bank_code",
			DBColumn:  "bank_code",
		},
		{
			CSVColumn: "bank_name",
			DBColumn:  "bank_name",
		},
		{
			CSVColumn: "bank_phonetic_name",
			DBColumn:  "bank_name_phonetic",
		},
		{
			CSVColumn: "is_archived",
			DBColumn:  "is_archived",
		},
	}

	res, err := exporter.ExportBatch(entities, colMap)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("exporter.ExportBatch err: %v", err))
	}

	csvBytes := exporter.ToCSV(res)

	return &invoice_pb.ExportBankResponse{
		Data: csvBytes,
	}, nil
}

package services

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/exporter"
	export_entities "github.com/manabie-com/backend/internal/invoicemgmt/export_entities"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *ExportMasterDataService) ExportBankBranch(ctx context.Context, req *invoice_pb.ExportBankBranchRequest) (*invoice_pb.ExportBankBranchResponse, error) {

	bankBranches, err := s.BankBranchRepo.FindExportableBankBranches(ctx, s.DB)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("s.BankBranchRepo.FindExportableBankBranches err: %v", err))
	}

	entities := make([]database.Entity, len(bankBranches))

	for i, e := range bankBranches {
		entities[i] = &export_entities.BankBranchExport{
			BankBranchID:           e.BankBranchID,
			BankBranchCode:         e.BankBranchCode,
			BankBranchName:         e.BankBranchName,
			BankBranchPhoneticName: e.BankBranchPhoneticName,
			BankCode:               e.BankCode,
			IsArchived:             e.IsArchived,
		}
	}

	colMap := []exporter.ExportColumnMap{
		{
			CSVColumn: "bank_branch_id",
			DBColumn:  "bank_branch_id",
		},
		{
			CSVColumn: "bank_branch_code",
			DBColumn:  "bank_branch_code",
		},
		{
			CSVColumn: "bank_branch_name",
			DBColumn:  "bank_branch_name",
		},
		{
			CSVColumn: "bank_branch_phonetic_name",
			DBColumn:  "bank_branch_phonetic_name",
		},
		{
			CSVColumn: "bank_code",
			DBColumn:  "bank_code",
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
	return &invoice_pb.ExportBankBranchResponse{
		Data: csvBytes,
	}, nil
}

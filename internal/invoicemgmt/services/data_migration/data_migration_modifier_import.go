package services

import (
	"context"
	"fmt"
	"log"

	"github.com/manabie-com/backend/internal/golibs/database"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type insertEntityFunc func(*DataMigrationModifierService, context.Context, database.QueryExecer, [][]string) ([]*invoice_pb.ImportDataMigrationResponse_ImportMigrationDataError, error)

var (
	insertPaymentDataMigration insertEntityFunc = (*DataMigrationModifierService).InsertPaymentDataMigration
	insertInvoiceDataMigration insertEntityFunc = (*DataMigrationModifierService).InsertInvoiceDataMigration
)

func (s *DataMigrationModifierService) ImportDataMigration(ctx context.Context, req *invoice_pb.ImportDataMigrationRequest) (*invoice_pb.ImportDataMigrationResponse, error) {
	lines, err := validateHeaderColumnRequest(req)
	if err != nil {
		return nil, err
	}

	entityName := req.EntityName
	var errLines []*invoice_pb.ImportDataMigrationResponse_ImportMigrationDataError
	// remove header row
	csvLines := lines[1:]

	switch entityName {
	case invoice_pb.DataMigrationEntityName_PAYMENT_ENTITY:
		errLines, err = insertPaymentDataMigration(s, ctx, s.DB, csvLines)
	case invoice_pb.DataMigrationEntityName_INVOICE_ENTITY:
		errLines, err = insertInvoiceDataMigration(s, ctx, s.DB, csvLines)
	default:
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid entity name: %v", entityName.String()))
	}

	if err != nil {
		return nil, err
	}

	if len(errLines) > 0 {
		for _, errLine := range errLines {
			log.Printf("Data Migration %v error: %v on RowNumber: %v", entityName, errLine.Error, errLine.RowNumber)
		}
	}

	return &invoice_pb.ImportDataMigrationResponse{
		EntityName: entityName,
		Errors:     errLines,
	}, nil
}

package invoicesvc

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/invoicemgmt/repositories"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *InvoiceModifierService) RetrieveInvoiceStatusCount(ctx context.Context, req *invoice_pb.RetrieveInvoiceStatusCountRequest) (*invoice_pb.RetrieveInvoiceStatusCountResponse, error) {
	sqlFilter, err := repositories.GenerateInvoiceDataWhereClause(req.InvoiceFilter, req.PaymentFilter, req.StudentName)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	invoiceStatuses, err := s.InvoiceRepo.RetrieveInvoiceStatusCount(ctx, s.DB, sqlFilter)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var totalItems int32
	invoiceStatusMap := map[string]int32{
		invoice_pb.InvoiceStatus_DRAFT.String():    0,
		invoice_pb.InvoiceStatus_ISSUED.String():   0,
		invoice_pb.InvoiceStatus_PAID.String():     0,
		invoice_pb.InvoiceStatus_REFUNDED.String(): 0,
		invoice_pb.InvoiceStatus_VOID.String():     0,
	}

	// check if array length is greater than 0
	if len(invoiceStatuses) > 0 && invoiceStatuses != nil {
		for invoiceStatus, statusCount := range invoiceStatuses {
			_, ok := invoiceStatusMap[invoiceStatus]
			if !ok {
				return nil, status.Error(codes.Internal, fmt.Sprintf("invalid invoice status: %s", invoiceStatus))
			}
			invoiceStatusMap[invoiceStatus] = statusCount
			totalItems += statusCount
		}
	}

	return &invoice_pb.RetrieveInvoiceStatusCountResponse{
		TotalItems: totalItems,
		InvoiceStatusCount: &invoice_pb.RetrieveInvoiceStatusCountResponse_InvoiceStatusCount{
			TotalDraft:    invoiceStatusMap[invoice_pb.InvoiceStatus_DRAFT.String()],
			TotalIssued:   invoiceStatusMap[invoice_pb.InvoiceStatus_ISSUED.String()],
			TotalPaid:     invoiceStatusMap[invoice_pb.InvoiceStatus_PAID.String()],
			TotalRefunded: invoiceStatusMap[invoice_pb.InvoiceStatus_REFUNDED.String()],
			TotalVoid:     invoiceStatusMap[invoice_pb.InvoiceStatus_VOID.String()],
		},
	}, nil
}

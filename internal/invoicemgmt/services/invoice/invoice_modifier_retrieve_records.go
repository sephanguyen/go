package invoicesvc

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *InvoiceModifierService) RetrieveInvoiceRecords(ctx context.Context, req *invoice_pb.RetrieveInvoiceRecordsRequest) (*invoice_pb.RetrieveInvoiceRecordsResponse, error) {
	limit := database.Int8(constant.PageLimit)
	offset := database.Int8(0)

	if req.Paging != nil && req.Paging.Limit != 0 {
		_ = limit.Set(req.Paging.Limit)
		_ = offset.Set(req.Paging.GetOffsetInteger())
	}
	invoiceRecords, err := s.InvoiceRepo.RetrieveRecordsByStudentID(ctx, s.DB, req.StudentId, limit, offset)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if len(invoiceRecords) == 0 {
		return &invoice_pb.RetrieveInvoiceRecordsResponse{}, nil
	}
	nextPage := &cpb.Paging{
		Limit: uint32(limit.Int),
		Offset: &cpb.Paging_OffsetInteger{
			OffsetInteger: limit.Int + offset.Int,
		},
	}
	responseItems := make([]*invoice_pb.InvoiceRecord, 0, len(invoiceRecords))
	for _, inv := range invoiceRecords {

		getExactValueWithDecimalPlaces, err := utils.GetFloat64ExactValueAndDecimalPlaces(inv.Total, "2")
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		invoiceItem := &invoice_pb.InvoiceRecord{
			DueDate:         timestamppb.New(pgtype.Timestamptz{Status: pgtype.Null}.Time),
			InvoiceIdString: inv.InvoiceID.String,
			InvoiceStatus:   invoice_pb.InvoiceStatus(invoice_pb.InvoiceStatus_value[inv.Status.String]),
			Total:           getExactValueWithDecimalPlaces,
		}

		// get latest payment due date record
		payment, err := s.PaymentRepo.GetLatestPaymentDueDateByInvoiceID(ctx, s.DB, inv.InvoiceID.String)
		if err != nil && err != pgx.ErrNoRows {
			// if no payment record is not the error, throw internal error
			return nil, status.Error(codes.Internal, err.Error())
		}

		if payment != nil {
			invoiceItem.DueDate = timestamppb.New(payment.PaymentDueDate.Time)
		}

		responseItems = append(responseItems, invoiceItem)
	}
	return &invoice_pb.RetrieveInvoiceRecordsResponse{
		InvoiceRecords: responseItems,
		NextPage:       nextPage,
	}, nil
}

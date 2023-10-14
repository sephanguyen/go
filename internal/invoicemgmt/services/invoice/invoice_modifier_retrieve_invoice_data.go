package invoicesvc

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/invoicemgmt/repositories"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/jackc/pgtype"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *InvoiceModifierService) RetrieveInvoiceData(ctx context.Context, req *invoice_pb.RetrieveInvoiceDataRequest) (*invoice_pb.RetrieveInvoiceDataResponse, error) {
	limit, offset := s.getLimitOffset(req)

	sqlFilter, err := repositories.GenerateInvoiceDataWhereClause(req.InvoiceFilter, req.PaymentFilter, req.StudentName)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	invoiceRecords, err := s.InvoiceRepo.RetrieveInvoiceData(ctx, s.DB, limit, offset, sqlFilter)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	invoiceData := make([]*invoice_pb.InvoiceData, 0)

	for _, invoiceRecord := range invoiceRecords {
		exactSubTotal, err := utils.GetFloat64ExactValueAndDecimalPlaces(invoiceRecord.Invoice.SubTotal, "2")
		if err != nil {
			return nil, fmt.Errorf("cannot assign invoice sub total: %v to float data type", invoiceRecord.Invoice.SubTotal)
		}

		exactTotal, err := utils.GetFloat64ExactValueAndDecimalPlaces(invoiceRecord.Invoice.Total, "2")
		if err != nil {
			return nil, fmt.Errorf("cannot assign invoice total: %v to float data type", invoiceRecord.Invoice.Total)
		}

		exactOutstandingBalance, err := utils.GetFloat64ExactValueAndDecimalPlaces(invoiceRecord.Invoice.OutstandingBalance, "2")
		if err != nil {
			return nil, fmt.Errorf("cannot assign invoice outstanding balance: %v to float data type", invoiceRecord.Invoice.OutstandingBalance)
		}

		exactAmountPaid, err := utils.GetFloat64ExactValueAndDecimalPlaces(invoiceRecord.Invoice.AmountPaid, "2")
		if err != nil {
			return nil, fmt.Errorf("cannot assign invoice amount paid: %v to float data type", invoiceRecord.Invoice.AmountPaid)
		}

		invoiceDataMap := &invoice_pb.InvoiceData{
			InvoiceDataDetail: &invoice_pb.InvoiceData_InvoiceDataDetail{
				InvoiceId:             invoiceRecord.Invoice.InvoiceID.String,
				InvoiceSequenceNumber: invoiceRecord.Invoice.InvoiceSequenceNumber.Int,
				InvoiceStatus:         invoice_pb.InvoiceStatus(invoice_pb.InvoiceStatus_value[invoiceRecord.Invoice.Status.String]),
				StudentId:             invoiceRecord.Invoice.StudentID.String,
				SubTotal:              exactSubTotal,
				Total:                 exactTotal,
				OutstandingBalance:    exactOutstandingBalance,
				AmountPaid:            exactAmountPaid,
				InvoiceType:           invoice_pb.InvoiceType(invoice_pb.InvoiceType_value[invoiceRecord.Invoice.Type.String]),
				CreatedAt:             timestamppb.New(invoiceRecord.Invoice.CreatedAt.Time),
			},
		}

		if invoiceRecord.Payment != nil {
			invoicePaymentDetail := &invoice_pb.InvoiceData_InvoiceDataPaymentDetail{
				PaymentId:             invoiceRecord.Payment.PaymentID.String,
				PaymentSequenceNumber: invoiceRecord.Payment.PaymentSequenceNumber.Int,
				IsExported:            invoiceRecord.Payment.IsExported.Bool,
				PaymentDueDate:        timestamppb.New(invoiceRecord.Payment.PaymentDueDate.Time),
				PaymentExpiryDate:     timestamppb.New(invoiceRecord.Payment.PaymentExpiryDate.Time),
				PaymentMethod:         invoice_pb.PaymentMethod(invoice_pb.PaymentMethod_value[invoiceRecord.Payment.PaymentMethod.String]),
				PaymentStatus:         invoice_pb.PaymentStatus(invoice_pb.PaymentStatus_value[invoiceRecord.Payment.PaymentStatus.String]),
			}

			if invoiceRecord.Payment.PaymentDate.Status == pgtype.Present {
				invoicePaymentDetail.PaymentDate = timestamppb.New(invoiceRecord.Payment.PaymentDate.Time)
			}

			if invoiceRecord.Payment.Amount.Status == pgtype.Present {
				exactPaymentAmountPaid, err := utils.GetFloat64ExactValueAndDecimalPlaces(invoiceRecord.Payment.Amount, "2")
				if err != nil {
					return nil, fmt.Errorf("cannot assign payment amount: %v to float data type", invoiceRecord.Payment.Amount)
				}
				invoicePaymentDetail.Amount = exactPaymentAmountPaid
			}

			invoiceDataMap.InvoiceDataPaymentDetail = invoicePaymentDetail
		}

		invoiceDataMap.StudentName = invoiceRecord.UserBasicInfo.Name.String

		invoiceData = append(invoiceData, invoiceDataMap)
	}
	return &invoice_pb.RetrieveInvoiceDataResponse{
		InvoiceData:  invoiceData,
		NextPage:     utils.GetNextPaging(limit, offset),
		PreviousPage: utils.GetPrevPaging(limit, offset),
	}, nil
}

func (s *InvoiceModifierService) getLimitOffset(req *invoice_pb.RetrieveInvoiceDataRequest) (limit, offset pgtype.Int8) {
	limit = database.Int8(constant.PageLimit)
	offset = database.Int8(0)

	if req.Paging != nil && req.Paging.Limit != 0 {
		_ = limit.Set(req.Paging.Limit)
		_ = offset.Set(req.Paging.GetOffsetInteger())
	}

	return limit, offset
}

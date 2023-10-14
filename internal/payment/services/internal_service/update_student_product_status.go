package service

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/entities"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgx/v4"
)

func (s *InternalService) UpdateStudentProductStatus(ctx context.Context, req *pb.UpdateStudentProductStatusRequest) (res *pb.UpdateStudentProductStatusResponse, err error) {
	var (
		studentProducts   []*entities.StudentProduct
		studentProductIDs []string
		errors            []*pb.UpdateStudentProductStatusResponse_UpdateStudentProductStatusError
	)
	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) (err error) {
		studentProducts, err = s.studentProductService.GetStudentProductsByStudentProductLabel(ctx, tx, req.StudentProductLabel)
		if err != nil {
			errors = append(errors, &pb.UpdateStudentProductStatusResponse_UpdateStudentProductStatusError{Error: err.Error()})
			return
		}
		errors = make([]*pb.UpdateStudentProductStatusResponse_UpdateStudentProductStatusError, 0, len(studentProducts))
		for _, studentProduct := range studentProducts {
			if studentProduct.EndDate.Time.After(req.EffectiveDate.AsTime()) {
				continue
			}

			switch studentProduct.StudentProductLabel.String {
			case pb.StudentProductLabel_WITHDRAWAL_SCHEDULED.String():
				err = s.studentProductService.CancelStudentProduct(ctx, tx, studentProduct.StudentProductID.String)
			case pb.StudentProductLabel_GRADUATION_SCHEDULED.String():
				err = s.studentProductService.CancelStudentProduct(ctx, tx, studentProduct.StudentProductID.String)
			case pb.StudentProductLabel_PAUSE_SCHEDULED.String():
				err = s.studentProductService.PauseStudentProduct(ctx, tx, *studentProduct)
			}

			if err != nil {
				errors = append(errors, &pb.UpdateStudentProductStatusResponse_UpdateStudentProductStatusError{
					StudentProductId: studentProduct.StudentProductID.String,
					Error:            err.Error(),
				})
			}
			studentProductIDs = append(studentProductIDs, studentProduct.StudentProductID.String)
		}
		return nil
	})
	res = &pb.UpdateStudentProductStatusResponse{
		StudentProductIds: studentProductIDs,
		Errors:            errors,
	}
	return
}

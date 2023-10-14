package service

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/entities"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgx/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *InternalService) UpdateStudentPackage(ctx context.Context, _ *pb.UpdateStudentPackageForCronjobRequest) (res *pb.UpdateStudentPackageForCronjobResponse, err error) {
	var (
		errors          []*pb.UpdateStudentPackageForCronjobResponse_UpdateStudentPackageForCronjobError
		failed, succeed int32
	)

	studentPackages, err := s.studentPackageService.GetStudentPackagesForCronJob(ctx, s.DB)
	if err != nil {
		err = status.Errorf(codes.Internal, fmt.Sprintf("Error when getting student packages for cronjob: %v", err))
		return
	}
	errors = make([]*pb.UpdateStudentPackageForCronjobResponse_UpdateStudentPackageForCronjobError, 0, len(studentPackages))
	for _, studentPackage := range studentPackages {
		var (
			studentPackageOrderID      string
			currentStudentPackageOrder *entities.StudentPackageOrder
			studentPackageEvent        *npb.EventStudentPackage
		)
		err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
			studentPackageEvent, currentStudentPackageOrder, err = s.studentPackageService.UpsertStudentPackageDataForCronjob(ctx, tx, studentPackage)
			if err != nil {
				return status.Errorf(codes.Internal, fmt.Sprintf("Error when upserting student package data by student package order: %v", err))
			}
			if studentPackageEvent != nil {
				err = s.subscriptionService.PublishStudentPackageForCreateOrder(ctx, []*npb.EventStudentPackage{studentPackageEvent})
				if err != nil {
					return status.Errorf(codes.Internal, fmt.Sprintf("Error when publish student package events with err=%s", err))
				}
			}

			return nil
		})
		if err != nil {
			if currentStudentPackageOrder != nil {
				studentPackageOrderID = currentStudentPackageOrder.ID.String
				_ = currentStudentPackageOrder.ExecutedError.Set(err.Error())
				updateStudentPackageErr := s.studentPackageOrderService.UpdateExecuteError(ctx, s.DB, *currentStudentPackageOrder)
				if updateStudentPackageErr != nil {
					errors = append(errors, &pb.UpdateStudentPackageForCronjobResponse_UpdateStudentPackageForCronjobError{
						Error:                 fmt.Sprintf("Fail to update execute error of student package with student_package_id = %s, student_package_order_id = %s and err = %s", studentPackage.ID.String, studentPackageOrderID, updateStudentPackageErr.Error()),
						StudentPackageId:      studentPackage.ID.String,
						StudentPackageOrderId: studentPackageOrderID,
					})
				}
			}
			errors = append(errors, &pb.UpdateStudentPackageForCronjobResponse_UpdateStudentPackageForCronjobError{
				Error:                 fmt.Sprintf("Fail to update student package with student_package_id = %s, student_package_order_id = %s and err = %s", studentPackage.ID.String, studentPackageOrderID, err.Error()),
				StudentPackageId:      studentPackage.ID.String,
				StudentPackageOrderId: studentPackageOrderID,
			})

			failed++
			continue
		}
		succeed++
	}
	return &pb.UpdateStudentPackageForCronjobResponse{
		Successful: true,
		Successed:  succeed,
		Failed:     failed,
		Errors:     errors,
	}, nil
}

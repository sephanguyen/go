package service

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/utils"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *StudentPackageService) VoidStudentPackageAndStudentCourse(
	ctx context.Context,
	db database.QueryExecer,
	args utils.VoidStudentPackageArgs,
) (
	studentPackageEvents []*npb.EventStudentPackage,
	err error,
) {
	mapStudentCourseWithStudentPackageAccessPath, err := s.StudentPackageAccessPathRepo.GetMapStudentCourseKeyWithStudentPackageAccessPathByStudentIDs(
		ctx,
		db,
		[]string{args.StudentProduct.StudentID.String},
	)
	if err != nil {
		return
	}
	switch args.Order.OrderType.String {
	case pb.OrderType_ORDER_TYPE_NEW.String(),
		pb.OrderType_ORDER_TYPE_ENROLLMENT.String(),
		pb.OrderType_ORDER_TYPE_RESUME.String():
		return s.voidStudentPackageForCreateOrder(ctx, db, args, mapStudentCourseWithStudentPackageAccessPath)
	case pb.OrderType_ORDER_TYPE_GRADUATE.String(),
		pb.OrderType_ORDER_TYPE_WITHDRAWAL.String(),
		pb.OrderType_ORDER_TYPE_LOA.String():
		return s.voidStudentPackageForCancelOrder(ctx, db, args, mapStudentCourseWithStudentPackageAccessPath)
	case pb.OrderType_ORDER_TYPE_UPDATE.String():
		if args.IsCancel {
			return s.voidStudentPackageForCancelOrder(ctx, db, args, mapStudentCourseWithStudentPackageAccessPath)
		}
		return s.voidStudentPackageForUpdateOrder(ctx, db, args, mapStudentCourseWithStudentPackageAccessPath)
	default:
		err = status.Errorf(codes.FailedPrecondition, "wrong order type")
	}
	return
}

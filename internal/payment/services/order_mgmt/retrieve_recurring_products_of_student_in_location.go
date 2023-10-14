package ordermgmt

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/entities"
	orderService "github.com/manabie-com/backend/internal/payment/services/domain_service/order"
	studentProductService "github.com/manabie-com/backend/internal/payment/services/domain_service/student_product"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type RetrieveRecurringProductsOfStudentInLocation struct {
	DB                    database.Ext
	StudentProductService IStudentProductServiceForRetrieveRecurringProduct
	OrderService          IOrderServiceForRetrieveRecurringProduct
}

type IStudentProductServiceForRetrieveRecurringProduct interface {
	GetActiveRecurringProductsOfStudentInLocation(ctx context.Context, db database.QueryExecer, studentID string, locationID string) (studentProducts []entities.StudentProduct, err error)
	GetRecurringProductsOfStudentInLocationForLOA(ctx context.Context, db database.QueryExecer, studentID string, locationID string) (studentProducts []entities.StudentProduct, err error)
	GetStudentProductsByStudentProductIDs(ctx context.Context, db database.Ext, studentProductIDs []string) (studentProducts []entities.StudentProduct, err error)
}

type IOrderServiceForRetrieveRecurringProduct interface {
	GetStudentProductIDsForResume(ctx context.Context, db database.QueryExecer, studentID string, locationID string) (studentProductIDs []string, err error)
}

func (s *RetrieveRecurringProductsOfStudentInLocation) RetrieveRecurringProductsOfStudentInLocation(ctx context.Context, req *pb.RetrieveRecurringProductsOfStudentInLocationRequest) (res *pb.RetrieveRecurringProductsOfStudentInLocationResponse, err error) {
	var studentProducts []entities.StudentProduct

	switch req.OrderType.String() {
	case pb.OrderType_ORDER_TYPE_WITHDRAWAL.String(), pb.OrderType_ORDER_TYPE_GRADUATE.String():
		studentProducts, err = s.StudentProductService.GetActiveRecurringProductsOfStudentInLocation(ctx, s.DB, req.StudentId, req.LocationId)
		if err != nil {
			err = fmt.Errorf("error when getting student products of student %s in location %s: %v", req.StudentId, req.LocationId, err)
			return
		}
	case pb.OrderType_ORDER_TYPE_LOA.String():
		studentProducts, err = s.StudentProductService.GetRecurringProductsOfStudentInLocationForLOA(ctx, s.DB, req.StudentId, req.LocationId)
		if err != nil {
			err = fmt.Errorf("error when getting student products of student %s in location %s for LOA: %v", req.StudentId, req.LocationId, err)
			return
		}
	case pb.OrderType_ORDER_TYPE_RESUME.String():
		var studentProductIDs []string
		studentProductIDs, err = s.OrderService.GetStudentProductIDsForResume(ctx, s.DB, req.StudentId, req.LocationId)
		if err != nil {
			err = fmt.Errorf("error when getting student product ID of student %s in location %s for resume: %v", req.StudentId, req.LocationId, err)
			return nil, err
		}

		if len(studentProductIDs) != 0 {
			studentProducts, err = s.StudentProductService.GetStudentProductsByStudentProductIDs(ctx, s.DB, studentProductIDs)
			if err != nil {
				err = fmt.Errorf("error when getting student products of student %s in location %s for resume: %v", req.StudentId, req.LocationId, err)
				return
			}
		}
	default:
		err = fmt.Errorf("invalid orderType: %s", req.OrderType.String())
		return
	}

	activeProductsInLocation := []*pb.RetrieveRecurringProductsOfStudentInLocationResponse_StudentProduct{}

	for _, studentProduct := range studentProducts {
		product := &pb.RetrieveRecurringProductsOfStudentInLocationResponse_StudentProduct{
			StudentProductId:    studentProduct.StudentProductID.String,
			StudentId:           studentProduct.StudentID.String,
			LocationId:          studentProduct.LocationID.String,
			ProductStatus:       studentProduct.ProductStatus.String,
			StudentProductLabel: studentProduct.StudentProductLabel.String,
			StartDate:           &timestamppb.Timestamp{Seconds: studentProduct.StartDate.Time.Unix()},
			EndDate:             &timestamppb.Timestamp{Seconds: studentProduct.EndDate.Time.Unix()},
		}

		activeProductsInLocation = append(activeProductsInLocation, product)
	}

	res = &pb.RetrieveRecurringProductsOfStudentInLocationResponse{
		StudentProductInLocation: activeProductsInLocation,
	}
	return
}

func NewRetrieveRecurringProductsOfStudentInLocation(db database.Ext) *RetrieveRecurringProductsOfStudentInLocation {
	return &RetrieveRecurringProductsOfStudentInLocation{
		DB:                    db,
		StudentProductService: studentProductService.NewStudentProductService(),
		OrderService:          orderService.NewOrderService(),
	}
}

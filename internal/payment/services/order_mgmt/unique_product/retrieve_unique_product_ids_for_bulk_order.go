package unique_product

import (
	"context"

	"github.com/manabie-com/backend/internal/payment/entities"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *UniqueProduct) RetrieveListOfUniqueProductIDForBulkOrder(ctx context.Context, req *pb.RetrieveListOfUniqueProductIDForBulkOrderRequest) (res *pb.RetrieveListOfUniqueProductIDForBulkOrderResponse, err error) {

	mapStudentProductOfUniqueProducts, err := s.StudentProductService.GetUniqueProductsByStudentIDs(ctx, s.DB, req.StudentIds)
	if err != nil {
		return
	}
	uniqueProductOfStudents := []*pb.RetrieveListOfUniqueProductIDForBulkOrderResponse_UniqueProductOfStudent{}
	for _, student := range req.StudentIds {
		studentProductOfUniqueProducts := mapStudentProductOfUniqueProducts[student]
		var productDetails []*pb.RetrieveListOfUniqueProductIDForBulkOrderResponse_ProductInfo
		for _, studentProductOfUniqueProduct := range studentProductOfUniqueProducts {
			productDetail, err := s.formatForProductDetailsForBulkOrder(ctx, studentProductOfUniqueProduct)
			if err != nil {
				return nil, err
			}
			if productDetail == nil {
				continue
			}
			productDetails = append(productDetails, productDetail)
		}
		uniqueProductOfStudents = append(uniqueProductOfStudents, &pb.RetrieveListOfUniqueProductIDForBulkOrderResponse_UniqueProductOfStudent{
			StudentId:      student,
			ProductDetails: productDetails,
		})
	}

	res = &pb.RetrieveListOfUniqueProductIDForBulkOrderResponse{
		UniqueProductOfStudent: uniqueProductOfStudents,
	}
	return
}

func (s *UniqueProduct) formatForProductDetailsForBulkOrder(ctx context.Context, studentProduct *entities.StudentProduct) (productDetail *pb.RetrieveListOfUniqueProductIDForBulkOrderResponse_ProductInfo, err error) {
	if studentProduct.EndDate.Status == pgtype.Present {
		if studentProduct.ProductStatus.String != pb.StudentProductStatus_CANCELLED.String() {
			productDetail = &pb.RetrieveListOfUniqueProductIDForBulkOrderResponse_ProductInfo{
				ProductId: studentProduct.ProductID.String,
			}
		} else {
			_, err := s.PackageService.GetByIDForUniqueProduct(ctx, s.DB, studentProduct.ProductID.String)
			if err != nil {
				if err != pgx.ErrNoRows {
					return nil, err
				}
			} else {
				return nil, nil
			}

			endTime, err := s.StudentProductService.EndDateOfUniqueRecurringProduct(ctx, s.DB, studentProduct.ProductID.String, studentProduct.EndDate.Time)
			if err != nil {
				return nil, err
			}
			productDetail = &pb.RetrieveListOfUniqueProductIDForBulkOrderResponse_ProductInfo{
				ProductId: studentProduct.ProductID.String,
				EndTime:   timestamppb.New(endTime),
			}
		}
	} else if studentProduct.ProductStatus.String != pb.StudentProductStatus_CANCELLED.String() {
		productDetail = &pb.RetrieveListOfUniqueProductIDForBulkOrderResponse_ProductInfo{
			ProductId: studentProduct.ProductID.String,
		}
	}

	return
}

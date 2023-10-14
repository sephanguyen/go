package unique_product

import (
	"context"

	"github.com/manabie-com/backend/internal/payment/entities"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *UniqueProduct) RetrieveListOfUniqueProductIDs(ctx context.Context, req *pb.RetrieveListOfUniqueProductIDsRequest) (res *pb.RetrieveListOfUniqueProductIDsResponse, err error) {
	var (
		studentProductOfUniqueProducts []*entities.StudentProduct
		productDetails                 []*pb.RetrieveListOfUniqueProductIDsResponse_ProductInfo
	)
	studentProductOfUniqueProducts, err = s.StudentProductService.GetUniqueProductsByStudentID(ctx, s.DB, req.StudentId)
	if err != nil {
		return
	}
	for _, studentProductOfUniqueProduct := range studentProductOfUniqueProducts {
		productDetail, err := s.formatForProductDetails(ctx, studentProductOfUniqueProduct)
		if err != nil {
			return nil, err
		}
		if productDetail == nil {
			continue
		}
		productDetails = append(productDetails, productDetail)
	}

	res = &pb.RetrieveListOfUniqueProductIDsResponse{
		ProductDetails: productDetails,
	}
	return
}

func (s *UniqueProduct) formatForProductDetails(ctx context.Context, studentProduct *entities.StudentProduct) (productDetail *pb.RetrieveListOfUniqueProductIDsResponse_ProductInfo, err error) {
	if studentProduct.EndDate.Status == pgtype.Present {
		if studentProduct.ProductStatus.String != pb.StudentProductStatus_CANCELLED.String() {
			productDetail = &pb.RetrieveListOfUniqueProductIDsResponse_ProductInfo{
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
			productDetail = &pb.RetrieveListOfUniqueProductIDsResponse_ProductInfo{
				ProductId: studentProduct.ProductID.String,
				EndTime:   timestamppb.New(endTime),
			}
		}
	} else if studentProduct.ProductStatus.String != pb.StudentProductStatus_CANCELLED.String() {
		productDetail = &pb.RetrieveListOfUniqueProductIDsResponse_ProductInfo{
			ProductId: studentProduct.ProductID.String,
		}
	}

	return
}

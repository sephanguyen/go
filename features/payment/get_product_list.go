package payment

import (
	"context"
	"fmt"
	"time"

	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"
)

func (s *suite) prepareDataForGetProductList(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	defaultPrepareDataSettings := PrepareDataForCreatingOrderSettings{
		insertPackage:      true,
		insertMaterial:     true,
		insertFee:          true,
		insertProductGrade: true,
		insertTax:          true,
		insertLocation:     false,
	}
	_, _, _, _, _, err := s.insertAllDataForInsertOrder(ctx, defaultPrepareDataSettings, "get product list test")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) getProductList(ctx context.Context, userGroup, productListFilter string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, userGroup)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	switch productListFilter {
	case "without filter":
		req := &pb.RetrieveListOfProductsRequest{
			Filter: &pb.RetrieveListOfProductsFilter{
				ProductTypes:  nil,
				StudentGrades: nil,
			},
			Keyword:       "",
			ProductStatus: 0,
			Paging: &cpb.Paging{
				Limit:  10,
				Offset: &cpb.Paging_OffsetInteger{OffsetInteger: 5},
			},
		}

		client := pb.NewOrderServiceClient(s.PaymentConn)
		stepState.RequestSentAt = time.Now()
		resp, err := client.RetrieveListOfProducts(contextWithToken(ctx), req)
		if err != nil {
			stepState.ResponseErr = err
			return StepStateToContext(ctx, stepState), err
		}
		stepState.Response = resp

		if stepState.ResponseErr != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("an unexpected error returned")
		}
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	case "with empty filter":
		req := &pb.RetrieveListOfProductsRequest{
			Filter: &pb.RetrieveListOfProductsFilter{
				ProductTypes:  []*pb.ProductSpecificType{},
				StudentGrades: []string{},
			},
			Keyword:       "",
			ProductStatus: 0,
			Paging: &cpb.Paging{
				Limit:  10,
				Offset: &cpb.Paging_OffsetInteger{OffsetInteger: 0},
			},
		}

		client := pb.NewOrderServiceClient(s.PaymentConn)
		stepState.RequestSentAt = time.Now()
		resp, err := client.RetrieveListOfProducts(contextWithToken(ctx), req)
		if err != nil {
			stepState.ResponseErr = err
			return StepStateToContext(ctx, stepState), err
		}
		stepState.Response = resp

		if stepState.ResponseErr != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("an unexpected error returned")
		}
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	case "with product type filter":
		req := &pb.RetrieveListOfProductsRequest{
			Filter: &pb.RetrieveListOfProductsFilter{
				ProductTypes: []*pb.ProductSpecificType{
					{
						ProductType:  pb.ProductType_PRODUCT_TYPE_MATERIAL,
						PackageType:  0,
						MaterialType: pb.MaterialType_MATERIAL_TYPE_ONE_TIME,
						FeeType:      0,
					},
				},
				StudentGrades: []string{},
			},
			Keyword:       "",
			ProductStatus: 0,
			Paging: &cpb.Paging{
				Limit:  10,
				Offset: &cpb.Paging_OffsetInteger{OffsetInteger: 0},
			},
		}

		client := pb.NewOrderServiceClient(s.PaymentConn)
		stepState.RequestSentAt = time.Now()
		resp, err := client.RetrieveListOfProducts(contextWithToken(ctx), req)
		if err != nil {
			stepState.ResponseErr = err
			return StepStateToContext(ctx, stepState), err
		}
		stepState.Response = resp

		if stepState.ResponseErr != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("an unexpected error returned")
		}
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) getProductListSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if len(stepState.Response.(*pb.RetrieveListOfProductsResponse).Items) == 0 {
		err := fmt.Errorf("failed to get product list")
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}

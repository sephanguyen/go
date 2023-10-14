package payment

import (
	"context"
	"time"

	commonpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	paymentpb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GrpClient interface {
	RetrieveListOfOrders(ctx context.Context, in *paymentpb.RetrieveListOfOrdersRequest, opts ...grpc.CallOption) (*paymentpb.RetrieveListOfOrdersResponse, error)
}

// GetOrderList ctx must already has grpc token
func GetOrderList(ctx context.Context, paymentSvc GrpClient) (res *paymentpb.RetrieveListOfOrdersResponse, err error) {
	req := &paymentpb.RetrieveListOfOrdersRequest{
		CurrentTime: timestamppb.New(time.Now()),
		OrderStatus: paymentpb.OrderStatus_ORDER_STATUS_ALL,
		Filter:      nil,
		Paging: &commonpb.Paging{
			Limit: 100,
			Offset: &commonpb.Paging_OffsetInteger{
				OffsetInteger: 0,
			},
		},
	}

	res, err = paymentSvc.RetrieveListOfOrders(ctx, req)
	if err != nil {
		return nil, err
	}

	return
}

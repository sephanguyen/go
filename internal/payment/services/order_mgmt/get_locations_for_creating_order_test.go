package ordermgmt

import (
	"context"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/utils"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"
	mockServices "github.com/manabie-com/backend/mock/payment/services/order_mgmt"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetLocationsForCreatingOrderService_GetLocationsForCreatingOrder(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	var (
		db              *mockDb.Ext
		locationService *mockServices.ILocationServiceForCreatingOrder
	)

	expectedLocations := []*pb.LocationInfo{
		{
			LocationId:   "location_id_1",
			LocationName: "location_name_1",
		},
		{
			LocationId:   "location_id_2",
			LocationName: "location_name_2",
		},
	}
	testCases := []utils.TestCase{
		{
			Name: "Fail case: Error when get lowest granted locations for creating order ",
			Ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			Req: &pb.GetLocationsForCreatingOrderRequest{
				Name:  "location_name_1",
				Limit: 10,
			},
			ExpectedResp: &pb.GetLocationsForCreatingOrderResponse{
				LocationInfos: []*pb.LocationInfo{},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				locationService.On("GetLowestGrantedLocationsForCreatingOrder", mock.Anything, mock.Anything, mock.Anything).Return([]*pb.LocationInfo{}, constant.ErrDefault)
			},
		},
		{
			Name:         "Happy case",
			Ctx:          interceptors.ContextWithUserID(ctx, "user-id"),
			Req:          &pb.GetLocationsForCreatingOrderRequest{},
			ExpectedResp: &pb.GetLocationsForCreatingOrderResponse{LocationInfos: expectedLocations},
			Setup: func(ctx context.Context) {
				locationService.On("GetLowestGrantedLocationsForCreatingOrder", mock.Anything, mock.Anything, mock.Anything).Return([]*pb.LocationInfo{
					{
						LocationId:   "location_id_1",
						LocationName: "location_name_1",
					},
					{
						LocationId:   "location_id_2",
						LocationName: "location_name_2",
					},
				}, nil)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			locationService = new(mockServices.ILocationServiceForCreatingOrder)

			testCase.Setup(testCase.Ctx)
			s := &GetLocationsForCreatingOrder{
				DB:              db,
				LocationService: locationService,
			}
			req := testCase.Req.(*pb.GetLocationsForCreatingOrderRequest)
			resp, err := s.GetLocationsForCreatingOrder(testCase.Ctx, req)
			if testCase.ExpectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, testCase.ExpectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				expectedResp := testCase.ExpectedResp.(*pb.GetLocationsForCreatingOrderResponse)
				assert.Equal(t, len(expectedResp.LocationInfos), len(resp.LocationInfos))
				for idx, expectedItem := range expectedResp.LocationInfos {
					item := resp.LocationInfos[idx]
					assert.Equal(t, expectedItem.LocationId, item.LocationId)
					assert.Equal(t, expectedItem.LocationName, item.LocationName)
				}
			}
		})
	}
}

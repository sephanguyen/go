package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestGhnService_CreateOrder(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var handler func(rw http.ResponseWriter, req *http.Request)

	req := &CreateOrderRequest{}
	respData := `
{
	"ErrorMessage": "",
	"OrderID": 268916,
	"PaymentTypeID": 4,
	"OrderCode": "236697NF",
	"ExtraFee": 0,
	"TotalServiceFee": 81400,
	"ExpectedDeliveryTime": "2017-09-22T23:00:00+07:00",
	"ClientHubID": 0,
	"SortCode": "N/A"
}`
	successResp := &CreateOrderResponse{}

	_ = json.Unmarshal([]byte(respData), successResp)
	var testcases = []TestCase{
		{
			name:         "Success",
			ctx:          ctx,
			req:          req,
			expectedResp: successResp,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				handler = func(rw http.ResponseWriter, req *http.Request) {
					if req.Method == http.MethodPost && req.URL.String() == "/api/v1/apiv3/CreateOrder" {
						rw.WriteHeader(http.StatusOK)
						_, _ = rw.Write([]byte(`
{
    "code": 1,
    "msg": "Success",
    "data": {
        "ErrorMessage": "",
        "OrderID": 268916,
        "PaymentTypeID": 4,
        "OrderCode": "236697NF",
        "ExtraFee": 0,
        "TotalServiceFee": 81400,
        "ExpectedDeliveryTime": "2017-09-22T23:00:00+07:00",
        "ClientHubID": 0,
        "SortCode": "N/A"
    }
}`))
					}
				}
			},
		},
		{
			name:         "Fail 400 Bad Request",
			ctx:          ctx,
			req:          req,
			expectedResp: (*CreateOrderResponse)(nil),
			expectedErr:  errors.New(`HTTP code: 400, [GHN-ERR80] Giá trị truyền vào không hợp lệ, data: null`),
			setup: func(ctx context.Context) {
				handler = func(rw http.ResponseWriter, req *http.Request) {
					if req.Method == http.MethodPost && req.URL.String() == "/api/v1/apiv3/CreateOrder" {
						rw.WriteHeader(http.StatusBadRequest)
						_, _ = rw.Write([]byte(`{
	"code": 0,
	"msg": "[GHN-ERR80] Giá trị truyền vào không hợp lệ",
	"data": null
}`))
					}
				}
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(tc.ctx)
			// Start a local HTTP server
			server := httptest.NewServer(http.HandlerFunc(handler))
			// Close the server when test finishes
			defer server.Close()

			s := NewGHNService()
			resp, err := s.CreateOrder(ctx, server.URL, tc.req.(*CreateOrderRequest))

			if tc.expectedErr != nil {
				assert.Equal(t, tc.expectedErr.Error(), errors.Cause(err).Error())
			} else {
				assert.Nil(t, err)
			}

			assert.Equal(t, tc.expectedResp, resp)
		})
	}
}

func TestGhnService_CancelOrder(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := &CancelOrderRequest{}
	var handler func(rw http.ResponseWriter, req *http.Request)

	var testcases = []TestCase{
		{
			name: "Success",
			ctx:  ctx,
			req:  req,
			expectedResp: &CancelOrderResponse{
				ErrorMessage: "",
				HubID:        0,
				OrderCode:    "",
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				handler = func(rw http.ResponseWriter, req *http.Request) {
					if req.Method == http.MethodPost && req.URL.String() == "/api/v1/apiv3/CancelOrder" {
						rw.WriteHeader(http.StatusOK)
						_, _ = rw.Write([]byte(`
{
    "code": 1,
    "msg": "Success",
    "data": {
        "ErrorMessage": "",
        "HubID": 0,
        "OrderCode": ""
    }
}`))
					}
				}
			},
		},
		{
			name:         "Fail 400 Bad Request",
			ctx:          ctx,
			req:          req,
			expectedResp: (*CancelOrderResponse)(nil),
			expectedErr:  errors.New(`HTTP code: 400, [GHN-ERR76] Trạng thái đơn hàng không cho phép hủy đơn hàng, data: null`),
			setup: func(ctx context.Context) {
				handler = func(rw http.ResponseWriter, req *http.Request) {
					if req.Method == http.MethodPost && req.URL.String() == "/api/v1/apiv3/CancelOrder" {
						rw.WriteHeader(http.StatusBadRequest)
						_, _ = rw.Write([]byte(`{
    "code": 0,
    "msg": "[GHN-ERR76] Trạng thái đơn hàng không cho phép hủy đơn hàng",
    "data": null
}`))
					}
				}
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(tc.ctx)
			// Start a local HTTP server
			server := httptest.NewServer(http.HandlerFunc(handler))
			// Close the server when test finishes
			defer server.Close()

			s := NewGHNService()

			resp, err := s.CancelOrder(ctx, server.URL, tc.req.(*CancelOrderRequest))

			if tc.expectedErr != nil {
				assert.Equal(t, tc.expectedErr.Error(), errors.Cause(err).Error())
			} else {
				assert.Nil(t, err)
			}

			assert.Equal(t, tc.expectedResp, resp)
		})
	}
}

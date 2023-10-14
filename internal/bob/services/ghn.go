package services

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

type ShippingOrderCost struct {
	ServiceID int `json:"ServiceID,omitempty"`
}

type CreateOrderRequest struct {
	Token                string               `json:"token,omitempty"`
	PaymentTypeID        int                  `json:"PaymentTypeID,omitempty"`
	FromDistrictID       int                  `json:"FromDistrictID,omitempty"`
	FromWardCode         string               `json:"FromWardCode,omitempty"`
	ToDistrictID         int                  `json:"ToDistrictID,omitempty"`
	ToWardCode           string               `json:"ToWardCode,omitempty"`
	Note                 string               `json:"Note,omitempty"`
	SealCode             string               `json:"SealCode,omitempty"`
	ExternalCode         string               `json:"ExternalCode,omitempty"`
	ClientContactName    string               `json:"ClientContactName,omitempty"`
	ClientContactPhone   string               `json:"ClientContactPhone,omitempty"`
	ClientAddress        string               `json:"ClientAddress,omitempty"`
	CustomerName         string               `json:"CustomerName,omitempty"`
	CustomerPhone        string               `json:"CustomerPhone,omitempty"`
	ShippingAddress      string               `json:"ShippingAddress,omitempty"`
	CoDAmount            int                  `json:"CoDAmount,omitempty"`
	NoteCode             string               `json:"NoteCode,omitempty"`
	InsuranceFee         int                  `json:"InsuranceFee,omitempty"`
	ClientHubID          int                  `json:"ClientHubID,omitempty"`
	ServiceID            int                  `json:"ServiceID,omitempty"`
	ToLatitude           float64              `json:"ToLatitude,omitempty"`
	ToLongitude          float64              `json:"ToLongitude,omitempty"`
	FromLat              float64              `json:"FromLat,omitempty"`
	FromLng              float64              `json:"FromLng,omitempty"`
	Content              string               `json:"Content,omitempty"`
	CouponCode           string               `json:"CouponCode,omitempty"`
	Weight               int                  `json:"Weight,omitempty"`
	Length               int                  `json:"Length,omitempty"`
	Width                int                  `json:"Width,omitempty"`
	Height               int                  `json:"Height,omitempty"`
	CheckMainBankAccount bool                 `json:"CheckMainBankAccount,omitempty"`
	ShippingOrderCosts   []*ShippingOrderCost `json:"ShippingOrderCosts,omitempty"`
	ReturnContactName    string               `json:"ReturnContactName,omitempty"`
	ReturnContactPhone   string               `json:"ReturnContactPhone,omitempty"`
	ReturnAddress        string               `json:"ReturnAddress,omitempty"`
	ReturnDistrictID     int                  `json:"ReturnDistrictID,omitempty"`
	ExternalReturnCode   string               `json:"ExternalReturnCode,omitempty"`
	IsCreditCreate       bool                 `json:"IsCreditCreate,omitempty"`
	AffiliateID          int                  `json:"AffiliateID,omitempty"`
}

type CreateOrderResponse struct {
	ErrorMessage         string    `json:"ErrorMessage,omitempty"`
	OrderID              int       `json:"OrderID,omitempty"`
	PaymentTypeID        int       `json:"PaymentTypeID,omitempty"`
	OrderCode            string    `json:"OrderCode,omitempty"`
	ExtraFee             int       `json:"ExtraFee,omitempty"`
	TotalServiceFee      int       `json:"TotalServiceFee,omitempty"`
	ExpectedDeliveryTime time.Time `json:"ExpectedDeliveryTime,omitempty"`
	ClientHubID          int       `json:"ClientHubID,omitempty"`
	SortCode             string    `json:"SortCode,omitempty"`
}

type CancelOrderRequest struct {
	Token     string `json:"token,omitempty"`
	OrderCode string `json:"OrderCode,omitempty"`
}

type CancelOrderResponse struct {
	ErrorMessage string `json:"ErrorMessage,omitempty"`
	HubID        int    `json:"HubID,omitempty"`
	OrderCode    string `json:"OrderCode,omitempty"`
}

type FindAvailableServicesRequest struct {
	Token          string `json:"token,omitempty"`
	FromDistrictID int    `json:"FromDistrictID,omitempty"`
	ToDistrictID   int    `json:"ToDistrictID,omitempty"`
}

type AvailableService struct {
	ExpectedDeliveryTime time.Time `json:"ExpectedDeliveryTime"`
	Extras               []struct {
		MaxValue   int    `json:"MaxValue"`
		Name       string `json:"Name"`
		ServiceFee int    `json:"ServiceFee"`
		ServiceID  int    `json:"ServiceID"`
	} `json:"Extras"`
	Name       string `json:"Name"`
	ServiceFee int    `json:"ServiceFee"`
	ServiceID  int    `json:"ServiceID"`
}

type FindAvailableServicesResponse struct {
	Data []*AvailableService
}

type GHNService interface {
	CreateOrder(ctx context.Context, domain string, req *CreateOrderRequest) (*CreateOrderResponse, error)
	CancelOrder(ctx context.Context, domain string, req *CancelOrderRequest) (*CancelOrderResponse, error)
	FindAvailableServices(ctx context.Context, domain string, req *FindAvailableServicesRequest) (*FindAvailableServicesResponse, error)
}

type ghnService struct {
	client *http.Client
}

func NewGHNService() GHNService {
	return &ghnService{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (rcv *ghnService) CreateOrder(ctx context.Context, domain string, req *CreateOrderRequest) (*CreateOrderResponse, error) {
	data, err := rcv.doRequest(ctx, domain+"/api/v1/apiv3/CreateOrder", req)
	if err != nil {
		return nil, errors.Wrap(err, "rcv.doRequest")
	}

	resp := new(CreateOrderResponse)
	err = json.Unmarshal(data, resp)
	if err != nil {
		return nil, errors.Wrap(err, "json.Unmarshal")
	}

	return resp, nil
}

func (rcv *ghnService) CancelOrder(ctx context.Context, domain string, req *CancelOrderRequest) (*CancelOrderResponse, error) {
	data, err := rcv.doRequest(ctx, domain+"/api/v1/apiv3/CancelOrder", req)
	if err != nil {
		return nil, errors.Wrap(err, "rcv.doRequest")
	}

	resp := new(CancelOrderResponse)
	err = json.Unmarshal(data, resp)
	if err != nil {
		return nil, errors.Wrap(err, "json.Unmarshal")
	}

	return resp, nil
}

func (rcv *ghnService) FindAvailableServices(ctx context.Context, domain string, req *FindAvailableServicesRequest) (*FindAvailableServicesResponse, error) {
	data, err := rcv.doRequest(ctx, domain+"/api/v1/apiv3/FindAvailableServices", req)
	if err != nil {
		return nil, errors.Wrap(err, "rcv.doRequest")
	}

	resp := new(FindAvailableServicesResponse)
	err = json.Unmarshal(data, &resp.Data)
	if err != nil {
		return nil, errors.Wrap(err, "json.Unmarshal")
	}

	return resp, nil
}

func (rcv *ghnService) doRequest(ctx context.Context, url string, req interface{}) ([]byte, error) {
	reqBody := new(bytes.Buffer)
	err := json.NewEncoder(reqBody).Encode(req)
	if err != nil {
		return nil, errors.Wrap(err, "json Encode")
	}

	request, err := http.NewRequest(
		http.MethodPost,
		url,
		reqBody,
	)
	if err != nil {
		return nil, errors.Wrap(err, "http.NewRequest")
	}

	request = request.WithContext(ctx)

	request.Header.Add("Accept", "application/json")
	request.Header.Add("Content-Type", "application/json")

	resp, err := rcv.client.Do(request)
	if err != nil {
		return nil, errors.Wrap(err, "client.Do")
	}

	defer resp.Body.Close()

	// decode response body
	var body struct {
		Msg  string          `json:"msg,omitempty"`
		Data json.RawMessage `json:"data,omitempty"`
	}
	err = json.NewDecoder(resp.Body).Decode(&body)
	if err != nil {
		return nil, errors.Wrap(err, "json.Decode")
	}

	// unmarshal to create order response
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("HTTP code: %d, %s, data: %s", resp.StatusCode, body.Msg, string(body.Data))
	}

	return body.Data, err
}

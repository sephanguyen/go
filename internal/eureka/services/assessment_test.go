package services

import (
	"context"
	"errors"
	"testing"

	"github.com/manabie-com/backend/internal/eureka/configurations"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

func TestAssessmentService_NewAssessmentService(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Name             string
		DB               database.Ext
		LearnosityConfig *configurations.LearnosityConfig
		ExpectedOutput   *AssessmentService
	}{
		{
			Name: "happy case",
			DB:   nil,
			LearnosityConfig: &configurations.LearnosityConfig{
				ConsumerKey:    "consumer_key",
				ConsumerSecret: "consumer_secret",
			},
			ExpectedOutput: &AssessmentService{
				DB: nil,
				LearnosityConfig: &configurations.LearnosityConfig{
					ConsumerKey:    "consumer_key",
					ConsumerSecret: "consumer_secret",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			service := NewAssessmentService(tc.DB, tc.LearnosityConfig)
			assert.Equal(t, tc.ExpectedOutput, service)
		})
	}
}

func TestAssessmentService_GetSignedRequest(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctx = metadata.NewIncomingContext(ctx, metadata.MD{"origin": []string{"https://staging.manabie.io"}})

	service := &AssessmentService{
		LearnosityConfig: &configurations.LearnosityConfig{
			ConsumerKey:    "consumer_key",
			ConsumerSecret: "consumer_secret",
		},
	}

	testCases := []struct {
		Name             string
		Ctx              context.Context
		Request          any
		ExpectedResponse any
		ExpectedError    error
	}{
		{
			Name: "happy case: have domain in request",
			Ctx:  interceptors.ContextWithUserID(ctx, "user_id"),
			Request: &sspb.GetSignedRequestRequest{
				RequestData: "{\"limit\":5}",
				Domain:      "localhost",
			},
			ExpectedResponse: &sspb.GetSignedRequestResponse{
				SignedRequest: "{\"request\":\"{\\\"limit\\\":5}\",\"security\":{\"consumer_key\":\"consumer_key\",\"domain\":\"localhost\",\"signature\":\"",
			},
			ExpectedError: nil,
		},
		{
			Name: "happy case: empty domain - get domain from origin header",
			Ctx:  interceptors.ContextWithUserID(ctx, "user_id"),
			Request: &sspb.GetSignedRequestRequest{
				RequestData: "{\"limit\":5}",
			},
			ExpectedResponse: &sspb.GetSignedRequestResponse{
				SignedRequest: "{\"request\":\"{\\\"limit\\\":5}\",\"security\":{\"consumer_key\":\"consumer_key\",\"domain\":\"staging.manabie.io\",\"signature\":\"",
			},
			ExpectedError: nil,
		},
		{
			Name: "empty request data",
			Ctx:  interceptors.ContextWithUserID(ctx, "user_id"),
			Request: &sspb.GetSignedRequestRequest{
				RequestData: "",
				Domain:      "localhost",
			},
			ExpectedResponse: nil,
			ExpectedError:    errors.New("req must have RequestData"),
		},
		{
			Name: "empty user id",
			Ctx:  ctx,
			Request: &sspb.GetSignedRequestRequest{
				RequestData: "{\"limit\":5}",
				Domain:      "localhost",
			},
			ExpectedResponse: nil,
			ExpectedError:    errors.New("Security.UserID"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			response, err := service.GetSignedRequest(tc.Ctx, tc.Request.(*sspb.GetSignedRequestRequest))
			if err != nil {
				assert.Contains(t, err.Error(), tc.ExpectedError.Error())
			}
			if response != nil {
				assert.Contains(t, response.SignedRequest, tc.ExpectedResponse.(*sspb.GetSignedRequestResponse).SignedRequest)
			}
		})
	}
}

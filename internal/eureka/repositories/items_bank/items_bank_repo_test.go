package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/eureka/configurations"
	entities "github.com/manabie-com/backend/internal/eureka/entities/items_bank"
	"github.com/manabie-com/backend/internal/golibs/learnosity"
	"github.com/manabie-com/backend/internal/golibs/stringutil"
	mock_learnosity "github.com/manabie-com/backend/mock/golibs/learnosity"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestItemsBankRepo_GetExistedIDs(t *testing.T) {
	t.Parallel()
	mockHTTP := &mock_learnosity.HTTP{}
	dataAPI := &mock_learnosity.DataAPI{}
	mockAPI := &ItemsBankRepo{
		LearnosityConfig: &configurations.LearnosityConfig{
			ConsumerKey:    "consumer_key",
			ConsumerSecret: "consumer_secret",
		},
		HTTP:    mockHTTP,
		DataAPI: dataAPI,
	}
	ctx := context.Background()
	testCases := []struct {
		Name             string
		Ctx              context.Context
		Setup            func(ctx context.Context)
		InputItemIDs     []string
		ExpectedResponse []string
		ExpectedError    error
	}{
		{
			Name: "happy case",
			Ctx:  ctx,
			Setup: func(ctx context.Context) {
				dataAPI.On("Request",
					mock.Anything,
					mock.Anything,
					learnosity.EndpointDataAPIGetItems,
					mock.Anything,
					mock.MatchedBy(func(req learnosity.Request) bool {
						refs := req["references"].([]string)
						expected := []string{"00000001", "00000003"}
						return stringutil.SliceElementsMatch(refs, expected)
					}),
					learnosity.ActionGet).Once().Return(
					&learnosity.Result{
						Meta: map[string]any{
							"status":  true,
							"records": float64(0),
						},
						Data: json.RawMessage{},
					}, nil)

			},
			InputItemIDs:     []string{"00000001", "00000003"},
			ExpectedResponse: []string{},
			ExpectedError:    nil,
		},
		{
			Name: "error case - find a existed item id",
			Ctx:  ctx,
			Setup: func(ctx context.Context) {
				dataAPI.On("Request",
					mock.Anything,
					mock.Anything,
					learnosity.EndpointDataAPIGetItems,
					mock.Anything,
					mock.MatchedBy(func(req learnosity.Request) bool {
						refs := req["references"].([]string)
						expected := []string{"00000001", "00000003"}
						return stringutil.SliceElementsMatch(refs, expected)
					}),
					learnosity.ActionGet).Once().Return(
					&learnosity.Result{
						Meta: map[string]any{
							"status":  true,
							"records": float64(1),
						},
						Data: json.RawMessage(`[{"reference":"00000001"}]`),
					}, nil)

			},
			InputItemIDs:     []string{"00000001", "00000003"},
			ExpectedResponse: []string{"00000001"},
			ExpectedError:    nil,
		},
		{
			Name: "error case ",
			Ctx:  ctx,
			Setup: func(ctx context.Context) {
				dataAPI.On("Request",
					mock.Anything,
					mock.Anything,
					learnosity.EndpointDataAPIGetItems,
					mock.Anything,
					mock.MatchedBy(func(req learnosity.Request) bool {
						refs := req["references"].([]string)
						expected := []string{"00000001", "00000003"}
						return stringutil.SliceElementsMatch(refs, expected)
					}),
					learnosity.ActionGet).Once().Return(
					nil, fmt.Errorf("error"))

			},
			InputItemIDs:     []string{"00000001", "00000003"},
			ExpectedResponse: nil,
			ExpectedError:    fmt.Errorf("error"),
		},
	}

	for _, tc := range testCases {
		tc.Setup(tc.Ctx)
		t.Run(tc.Name, func(t *testing.T) {
			res, err := mockAPI.GetExistedIDs(tc.Ctx, tc.InputItemIDs)
			if err != nil {
				assert.Contains(t, err.Error(), tc.ExpectedError.Error())
			}
			result := stringutil.SliceElementsMatch(res, tc.ExpectedResponse)
			assert.Equal(t, result, true)

		})
	}

}

func TestItemsBankRepo_UploadContentData(t *testing.T) {
	t.Parallel()
	mockHTTP := &mock_learnosity.HTTP{}
	dataAPI := &mock_learnosity.DataAPI{}
	mockAPI := &ItemsBankRepo{
		LearnosityConfig: &configurations.LearnosityConfig{
			ConsumerKey:    "consumer_key",
			ConsumerSecret: "consumer_secret",
		},
		HTTP:    mockHTTP,
		DataAPI: dataAPI,
	}
	testCases := []struct {
		Name          string
		Ctx           context.Context
		Setup         func(ctx context.Context)
		Items         map[string]*entities.ItemsBankItem
		Questions     []*entities.ItemsBankQuestion
		OrgId         string
		ExpectedError error
	}{}

	for _, tc := range testCases {
		tc.Setup(tc.Ctx)
		t.Run(tc.Name, func(t *testing.T) {
			_, err := mockAPI.UploadContentData(tc.Ctx, tc.OrgId, tc.Items, tc.Questions)
			if err != nil {
				assert.Contains(t, err.Error(), tc.ExpectedError.Error())
			}
		})
	}

}

// Todo @hohieuu: update test for upload data endpoint

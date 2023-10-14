package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/zoom"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/zoom/domain"
	mock_clients "github.com/manabie-com/backend/mock/lessonmgmt/zoom/clients"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type TestCase struct {
	name         string
	ctx          context.Context
	req          interface{}
	expectedResp interface{}
	expectedErr  error
	setup        func(ctx context.Context)
}

func TestExternalConfigService_GetConfigByResource(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockConfigurationClient := &mock_clients.MockConfigurationClient{}
	secretKey := "452948404D635166546A576D5A7134743777217A25432A462D4A614E64526755"
	s := InitExternalConfigService(mockConfigurationClient, secretKey)
	organizationID := "id"
	configurationsData := []*mpb.Configuration{{Id: organizationID, ConfigValue: `{"account_id":"0cdf842b4bec3c3a32e7cd71602fe44b720eed02e0096b2ca999c0d89700bc178d7923de31431634653489a2add4f7f5e2b3","client_id":"6ce7a71ef2e5fa7e76104d34e73c4b8aef56b8d0bc5ff830cba9096be39057e75ae41c8e153c2536698fea02156b6b6538","client_secret":"95ebc7a226d1562688e0a9736a2faa724dc22a3f60dd574b0388e3736540b9140cf97e83469a1672e57663e1dc24653ca1c827ce2f103a28dfb5989b"}`}}
	tc := []TestCase{
		{
			name:        "should get zoom config success",
			ctx:         auth.InjectFakeJwtToken(ctx, organizationID),
			req:         nil,
			expectedErr: nil,
			expectedResp: &domain.ZoomConfig{
				AccountID:    "gWU7ykqSQzergk2qcVi-Ow",
				ClientID:     "joPIbcXsR5y_JaehPT3Fg",
				ClientSecret: "LOX31mK4KXmbCXV7zav123Eu0ugHiE2J",
			},
			setup: func(ctx context.Context) {
				getConfigReq := &mpb.GetConfigurationsRequest{
					Keyword: zoom.KeyZoomConfig,
					Paging: &cpb.Paging{
						Limit: 1,
					},
				}
				mockConfigurationClient.On("GetConfigurations", ctx, getConfigReq).
					Return(&mpb.GetConfigurationsResponse{Items: configurationsData}, nil).Once()
			},
		},
		{
			name:         "should throw error if not found config zoom for org",
			ctx:          auth.InjectFakeJwtToken(ctx, "id2"),
			req:          nil,
			expectedErr:  fmt.Errorf("not found config for org: %s", "id2"),
			expectedResp: nil,
			setup: func(ctx context.Context) {
				getConfigReq := &mpb.GetConfigurationsRequest{
					Keyword: zoom.KeyZoomConfig,
					Paging: &cpb.Paging{
						Limit: 1,
					},
				}
				mockConfigurationClient.On("GetConfigurations", ctx, getConfigReq).
					Return(&mpb.GetConfigurationsResponse{Items: []*mpb.Configuration{}}, nil).Once()
			},
		},
		{
			name:        "should not re-call when config exists",
			ctx:         auth.InjectFakeJwtToken(ctx, organizationID),
			req:         nil,
			expectedErr: nil,
			expectedResp: &domain.ZoomConfig{
				AccountID:    "gWU7ykqSQzergk2qcVi-Ow",
				ClientID:     "joPIbcXsR5y_JaehPT3Fg",
				ClientSecret: "LOX31mK4KXmbCXV7zav123Eu0ugHiE2J",
			},
			setup: func(ctx context.Context) {

			},
		},
	}

	for _, testCase := range tc {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			resp, err := s.GetConfigByResource(testCase.ctx)
			if testCase.expectedErr != nil {

				assert.Error(t, err)
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expectedResp, resp)
			}
			mock.AssertExpectationsForObjects(t, mockConfigurationClient)
		})
	}
}

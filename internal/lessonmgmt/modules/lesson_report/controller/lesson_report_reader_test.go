package controller

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/bob/constants"
	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/lessonmgmt/configurations"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	mock_repositories "github.com/manabie-com/backend/mock/bob/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRetrievePartnerDomain(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	// mock structs
	db := &mock_database.Ext{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	configRepo := &mock_repositories.MockConfigRepo{}
	validResourcePath := fmt.Sprint(constants.ManabieSchool)

	tcs := []struct {
		name        string
		reqUserID   string
		req         *lpb.GetPartnerDomainRequest
		setup       func(context.Context)
		expectedRes *lpb.GetPartnerDomainResponse
		hasError    bool
	}{
		{
			name:      "School admin get partner domain bo has in configs",
			reqUserID: "admin-1",
			req: &lpb.GetPartnerDomainRequest{
				Type: lpb.DomainType_DOMAIN_TYPE_BO,
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				configRepo.On("Retrieve",
					mock.Anything, mock.Anything,
					database.Text("COUNTRY_MASTER"),
					database.Text("lesson"),
					database.TextArray([]string{"domain_local_bo_-2147483648"})).
					Once().
					Return([]*entities.Config{
						{
							Key:     database.Text("domain_local_bo_-2147483648"),
							Group:   database.Text("lesson"),
							Country: database.Text("COUNTRY_MASTER"),
							Value:   database.Text("https://staging-jprep-school-portal.web.app/"),
						},
					}, nil)
			},
			expectedRes: &lpb.GetPartnerDomainResponse{
				Domain: "https://staging-jprep-school-portal.web.app/",
			},
		},
		{
			name:      "School admin get partner domain teacher has in configs",
			reqUserID: "admin-1",
			req: &lpb.GetPartnerDomainRequest{
				Type: lpb.DomainType_DOMAIN_TYPE_TEACHER,
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				configRepo.On("Retrieve",
					mock.Anything, mock.Anything,
					database.Text("COUNTRY_MASTER"),
					database.Text("lesson"),
					database.TextArray([]string{"domain_local_teacher_-2147483648"})).
					Once().
					Return([]*entities.Config{
						{
							Key:     database.Text("domain_local_teacher_-2147483648"),
							Group:   database.Text("lesson"),
							Country: database.Text("COUNTRY_MASTER"),
							Value:   database.Text("https://teacher.com/"),
						},
					}, nil)
			},
			expectedRes: &lpb.GetPartnerDomainResponse{
				Domain: "https://teacher.com/",
			},
		},
		{
			name:      "School admin get partner domain teacher has in configs",
			reqUserID: "admin-1",
			req: &lpb.GetPartnerDomainRequest{
				Type: lpb.DomainType_DOMAIN_TYPE_LEARNER,
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				configRepo.On("Retrieve",
					mock.Anything, mock.Anything,
					database.Text("COUNTRY_MASTER"),
					database.Text("lesson"),
					database.TextArray([]string{"domain_local_learner_-2147483648"})).
					Once().
					Return([]*entities.Config{
						{
							Key:     database.Text("domain_local_learner_-2147483648"),
							Group:   database.Text("lesson"),
							Country: database.Text("COUNTRY_MASTER"),
							Value:   database.Text("https://learner.com/"),
						},
					}, nil)
			},
			expectedRes: &lpb.GetPartnerDomainResponse{
				Domain: "https://learner.com/",
			},
		},
		{
			name:      "School admin get partner domain bo without configs",
			reqUserID: "admin-1",
			req: &lpb.GetPartnerDomainRequest{
				Type: lpb.DomainType_DOMAIN_TYPE_BO,
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				configRepo.On("Retrieve",
					mock.Anything, mock.Anything,
					database.Text("COUNTRY_MASTER"),
					database.Text("lesson"),
					database.TextArray([]string{"domain_local_bo_-2147483648"})).
					Once().
					Return([]*entities.Config{}, nil)
			},
			expectedRes: &lpb.GetPartnerDomainResponse{
				Domain: "https://staging-jprep-school-portal.web.app/",
			},
		},
		{
			name:      "School admin get partner domain teacher without configs",
			reqUserID: "admin-1",
			req: &lpb.GetPartnerDomainRequest{
				Type: lpb.DomainType_DOMAIN_TYPE_TEACHER,
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				configRepo.On("Retrieve",
					mock.Anything, mock.Anything,
					database.Text("COUNTRY_MASTER"),
					database.Text("lesson"),
					database.TextArray([]string{"domain_local_teacher_-2147483648"})).
					Once().
					Return([]*entities.Config{}, nil)
			},
			expectedRes: &lpb.GetPartnerDomainResponse{
				Domain: "https://staging-jprep-teacher.web.app/",
			},
		},
		{
			name:      "School admin get partner domain learner without configs",
			reqUserID: "admin-1",
			req: &lpb.GetPartnerDomainRequest{
				Type: lpb.DomainType_DOMAIN_TYPE_LEARNER,
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				configRepo.On("Retrieve",
					mock.Anything, mock.Anything,
					database.Text("COUNTRY_MASTER"),
					database.Text("lesson"),
					database.TextArray([]string{"domain_local_learner_-2147483648"})).
					Once().
					Return([]*entities.Config{}, nil)
			},
			expectedRes: &lpb.GetPartnerDomainResponse{
				Domain: "https://staging-jprep-learner.web.app/",
			},
		},
		{
			name:      "Teacher get partner domain bo has in configs",
			reqUserID: "teacher-1",
			req: &lpb.GetPartnerDomainRequest{
				Type: lpb.DomainType_DOMAIN_TYPE_BO,
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				configRepo.On("Retrieve",
					mock.Anything, mock.Anything,
					database.Text("COUNTRY_MASTER"),
					database.Text("lesson"),
					database.TextArray([]string{"domain_local_bo_-2147483648"})).
					Once().
					Return([]*entities.Config{
						{
							Key:     database.Text("domain_local_bo_-2147483648"),
							Group:   database.Text("lesson"),
							Country: database.Text("COUNTRY_MASTER"),
							Value:   database.Text("https://staging-jprep-school-portal.web.app/"),
						},
					}, nil)
			},
			expectedRes: &lpb.GetPartnerDomainResponse{
				Domain: "https://staging-jprep-school-portal.web.app/",
			},
		},
		{
			name:      "Student get partner domain bo has in configs",
			reqUserID: "learner-1",
			req: &lpb.GetPartnerDomainRequest{
				Type: lpb.DomainType_DOMAIN_TYPE_BO,
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				configRepo.On("Retrieve",
					mock.Anything, mock.Anything,
					database.Text("COUNTRY_MASTER"),
					database.Text("lesson"),
					database.TextArray([]string{"domain_local_bo_-2147483648"})).
					Once().
					Return([]*entities.Config{
						{
							Key:     database.Text("domain_local_bo_-2147483648"),
							Group:   database.Text("lesson"),
							Country: database.Text("COUNTRY_MASTER"),
							Value:   database.Text("https://staging-jprep-school-portal.web.app/"),
						},
					}, nil)
			},
			expectedRes: &lpb.GetPartnerDomainResponse{
				Domain: "https://staging-jprep-school-portal.web.app/",
			},
		},

		{
			name:      "User get domain error when get config error",
			reqUserID: "admin-1",
			req: &lpb.GetPartnerDomainRequest{
				Type: lpb.DomainType_DOMAIN_TYPE_BO,
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				configRepo.On("Retrieve",
					mock.Anything, mock.Anything,
					database.Text("COUNTRY_MASTER"),
					database.Text("lesson"),
					database.TextArray([]string{"domain_local_bo_-2147483648"})).
					Once().
					Return([]*entities.Config{}, errors.New("error"))
			},
			expectedRes: &lpb.GetPartnerDomainResponse{
				Domain: "",
			},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			claim := &interceptors.CustomClaims{
				Manabie: &interceptors.ManabieClaims{
					UserID:       tc.reqUserID,
					ResourcePath: validResourcePath,
				},
			}

			ctxT := interceptors.ContextWithJWTClaims(ctx, claim)
			tc.setup(ctxT)
			srv := LessonReportReaderService{
				cfg: &configurations.Config{
					Common: configs.CommonConfig{
						Environment: "local",
					},
					Partner: configs.PartnerConfig{
						DomainBo:      "https://staging-jprep-school-portal.web.app/",
						DomainTeacher: "https://staging-jprep-teacher.web.app/",
						DomainLearner: "https://staging-jprep-learner.web.app/",
					},
				},
				wrapperConnection: wrapperConnection,
				configRepo:        configRepo,
			}
			actualRes, err := srv.RetrievePartnerDomain(ctxT, tc.req)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.EqualValues(t, tc.expectedRes, actualRes)
				mock.AssertExpectationsForObjects(t, db, configRepo, mockUnleashClient)
			}
		})
	}
}

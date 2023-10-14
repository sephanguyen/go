package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	mock_repositories "github.com/manabie-com/backend/mock/mastermgmt/modules/organization/repositories"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb_ms "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TestCase struct {
	name        string
	ctx         context.Context
	req         interface{}
	expectedErr error
	setup       func(ctx context.Context)
}

// ErrOrganizationEnrollmentStatusNotAllowedTobeNone returned when create or edit student
var ErrOrganizationEnrollmentStatusNotAllowedTobeNone = errors.New("organization enrollment status not allowed to be ORGANIZATION_ENROLLMENT_STATUS_NONE")

func TestCreateOrganization(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	organizationRepo := new(mock_repositories.MockOrganizationRepo)

	tx := new(mock_database.Tx)
	db := new(mock_database.Ext)
	jsm := new(mock_nats.JetStreamManagement)

	s := OrganizationService{
		DB:               db,
		JSM:              jsm,
		OrganizationRepo: organizationRepo,
	}

	testCases := []TestCase{
		{
			name: "create organization success",
			ctx:  ctx,
			req: &pb_ms.CreateOrganizationRequest{Organization: &pb_ms.Organization{
				OrganizationId:   idutil.ULIDNow(),
				TenantId:         idutil.ULIDNow(),
				OrganizationName: "organization name test",
				DomainName:       "domain-test",
				LogoUrl:          "logo-url",
				CountryCode:      cpb.Country(pb.COUNTRY_VN),
			}},
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				organizationRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				jsm.On("TracedPublish", mock.Anything, "publishOrganizationEvent", mock.Anything, mock.Anything).Once().Return(nil, nil)
			},
		},
		{
			name: "trace publish error",
			ctx:  ctx,
			req: &pb_ms.CreateOrganizationRequest{Organization: &pb_ms.Organization{
				OrganizationId:   idutil.ULIDNow(),
				TenantId:         idutil.ULIDNow(),
				OrganizationName: "organization name test",
				DomainName:       "domain-test",
				LogoUrl:          "logo-url",
				CountryCode:      cpb.Country(pb.COUNTRY_VN),
			}},
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				organizationRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
				jsm.On("TracedPublish", mock.Anything, "publishOrganizationEvent", mock.Anything, mock.Anything).Once().Return(nil, errors.New("publishOrganizationEvent: s.JSM.Publish failed"))
			},
			expectedErr: errors.New("publishOrganizationEvent: s.JSM.Publish failed"),
		},
		{
			name:        "cannot create organization if organization empty",
			ctx:         ctx,
			req:         &pb_ms.CreateOrganizationRequest{Organization: nil},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Errorf("organization invalid params").Error()),
			setup:       func(ctx context.Context) {},
		},
		{
			name: "cannot create organization if tenantID empty",
			ctx:  ctx,
			req: &pb_ms.CreateOrganizationRequest{Organization: &pb_ms.Organization{
				OrganizationId:   idutil.ULIDNow(),
				OrganizationName: "Manabie School",
				DomainName:       "domain-test",
				LogoUrl:          "logo-url",
				CountryCode:      cpb.Country(pb.COUNTRY_VN),
			}},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Errorf("tenant id cannot be empty").Error()),
			setup:       func(ctx context.Context) {},
		},
		{
			name: "cannot create organization if organization name empty",
			ctx:  ctx,
			req: &pb_ms.CreateOrganizationRequest{Organization: &pb_ms.Organization{
				OrganizationId:   idutil.ULIDNow(),
				TenantId:         idutil.ULIDNow(),
				OrganizationName: "",
				DomainName:       "domain-test",
				LogoUrl:          "logo-url",
				CountryCode:      cpb.Country(pb.COUNTRY_VN),
			}},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Errorf("organization name cannot be empty").Error()),
			setup:       func(ctx context.Context) {},
		},
		{
			name: "cannot create organization if organization name empty",
			ctx:  ctx,
			req: &pb_ms.CreateOrganizationRequest{Organization: &pb_ms.Organization{
				OrganizationId:   idutil.ULIDNow(),
				TenantId:         idutil.ULIDNow(),
				OrganizationName: "",
				DomainName:       "domain-test",
				LogoUrl:          "logo-url",
				CountryCode:      cpb.Country(pb.COUNTRY_VN),
			}},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Errorf("organization name cannot be empty").Error()),
			setup:       func(ctx context.Context) {},
		},
		{
			name: "fail create new organziation",
			ctx:  ctx,
			req: &pb_ms.CreateOrganizationRequest{Organization: &pb_ms.Organization{
				OrganizationId:   idutil.ULIDNow(),
				TenantId:         idutil.ULIDNow(),
				OrganizationName: "organization name test",
				DomainName:       "domain-test",
				LogoUrl:          "logo-url",
				CountryCode:      cpb.Country(pb.COUNTRY_VN),
			}},
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				organizationRepo.On("Create", ctx, tx, mock.Anything).Once().Return(errors.New("error create new organization"))
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
			expectedErr: status.Error(codes.Unknown, fmt.Errorf("error create new organization").Error()),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			claim := &interceptors.CustomClaims{
				Manabie: &interceptors.ManabieClaims{
					ResourcePath: "1",
				},
			}
			testCase.ctx = interceptors.ContextWithJWTClaims(ctx, claim)

			testCase.setup(testCase.ctx)

			_, err := s.CreateOrganization(testCase.ctx, testCase.req.(*pb_ms.CreateOrganizationRequest))
			if err != nil {
				fmt.Println(err)
			}

			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
			mock.AssertExpectationsForObjects(t, organizationRepo, db, tx, jsm)
		})
	}
}

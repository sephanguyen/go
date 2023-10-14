package controller

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/appsmith/domain"
	mock_alert "github.com/manabie-com/backend/mock/golibs/alert"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TestCase struct {
	name         string
	ctx          context.Context
	req          interface{}
	expectedResp interface{}
	expectedErr  error
	setup        func(ctx context.Context)
}

func TestAppsmithService_GetPageInfoBySlug(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()
	newPageRepo := &MockNewPageRepo{}

	mockAlert := &mock_alert.SlackFactory{}
	s := NewAppsmithService(mt.DB, newPageRepo, "local", "manabie", mockAlert)

	tc := []TestCase{
		{
			name:        "config found",
			ctx:         interceptors.ContextWithUserID(ctx, "id"),
			req:         &mpb.GetPageInfoBySlugRequest{Slug: "slug-1", ApplicationId: "app-1", BranchName: "brand-1"},
			expectedErr: nil,
			expectedResp: &mpb.GetPageInfoBySlugResponse{
				Id:            "id-1",
				ApplicationId: "app-1",
			},
			setup: func(ctx context.Context) {
				newPageRepo.On("GetBySlug", ctx, mt.DB, "slug-1", "app-1", "brand-1").
					Return(&domain.NewPage{
						ID:            "id-1",
						ApplicationID: "app-1",
					}, nil).Times(1)

			},
		},
		{
			name:         "internal err",
			ctx:          interceptors.ContextWithUserID(ctx, "id"),
			req:          &mpb.GetPageInfoBySlugRequest{Slug: "slug-1", ApplicationId: "app-1", BranchName: "brand-1"},
			expectedErr:  status.Error(codes.Internal, "internal err"),
			expectedResp: &mpb.GetConfigurationByKeyResponse{},
			setup: func(ctx context.Context) {
				newPageRepo.On("GetBySlug", ctx, mt.DB, "slug-1", "app-1", "brand-1").
					Return(nil, errors.New("internal err")).Times(1)
				mockAlert.On("Send", mock.Anything).Once().Return(nil)
			},
		},
	}

	for _, testCase := range tc {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*mpb.GetPageInfoBySlugRequest)
			resp, err := s.GetPageInfoBySlug(testCase.ctx, req)
			if testCase.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}

func TestAppsmithService_GetSchemaByID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()
	newPageRepo := &MockNewPageRepo{}
	mockAlert := &mock_alert.SlackFactory{}
	s := NewAppsmithService(mt.DB, newPageRepo, "local", "manabie", mockAlert)

	tc := []TestCase{
		{
			name:        "schema found",
			ctx:         interceptors.ContextWithUserID(ctx, "id"),
			req:         &mpb.GetSchemaNameByWorkspaceIDRequest{WorkspaceId: "635f5299b3ce396b06d52db8"},
			expectedErr: nil,
			expectedResp: &mpb.GetSchemaNameByWorkspaceIDResponse{
				Schema: "architecture",
			},
		},
		{
			name:        "schema not found",
			ctx:         interceptors.ContextWithUserID(ctx, "id"),
			req:         &mpb.GetSchemaNameByWorkspaceIDRequest{WorkspaceId: "id"},
			expectedErr: nil,
			expectedResp: &mpb.GetSchemaNameByWorkspaceIDResponse{
				Schema: "",
			},
		},
	}

	for _, testCase := range tc {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			req := testCase.req.(*mpb.GetSchemaNameByWorkspaceIDRequest)
			resp, err := s.GetSchemaByWorkspaceID(testCase.ctx, req)
			if testCase.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}

type MockNewPageRepo struct {
	mock.Mock
}

func (r *MockNewPageRepo) GetBySlug(arg1 context.Context, arg2 *mongo.Database, arg3, arg4, arg5 string) (*domain.NewPage, error) {
	args := r.Called(arg1, arg2, arg3, arg4, arg5)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.NewPage), args.Error(1)
}

type Ext struct {
	mock.Mock
}

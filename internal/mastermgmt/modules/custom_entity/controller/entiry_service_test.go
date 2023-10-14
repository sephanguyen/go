package controller

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"github.com/jackc/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func TestCustomEntityService_ExecCustomEntity(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	s := &CustomEntityService{
		DB: db,
	}

	tc := []TestCase{
		{
			name:        "successfully!",
			ctx:         interceptors.ContextWithUserID(ctx, "id"),
			req:         &mpb.ExecuteCustomEntityRequest{Sql: "CREATE TABLE public.abc"},
			expectedErr: nil,
			expectedResp: &mpb.ExecuteCustomEntityResponse{
				Success: true,
				Error:   "",
			},
			setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("Exec", mock.Anything, mock.AnythingOfType("string")).Once().Return(cmdTag, nil)
			},
		},
		{
			name:         "Error",
			ctx:          interceptors.ContextWithUserID(ctx, "id"),
			req:          &mpb.ExecuteCustomEntityRequest{Sql: "CREATE TABLE public.abc"},
			expectedErr:  status.Error(codes.Internal, fmt.Errorf("Permission denied").Error()),
			expectedResp: &mpb.GetConfigurationByKeyResponse{},
			setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("Exec", mock.Anything, mock.AnythingOfType("string")).Once().Return(cmdTag, fmt.Errorf("Permission denied"))
			},
		},
	}

	for _, testCase := range tc {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*mpb.ExecuteCustomEntityRequest)
			resp, err := s.ExecuteCustomEntity(testCase.ctx, req)
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

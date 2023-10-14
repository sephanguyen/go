package service

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/classdo/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/classdo/infrastructure/repo"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/lessonmgmt/classdo/repositories"
	mock_clients "github.com/manabie-com/backend/mock/lessonmgmt/zoom/clients"

	"github.com/jackc/pgtype"
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

func TestPortForwardClassDoService_PortForwardClassDo(t *testing.T) {

	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	mockHTTPClient := &mock_clients.MockHTTPClient{}
	config := &configs.ClassDoConfig{
		SecretKey: "fakeSecretKey",
		Endpoint:  "fakeEndpoint",
	}
	mockRepo := &mock_repositories.MockClassDoAccountRepo{}
	s := NewPortForwardClassDoService(config, db, mockHTTPClient, mockRepo)

	classDoID := "ID"
	classDoAccount := &repo.ClassDoAccount{
		ClassDoID:     pgtype.Text{String: classDoID, Status: pgtype.Present},
		ClassDoAPIKey: pgtype.Text{String: "fakeApiKey", Status: pgtype.Present},
	}
	body := `{"body":"fakeBody"}`
	response := `{"response":"fakeResponse"}`
	tc := []TestCase{
		{
			name: "success",
			ctx:  ctx,
			req: &domain.PortForwardClassDoRequest{
				ClassDoID: classDoID,
				Body:      body,
			},
			expectedResp: &domain.PortForwardClassDoResponse{
				Response: response,
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockRepo.On("GetClassDoAccountByID", ctx, db, classDoID).Return(classDoAccount, nil).Once()
				mockHTTPClient.On("SendRequest", ctx, mock.Anything).
					Return(&http.Response{Body: ioutil.NopCloser(bytes.NewBuffer([]byte(response)))}, nil).Once()
			},
		},
		{
			name: "GetClassDoAccountByID failed",
			ctx:  ctx,
			req: &domain.PortForwardClassDoRequest{
				ClassDoID: classDoID,
				Body:      body,
			},
			expectedResp: &domain.PortForwardClassDoResponse{
				Response: response,
			},
			expectedErr: fmt.Errorf("GetClassDoAccountByID failed"),
			setup: func(ctx context.Context) {
				mockRepo.On("GetClassDoAccountByID", ctx, db, classDoID).Return(nil, fmt.Errorf("GetClassDoAccountByID failed")).Once()
			},
		},
		{
			name: "port forward to class do grapql failed",
			ctx:  ctx,
			req: &domain.PortForwardClassDoRequest{
				ClassDoID: classDoID,
				Body:      body,
			},
			expectedResp: &domain.PortForwardClassDoResponse{
				Response: response,
			},
			expectedErr: fmt.Errorf("port forward grapql classdo failed: %s", "err"),
			setup: func(ctx context.Context) {
				mockRepo.On("GetClassDoAccountByID", ctx, db, classDoID).Return(classDoAccount, nil).Once()
				mockHTTPClient.On("SendRequest", ctx, mock.Anything).
					Return(nil, fmt.Errorf("err")).Once()
			},
		},
	}

	for _, testCase := range tc {
		tCase := testCase
		t.Run(tCase.name, func(t *testing.T) {
			tCase.setup(tCase.ctx)

			resp, err := s.PortForwardClassDo(tCase.ctx, tCase.req.(*domain.PortForwardClassDoRequest))
			if tCase.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tCase.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tCase.expectedResp, resp)
			}

			mock.AssertExpectationsForObjects(t, db, mockHTTPClient, mockRepo)
		})
	}
}

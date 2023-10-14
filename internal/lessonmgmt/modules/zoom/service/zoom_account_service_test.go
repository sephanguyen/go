package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/exporter"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/zoom/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/zoom/infrastructure/repo"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_lesson_repositories "github.com/manabie-com/backend/mock/lessonmgmt/lesson/repositories"
	mock_repositories "github.com/manabie-com/backend/mock/lessonmgmt/zoom/repositories"
	mock_service "github.com/manabie-com/backend/mock/lessonmgmt/zoom/service"

	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestZoomAccountService_ImportZoomAccount(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db := &mock_database.Ext{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	mockZoomAccountRepo := &mock_repositories.MockZoomAccountRepo{}
	mockLessonRepo := &mock_lesson_repositories.MockLessonRepo{}
	tx := &mock_database.Tx{}

	mockZoomService := &mock_service.MockZoomService{}
	s := NewZoomAccountService(
		wrapperConnection,
		mockZoomService,
		mockZoomAccountRepo,
		mockLessonRepo,
	)

	tc := []TestCase{
		{
			name: "should import zoom account success",
			ctx:  ctx,
			req: &lpb.ImportZoomAccountRequest{
				Payload: []byte(fmt.Sprintf(`zoom_id,zoom_username,school_id,Action
				pid_1,name1,1,Upsert
				pid_2,name2,1,Upsert`)),
			},
			expectedErr: nil,
			expectedResp: &lpb.ImportZoomAccountResponse{
				Errors: []*lpb.ImportZoomAccountResponse_ImportZoomAccountError{},
			},
			setup: func(ctx context.Context) {
				zoomUserResponse := domain.UserZoomResponse{
					PageCount:  1,
					PageNumber: 1,
					Users: []domain.ZoomUserInfo{
						{Email: "name1", Status: domain.ZoomUserStatusActive}, {Email: "name2", Status: domain.ZoomUserStatusActive},
					},
				}
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				mockZoomService.On("RetryGetListUsers", ctx, &domain.ZoomGetListUserRequest{
					PageNumber: 1,
					PageSize:   300,
				}).Once().Return(&zoomUserResponse, nil)
				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				mockZoomAccountRepo.On("Upsert", mock.Anything, tx,
					mock.Anything).
					Once().Return(nil)

			},
		},
		{
			name: "should call delete zoom of lesson when delete zoom account success",
			ctx:  ctx,
			req: &lpb.ImportZoomAccountRequest{
				Payload: []byte(fmt.Sprintf(`zoom_id,zoom_username,school_id,Action
				pid_1,name1,1,Delete
				pid_2,name2,1,Upsert`)),
			},
			expectedErr: nil,
			expectedResp: &lpb.ImportZoomAccountResponse{
				Errors: []*lpb.ImportZoomAccountResponse_ImportZoomAccountError{},
			},
			setup: func(ctx context.Context) {
				zoomUserResponse := domain.UserZoomResponse{
					PageCount:  1,
					PageNumber: 1,
					Users: []domain.ZoomUserInfo{
						{Email: "name1", Status: domain.ZoomUserStatusActive}, {Email: "name2", Status: domain.ZoomUserStatusActive},
					},
				}
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				mockZoomService.On("RetryGetListUsers", ctx, &domain.ZoomGetListUserRequest{
					PageNumber: 1,
					PageSize:   300,
				}).Once().Return(&zoomUserResponse, nil)
				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				mockZoomAccountRepo.On("Upsert", mock.Anything, tx,
					mock.Anything).
					Once().Return(nil)

				mockLessonRepo.On("RemoveZoomLinkOfLesson", mock.Anything, tx,
					[]string{"pid_1"}).
					Once().Return(nil)
			},
		},
		{
			name: "should response a file error when exists a zoom account invalid",
			ctx:  ctx,
			req: &lpb.ImportZoomAccountRequest{
				Payload: []byte(fmt.Sprintf(`zoom_id,zoom_username,school_id,Action
				pid_1,name1,1,Upsert
				pid_2,,1,Upsert`)),
			},
			expectedErr: nil,
			expectedResp: &lpb.ImportZoomAccountResponse{
				Errors: []*lpb.ImportZoomAccountResponse_ImportZoomAccountError{
					{RowNumber: 3, Error: "invalid zoom account detail: email could not be empty"},
				},
			},
			setup: func(ctx context.Context) {
				zoomUserResponse := domain.UserZoomResponse{
					PageCount:  1,
					PageNumber: 1,
					Users: []domain.ZoomUserInfo{
						{Email: "name1", Status: domain.ZoomUserStatusActive}, {Email: "name2", Status: domain.ZoomUserStatusActive},
					},
				}
				mockZoomService.On("RetryGetListUsers", ctx, &domain.ZoomGetListUserRequest{
					PageNumber: 1,
					PageSize:   300,
				}).Once().Return(&zoomUserResponse, nil)

			},
		},
		{
			name: "should throw error when file empty",
			ctx:  ctx,
			req: &lpb.ImportZoomAccountRequest{
				Payload: []byte(fmt.Sprintf(``)),
			},
			expectedErr:  status.Error(codes.InvalidArgument, "no data in csv file"),
			expectedResp: nil,
			setup: func(ctx context.Context) {
			},
		},
	}

	for _, testCase := range tc {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			resp, err := s.ImportZoomAccount(testCase.ctx, testCase.req.(*lpb.ImportZoomAccountRequest))
			if testCase.expectedErr != nil {

				assert.Error(t, err)
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expectedResp, resp)
			}
			mock.AssertExpectationsForObjects(t, db, mockZoomAccountRepo, mockLessonRepo, mockUnleashClient)
		})
	}
}

func TestZoomAccountService_ExportZoomAccount(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db := &mock_database.Ext{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")

	mockZoomAccountRepo := &mock_repositories.MockZoomAccountRepo{}
	mockLessonRepo := &mock_lesson_repositories.MockLessonRepo{}

	mockZoomService := &mock_service.MockZoomService{}
	s := NewZoomAccountService(
		wrapperConnection,
		mockZoomService,
		mockZoomAccountRepo,
		mockLessonRepo,
	)
	exportCols := []exporter.ExportColumnMap{
		{
			DBColumn:  "zoom_id",
			CSVColumn: domain.ZoomIDLabel,
		},
		{
			DBColumn:  "email",
			CSVColumn: domain.ZoomUsernameLabel,
		},
	}
	exportable := sliceutils.Map([]*repo.ZoomAccount{}, func(d *repo.ZoomAccount) database.Entity {
		return d
	})

	str, _ := exporter.ExportBatch(exportable, exportCols)

	tc := []TestCase{
		{
			name:        "should export zoom account success",
			ctx:         ctx,
			req:         &lpb.ExportZoomAccountRequest{},
			expectedErr: nil,
			expectedResp: &lpb.ExportZoomAccountResponse{
				Data: exporter.ToCSV(str),
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				mockZoomAccountRepo.On("GetAllZoomAccount", mock.Anything, mock.Anything,
					mock.Anything).
					Once().Return([]*repo.ZoomAccount{}, nil)

			},
		},
	}

	for _, testCase := range tc {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			resp, err := s.ExportZoomAccount(testCase.ctx, testCase.req.(*lpb.ExportZoomAccountRequest))
			if testCase.expectedErr != nil {

				assert.Error(t, err)
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expectedResp, resp)
			}
			mock.AssertExpectationsForObjects(t, db, mockZoomAccountRepo, mockLessonRepo, mockUnleashClient)
		})
	}
}

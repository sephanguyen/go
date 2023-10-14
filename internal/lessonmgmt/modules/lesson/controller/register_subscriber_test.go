package controller

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_repositories "github.com/manabie-com/backend/mock/lessonmgmt/lesson/repositories"
	mock_repositories_report "github.com/manabie-com/backend/mock/lessonmgmt/lesson_report/repositories"
	mock_repo "github.com/manabie-com/backend/mock/mastermgmt/modules/class/infrastructure/repo"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

const errSubString = "error subscribing to subject"

func TestRegisterLockLessonSubscriptionHandler(t *testing.T) {
	t.Parallel()
	_, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	jsm := new(mock_nats.JetStreamManagement)
	lessonRepo := new(mock_repositories.MockLessonRepo)
	db := &mock_database.Ext{}
	unleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, unleashClient, "local")

	type TestCase struct {
		name        string
		expectedErr error
		setup       func(ctx context.Context)
	}

	testCases := []TestCase{
		{
			name: "success",
			setup: func(ctx context.Context) {
				unleashClient.
					On("IsFeatureEnabled", mock.Anything, mock.Anything).
					Return(true, nil).Once()
				jsm.On("QueueSubscribe", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)

			},
		},
		{
			name:        "failed",
			expectedErr: fmt.Errorf(errSubString),
			setup: func(ctx context.Context) {
				unleashClient.
					On("IsFeatureEnabled", mock.Anything, mock.Anything).
					Return(true, nil).Once()
				jsm.On("QueueSubscribe", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf(errSubString))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctxRP := golibs.ResourcePathToCtx(context.Background(), "school-id")
			tc.setup(ctxRP)

			err := RegisterLockLessonSubscriptionHandler(jsm, zap.NewNop(), wrapperConnection, lessonRepo, "local", unleashClient)
			if tc.expectedErr != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mock.AssertExpectationsForObjects(t, unleashClient)
		})
	}
}

func TestRegisterStudentClassSubscriptionHandler(t *testing.T) {
	t.Parallel()
	_, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	jsm := new(mock_nats.JetStreamManagement)
	lessonRepo := new(mock_repositories.MockLessonRepo)
	db := &mock_database.Ext{}
	unleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, unleashClient, "local")
	lessonMemberRepo := new(mock_repositories.MockLessonMemberRepo)
	classMemberRepo := new(mock_repo.MockClassMemberRepo)
	lessonReportRepo := new(mock_repositories_report.MockLessonReportRepo)
	type TestCase struct {
		name        string
		expectedErr error
		setup       func(ctx context.Context)
	}

	testCases := []TestCase{
		{
			name: "success",
			setup: func(ctx context.Context) {
				jsm.On("QueueSubscribe", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
			},
		},
		{
			name:        "failed",
			expectedErr: fmt.Errorf(errSubString),
			setup: func(ctx context.Context) {
				jsm.On("QueueSubscribe", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf(errSubString))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctxRP := golibs.ResourcePathToCtx(context.Background(), "school-id")
			tc.setup(ctxRP)
			err := RegisterStudentClassSubscriptionHandler(jsm, zap.NewNop(), db, wrapperConnection, lessonRepo, lessonMemberRepo, classMemberRepo, lessonReportRepo)
			if tc.expectedErr != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mock.AssertExpectationsForObjects(t, unleashClient)
		})
	}
}

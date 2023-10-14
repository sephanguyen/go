package multitenant

import (
	"context"
	"testing"
	"time"

	mocks "github.com/manabie-com/backend/mock/golibs/auth"

	"firebase.google.com/go/v4/auth"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func aMockTenantManager() *tenantManager {
	tm := defaultTenantManager()
	tm.gcpUtils = new(mocks.GCPUtils)
	tm.gcpApp = nil
	tm.gcpTenantManager = new(mocks.GCPTenantManager)

	return tm
}

func TestNewTenantManager(t *testing.T) {
	t.Parallel()

	var tm *tenantManager

	testCases := []struct {
		name                           string
		expectedUserBatchSize          int
		expectedUserBatchImportTimeout time.Duration
		setupFunc                      func(ctx context.Context)
	}{
		{
			name:                           "init tenant manager by default",
			expectedUserBatchSize:          DefaultUserBatchSize,
			expectedUserBatchImportTimeout: DefaultUserBatchImportTimeout,
			setupFunc: func(ctx context.Context) {
				tm = defaultTenantManager()
			},
		},
		{
			name:                           "init tenant manager with options",
			expectedUserBatchSize:          100,
			expectedUserBatchImportTimeout: time.Second,
			setupFunc: func(ctx context.Context) {
				tm = defaultTenantManager(WithUserBatchSize(100), WithUserBatchImportTimeout(time.Second))
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setupFunc(ctx)

			assert.Equal(t, testCase.expectedUserBatchSize, tm.userBatchSize)
			assert.Equal(t, testCase.expectedUserBatchImportTimeout.Nanoseconds(), tm.userBatchImportTimeout.Nanoseconds())
		})
	}
}

func TestTenantManager_Tenant(t *testing.T) {
	t.Parallel()

	var tm *tenantManager

	testCases := []struct {
		name          string
		inputTenantID string
		expectedErr   error
		setupFunc     func(ctx context.Context)
	}{
		{
			name:          "get tenant with empty tenant id",
			inputTenantID: "",
			expectedErr:   ErrTenantIDIsEmpty,
			setupFunc: func(ctx context.Context) {
				tm = aMockTenantManager()
			},
		},
		{
			name:          "get tenant with non-existing tenant id",
			inputTenantID: "existingTenantID",
			expectedErr:   ErrTenantNotFound,
			setupFunc: func(ctx context.Context) {
				tm = aMockTenantManager()

				tm.gcpTenantManager.(*mocks.GCPTenantManager).On("Tenant", ctx, "existingTenantID").Once().Return(new(auth.Tenant), errors.New("a error occurs"))
				tm.gcpUtils.(*mocks.GCPUtils).On("IsTenantNotFound", mock.Anything).Once().Return(true)
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setupFunc(ctx)

			tenant, actualErr := tm.Tenant(ctx, testCase.inputTenantID)

			assert.Equal(t, testCase.expectedErr, actualErr)
			if testCase.expectedErr == nil {
				assert.NotNil(t, tenant)
			} else {
				assert.Nil(t, tenant)
			}
		})
	}
}

func TestTenantManager_TenantClient(t *testing.T) {
	t.Parallel()

	var tm *tenantManager

	testCases := []struct {
		name          string
		inputTenantID string
		expectedErr   error
		setupFunc     func(ctx context.Context)
	}{
		{
			name:          "get tenant client with empty tenant id",
			inputTenantID: "",
			expectedErr:   ErrTenantIDIsEmpty,
			setupFunc: func(ctx context.Context) {
				tm = aMockTenantManager()
			},
		},
		{
			name:          "get tenant client with non-existing tenant id",
			inputTenantID: "existingTenantID",
			expectedErr:   ErrTenantNotFound,
			setupFunc: func(ctx context.Context) {
				tm = aMockTenantManager()
				//tenantClient := NewTenantClient(new(mocks.GCPTenantClient))

				tm.gcpTenantManager.(*mocks.GCPTenantManager).On("AuthForTenant", "existingTenantID").Once().Return(new(auth.TenantClient), errors.New("a error occurs"))
				tm.gcpUtils.(*mocks.GCPUtils).On("IsTenantNotFound", mock.Anything).Once().Return(true)
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setupFunc(ctx)

			tenantClient, actualErr := tm.TenantClient(ctx, testCase.inputTenantID)

			assert.Nil(t, tenantClient)
			assert.Equal(t, testCase.expectedErr, actualErr)
		})
	}
}

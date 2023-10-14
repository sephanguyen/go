package multitenant

import (
	"context"
	"sync"
	"time"

	"github.com/manabie-com/backend/internal/usermgmt/pkg/gcp"

	"github.com/pkg/errors"
)

type TenantManager interface {
	TenantClient(ctx context.Context, tenantID string) (TenantClient, error)
	//CreateTenant(ctx context.Context, tenantToCreate TenantInfo) (Tenant, error)
	//DeleteTenant(ctx context.Context, tenantID string) error
}

type tenantManager struct {
	tenants map[string]TenantClient

	gcpApp           *gcp.App
	gcpTenantManager GCPTenantManager
	gcpUtils         GCPUtils

	// options
	userBatchSize          int
	userBatchImportTimeout time.Duration
	//userBatchImportInterval time.Duration

	mutex sync.RWMutex
	//shutdown chan chan error

	secondaryTenantConfigProvider gcp.TenantConfigProvider
}

func defaultTenantManager(opts ...TenantManagerOption) *tenantManager {
	tm := &tenantManager{
		tenants:                make(map[string]TenantClient),
		userBatchSize:          DefaultUserBatchSize,
		userBatchImportTimeout: DefaultUserBatchImportTimeout,
		gcpUtils:               NewGCPUtils(),
	}
	// apply options to instance
	for _, opt := range opts {
		opt(tm)
	}
	return tm
}

func NewTenantManagerFromGCP(ctx context.Context, gcpApp *gcp.App, opts ...TenantManagerOption) (TenantManager, error) {
	authClient, err := gcpApp.Auth(ctx)
	if err != nil {
		return nil, err
	}

	tm := defaultTenantManager(opts...)
	tm.gcpApp = gcpApp
	tm.gcpTenantManager = authClient.TenantManager

	/*go tenantManager.run()
	go tenantManager.userBatchProducer()
	go tenantManager.userBatchConsumer()*/

	return tm, nil
}

// Tenant gets tenant info by tenant id
// return ErrTenantNotFound if tenant with that id is not exists
func (tm *tenantManager) Tenant(ctx context.Context, tenantID string) (Tenant, error) {
	if tenantID == "" {
		return nil, ErrTenantIDIsEmpty
	}

	gcpTenant, err := tm.gcpTenantManager.Tenant(ctx, tenantID)
	if err != nil {
		if tm.gcpUtils.IsTenantNotFound(err) {
			return nil, ErrTenantNotFound
		}
		return nil, err
	}

	return newTenantFromGCPTenant(gcpTenant), nil
}

// TenantClient get tenant client to interact with tenant
// return ErrTenantNotFound if tenant with that id is not exists
func (tm *tenantManager) TenantClient(ctx context.Context, tenantID string) (TenantClient, error) {
	if tenantID == "" {
		return nil, ErrTenantIDIsEmpty
	}

	cachedTenant, isTenantCached := tm.tenants[tenantID]
	if !isTenantCached {
		gcpTenantClient, err := tm.gcpTenantManager.AuthForTenant(tenantID)
		if err != nil {
			if tm.gcpUtils.IsTenantNotFound(err) {
				return nil, ErrTenantNotFound
			}
			return nil, err
		}

		config, err := tm.gcpApp.GetTenantConfig(ctx, tenantID)
		if err != nil {
			return nil, errors.Wrap(err, "gcpApp.GetTenantConfig")
		}

		if err := IsScryptHashValid(config.HashConfig); err != nil {
			if tm.secondaryTenantConfigProvider == nil {
				return nil, errors.Wrap(err, "secondaryTenantConfigProvider is nil")
			}

			fallbackTenantConfig, err := tm.secondaryTenantConfigProvider.GetTenantConfig(ctx, tenantID)
			if err != nil {
				return nil, errors.Wrap(err, "secondaryTenantConfigProvider.GetTenantConfig")
			}
			if err := IsScryptHashValid(fallbackTenantConfig.HashConfig); err != nil {
				return nil, errors.Wrap(err, "secondaryTenantConfigProvider.GetTenantConfig hash config is empty")
			}
			config.HashConfig = fallbackTenantConfig.HashConfig
		}

		tenantClient := &tenantClient{
			tenantID:   tenantID,
			HashConfig: config.HashConfig,
			gcpClient:  gcpTenantClient,
			gcpUtils:   tm.gcpUtils,
		}

		tm.mutex.Lock()
		tm.tenants[tenantID] = tenantClient
		tm.mutex.Unlock()
		cachedTenant = tenantClient
	}
	return cachedTenant, nil
}

/*func (tm *tenantManager) CreateTenant(ctx context.Context, tenantToCreate TenantInfo) (Tenant, error) {
	if tenantToCreate.GetDisplayName() == "" {
		return nil, errors.New("tenant name is empty")
	}

	t := new(auth.TenantToCreate).
		DisplayName(tenantToCreate.GetDisplayName()).
		AllowPasswordSignUp(tenantToCreate.GetPasswordSignUpAllowed()).
		EnableEmailLinkSignIn(tenantToCreate.GetEmailLinkSignInEnabled())

	createdTenant, err := tm.gcpTenantManager.CreateTenant(ctx, t)
	if err != nil {
		return nil, err
	}

	tenant := &tenant{
		id:                     createdTenant.ID,
		displayName:            createdTenant.DisplayName,
		passwordSignUpAllowed:  createdTenant.AllowPasswordSignUp,
		emailLinkSignInEnabled: createdTenant.EnableEmailLinkSignIn,
	}

	return tenant, nil
}

func (tm *tenantManager) DeleteTenant(ctx context.Context, tenantID string) error {
	return tm.gcpTenantManager.DeleteTenant(ctx, tenantID)
}*/

// Implement later
/*func (tm *tenantManager) Shutdown(ctx context.Context) error {
	shutdownReq := make(chan error)
	select {
	case <-ctx.Done():
		return ctx.Err()
	case tm.shutdown <- shutdownReq:
		return <-shutdownReq
	}
}*/

package multitenant

import (
	"time"

	"github.com/manabie-com/backend/internal/golibs/gcp"
)

const (
	DefaultUserBatchSize          = 1000
	DefaultUserBatchImportTimeout = 5 * time.Second
)

type TenantManagerOption func(identityPlatform *tenantManager)

func WithSecondaryTenantConfigProvider(tenantConfigProvider gcp.TenantConfigProvider) TenantManagerOption {
	return func(tenantManager *tenantManager) {
		tenantManager.secondaryTenantConfigProvider = tenantConfigProvider
	}
}

func WithUserBatchSize(userBatchSize int) TenantManagerOption {
	return func(tenantManager *tenantManager) {
		tenantManager.userBatchSize = userBatchSize
	}
}

func WithUserBatchImportTimeout(timeout time.Duration) TenantManagerOption {
	return func(tenantManager *tenantManager) {
		tenantManager.userBatchImportTimeout = timeout
	}
}

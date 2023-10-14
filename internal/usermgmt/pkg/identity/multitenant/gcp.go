package multitenant

import (
	"context"

	"firebase.google.com/go/v4/auth"
)

type GCPTenantClient interface {
	GetUserByEmail(ctx context.Context, email string) (*auth.UserRecord, error)
	GetUser(ctx context.Context, uid string) (*auth.UserRecord, error)
	CreateUser(ctx context.Context, user *auth.UserToCreate) (*auth.UserRecord, error)
	ImportUsers(ctx context.Context, users []*auth.UserToImport, opts ...auth.UserImportOption) (*auth.UserImportResult, error)
	UpdateUser(ctx context.Context, uid string, user *auth.UserToUpdate) (ur *auth.UserRecord, err error)
	Users(ctx context.Context, nextPageToken string) *auth.UserIterator
	CustomToken(ctx context.Context, uid string) (string, error)
}

type GCPTenantManager interface {
	Tenant(ctx context.Context, tenantID string) (*auth.Tenant, error)
	AuthForTenant(tenantID string) (*auth.TenantClient, error)
	CreateTenant(ctx context.Context, tenant *auth.TenantToCreate) (*auth.Tenant, error)
	DeleteTenant(ctx context.Context, tenantID string) error
}

type GCPPager interface {
	NextPage(interface{}) (nextPageToken string, err error)
}

type GCPUtils interface {
	IsTenantNotFound(err error) bool
	IsUserNotFound(err error) bool
}

type gcpUtils struct{}

func (utils *gcpUtils) IsTenantNotFound(err error) bool {
	return auth.IsTenantNotFound(err)
}

func (utils *gcpUtils) IsUserNotFound(err error) bool {
	return auth.IsUserNotFound(err)
}

func NewGCPUtils() GCPUtils {
	return new(gcpUtils)
}

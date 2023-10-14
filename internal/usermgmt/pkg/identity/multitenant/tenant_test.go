package multitenant

import (
	"testing"

	"firebase.google.com/go/v4/auth"
	"github.com/stretchr/testify/assert"
)

func aValidGCPTenant() *auth.Tenant {
	tenant := &auth.Tenant{
		ID:                    "example-id",
		DisplayName:           "example-display-name",
		AllowPasswordSignUp:   true,
		EnableEmailLinkSignIn: true,
	}
	return tenant
}

func aValidTenant() *tenant {
	tenant := &tenant{
		id:                     "example-id",
		displayName:            "example-display-name",
		passwordSignUpAllowed:  true,
		emailLinkSignInEnabled: true,
	}
	return tenant
}

func TestTenant_TenantImpl(t *testing.T) {
	t.Parallel()

	tenant := aValidTenant()

	assert.Equal(t, tenant.id, tenant.GetID())
	assert.Equal(t, tenant.displayName, tenant.GetDisplayName())
	assert.Equal(t, tenant.passwordSignUpAllowed, tenant.GetPasswordSignUpAllowed())
	assert.Equal(t, tenant.emailLinkSignInEnabled, tenant.GetEmailLinkSignInEnabled())
}

func TestNewTenantFromGCPTenant(t *testing.T) {
	t.Parallel()

	gcpAuthTenant := aValidGCPTenant()

	tenant := newTenantFromGCPTenant(gcpAuthTenant)

	assert.Equal(t, gcpAuthTenant.ID, tenant.GetID())
	assert.Equal(t, gcpAuthTenant.DisplayName, tenant.GetDisplayName())
	assert.Equal(t, gcpAuthTenant.AllowPasswordSignUp, tenant.GetPasswordSignUpAllowed())
	assert.Equal(t, gcpAuthTenant.EnableEmailLinkSignIn, tenant.GetEmailLinkSignInEnabled())
}

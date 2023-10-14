package multitenant

import "firebase.google.com/go/v4/auth"

type TenantIdentifier interface {
	GetID() string
}

type TenantInfo interface {
	GetDisplayName() string
	GetPasswordSignUpAllowed() bool
	GetEmailLinkSignInEnabled() bool
}

// Tenant represents a tenant information
type Tenant interface {
	TenantIdentifier
	TenantInfo
}

type tenant struct {
	id                     string
	displayName            string
	passwordSignUpAllowed  bool
	emailLinkSignInEnabled bool
}

func (t *tenant) GetID() string {
	return t.id
}

func (t *tenant) GetDisplayName() string {
	return t.displayName
}

func (t *tenant) GetPasswordSignUpAllowed() bool {
	return t.passwordSignUpAllowed
}

func (t *tenant) GetEmailLinkSignInEnabled() bool {
	return t.emailLinkSignInEnabled
}

type Tenants []Tenant

func newTenantFromGCPTenant(gcpTenant *auth.Tenant) Tenant {
	tenant := &tenant{
		id:                     gcpTenant.ID,
		displayName:            gcpTenant.DisplayName,
		passwordSignUpAllowed:  gcpTenant.AllowPasswordSignUp,
		emailLinkSignInEnabled: gcpTenant.EnableEmailLinkSignIn,
	}
	return tenant
}

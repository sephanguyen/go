package multitenant

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

type TenantOption func(*tenant)

func NewTenant(opts ...TenantOption) Tenant {
	t := &tenant{}

	// Loop through each option
	for _, opt := range opts {
		// Call the option giving the instantiated
		// *House as the argument
		opt(t)
	}
	return t
}

func WithTenantID(id string) TenantOption {
	return func(h *tenant) {
		h.id = id
	}
}

func WithTenantDisplayName(displayName string) TenantOption {
	return func(h *tenant) {
		h.displayName = displayName
	}
}

func WithTenantPasswordSignUpAllowed(passwordSignUpAllowed bool) TenantOption {
	return func(h *tenant) {
		h.passwordSignUpAllowed = passwordSignUpAllowed
	}
}

func WithTenantEmailLinkSignInEnabled(emailLinkSignInEnabled bool) TenantOption {
	return func(h *tenant) {
		h.emailLinkSignInEnabled = emailLinkSignInEnabled
	}
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

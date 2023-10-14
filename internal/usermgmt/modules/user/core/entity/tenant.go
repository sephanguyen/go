package entity

type Tenant struct {
	ID                     string
	DisplayName            string
	PasswordSignUpAllowed  bool
	EmailLinkSignInEnabled bool
}

func (t *Tenant) GetID() string {
	return t.ID
}

func (t *Tenant) GetDisplayName() string {
	return t.DisplayName
}

func (t *Tenant) GetPasswordSignUpAllowed() bool {
	return t.PasswordSignUpAllowed
}

func (t *Tenant) GetEmailLinkSignInEnabled() bool {
	return t.EmailLinkSignInEnabled
}

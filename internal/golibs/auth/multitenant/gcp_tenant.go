package multitenant

import (
	"firebase.google.com/go/v4/auth"
)

func newTenantFromGCPTenant(gcpTenant *auth.Tenant) Tenant {
	tenant := NewTenant(
		WithTenantID(gcpTenant.ID),
		WithTenantDisplayName(gcpTenant.DisplayName),
		WithTenantPasswordSignUpAllowed(gcpTenant.AllowPasswordSignUp),
		WithTenantEmailLinkSignInEnabled(gcpTenant.EnableEmailLinkSignIn),
	)
	return tenant
}

//Temporarily disable for now
/*func newGCPAuthClientFromCredentialFile(ctx context.Context, credentialsFile string) (*auth.Client, error) {
	app, err := firebase.NewApp(ctx, nil, option.WithCredentialsFile(credentialsFile))
	if err != nil {
		log.Println("error initializing app:", err)
	}

	authClient, err := app.Auth(ctx)
	if err != nil {
		log.Println("error getting Auth client:", err)
	}

	return authClient, err
}*/

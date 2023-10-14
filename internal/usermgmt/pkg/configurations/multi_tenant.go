package configurations

//MultiTenantConfig contains information to init multi tenant application
//After we unified the configuration, we can remove this implementation
//and implement GetGCPProjectID() and GetGCPServiceAccountID() for Config directly after
type MultiTenantConfig struct {
	projectID        string
	serviceAccountID string
}

func (config MultiTenantConfig) GetGCPProjectID() string {
	return config.projectID
}

func (config MultiTenantConfig) GetGCPServiceAccountID() string {
	return config.serviceAccountID
}

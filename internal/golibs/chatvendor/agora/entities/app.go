package entities

type AppInfo struct {
	ID                                    uint64 `json:"id"`
	UUID                                  string `json:"uuid"`
	Type                                  string `json:"type"`
	Name                                  string `json:"name"`
	Created                               uint64 `json:"created"`
	Modified                              uint64 `json:"modified"`
	ApplicationName                       string `json:"applicationName"`
	OrganizationName                      string `json:"organizationName"`
	ProductName                           string `json:"productName"`
	AppStatus                             string `json:"app_status"`
	AppDesc                               string `json:"appDesc"`
	TTL                                   uint64 `json:"ttl"`
	AllowOpenRegistration                 bool   `json:"allow_open_registration"`
	AllowAppForceNotification             bool   `json:"allow_app_force_notification"`
	RegistrationRequiresEmailConfirmation bool   `json:"registration_requires_email_confirmation"`
	RegistrationRequiresAdminApproval     bool   `json:"registration_requires_admin_approval"`
	NotifyAdminOfNewUsers                 bool   `json:"notify_admin_of_new_users"`
	AppStatusUpdateTimestamp              uint64 `json:"app_status_update_timestamp"`
}

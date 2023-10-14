package form_partner

type EvictionPartner interface {
	getMapNewField() map[string][]string
	GetConfigFormId() string
	SetConfigFormId(formId string)
	GetConfigForm() string
	GetConfigFormName() string
}

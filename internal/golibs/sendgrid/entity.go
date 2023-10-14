package sendgrid

// A struct to encapsulate a Sender/Recipient entity apart from SendGrid's entities
// for when we need to adopt new fields and requirements
type Correspondent struct {
	Name    string
	Address string

	// If this field is empty, use the root dynamic data instead
	//
	// Note: DynamicData and SubstitutionData are mutually excluded and should not be used together
	DynamicData map[string]interface{}

	// If this field is empty, use the root substitution data instead
	//
	// Note: DynamicData and SubstitutionData are mutually excluded and should not be used together
	SubstitutionData map[string]string

	// CustomArguments support adding custom fields and values into each Recipient entity
	CustomArguments []map[string]string
}

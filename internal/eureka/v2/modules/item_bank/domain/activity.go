package domain

type Activity struct {
	Reference string
	Data      ActivityData
	Tags      Tags
}

type Config struct {
	Regions string
}

type ActivityData struct {
	// Items can be an array of string, or array of objects :(
	Items         []any
	Config        Config
	RenderingType string
}

type Tags struct {
	Tenant []string `json:"tenant,omitempty"`
}

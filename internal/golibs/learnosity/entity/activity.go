package entity

const (
	RenderingTypeAssess = "assess"
)

type Activity struct {
	Reference string       `json:"reference,omitempty"`
	Data      ActivityData `json:"data,omitempty"`
	Tags      Tags         `json:"tags,omitempty"`
}

type Config struct {
	Regions string `json:"regions,omitempty"`
}

type ActivityData struct {
	// Items can be an array of string, or array of objects :(
	Items         []any  `json:"items,omitempty"`
	Config        Config `json:"config,omitempty"`
	RenderingType string `json:"rendering_type,omitempty"`
}

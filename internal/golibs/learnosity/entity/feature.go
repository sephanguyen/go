package entity

type Data struct {
	Heading string `json:"heading"`
	Content string `json:"content"`
	Type    string `json:"type"`
	IsMath  bool   `json:"is_math,omitempty"`
}
type Feature struct {
	Type      string `json:"type"`
	Reference string `json:"reference"`
	Data      Data   `json:"data"`
}

func NewPassageFeature(heading, content, reference string, isMath bool) *Feature {
	featureType := "sharedpassage"
	return &Feature{
		Type:      featureType,
		Reference: reference,
		Data: Data{
			Heading: heading,
			Content: content,
			Type:    featureType,
			IsMath:  isMath,
		},
	}
}

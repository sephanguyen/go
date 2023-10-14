package entities

type UnleashFeatureEntity struct {
	Strategies []struct {
		Name        string        `json:"name"`
		Constraints []interface{} `json:"constraints"`
		Parameters  struct {
			Environments string `json:"environments"`
		} `json:"parameters"`
	} `json:"strategies"`
	ImpressionData bool          `json:"impressionData"`
	Enabled        bool          `json:"enabled"`
	Name           string        `json:"name"`
	Description    string        `json:"description"`
	Project        string        `json:"project"`
	Stale          bool          `json:"stale"`
	Type           string        `json:"type"`
	Variants       []interface{} `json:"variants"`
}

package entity

const (
	ItemStatusPublished = "published"
)

type Definition struct {
	Widgets []Reference `json:"widgets,omitempty"`
}

type Item struct {
	Reference  string      `json:"reference,omitempty"`
	Status     string      `json:"status,omitempty"`
	Metadata   interface{} `json:"metadata,omitempty"`
	Definition Definition  `json:"definition,omitempty"`
	Questions  []Reference `json:"questions,omitempty"`
	Features   []Reference `json:"features,omitempty"`
	Tags       Tags        `json:"tags,omitempty"`
}

package entity

const (
	StatusPublished   = "published"
	StatusUnpublished = "unpublished"
	StatusArchived    = "archived"
)

type Reference struct {
	Reference string `json:"reference,omitempty"`
}

type Tags struct {
	Tenant []string `json:"tenant,omitempty"`
}

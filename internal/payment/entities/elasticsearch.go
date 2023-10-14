package entities

type ESIndexMigration struct {
	IndexName    string `json:"index_name"`
	IndexVersion string `json:"index_version"`
}

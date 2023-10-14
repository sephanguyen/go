package rls

// use only for testing with custom policy

type TemplateHasuraRole struct {
	Name   string       `json:"name" yaml:"name"`
	Filter *interface{} `json:"filter,omitempty" yaml:"filter,omitempty"`
	Check  *interface{} `json:"check,omitempty" yaml:"check,omitempty"`
}

type TemplateHasuraPolicy struct {
	SelectPermission *[]TemplateHasuraRole `json:"select_permission" yaml:"selectPermission"`
	InsertPermission *[]TemplateHasuraRole `json:"insert_permission" yaml:"insertPermission"`
	DeletePermission *[]TemplateHasuraRole `json:"delete_permission" yaml:"deletePermission"`
	UpdatePermission *[]TemplateHasuraRole `json:"update_permission" yaml:"updatePermission"`

	ArrayCustomRelationship  *[]ArrayCustomRelationship  `json:"array_custom_relationship,omitempty" yaml:"arrayCustomRelationship,omitempty"`
	ObjectCustomRelationship *[]ObjectCustomRelationship `json:"object_custom_relationship,omitempty" yaml:"objectCustomRelationship,omitempty"`
}

type ArrayCustomRelationship struct {
	TableName    string                   `json:"table_name" yaml:"tableName"`
	ManualConfig HasuraArrayRelationships `json:"manual_config" yaml:"manualConfig"`
}

type ObjectCustomRelationship struct {
	TableName    string                    `json:"table_name" yaml:"tableName"`
	ManualConfig HasuraObjectRelationships `json:"manual_config" yaml:"manualConfig"`
}

type TemplatePostgresPolicy struct {
	Name      string `json:"name" yaml:"name"`
	Using     string `json:"using" yaml:"using"`
	WithCheck string `json:"with_check" yaml:"withCheck"`
	For       string `json:"for" yaml:"for"`
}
type TemplatesPolicy struct {
	UseCustomPolicy       *bool                     `json:"use_custom_policy" yaml:"useCustomPolicy"`
	HasuraPolicy          *TemplateHasuraPolicy     `json:"hasura_policies" yaml:"hasuraPolicy"`
	PostgresPolicy        *[]TemplatePostgresPolicy `json:"postgres_policies" yaml:"postgresPolicy"`
	UseCustomHasuraPolicy *bool                     `json:"use_custom_hasura_policy" yaml:"useCustomHasuraPolicy"`
	PostgresPolicyVersion *int                      `json:"postgres_policy_version,omitempty" yaml:"postgresPolicyVersion"`
}

type Template struct {
	Template         string                   `yaml:"template"`
	TableName        string                   `yaml:"tableName"`
	AccessPathTable  *TemplateAccessPathTable `yaml:"accessPathTable"`
	LocationCol      *string                  `yaml:"locationCol"`
	PermissionPrefix *string                  `yaml:"permissionPrefix"`
	Permissions      *TemplatePermission      `yaml:"permissions"`
	OwnerCol         *string                  `yaml:"ownerCol"`

	UseCustomPolicy *bool                     `yaml:"useCustomPolicy"`
	HasuraPolicy    *TemplateHasuraPolicy     `yaml:"hasuraPolicy"`
	PostgresPolicy  *[]TemplatePostgresPolicy `yaml:"postgresPolicy"`

	UseCustomHasuraPolicy *bool `yaml:"useCustomHasuraPolicy"`

	PostgresPolicyVersion *int `yaml:"postgresPolicyVersion"`
}

type TemplateAccessPathTable struct {
	Name          string             `yaml:"name"`
	ColumnMapping *map[string]string `yaml:"columnMapping"`
}

type TemplatePermission struct {
	Postgres *[]string `yaml:"postgres"`
	Hasura   *[]string `yaml:"hasura"`
}

type TemplateFile struct {
	Templates    *[]Template
	FileDir      string
	DatabaseName string
}

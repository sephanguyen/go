package rls

type HasuraForeignKeyConstraintOn struct {
	Column string            `yaml:"column"`
	Table  HasuraTableSchema `yaml:"table"`
}

type HasuraRemoteTable struct {
	Schema string `yaml:"schema"`
	Name   string `yaml:"name"`
}

type HasuraManualConfiguration struct {
	RemoteTable   HasuraRemoteTable `yaml:"remote_table"`
	ColumnMapping map[string]string `yaml:"column_mapping"`
}

type HasuraUsingObjectRelationships struct {
	ForeignKeyConstraintOn string                     `yaml:"foreign_key_constraint_on,omitempty"`
	ManualConfiguration    *HasuraManualConfiguration `yaml:"manual_configuration,omitempty"`
}

type HasuraUsing struct {
	ForeignKeyConstraintOn *HasuraForeignKeyConstraintOn `yaml:"foreign_key_constraint_on,omitempty"`
	ManualConfiguration    *HasuraManualConfiguration    `yaml:"manual_configuration,omitempty"`
}

type HasuraObjectRelationships struct {
	Name  string                          `yaml:"name"`
	Using *HasuraUsingObjectRelationships `yaml:"using,omitempty"`
}

type HasuraArrayRelationships struct {
	Name  string       `yaml:"name"`
	Using *HasuraUsing `yaml:"using,omitempty"`
}

type HasuraPermission struct {
	Columns           []string     `yaml:"columns"`
	Filter            *interface{} `yaml:"filter,omitempty"`
	Limit             int          `yaml:"limit,omitempty"`
	AllowAggregations bool         `yaml:"allow_aggregations,omitempty"`
}

type HasuraInsertPermission struct {
	Check       *interface{}       `yaml:"check,omitempty"`
	Set         *map[string]string `yaml:"set,omitempty"`
	Columns     []string           `yaml:"columns"`
	Filter      *interface{}       `yaml:"filter,omitempty"`
	BackendOnly *interface{}       `yaml:"backend_only,omitempty"`
}

type HasuraDeletePermission struct {
	Check  *interface{} `yaml:"check,omitempty"`
	Filter *interface{} `yaml:"filter,omitempty"`
}

type HasuraInsertPermissions struct {
	Role       string                  `yaml:"role"`
	Permission *HasuraInsertPermission `yaml:"permission,omitempty"`
}

type HasuraDeletePermissions struct {
	Role       string                  `yaml:"role"`
	Permission *HasuraDeletePermission `yaml:"permission,omitempty"`
}
type HasuraSelectPermissions struct {
	Role       string            `yaml:"role"`
	Permission *HasuraPermission `yaml:"permission,omitempty"`
}

type HasuraTableSchema struct {
	Schema string `yaml:"schema"`
	Name   string `yaml:"name"`
}

type HasuraTable struct {
	Table               HasuraTableSchema            `yaml:"table"`
	ObjectRelationships *[]HasuraObjectRelationships `yaml:"object_relationships,omitempty"`
	ArrayRelationships  *[]HasuraArrayRelationships  `yaml:"array_relationships,omitempty"`
	InsertPermissions   *[]HasuraInsertPermissions   `yaml:"insert_permissions,omitempty"`
	SelectPermissions   *[]HasuraSelectPermissions   `yaml:"select_permissions,omitempty"`
	UpdatePermissions   *[]HasuraInsertPermissions   `yaml:"update_permissions,omitempty"`
	DeletePermissions   *[]HasuraDeletePermissions   `yaml:"delete_permissions,omitempty"`
}

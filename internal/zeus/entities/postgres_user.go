package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type PostgresNamespace struct {
	Namespace        pgtype.Name
	AccessPrivileges pgtype.TextArray
}

// FieldMap return a map of field name and pointer to field
func (e *PostgresNamespace) FieldMap() ([]string, []interface{}) {
	return []string{
			"nspname", "nspacl",
		}, []interface{}{
			&e.Namespace, &e.AccessPrivileges,
		}
}

// TableName returns "students"
func (e *PostgresNamespace) TableName() string {
	return "pg_catalog.pg_namespace"
}

type PostgresNamespaces []*PostgresNamespace

func (u *PostgresNamespaces) Add() database.Entity {
	e := &PostgresNamespace{}
	*u = append(*u, e)

	return e
}

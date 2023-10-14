package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type PostgresUser struct {
	UserName     pgtype.Name
	UseCreateDB  pgtype.Bool
	UseSuper     pgtype.Bool
	UseRepl      pgtype.Bool
	UseByPassRLS pgtype.Bool
}

// FieldMap return a map of field name and pointer to field
func (e *PostgresUser) FieldMap() ([]string, []interface{}) {
	return []string{
			"usename", "usecreatedb", "usesuper", "userepl", "usebypassrls",
		}, []interface{}{
			&e.UserName, &e.UseCreateDB, &e.UseSuper, &e.UseRepl, &e.UseByPassRLS,
		}
}

// TableName returns "students"
func (e *PostgresUser) TableName() string {
	return "pg_catalog.pg_user"
}

type PostgresUsers []*PostgresUser

func (u *PostgresUsers) Add() database.Entity {
	e := &PostgresUser{}
	*u = append(*u, e)

	return e
}

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

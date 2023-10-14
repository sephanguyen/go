package database

import "github.com/jackc/pgtype"

type tablePolicy struct {
	Name                pgtype.Text      `json:"tablename"`
	PolicyName          pgtype.Text      `json:"policyname,omitempty"`
	Qual                pgtype.Text      `json:"qual,omitempty"`
	WithCheck           pgtype.Text      `json:"with_check,omitempty"`
	RelrowSecurity      pgtype.Bool      `json:"relrowsecurity,omitempty"`
	Relforcerowsecurity pgtype.Bool      `json:"relforcerowsecurity,omitempty"`
	Permissive          pgtype.Text      `json:"permissive,omitempty"`
	Roles               pgtype.TextArray `json:"roles,omitempty"`
}

// FieldMap returns fields' names and values of table entity.
func (p *tablePolicy) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"tablename", "policyname", "qual", "with_check", "relrowsecurity", "relforcerowsecurity", "permissive", "roles"}
	values = []interface{}{&p.Name, &p.PolicyName, &p.Qual, &p.WithCheck, &p.RelrowSecurity, &p.Relforcerowsecurity, &p.Permissive, &p.Roles}
	return
}

func (p *tablePolicy) TableName() string {
	return ""
}

type tablePolicies []*tablePolicy

func (p *tablePolicies) Add() Entity {
	e := &tablePolicy{}
	*p = append(*p, e)
	return e
}

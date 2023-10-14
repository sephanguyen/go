package repo

import "github.com/jackc/pgtype"

type Class struct {
	ClassID pgtype.Text
	Name    pgtype.Text
}

func (c *Class) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"class_id", "name"}
	values = []interface{}{&c.ClassID, &c.Name}
	return
}

func (c *Class) TableName() string {
	return "class"
}

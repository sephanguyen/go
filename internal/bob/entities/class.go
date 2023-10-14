package entities

import (
	"math/rand"

	"github.com/jackc/pgtype"
	"github.com/manabie-com/backend/internal/golibs/database"
)

const (
	ClassStatusActive   = "CLASS_STATUS_ACTIVE"
	ClassStatusInactive = "CLASS_STATUS_INACTIVE"
)

type Class struct {
	ID            pgtype.Int4 `sql:"class_id,pk"`
	Code          pgtype.Text `sql:"class_code"`
	SchoolID      pgtype.Int4 `sql:"school_id"`
	Avatar        pgtype.Text
	Name          pgtype.Text
	Subjects      pgtype.TextArray
	Grades        pgtype.Int4Array
	PlanID        pgtype.Text `sql:"plan_id"`
	Country       pgtype.Text
	PlanExpiredAt pgtype.Timestamptz
	PlanDuration  pgtype.Int2
	Status        pgtype.Text
	CreatedAt     pgtype.Timestamptz
	UpdatedAt     pgtype.Timestamptz
	DeletedAt     pgtype.Timestamptz
}

func (t *Class) FieldMap() ([]string, []interface{}) {
	return []string{
			"class_id", "class_code", "school_id", "avatar", "name", "subjects", "grades", "plan_id", "country", "plan_expired_at", "plan_duration", "status", "updated_at", "created_at", "deleted_at",
		}, []interface{}{
			&t.ID, &t.Code, &t.SchoolID, &t.Avatar, &t.Name, &t.Subjects, &t.Grades, &t.PlanID, &t.Country, &t.PlanExpiredAt, &t.PlanDuration, &t.Status, &t.UpdatedAt, &t.CreatedAt, &t.DeletedAt,
		}
}

func (t *Class) TableName() string {
	return "classes"
}

const codeLetters = "ABCDEFGHJKLMNPQRSTUVWXYZ2345689"

func GenerateClassCode(n int) string {
	b := make([]byte, n)
	total := len(codeLetters)
	for i := range b {
		b[i] = codeLetters[rand.Intn(total)]
	}

	return string(b)
}

type Classes []*Class

func (u *Classes) Add() database.Entity {
	e := &Class{}
	*u = append(*u, e)

	return e
}

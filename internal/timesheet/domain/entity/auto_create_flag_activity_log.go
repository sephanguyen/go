package entity

import (
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type AutoCreateFlagActivityLog struct {
	ID         pgtype.Text
	StaffID    pgtype.Text
	ChangeTime pgtype.Timestamptz
	FlagOn     pgtype.Bool
	CreatedAt  pgtype.Timestamptz
	UpdatedAt  pgtype.Timestamptz
	DeletedAt  pgtype.Timestamptz
}

func (a *AutoCreateFlagActivityLog) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"id",
		"staff_id",
		"change_time",
		"flag_on",
		"created_at",
		"updated_at",
		"deleted_at",
	}
	values = []interface{}{
		&a.ID,
		&a.StaffID,
		&a.ChangeTime,
		&a.FlagOn,
		&a.CreatedAt,
		&a.UpdatedAt,
		&a.DeletedAt,
	}
	return
}

func (*AutoCreateFlagActivityLog) TableName() string {
	return "auto_create_flag_activity_log"
}

func (a *AutoCreateFlagActivityLog) PrimaryKey() string {
	return "id"
}

func (a *AutoCreateFlagActivityLog) PreInsert() error {
	now := time.Now().In(timeutil.Timezone(pb.COUNTRY_JP))
	dateNow := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return multierr.Combine(
		a.CreatedAt.Set(now),
		a.UpdatedAt.Set(now),
		a.DeletedAt.Set(nil),
		a.ChangeTime.Set(dateNow),
	)
}

func (a *AutoCreateFlagActivityLog) PreUpdate() error {
	now := time.Now()
	return multierr.Combine(
		a.UpdatedAt.Set(now),
		a.DeletedAt.Set(nil),
	)
}

type AutoCreateFlagActivityLogs []*AutoCreateFlagActivityLog

func (la *AutoCreateFlagActivityLogs) Add() database.Entity {
	e := &AutoCreateFlagActivityLog{}
	*la = append(*la, e)

	return e
}

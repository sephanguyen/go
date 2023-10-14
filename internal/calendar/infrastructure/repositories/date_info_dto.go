package repositories

import (
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/calendar/domain/dto"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type DateInfo struct {
	Date        pgtype.Date
	LocationID  pgtype.Text
	DateTypeID  pgtype.Text
	OpeningTime pgtype.Text
	Status      pgtype.Text
	CreatedAt   pgtype.Timestamptz
	UpdatedAt   pgtype.Timestamptz
	DeletedAt   pgtype.Timestamptz
	TimeZone    pgtype.Text
}

func NewDateInfo(values map[string]interface{}) (*DateInfo, error) {
	dateInfo := &DateInfo{}
	database.AllNullEntity(dateInfo)
	var err error

	if date, ok := values["date"]; ok {
		err = multierr.Append(err, dateInfo.Date.Set(date))
	}
	if locationID, ok := values["location_id"]; ok {
		err = multierr.Append(err, dateInfo.LocationID.Set(locationID))
	}
	if dateTypeID, ok := values["day_type_id"]; ok {
		err = multierr.Append(err, dateInfo.DateTypeID.Set(dateTypeID))
	}
	if openingTime, ok := values["opening_time"]; ok {
		err = multierr.Append(err, dateInfo.OpeningTime.Set(openingTime))
	}
	if status, ok := values["status"]; ok {
		err = multierr.Append(err, dateInfo.Status.Set(status))
	}
	if createdAt, ok := values["created_at"]; ok {
		err = multierr.Append(err, dateInfo.CreatedAt.Set(createdAt))
	}
	if updatedAt, ok := values["updated_at"]; ok {
		err = multierr.Append(err, dateInfo.UpdatedAt.Set(updatedAt))
	}
	if timezone, ok := values["time_zone"]; ok {
		err = multierr.Append(err, dateInfo.TimeZone.Set(timezone))
	}

	return dateInfo, err
}

func (d *DateInfo) TableName() string {
	return "day_info"
}

func (d *DateInfo) FieldMap() ([]string, []interface{}) {
	return []string{
			"date",
			"location_id",
			"day_type_id",
			"opening_time",
			"status",
			"created_at",
			"updated_at",
			"deleted_at",
			"time_zone",
		}, []interface{}{
			&d.Date,
			&d.LocationID,
			&d.DateTypeID,
			&d.OpeningTime,
			&d.Status,
			&d.CreatedAt,
			&d.UpdatedAt,
			&d.DeletedAt,
			&d.TimeZone,
		}
}

func (d *DateInfo) ExportFieldMap() (fields []string, values []interface{}) {
	return []string{
			"date",
			"location_id",
			"day_type_id",
			"opening_time",
			"time_zone",
			"status",
		}, []interface{}{
			&d.Date,
			&d.LocationID,
			&d.DateTypeID,
			&d.OpeningTime,
			&d.TimeZone,
			&d.Status,
		}
}

func (d *DateInfo) ConvertToDTO() *dto.DateInfo {
	return &dto.DateInfo{
		Date:        d.Date.Time,
		LocationID:  d.LocationID.String,
		DateTypeID:  d.DateTypeID.String,
		OpeningTime: d.OpeningTime.String,
		Status:      d.Status.String,
		CreatedAt:   d.CreatedAt.Time,
		UpdatedAt:   d.UpdatedAt.Time,
		DeletedAt:   &d.DeletedAt.Time,
		TimeZone:    d.TimeZone.String,
	}
}

func (d *DateInfo) PreUpsert() error {
	now := time.Now()

	if err := multierr.Combine(
		d.CreatedAt.Set(now),
		d.UpdatedAt.Set(now),
		d.DeletedAt.Set(nil),
	); err != nil {
		return err
	}

	// Optional field: set Date Type ID as null if empty
	if len(strings.TrimSpace(d.DateTypeID.String)) == 0 {
		if err := d.DateTypeID.Set(nil); err != nil {
			return err
		}
	}

	// Optional field: set Open Time as null if empty
	if len(strings.TrimSpace(d.OpeningTime.String)) == 0 {
		if err := d.OpeningTime.Set(nil); err != nil {
			return err
		}
	}
	return nil
}

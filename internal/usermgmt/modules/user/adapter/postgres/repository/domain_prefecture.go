package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"

	"github.com/pkg/errors"
)

type DomainPrefectureRepo struct{}

type PrefectureAttribute struct {
	ID      field.String
	Name    field.String
	Code    field.String
	Country field.String
}

type Prefecture struct {
	PrefectureAttribute

	CreatedAt field.Time
	UpdatedAt field.Time
	DeletedAt field.Time
}

func NewPrefecture(p entity.DomainPrefecture) *Prefecture {
	now := field.NewTime(time.Now())
	return &Prefecture{
		PrefectureAttribute: PrefectureAttribute{
			ID:      p.PrefectureID(),
			Code:    p.PrefectureCode(),
			Name:    p.Name(),
			Country: p.Country(),
		},
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: field.NewNullTime(),
	}
}

func (p *Prefecture) PrefectureID() field.String {
	return p.PrefectureAttribute.ID
}
func (p *Prefecture) PrefectureCode() field.String {
	return p.PrefectureAttribute.Code
}
func (p *Prefecture) Name() field.String {
	return p.PrefectureAttribute.Name
}
func (p *Prefecture) Country() field.String {
	return p.PrefectureAttribute.Country
}

func (*Prefecture) TableName() string {
	return "prefecture"
}

func (p *Prefecture) FieldMap() ([]string, []interface{}) {
	return []string{
			"prefecture_id",
			"prefecture_code",
			"country",
			"name",
			"created_at",
			"updated_at",
			"deleted_at",
		}, []interface{}{
			&p.PrefectureAttribute.ID,
			&p.PrefectureAttribute.Code,
			&p.PrefectureAttribute.Country,
			&p.PrefectureAttribute.Name,
			&p.CreatedAt,
			&p.UpdatedAt,
			&p.DeletedAt,
		}
}

func (r *DomainPrefectureRepo) GetByPrefectureCodes(ctx context.Context, db database.QueryExecer, prefectureCodes []string) (entity.DomainPrefectures, error) {
	ctx, span := interceptors.StartSpan(ctx, "DomainPrefectureRepo.GetByPrefectureCodes")
	defer span.End()

	stmt := `SELECT %s FROM %s WHERE prefecture_code = ANY($1) AND deleted_at IS NULL`
	prefecture := NewPrefecture(entity.DefaultDomainPrefecture{})

	fieldNames, _ := prefecture.FieldMap()
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(fieldNames, ","),
		prefecture.TableName(),
	)

	rows, err := db.Query(
		ctx,
		stmt,
		database.TextArray(prefectureCodes),
	)
	if err != nil {
		return nil, InternalError{
			RawError: err,
		}
	}

	defer rows.Close()

	var result entity.DomainPrefectures
	for rows.Next() {
		item := NewPrefecture(entity.DefaultDomainPrefecture{})

		_, fieldValues := item.FieldMap()

		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, InternalError{
				RawError: fmt.Errorf("rows.Scan: %w", err),
			}
		}

		result = append(result, item)
	}
	return result, nil
}

func (r *DomainPrefectureRepo) GetByIDs(ctx context.Context, db database.QueryExecer, ids []string) ([]entity.DomainPrefecture, error) {
	ctx, span := interceptors.StartSpan(ctx, "DomainPrefectureRepo.GetByIDs")
	defer span.End()

	stmt := `SELECT %s FROM %s WHERE prefecture_id = ANY($1) AND deleted_at IS NULL`
	prefecture := NewPrefecture(entity.DefaultDomainPrefecture{})

	fieldNames, _ := prefecture.FieldMap()
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(fieldNames, ","),
		prefecture.TableName(),
	)

	rows, err := db.Query(
		ctx,
		stmt,
		database.TextArray(ids),
	)
	if err != nil {
		return nil, InternalError{RawError: errors.Wrap(err, "db.Query")}
	}

	defer rows.Close()

	var result []entity.DomainPrefecture
	for rows.Next() {
		item := NewPrefecture(entity.DefaultDomainPrefecture{})

		_, fieldValues := item.FieldMap()

		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, InternalError{RawError: errors.Wrap(err, "rows.Scan")}
		}

		result = append(result, item)
	}
	return result, nil
}

func (r *DomainPrefectureRepo) GetAll(ctx context.Context, db database.QueryExecer) ([]entity.DomainPrefecture, error) {
	ctx, span := interceptors.StartSpan(ctx, "DomainPrefectureRepo.GetAll")
	defer span.End()

	stmt := `SELECT %s FROM %s WHERE deleted_at IS NULL`
	prefecture := NewPrefecture(entity.DefaultDomainPrefecture{})

	fieldNames, _ := prefecture.FieldMap()
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(fieldNames, ","),
		prefecture.TableName(),
	)

	rows, err := db.Query(
		ctx,
		stmt,
	)
	if err != nil {
		return nil, InternalError{RawError: errors.Wrap(err, "db.Query")}
	}

	defer rows.Close()

	var result []entity.DomainPrefecture
	for rows.Next() {
		item := NewPrefecture(entity.DefaultDomainPrefecture{})

		_, fieldValues := item.FieldMap()

		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, InternalError{RawError: errors.Wrap(err, "rows.Scan")}
		}

		result = append(result, item)
	}
	return result, nil
}

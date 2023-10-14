package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/discount/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/constant"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
)

type UserDiscountTagRepo struct {
}

func (r *UserDiscountTagRepo) GetDiscountTagsByUserIDAndLocationID(
	ctx context.Context,
	db database.QueryExecer,
	userID string,
	locationID string,
) (
	userDiscountTags []*entities.UserDiscountTag,
	err error,
) {
	userDiscountTag := entities.UserDiscountTag{}
	userDiscountTagFieldNames, _ := userDiscountTag.FieldMap()
	stmt := `SELECT %s
		FROM 
			%s
		WHERE 
			user_id = $1
		AND
		(
			location_id = $2
		OR
			location_id IS NULL
		)
		AND
			deleted_at IS NULL`

	stmt = fmt.Sprintf(
		stmt,
		strings.Join(userDiscountTagFieldNames, ","),
		userDiscountTag.TableName(),
	)

	rows, err := db.Query(ctx, stmt, userID, locationID)
	if err != nil {
		return
	}

	defer rows.Close()

	userDiscountTags = []*entities.UserDiscountTag{}
	for rows.Next() {
		userDiscountTag := new(entities.UserDiscountTag)
		_, fieldValues := userDiscountTag.FieldMap()
		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf(constant.RowScanError, err)
		}
		userDiscountTags = append(userDiscountTags, userDiscountTag)
	}
	return userDiscountTags, nil
}

func (r *UserDiscountTagRepo) GetDiscountEligibilityOfStudentProduct(
	ctx context.Context,
	db database.QueryExecer,
	userID string,
	locationID string,
	productID string,
) (
	userDiscountTags []*entities.UserDiscountTag,
	err error,
) {
	userDiscountTag := entities.UserDiscountTag{}
	userDiscountTagFieldNames, _ := userDiscountTag.FieldMap()
	stmt := `SELECT %s
		FROM 
			%s
		WHERE 
			user_id = $1
		AND
			(location_id = $2 OR location_id IS NULL)
		AND
			(product_id = $3 OR product_id IS NULL)
		AND
			deleted_at IS NULL
		AND (
			start_date IS NULL
			OR
			(start_date <= NOW() AND end_date >= NOW())
		)`

	stmt = fmt.Sprintf(
		stmt,
		strings.Join(userDiscountTagFieldNames, ","),
		userDiscountTag.TableName(),
	)

	rows, err := db.Query(ctx, stmt, userID, locationID, productID)
	if err != nil {
		return
	}

	defer rows.Close()

	userDiscountTags = []*entities.UserDiscountTag{}
	for rows.Next() {
		userDiscountTag := new(entities.UserDiscountTag)
		_, fieldValues := userDiscountTag.FieldMap()
		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf(constant.RowScanError, err)
		}
		userDiscountTags = append(userDiscountTags, userDiscountTag)
	}
	return userDiscountTags, nil
}

func (r *UserDiscountTagRepo) GetDiscountTagsWithActivityOnDate(
	ctx context.Context,
	db database.QueryExecer,
	timestamp time.Time,
) (
	userDiscountTags []*entities.UserDiscountTag,
	err error,
) {
	userDiscountTag := entities.UserDiscountTag{}
	userDiscountTagFieldNames, _ := userDiscountTag.FieldMap()
	stmt := `SELECT %s
		FROM 
			%s
		WHERE
			(
				start_date IS NULL AND
				(created_at >= $1 AND created_at <= $2) OR
				(updated_at >= $1 AND updated_at <= $2)
			)
		OR
			(
				(start_date >= $1 AND start_date <= $2) OR
				(end_date >= $1 AND end_date <= $2)
			)
		OR
			(deleted_at >= $1 AND deleted_at <= $2)`

	stmt = fmt.Sprintf(
		stmt,
		strings.Join(userDiscountTagFieldNames, ","),
		userDiscountTag.TableName(),
	)

	rows, err := db.Query(ctx, stmt, timestamp.AddDate(0, 0, -1), timestamp)
	if err != nil {
		return
	}

	defer rows.Close()

	userDiscountTags = []*entities.UserDiscountTag{}
	for rows.Next() {
		userDiscountTag := new(entities.UserDiscountTag)
		_, fieldValues := userDiscountTag.FieldMap()
		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf(constant.RowScanError, err)
		}
		userDiscountTags = append(userDiscountTags, userDiscountTag)
	}
	return userDiscountTags, nil
}

func (r *UserDiscountTagRepo) GetUserIDsWithActivityOnDate(
	ctx context.Context,
	db database.QueryExecer,
	timestamp time.Time,
) (
	userIDs []pgtype.Text,
	err error,
) {
	userDiscountTag := entities.UserDiscountTag{}
	stmt := `SELECT DISTINCT user_id
	FROM 
		%s
	WHERE
		(
			start_date IS NULL AND
			(created_at >= $1 AND created_at <= $2) OR
			(updated_at >= $1 AND updated_at <= $2)
		)
	OR
		(
			(start_date >= $1 AND start_date <= $2) OR
			(end_date >= $1 AND end_date <= $2)
		)
	OR
		(deleted_at >= $1 AND deleted_at <= $2)`

	stmt = fmt.Sprintf(
		stmt,
		userDiscountTag.TableName(),
	)

	rows, err := db.Query(ctx, stmt, timestamp.AddDate(0, 0, -1), timestamp)
	if err != nil {
		return
	}

	defer rows.Close()

	userIDs = []pgtype.Text{}
	for rows.Next() {
		var userID pgtype.Text
		err := rows.Scan(&userID)
		if err != nil {
			return nil, fmt.Errorf(constant.RowScanError, err)
		}
		userIDs = append(userIDs, userID)
	}
	return userIDs, nil
}

func (r *UserDiscountTagRepo) Create(ctx context.Context, db database.QueryExecer, e *entities.UserDiscountTag) error {
	ctx, span := interceptors.StartSpan(ctx, "UserDiscountTag.Create")
	defer span.End()

	now := time.Now()
	if err := multierr.Combine(
		e.CreatedAt.Set(now),
		e.UpdatedAt.Set(now),
		e.DeletedAt.Set(nil),
	); err != nil {
		return fmt.Errorf("multierr.Combine UpdatedAt.Set CreatedAt.Set DeletedAt.Set: %w", err)
	}

	cmdTag, err := database.InsertExcept(ctx, e, []string{"resource_path"}, db.Exec)
	if err != nil {
		return fmt.Errorf("err insert UserDiscountTag: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err insert UserDiscountTag: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}

func (r *UserDiscountTagRepo) GetActiveDiscountTagIDsByDateAndUserID(
	ctx context.Context,
	db database.QueryExecer,
	timestamp time.Time,
	userID string,
) (
	userDiscountTagIDs []string,
	err error,
) {
	userDiscountTag := entities.UserDiscountTag{}
	stmt := `SELECT DISTINCT discount_tag_id
		FROM 
			%s
		WHERE
			user_id = $1
		AND
			deleted_at IS NULL
		AND (
			(start_date IS NULL AND created_at <= $2)
			OR
			(start_date <= $2 AND end_date > $2)
		)`

	stmt = fmt.Sprintf(
		stmt,
		userDiscountTag.TableName(),
	)

	rows, err := db.Query(ctx, stmt, userID, timestamp)
	userDiscountTagIDs = []string{}

	if err != nil {
		if err == pgx.ErrNoRows {
			return userDiscountTagIDs, nil
		}
		return
	}

	defer rows.Close()

	for rows.Next() {
		var discountTagID string

		err := rows.Scan(
			&discountTagID,
		)
		if err != nil {
			return userDiscountTagIDs, fmt.Errorf("row.Scan: %w", err)
		}

		userDiscountTagIDs = append(userDiscountTagIDs, discountTagID)
	}
	return userDiscountTagIDs, nil
}

func (r *UserDiscountTagRepo) SoftDeleteByTypesAndUserID(ctx context.Context, db database.QueryExecer, userID string, discountTypes pgtype.TextArray) error {
	ctx, span := interceptors.StartSpan(ctx, "UserDiscountTagRepo.SoftDeleteByTypesAndUserID")
	defer span.End()

	e := &entities.UserDiscountTag{}

	now := time.Now()

	stmt := fmt.Sprintf(`
		UPDATE
			%s
		SET
			deleted_at = $1,
			updated_at = $2
		WHERE
			discount_type = ANY($3::_TEXT)
		AND
			user_id = $4 
		AND
			deleted_at IS NULL
		AND
			EXISTS (SELECT 1 FROM %s WHERE discount_type = ANY($3::_TEXT) AND user_id = $4 AND deleted_at IS NULL);`,
		e.TableName(), e.TableName())

	cmdTag, err := db.Exec(ctx, stmt, now, now, discountTypes, userID)
	if err != nil {
		return fmt.Errorf("err delete UserDiscountTagRepo: %w", err)
	}

	if cmdTag.RowsAffected() != int64(len(discountTypes.Elements)) {
		return fmt.Errorf("err delete UserDiscountTagRepo: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}

func (r *UserDiscountTagRepo) GetDiscountTagsByUserID(
	ctx context.Context,
	db database.QueryExecer,
	userID string,
) (
	userDiscountTags []*entities.UserDiscountTag,
	err error,
) {
	userDiscountTag := entities.UserDiscountTag{}
	userDiscountTagFieldNames, _ := userDiscountTag.FieldMap()
	stmt := `SELECT %s
		FROM 
			%s
		WHERE 
			user_id = $1
		AND
			deleted_at IS NULL`

	stmt = fmt.Sprintf(
		stmt,
		strings.Join(userDiscountTagFieldNames, ","),
		userDiscountTag.TableName(),
	)

	rows, err := db.Query(ctx, stmt, userID)
	if err != nil {
		return
	}

	defer rows.Close()

	userDiscountTags = []*entities.UserDiscountTag{}
	for rows.Next() {
		userDiscountTag := new(entities.UserDiscountTag)
		_, fieldValues := userDiscountTag.FieldMap()
		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf(constant.RowScanError, err)
		}
		userDiscountTags = append(userDiscountTags, userDiscountTag)
	}
	return userDiscountTags, nil
}

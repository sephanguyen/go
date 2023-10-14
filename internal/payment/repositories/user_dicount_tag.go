package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
)

type UserDiscountTagRepo struct{}

func (r *UserDiscountTagRepo) GetDiscountTagByUserIDAndDiscountTagID(
	ctx context.Context,
	db database.QueryExecer,
	userID string,
	discountTagID string,
) (
	userDiscountTags []*entities.UserDiscountTag,
	err error,
) {
	userDiscountTag := entities.UserDiscountTag{}
	userDiscountTagFieldNames, _ := userDiscountTag.FieldMap()
	now := time.Now()

	stmt := `SELECT %s
			FROM 
				%s
			WHERE
				user_id = $1
			AND
				discount_tag_id = $2
			AND
				start_date <= $3
			AND
				(end_date IS NULL
			OR
				end_date > $3)
			AND
				deleted_at IS NULL`

	stmt = fmt.Sprintf(
		stmt,
		strings.Join(userDiscountTagFieldNames, ","),
		userDiscountTag.TableName(),
	)

	rows, err := db.Query(ctx, stmt, userID, discountTagID, now)
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

	return
}

func (r *UserDiscountTagRepo) GetAvailableDiscountTagIDsByUserID(ctx context.Context, db database.QueryExecer, userID string) (usersDiscountTagIDs []string, err error) {
	// Define the SQL statement with placeholders
	stmt := `SELECT DISTINCT discount_tag_id
        FROM %s
        WHERE user_id = $1
        AND deleted_at IS NULL
        AND (
            (start_date IS NULL AND created_at <= NOW())
            OR
            (start_date <= NOW() AND end_date > NOW())
        )`

	userDiscountTag := entities.UserDiscountTag{}
	tableName := userDiscountTag.TableName()

	stmt = fmt.Sprintf(stmt, tableName)

	// Execute the query and retrieve the rows
	rows, err := db.Query(ctx, stmt, userID)
	if err != nil {
		return nil, fmt.Errorf("error while querying the database: %v", err)
	}
	defer rows.Close()

	// Collect the results
	for rows.Next() {
		var userDiscountTagID string
		if err := rows.Scan(&userDiscountTagID); err != nil {
			return nil, fmt.Errorf("error while scanning rows: %v", err)
		}
		usersDiscountTagIDs = append(usersDiscountTagIDs, userDiscountTagID)
	}
	return usersDiscountTagIDs, nil
}

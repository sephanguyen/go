package repositories

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
)

type ClassMemberRepo struct{}

func (r *ClassMemberRepo) UpdateStatus(ctx context.Context, db database.QueryExecer, userIDs []string, classIDs []int32, userGroup, status string) error {
	query := "UPDATE class_members SET status = $1 WHERE user_id = ANY($2) AND class_id = ANY($3) "
	args := []interface{}{status, database.TextArray(userIDs), database.Int4Array(classIDs)}

	if userGroup != "" {
		query += "AND user_group = ? "
		args = append(args, userGroup)
	}

	_, err := db.Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}

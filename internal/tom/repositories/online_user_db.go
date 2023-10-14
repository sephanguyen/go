package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	entities "github.com/manabie-com/backend/internal/tom/domain/core"

	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
)

type OnlineUserDBRepo struct {
}

func (r *OnlineUserDBRepo) Find(ctx context.Context, db database.QueryExecer, userIDs pgtype.TextArray, since pgtype.Timestamptz) (mapNodeUserIDs map[pgtype.Text][]string, err error) {
	ctx, span := interceptors.StartSpan(ctx, "OnlineUserDBRepo.Find")
	defer span.End()

	selectStmt := `
SELECT DISTINCT user_id, node_name
FROM online_users
WHERE user_id = ANY($1) AND last_active_at >= $2`

	rows, err := db.Query(ctx, selectStmt, &userIDs, &since)
	if err != nil {
		return nil, errors.Wrap(err, "r.DB.QueryEx(")
	}
	defer rows.Close()

	mapNodeUserIDs = make(map[pgtype.Text][]string)
	for rows.Next() {
		var (
			userID, nodeName pgtype.Text
		)
		if err := rows.Scan(&userID, &nodeName); err != nil {
			return nil, errors.Wrap(err, "row.Scan")
		}

		mapNodeUserIDs[nodeName] = append(mapNodeUserIDs[nodeName], userID.String)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "row.Err")
	}

	return mapNodeUserIDs, nil
}

func (r *OnlineUserDBRepo) Insert(ctx context.Context, db database.QueryExecer, e *entities.OnlineUser) error {
	ctx, span := interceptors.StartSpan(ctx, "OnlineUserDBRepo.Insert")
	defer span.End()

	now := time.Now()
	e.UpdatedAt.Set(now)
	e.CreatedAt.Set(now)

	cmdTag, err := Insert(ctx, e, db.Exec)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() != 1 {
		return errors.New("cannot insert new online_users")
	}

	return nil
}

func (r *OnlineUserDBRepo) SetActive(ctx context.Context, db database.QueryExecer, id pgtype.Text) error {
	ctx, span := interceptors.StartSpan(ctx, "OnlineUserDBRepo.SetActive")
	defer span.End()

	sql := "UPDATE online_users SET updated_at = NOW(), last_active_at = NOW() WHERE online_user_id = $1"
	cmdTag, err := db.Exec(ctx, sql, &id)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("not found online_users to update")
	}

	return nil
}

func (r *OnlineUserDBRepo) Delete(ctx context.Context, db database.QueryExecer, id pgtype.Text) error {
	ctx, span := interceptors.StartSpan(ctx, "OnlineUserDBRepo.Delete")
	defer span.End()

	sql := "DELETE FROM online_users WHERE online_user_id = $1"
	_, err := db.Exec(ctx, sql, &id)
	if err != nil {
		return err
	}

	return nil
}

func (r *OnlineUserDBRepo) DeleteByNode(ctx context.Context, db database.QueryExecer, node pgtype.Text) error {
	ctx, span := interceptors.StartSpan(ctx, "OnlineUserDBRepo.DeleteByNode")
	defer span.End()

	sql := "DELETE FROM online_users WHERE node_name = $1"
	_, err := db.Exec(ctx, sql, &node)
	if err != nil {
		return err
	}

	return nil
}

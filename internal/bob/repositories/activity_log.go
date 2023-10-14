package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type ActivityLogRepo struct{}

func (rcv *ActivityLogRepo) Create(ctx context.Context, db database.QueryExecer, e *entities_bob.ActivityLog) error {
	ctx, span := interceptors.StartSpan(ctx, "ActivityLogRepo.Create")
	defer span.End()

	now := time.Now()
	_ = e.UpdatedAt.Set(now)
	_ = e.CreatedAt.Set(now)
	_ = e.DeletedAt.Set(nil)

	cmdTag, err := database.Insert(ctx, e, db.Exec)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() != 1 {
		return errors.New("cannot insert new ActivityLogRepo")
	}

	return nil
}

func (rcv *ActivityLogRepo) RetrieveLastCheckPromotionCode(ctx context.Context, tx database.QueryExecer, studentID pgtype.Text) (string, error) {
	al := &entities_bob.ActivityLog{}
	alFields := database.GetFieldNames(al)

	query := fmt.Sprintf(`SELECT %s FROM activity_logs WHERE user_id = $1 AND action_type IN ('use_promotion_code', 'use_activation_code')
	ORDER BY created_at DESC LIMIT 1`, strings.Join(alFields, ","))
	if err := tx.QueryRow(ctx, query, &studentID).Scan(database.GetScanFields(al, alFields)...); err != nil {
		if err == pgx.ErrNoRows {
			return "", ErrNoRows
		}
		return "", errors.Wrap(err, "tx.QueryRowEx.Scan")
	}

	var payload struct {
		PromotionCode  string `json:"promotion_code"`
		ActivationCode string `json:"activation_code"`
	}
	if err := al.Payload.AssignTo(&payload); err != nil {
		return "", errors.Wrap(err, "al.Payload.AssignTo")
	}
	if payload.PromotionCode != "" {
		return payload.PromotionCode, nil
	} else if payload.ActivationCode != "" {
		return payload.ActivationCode, nil
	} else {
		return "", nil
	}
}

func (rcv *ActivityLogRepo) BulkImport(ctx context.Context, db database.QueryExecer, logs []*entities_bob.ActivityLog) error {
	queueFn := func(b *pgx.Batch, e *entities_bob.ActivityLog) {
		fieldNames := []string{"activity_log_id", "user_id", "action_type", "payload", "created_at", "updated_at"}
		placeHolders := "$1, $2, $3, $4, $5, $6"

		query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
			e.TableName(),
			strings.Join(fieldNames, ","),
			placeHolders,
		)
		b.Queue(query, database.GetScanFields(e, fieldNames)...)
	}

	b := &pgx.Batch{}
	var d pgtype.Timestamptz
	err := d.Set(time.Now())
	if err != nil {
		return fmt.Errorf("cannot set time for activity logs: %w", err)
	}

	for _, each := range logs {
		err = each.ID.Set(idutil.ULIDNow())
		if err != nil {
			return fmt.Errorf("cannot set id for activity logs: %w", err)
		}
		each.CreatedAt = d
		each.UpdatedAt = d
		queueFn(b, each)
	}

	result := db.SendBatch(ctx, b)
	defer result.Close()

	for i := 0; i < b.Len(); i++ {
		_, err := result.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
	}

	return nil
}

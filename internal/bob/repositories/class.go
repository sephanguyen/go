package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type ClassRepo struct{}

func (rcv *ClassRepo) Create(ctx context.Context, db database.QueryExecer, e *entities_bob.Class) error {
	ctx, span := interceptors.StartSpan(ctx, "ClassRepo.Create")
	defer span.End()

	now := time.Now()
	_ = e.UpdatedAt.Set(now)
	_ = e.CreatedAt.Set(now)

	_, err := database.Insert(ctx, e, db.Exec)
	if err != nil {
		return fmt.Errorf("database.Insert: %w", err)
	}

	return nil
}

func (rcv *ClassRepo) GetNextID(ctx context.Context, db database.QueryExecer) (*pgtype.Int4, error) {
	var id pgtype.Int4
	sql := `SELECT class_id FROM classes ORDER BY class_id DESC`
	err := db.QueryRow(ctx, sql).Scan(&id)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("db.QueryRow: %w", err)
	}

	if err == pgx.ErrNoRows {
		id = database.Int4(0)
	}

	id.Int++

	return &id, nil
}

func (rcv *ClassRepo) Update(ctx context.Context, db database.QueryExecer, e *entities_bob.Class) error {
	ctx, span := interceptors.StartSpan(ctx, "ClassRepo.Update")
	defer span.End()

	now := time.Now()
	_ = e.UpdatedAt.Set(now)

	cmdTag, err := database.Update(ctx, e, db.Exec, "class_id")
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() != 1 {
		return errors.New("cannot update Class")
	}

	return nil
}

func (rcv *ClassRepo) FindByID(ctx context.Context, db database.QueryExecer, id pgtype.Int4) (*entities_bob.Class, error) {
	ctx, span := interceptors.StartSpan(ctx, "ClassRepo.Find")
	defer span.End()

	e := new(entities_bob.Class)
	fields := database.GetFieldNames(e)
	selectStmt := fmt.Sprintf("SELECT %s FROM %s WHERE class_id = $1", strings.Join(fields, ","), e.TableName())

	row := db.QueryRow(ctx, selectStmt, &id)

	if err := row.Scan(database.GetScanFields(e, fields)...); err != nil {
		return nil, err
	}

	return e, nil
}

func (rcv *ClassRepo) FindByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.Int4Array) ([]*entities_bob.Class, error) {
	ctx, span := interceptors.StartSpan(ctx, "ClassRepo.Find")
	defer span.End()

	classes := entities_bob.Classes{}
	e := new(entities_bob.Class)
	fields := database.GetFieldNames(e)
	selectStmt := fmt.Sprintf("SELECT %s FROM %s WHERE class_id = ANY($1) ORDER BY name", strings.Join(fields, ","), e.TableName())
	err := database.Select(ctx, db, selectStmt, &ids).ScanAll(&classes)
	if err != nil {
		return nil, err
	}
	return classes, nil
}

func (rcv *ClassRepo) FindJoined(ctx context.Context, db database.QueryExecer, userID pgtype.Text) ([]*entities_bob.Class, error) {
	ctx, span := interceptors.StartSpan(ctx, "ClassRepo.FindJoined")
	defer span.End()

	e := &entities_bob.Class{}
	fields := database.GetFieldNames(e)
	query := fmt.Sprintf("SELECT c.%s "+
		"FROM classes AS c JOIN class_members AS m ON c.class_id = m.class_id "+
		"AND m.user_id = $1 AND c.status = '%s' AND m.status = '%s'",
		strings.Join(fields, ", c."), entities_bob.ClassStatusActive, entities_bob.ClassMemberStatusActive)
	rows, err := db.Query(ctx, query, &userID)
	if err != nil {
		return nil, errors.Wrap(err, "r.DB.QueryEx")
	}
	defer rows.Close()

	var result []*entities_bob.Class
	for rows.Next() {
		e := &entities_bob.Class{}
		_, values := e.FieldMap()
		if err := rows.Scan(values...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		result = append(result, e)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return result, nil
}

func (rcv *ClassRepo) UpdateClassCode(ctx context.Context, db database.QueryExecer, classID pgtype.Int4, code pgtype.Text) error {
	ctx, span := interceptors.StartSpan(ctx, "ClassRepo.UpdateClassCode")
	defer span.End()

	stmt := "UPDATE classes SET class_code = $1, updated_at = now() WHERE class_id = $2"
	cmdTag, err := db.Exec(ctx, stmt, &code, &classID)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() != 1 {
		return errors.New("cannot update class code")
	}

	return nil
}

func (rcv *ClassRepo) FindByCode(ctx context.Context, db database.QueryExecer, code pgtype.Text) (*entities_bob.Class, error) {
	ctx, span := interceptors.StartSpan(ctx, "ClassRepo.FindByCode")
	defer span.End()

	e := new(entities_bob.Class)
	fields := database.GetFieldNames(e)
	selectStmt := fmt.Sprintf("SELECT %s FROM %s WHERE class_code = $1", strings.Join(fields, ","), e.TableName())

	row := db.QueryRow(ctx, selectStmt, &code)

	if err := row.Scan(database.GetScanFields(e, fields)...); err != nil {
		return nil, err
	}

	return e, nil
}

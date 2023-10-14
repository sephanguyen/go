package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
)

type RowScanner struct {
	err         error
	pgxRows     pgx.Rows
	skipRowNext bool
}

func (r *RowScanner) GetFieldNames() []string {
	fields := make([]string, 0, len(r.pgxRows.FieldDescriptions()))

	for _, f := range r.pgxRows.FieldDescriptions() {
		fields = append(fields, string(f.Name))
	}

	return fields
}

func (r *RowScanner) ScanFields(dst ...interface{}) error {
	if r.err != nil {
		return r.err
	}
	defer r.pgxRows.Close()

	if !r.skipRowNext && !r.pgxRows.Next() {
		err := r.pgxRows.Err()
		if err == nil {
			return pgx.ErrNoRows
		}

		return fmt.Errorf("rows.Err: %w", err)
	}

	if err := r.pgxRows.Scan(dst...); err != nil {
		return fmt.Errorf("rows.Scan: %w", err)
	}

	if err := r.pgxRows.Err(); err != nil {
		return fmt.Errorf("rows.Err: %w", err)
	}

	return nil
}

func (r *RowScanner) ScanOne(dst Entity) error {
	if r.err != nil {
		return r.err
	}

	if !r.pgxRows.Next() {
		defer r.pgxRows.Close()
		err := r.pgxRows.Err()
		if err == nil {
			return pgx.ErrNoRows
		}

		return fmt.Errorf("rows.Err: %w", err)
	}

	r.skipRowNext = true
	// must call rows.Next() before, if not, rows description will return empty fields
	fields := r.GetFieldNames()
	args := GetScanFields(dst, fields)
	return r.ScanFields(args...)
}

func (r *RowScanner) ScanAll(dst Entities) error {
	if r.err != nil {
		return r.err
	}
	defer r.pgxRows.Close()

	var fields []string
	for r.pgxRows.Next() {
		if len(fields) == 0 {
			fields = r.GetFieldNames()
		}

		e := dst.Add()

		if err := r.pgxRows.Scan(GetScanFields(e, fields)...); err != nil {
			return fmt.Errorf("rows.Scan: %w", err)
		}
	}

	if err := r.pgxRows.Err(); err != nil {
		return fmt.Errorf("rows.Err: %w", err)
	}

	return nil
}

func Select(ctx context.Context, db queryer, sql string, args ...interface{}) *RowScanner {
	rows, err := db.Query(ctx, sql, args...)
	if err != nil {
		err = fmt.Errorf("err db.Query: %w", err)
	}

	return &RowScanner{
		err:     err,
		pgxRows: rows,
	}
}

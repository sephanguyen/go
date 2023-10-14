package multitenant

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/pkg/errors"
)

type PGClass struct {
	RelName             string
	RelRowSecurity      bool
	RelForceRowSecurity bool
}

func getPgClass(ctx context.Context, db database.QueryExecer) ([]*PGClass, error) {
	stmt :=
		`
		SELECT
			relname, relrowsecurity, relforcerowsecurity
		FROM
			pg_class
		WHERE
			relname IN (
				SELECT
					table_name
				FROM
					information_schema.tables
				WHERE
					table_schema = 'public'
					AND table_name != 'schema_migrations'
					AND table_type = 'BASE TABLE'
			)
		`

	rows, err := db.Query(ctx, stmt)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}

	defer rows.Close()

	pgClasses := make([]*PGClass, 0)
	for rows.Next() {
		pgClass := new(PGClass)

		err := rows.Scan(&pgClass.RelName, &pgClass.RelRowSecurity, &pgClass.RelForceRowSecurity)
		if err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}

		pgClasses = append(pgClasses, pgClass)
	}

	return pgClasses, nil
}

func (s *suite) scannerScansOnAllTables(ctx context.Context) (context.Context, error) {
	nCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	databases := []string{"bob", "tom", "eureka", "fatima"}

	for _, db := range databases {
		dbConn := s.dbConnForSchema(db)
		if dbConn == nil {
			return nCtx, fmt.Errorf(`failed to get db conn for "%s" database`, db)
		}

		pgClasses, err := getPgClass(ctx, dbConn)
		if err != nil {
			return nCtx, errors.Wrap(err, "getPgClass")
		}
		s.Value = pgClasses
	}
	return nCtx, nil
}

func (s *suite) thoseTablesMustHasRlsEnabled(ctx context.Context) (context.Context, error) {
	pgClasses := s.Value.([]*PGClass)

	for _, pgClass := range pgClasses {
		if !pgClass.RelRowSecurity {
			return ctx, fmt.Errorf(`expected table "%s" has rls enabled but actual is not`, pgClass.RelName)
		}
	}

	return ctx, nil
}

func (s *suite) thoseTablesMustHasRlsForced(ctx context.Context) (context.Context, error) {
	pgClasses := s.Value.([]*PGClass)

	for _, pgClass := range pgClasses {
		if !pgClass.RelForceRowSecurity {
			return ctx, fmt.Errorf(`expected table "%s" has rls forced but actual is not`, pgClass.RelName)
		}
	}

	return ctx, nil
}

package multitenant

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
)

func (s *suite) createSomeTableWithRandomName(ctx context.Context) (context.Context, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	s.StepState.Random = "table_" + strings.ToLower(idutil.ULIDNow())
	databases := []string{"bob", "tom", "eureka", "fatima"}

	script := fmt.Sprintf(`CREATE TABLE %s(column1 text, column2 text)`, s.StepState.Random)

	for _, db := range databases {
		dbConn := s.dbConnForSchema(db)
		if dbConn == nil {
			return ctx, fmt.Errorf(`failed to get db conn for "%s" database`, db)
		}

		_, err := dbConn.Exec(ctx, script)
		if err != nil {
			return ctx, fmt.Errorf(`error when create table "%s" in db "%s": %v`, s.StepState.Random, db, err)
		}
	}

	return ctx, nil

}

func (s *suite) thoseTablesMustHaveColumnResourcePath(ctx context.Context) (context.Context, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	databases := []string{"bob", "tom", "eureka", "fatima"}

	script := fmt.Sprintf(`SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema='public' AND table_name='%s' AND column_name='%s')`, s.StepState.Random, "resource_path")

	for _, db := range databases {
		existColumn := false
		dbConn := s.dbConnForSchema(db)
		if dbConn == nil {
			return ctx, fmt.Errorf(`failed to get db conn for "%s" database`, db)
		}

		err := dbConn.QueryRow(ctx, script).Scan(&existColumn)
		if err != nil {
			return ctx, fmt.Errorf(`error when check column is existed in table "%s" at db "%s": %v`, s.StepState.Random, db, err)
		}

		if existColumn == false {
			return ctx, fmt.Errorf(`column resource_path in table "%s" at db "%s" is not exists`, s.StepState.Random, db)
		}

	}

	return ctx, nil
}

func getOnePgClass(ctx context.Context, db database.QueryExecer, tableName string) (*PGClass, error) {
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
			AND relname = $1
		`

	var result = &PGClass{}
	err := db.QueryRow(ctx, stmt, tableName).Scan(&result.RelName, &result.RelRowSecurity, &result.RelForceRowSecurity)
	if err != nil {
		return nil, fmt.Errorf(`error when query rows: %v`, err)
	}

	return result, nil
}

func (s *suite) thoseTablesMustHaveRlsEnabledAndRlsForced(ctx context.Context) (context.Context, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	databases := []string{"bob", "tom", "eureka", "fatima"}

	for _, db := range databases {
		dbConn := s.dbConnForSchema(db)
		if dbConn == nil {
			return ctx, fmt.Errorf(`failed to get db conn for "%s" database`, db)
		}
		pgClass, err := getOnePgClass(ctx, dbConn, s.StepState.Random)
		if err != nil {
			return ctx, fmt.Errorf(`error when getOnePgClass in table "%s" at db "%s": %v`, s.StepState.Random, db, err)
		}
		if !pgClass.RelRowSecurity {
			return ctx, fmt.Errorf(`expected table "%s" in db "%s" has rls enabled but actual is not`, pgClass.RelName, db)
		}
		if !pgClass.RelForceRowSecurity {
			return ctx, fmt.Errorf(`expected table "%s" in db "%s" has rls forced but actual is not`, pgClass.RelName, db)
		}
	}

	return ctx, nil

}

package multitenant

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/yasuo/constant"

	"golang.org/x/sync/errgroup"
)

func (s *suite) aRandomNumber() {
	s.Random = strconv.Itoa(rand.Int())
}

func (s *suite) aRandomTableInDB(ctx context.Context) (context.Context, error) {
	s.aRandomNumber()
	nCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	s.currentTableName = fmt.Sprintf("random_%s", s.Random)
	stmt := fmt.Sprintf(`CREATE TABLE public.%s (
		id text NOT NULL,
		resource_path text NULL
	);`, s.currentTableName)
	_, err := s.bobDB.Exec(nCtx, stmt)
	return ctx, err
}

func (s *suite) recordWithDifferentResourcePath(ctx context.Context, recordNum int) (context.Context, error) {
	s.UserResourcePath = make(map[string]string)
	nCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	stmt := fmt.Sprintf("INSERT INTO %s VALUES ($1, $2)", s.currentTableName)
	for i := 0; i < recordNum; i++ {
		userID := idutil.ULIDNow()
		//TO-DO this is school_id for now. Need to change it to organization
		resourcePath := strconv.Itoa(i)
		s.UserResourcePath[userID] = resourcePath
		_, err := s.bobDB.Exec(nCtx, stmt, userID, resourcePath)
		if err != nil {
			return ctx, err
		}
	}

	return ctx, nil
}

func (s *suite) rlsIsEnableForTable(ctx context.Context) (context.Context, error) {
	nCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	stmt := fmt.Sprintf(`CREATE POLICY rls_pc ON "%s" using (permission_check(resource_path, '%s')) with check (permission_check(resource_path, '%s'));`,
		s.currentTableName, s.currentTableName, s.currentTableName)
	_, err := s.bobDB.Exec(nCtx, stmt)
	if err != nil {
		return ctx, err
	}

	stmt = fmt.Sprintf(`ALTER TABLE %s ENABLE ROW LEVEL security;`, s.currentTableName)
	_, err = s.bobDB.Exec(nCtx, stmt)
	if err != nil {
		return ctx, err
	}

	stmt = fmt.Sprintf(`ALTER TABLE %s FORCE ROW LEVEL security;`, s.currentTableName)
	_, err = s.bobDB.Exec(nCtx, stmt)
	if err != nil {
		return ctx, err
	}

	return ctx, nil
}

func (s *suite) userCanOnlyFetchTheirData(ctx context.Context) (context.Context, error) {
	fCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	rlsStmt := fmt.Sprintf(`SELECT count(*) FROM %s`, s.currentTableName)
	isolationStmt := fmt.Sprintf(`SELECT id, resource_path FROM %s`, s.currentTableName)

	eg, egCtx := errgroup.WithContext(fCtx)
	for userID, resourcePath := range s.UserResourcePath {
		id := userID
		rp := resourcePath
		eg.Go(func() error {
			claim := &interceptors.CustomClaims{
				Manabie: &interceptors.ManabieClaims{
					ResourcePath: rp,
					DefaultRole:  constant.UserGroupStudent,
				},
			}
			ctx, cancel := context.WithTimeout(egCtx, 8*time.Second)
			defer cancel()

			ctx = interceptors.ContextWithJWTClaims(ctx, claim)

			var count int
			if err := s.bobDB.QueryRow(ctx, rlsStmt).Scan(&count); err != nil {
				return fmt.Errorf("error query rlsStmt: %w", err)
			}
			if count != 1 {
				return fmt.Errorf("rls is not enable for table %s. Count total: %d line", s.currentTableName, count)
			}
			rows, err := s.bobDB.Query(ctx, isolationStmt)
			if err != nil {
				return err
			}
			defer rows.Close()
			for rows.Next() {
				var scanID, scanRP string
				if err := rows.Scan(&scanID, &scanRP); err != nil {
					return err
				}
				if rp != scanRP {
					return fmt.Errorf("connection does not have isolation: expect resource_path: %s got %s", rp, scanRP)
				}
				if id != scanID {
					return fmt.Errorf("connection does not have isolation: expect id: %s got %s from table %s", id, scanID, s.currentTableName)
				}
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return ctx, err
	}

	return ctx, nil
}

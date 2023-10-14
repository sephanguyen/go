package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/draft/entities"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgx/v4"
)

// DraftRepo struct
type DraftRepo struct {
}

// GetTargetCoverage return coverage of target branch
func (d *DraftRepo) GetTargetCoverage(ctx context.Context, db database.QueryExecer, repository, branchName string, integration bool) (*entities.TargetCoverage, error) {
	targetCoverage := entities.TargetCoverage{}
	fieldNames, values := targetCoverage.FieldMap()
	querystm := fmt.Sprintf("SELECT %s FROM public.target_coverage WHERE repository=$1 AND branch_name=$2 AND integration=$3 LIMIT 1", strings.Join(fieldNames, ","))
	row := db.QueryRow(ctx, querystm, &repository, &branchName, &integration)
	err := row.Scan(values...)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("row.Scan: %v", err)
	}
	return &targetCoverage, nil
}

// AddHistory to update list history
func (d *DraftRepo) AddHistory(ctx context.Context, db database.QueryExecer, coverage float64, branchName, targetbranch, status, repository string, integration bool) error {
	querystm := "INSERT INTO public.history (branch_name,target_branch_name, coverage, time, status, repository, integration)  VALUES ($1, $2, $3, $4, $5, $6, $7)"
	_, err := db.Exec(ctx, querystm, &branchName, &targetbranch, &coverage, time.Now(), &status, &repository, &integration)
	if err != nil {
		return fmt.Errorf("db.Exec: %v", err)
	}
	return nil
}

// UpdateTargetCoverage to update target coverage
func (d *DraftRepo) UpdateTargetCoverage(ctx context.Context, db database.QueryExecer, coverage float64, id int64) error {
	querystm := "UPDATE public.target_coverage SET coverage = $1, update_at = $2 WHERE id = $3"
	_, err := db.Exec(ctx, querystm, &coverage, time.Now(), id)
	if err != nil {
		return err
	}
	return nil
}

func (d *DraftRepo) CreateTargetBranch(ctx context.Context, db database.QueryExecer, branchName, repository string, integration bool, key string, coverage float64) error {
	querystm := "INSERT INTO public.target_coverage (branch_name, coverage, repository, integration, update_at, secret_key) VALUES ($1, $2, $3, $4, $5, $6)"
	_, err := db.Exec(ctx, querystm, branchName, coverage, repository, integration, time.Now(), key)
	if err != nil {
		return fmt.Errorf("db.Exec: %v", err)
	}
	return nil
}

func (d *DraftRepo) GetGitHubEvent(ctx context.Context, db database.QueryExecer) (*entities.TargetCoverage, error) {
	targetCoverage := entities.TargetCoverage{}
	fieldNames, values := targetCoverage.FieldMap()
	querystm := fmt.Sprintf("SELECT %s FROM public.target_coverage", strings.Join(fieldNames, ","))
	row := db.QueryRow(ctx, querystm)
	err := row.Scan(values...)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("row.Scan: %v", err)
	}
	return &targetCoverage, nil
}

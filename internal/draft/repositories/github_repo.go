package repositories

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type GithubMergeStatusRepo struct{}

func (g *GithubMergeStatusRepo) SetRepoMergeStatus(ctx context.Context, db database.QueryExecer, org, repo string, isblocked bool) error {
	querystm := "UPDATE github_repo_state SET is_blocked=$1 WHERE org = $2 and repo = $3"
	_, err := db.Exec(ctx, querystm, database.Bool(isblocked), database.Text(org), database.Text(repo))
	if err != nil {
		return fmt.Errorf("db.Exec %w", err)
	}
	return nil
}

func (g *GithubMergeStatusRepo) GetRepoMergeStatus(ctx context.Context, db database.QueryExecer, org, repo string) (isblocked bool, err error) {
	querystm := "SELECT is_blocked FROM github_repo_state WHERE org = $1 and repo = $2"
	var isBlocked pgtype.Bool
	err = db.QueryRow(ctx, querystm, database.Text(org), database.Text(repo)).Scan(&isBlocked)
	if err != nil {
		return false, fmt.Errorf("db.QueryRow %w", err)
	}
	isblocked = isBlocked.Bool
	return
}

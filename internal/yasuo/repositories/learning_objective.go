package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/bob/entities"
	repositories_bob "github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
)

type LearningObjectiveRepo struct{}

func (r *LearningObjectiveRepo) FindByIDs(ctx context.Context, db database.Ext, loIDs []string) (map[string]*entities.LearningObjective, error) {
	var los entities.LearningObjectives
	e := &entities.LearningObjective{}
	fieldNames := database.GetFieldNames(e)
	stmt := "SELECT %s FROM %s WHERE lo_id = ANY($1) AND deleted_at IS NULL;"
	query := fmt.Sprintf(stmt, strings.Join(fieldNames, ", "), e.TableName())
	err := database.Select(ctx, db, query, loIDs).ScanAll(&los)
	if err != nil {
		return nil, fmt.Errorf("FindByIDs:%w", err)
	}
	mLos := make(map[string]*entities.LearningObjective)
	if len(los) > 0 {
		for _, v := range los {
			mLos[v.ID.String] = v
		}
	}

	return mLos, nil
}

func (r *LearningObjectiveRepo) SoftDeleteWithLoIDs(ctx context.Context, db database.QueryExecer, loIDs pgtype.TextArray) (int64, error) {
	query := `UPDATE learning_objectives SET deleted_at = NOW() WHERE lo_id = ANY($1::TEXT[]) AND deleted_at IS NULL`
	tag, err := db.Exec(ctx, query, &loIDs)
	if err != nil {
		return 0, fmt.Errorf("err db.Exec: %w", err)
	}
	return tag.RowsAffected(), nil
}

func (r *LearningObjectiveRepo) FindSchoolIDs(ctx context.Context, db database.QueryExecer, loIDs []string) ([]int32, error) {
	ctx, span := interceptors.StartSpan(ctx, "LearningObjectiveRepo.FindSchoolIDs")
	defer span.End()

	query := "SELECT school_id FROM learning_objectives WHERE deleted_at IS NULL AND lo_id = ANY($1)"
	pgIDs := database.TextArray(loIDs)

	schoolIDs := repositories_bob.EnSchoolIDs{}
	err := database.Select(ctx, db, query, &pgIDs).ScanAll(&schoolIDs)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}

	result := []int32{}
	for _, v := range schoolIDs {
		result = append(result, v.SchoolID)
	}

	return result, nil
}

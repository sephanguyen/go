package repositories

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
)

type GradeRepo struct{}

func (r *GradeRepo) GetByPartnerInternalIDs(ctx context.Context, db database.QueryExecer, partnerInternalIDs []string, orgID string) (map[string]string, error) {
	query := `
	SELECT grade_id, partner_internal_id
	FROM grade
	WHERE partner_internal_id = ANY($1::TEXT[]) and deleted_at is NULL AND resource_path=$2`
	rows, err := db.Query(ctx, query, database.TextArray(partnerInternalIDs), orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	gradeMap := make(map[string]string)
	for rows.Next() {
		var (
			gradeID           string
			partnerInternalID string
		)
		err = rows.Scan(&gradeID, &partnerInternalID)
		if err != nil {
			return nil, fmt.Errorf("failed scan: %v", err)
		}
		gradeMap[gradeID] = partnerInternalID
	}

	return gradeMap, nil
}

func (r *GradeRepo) GetGradesByOrg(ctx context.Context, db database.QueryExecer, orgID string) (map[string]string, error) {
	query := `
	SELECT g.grade_id, g.name
	FROM grade g
	WHERE deleted_at is NULL AND resource_path=$1`
	rows, err := db.Query(ctx, query, orgID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	gradeMap := make(map[string]string)
	for rows.Next() {
		var (
			gradeID   string
			gradeName string
		)
		err = rows.Scan(&gradeID, &gradeName)
		if err != nil {
			return nil, fmt.Errorf("failed scan: %v", err)
		}
		gradeMap[gradeID] = gradeName
	}

	return gradeMap, nil
}

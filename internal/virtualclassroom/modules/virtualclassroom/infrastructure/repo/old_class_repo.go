package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
)

type OldClassRepo struct{}

func (o *OldClassRepo) FindJoined(ctx context.Context, db database.QueryExecer, userID string) (domain.OldClasses, error) {
	ctx, span := interceptors.StartSpan(ctx, "OldClassRepo.FindJoined")
	defer span.End()

	oldClass := &OldClass{}
	fields, values := oldClass.FieldMap()

	query := fmt.Sprintf(`SELECT c.%s FROM classes c 
		JOIN class_members cm ON c.class_id = cm.class_id
		AND cm.user_id = $1 
		AND c.status = $2 
		AND cm.status = $3 `,
		strings.Join(fields, ", c."))

	rows, err := db.Query(ctx, query, &userID, string(domain.ClassStatusActive), string(domain.ClassMemberStatusActive))
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()

	var result domain.OldClasses
	for rows.Next() {
		if err := rows.Scan(values...); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		result = append(result, oldClass.ToOldClassDomain())
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return result, nil
}

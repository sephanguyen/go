package repositories

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/eureka/entities"
	dbeureka "github.com/manabie-com/backend/internal/eureka/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
)

type QuestionTagTypeRepo struct{}

func (r *QuestionTagTypeRepo) BulkUpsert(ctx context.Context, db database.QueryExecer, items []*entities.QuestionTagType) error {
	ctx, span := interceptors.StartSpan(ctx, "QuestionTagTypeRepo.BulkUpsert")
	defer span.End()

	const query = `
			INSERT INTO %s (%s) VALUES %s
				ON CONFLICT ON CONSTRAINT question_tag_type_id_pk
			DO UPDATE SET
				name = excluded.name,
				updated_at = NOW();
		`
	err := dbeureka.BulkUpsert(ctx, db, query, items)
	if err != nil {
		return fmt.Errorf("QuestionTagTypeRepo.BulkUpsert error: %s", err.Error())
	}
	return nil
}

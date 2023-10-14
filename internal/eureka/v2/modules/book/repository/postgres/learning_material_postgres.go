package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/book/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/modules/book/repository"
	"github.com/manabie-com/backend/internal/eureka/v2/modules/book/repository/postgres/dto"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"

	"github.com/jackc/pgx/v4"
)

type LearningMaterialRepo struct {
	DB database.Ext
}

var _ repository.LearningMaterialRepo = (*LearningMaterialRepo)(nil)

func (lm *LearningMaterialRepo) UpdatePublishStatusLearningMaterials(ctx context.Context, learningMaterials []domain.LearningMaterial) error {
	ctx, span := interceptors.StartSpan(ctx, "LearningMaterialRepo.UpdatePublishStatusLearningMaterials")
	defer span.End()

	learningMaterialHolder := dto.LearningMaterialDto{}
	table := learningMaterialHolder.TableName()

	stmt := fmt.Sprintf("UPDATE %s SET is_published = $2, updated_at = now() WHERE learning_material_id = $1::TEXT AND deleted_at IS NULL", table)

	batch := &pgx.Batch{}
	for _, learningMaterial := range learningMaterials {
		batch.Queue(stmt, learningMaterial.ID, learningMaterial.Published)
	}

	batchResults := lm.DB.SendBatch(ctx, batch)
	defer batchResults.Close()

	for i := 0; i < batch.Len(); i++ {
		ct, err := batchResults.Exec()

		if err != nil {
			return fmt.Errorf("updatePublishStatusLearningMaterials batchResults.Exec: %w", err)
		}

		if ct.RowsAffected() != 1 {
			return fmt.Errorf("updatePublishStatusLearningMaterials no item updated")
		}
	}

	return nil
}

func (lm *LearningMaterialRepo) GetByID(ctx context.Context, id string) (domain.LearningMaterial, error) {
	ctx, span := interceptors.StartSpan(ctx, "LearningMaterialRepo.GetByID")
	defer span.End()

	var result dto.LearningMaterialDto
	fields, _ := result.FieldMap()

	stmt := fmt.Sprintf(`
    SELECT %s
      FROM %s
     WHERE learning_material_id = $1
       AND deleted_at IS NULL;
	`, strings.Join(fields, ", "), result.TableName())

	if err := database.Select(ctx, lm.DB, stmt, database.Text(id)).ScanOne(&result); err != nil {
		return domain.LearningMaterial{}, fmt.Errorf("database.Select: %w", err)
	}

	return result.ToEntity(), nil
}

func (lm *LearningMaterialRepo) GetManyByIDs(ctx context.Context, ids []string) ([]domain.LearningMaterial, error) {
	ctx, span := interceptors.StartSpan(ctx, "LearningMaterialRepo.GetManyByIDs")
	defer span.End()

	buffer := 0
	idPlaceHolders := sliceutils.Map(ids, func(t string) string {
		buffer++
		return fmt.Sprintf("$%d", buffer)
	})
	idArr := strings.Join(idPlaceHolders, ",")

	var result dto.LearningMaterialDto
	fields, _ := result.FieldMap()

	query := fmt.Sprintf(`
    SELECT %s
      FROM %s
     WHERE learning_material_id IN (%s)
       AND deleted_at IS NULL;
	`, strings.Join(fields, ", "), result.TableName(), idArr)

	rows, err := lm.DB.Query(ctx, query, database.TextArray(ids))
	if err != nil {
		return nil, errors.NewDBError("LearningMaterialRepo.GetManyByIDs", err)
	}
	return scanLearningMaterials(rows)
}

func scanLearningMaterials(rows pgx.Rows) ([]domain.LearningMaterial, error) {
	var domainEntities []domain.LearningMaterial
	dtoEntity := &dto.LearningMaterialDto{}
	fields, _ := dtoEntity.FieldMap()

	defer rows.Close()
	for rows.Next() {
		l := new(dto.LearningMaterialDto)
		if err := rows.Scan(database.GetScanFields(l, fields)...); err != nil {
			return nil, errors.NewConversionError("LearningMaterialRepo.scanLearningMaterials", err)
		}
		domainEntities = append(domainEntities, l.ToEntity())
	}
	if err := rows.Err(); err != nil {
		return nil, errors.NewConversionError("LearningMaterialRepo.scanLearningMaterials", err)
	}
	return domainEntities, nil
}

package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

type LearningMaterialRepo struct{}

// Delete with soft delete
func (r *LearningMaterialRepo) Delete(ctx context.Context, db database.QueryExecer, lmID pgtype.Text) error {
	lm := &entities.LearningMaterial{}
	stmt := fmt.Sprintf(`UPDATE %s SET deleted_at = NOW() WHERE learning_material_id = $1::TEXT AND deleted_at IS NULL`, lm.TableName())

	cmd, err := db.Exec(ctx, stmt, lmID)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("not found any learning material to delete: %w", pgx.ErrNoRows)
	}

	return nil
}

func (r *LearningMaterialRepo) FindByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*entities.LearningMaterial, error) {
	ctx, span := interceptors.StartSpan(ctx, "LearningMaterialRepo.FindByIDs")
	defer span.End()
	lms := &entities.LearningMaterials{}
	e := &entities.LearningMaterial{}
	query := fmt.Sprintf("SELECT %s FROM %s WHERE learning_material_id = ANY($1) AND deleted_at IS NULL", strings.Join(database.GetFieldNames(e), ","), e.TableName())
	if err := database.Select(ctx, db, query, ids).ScanAll(lms); err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}
	return lms.Get(), nil
}

type LearningMaterialInfo struct {
	LearningMaterialID  pgtype.Text
	Type                pgtype.Text
	Name                pgtype.Text
	BookID              pgtype.Text
	ChapterID           pgtype.Text
	ChapterDisplayOrder pgtype.Int2
	TopicID             pgtype.Text
	TopicDisplayOrder   pgtype.Int2
	LmDisplayOrder      pgtype.Int2
	IsCompleted         pgtype.Bool
	HighestScore        pgtype.Int2
}

func (r *LearningMaterialRepo) FindInfoByStudyPlanItemIdentity(ctx context.Context, db database.QueryExecer, studentID, studyPlanID pgtype.Text, learningMaterialID pgtype.Text) ([]*LearningMaterialInfo, error) {
	ctx, span := interceptors.StartSpan(ctx, "LearningMaterialRepo.FindByStudyPlanItemIdentity")
	defer span.End()

	query := `
		select distinct on (student_id, study_plan_id, learning_material_id) 
			learning_material_id, 
			type, 
			name, 
			book_id, 
			chapter_id, 
			lalm.chapter_display_order, 
			lalm.topic_id, 
			lalm.topic_display_order, 
			lalm.lm_display_order,
			(select exists (select learning_material_id
								from get_student_completion_learning_material() gsclm
								where gsclm.student_id = lalm.student_id
								and gsclm.study_plan_id = lalm.study_plan_id
								and gsclm.learning_material_id = lm.learning_material_id)) as is_completed,
			(select coalesce((graded_points * 1.0 / total_points) * 100, 0)::smallint
				from max_graded_score() mgs
				where mgs.student_id = lalm.student_id
					and mgs.study_plan_id = lalm.study_plan_id
					and mgs.learning_material_id = lm.learning_material_id) as highest_score
		from list_available_learning_material() lalm
		join learning_material lm using (learning_material_id)
		where 
			student_id = $1::TEXT 
			and study_plan_id = $2::TEXT
			and learning_material_id = coalesce(nullif($3::TEXT, ''), learning_material_id)
		`
	rows, err := db.Query(ctx, query, &studentID, &studyPlanID, &learningMaterialID)
	if err != nil {
		return nil, fmt.Errorf("LearningMaterialRepo.FindByStudyPlanItemIdentity.Query: %w", err)
	}
	defer rows.Close()

	result := make([]*LearningMaterialInfo, 0)
	for rows.Next() {
		var lm LearningMaterialInfo
		if err := rows.Scan(&lm.LearningMaterialID, &lm.Type, &lm.Name, &lm.BookID, &lm.ChapterID, &lm.ChapterDisplayOrder, &lm.TopicID, &lm.TopicDisplayOrder, &lm.LmDisplayOrder, &lm.IsCompleted, &lm.HighestScore); err != nil {
			return nil, fmt.Errorf("LearningMaterialRepo.FindByStudyPlanItemIdentity.Scan: %w", err)
		}
		result = append(result, &lm)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("LearningMaterialRepo.FindByStudyPlanItemIdentity.Err: %w", err)
	}

	return result, nil
}

func (r *LearningMaterialRepo) UpdateDisplayOrders(ctx context.Context, db database.QueryExecer, lms []*entities.LearningMaterial) error {
	ctx, span := interceptors.StartSpan(ctx, "LearningMaterialRepo.UpdateDisplayOrders")
	defer span.End()
	queue := func(b *pgx.Batch, t *entities.LearningMaterial) {
		query := fmt.Sprintf("UPDATE %s SET display_order = $1, updated_at = now() WHERE learning_material_id = $2::TEXT AND deleted_at IS NULL", t.TableName())
		b.Queue(query, &t.DisplayOrder, &t.ID)
	}
	b := &pgx.Batch{}
	for _, lm := range lms {
		queue(b, lm)
	}
	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()
	for i := 0; i < len(lms); i++ {
		ct, err := batchResults.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
		if ct.RowsAffected() != 1 {
			return fmt.Errorf("LearningMaterials not updated")
		}
	}
	return nil
}

func (r *LearningMaterialRepo) UpdateName(ctx context.Context, db database.QueryExecer, lmID pgtype.Text, lmName pgtype.Text) (int64, error) {
	ctx, span := interceptors.StartSpan(ctx, "LearningMaterialRepo.UpdateLearningMaterialName")
	defer span.End()
	lm := &entities.LearningMaterial{}

	query := fmt.Sprintf("UPDATE %s SET name = $1, updated_at = now() WHERE learning_material_id = $2::TEXT AND deleted_at IS NULL", lm.TableName())
	cmd, err := db.Exec(ctx, query, lmName, lmID)
	if err != nil {
		return 0, fmt.Errorf("db.Exec: %w", err)
	}

	return cmd.RowsAffected(), nil
}

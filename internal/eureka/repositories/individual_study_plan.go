package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

type IndividualStudyPlan struct {
}

const insertIndividualStudyPlan = `INSERT INTO
	individual_study_plan (%s)
VALUES (%s)
ON CONFLICT (learning_material_id, student_id, study_plan_id)
DO UPDATE SET
	available_from = $4,
	available_to = $5,
	start_date = $6,
	end_date = $7,
	updated_at = $8,
	status = $11,
	school_date = $12
RETURNING study_plan_id`

func (m *IndividualStudyPlan) BulkUpdateTime(ctx context.Context, db database.QueryExecer, items []*entities.IndividualStudyPlan) error {
	queueFn := func(b *pgx.Batch, e *entities.IndividualStudyPlan) {
		query := `UPDATE individual_study_plan
		SET start_date = $3, end_date = $4, available_from = $5, available_to = $6, updated_at = now()
		WHERE study_plan_id = $1 and learning_material_id = $2`
		b.Queue(query, &e.ID, &e.LearningMaterialID, &e.StartDate, &e.EndDate, &e.AvailableFrom, &e.AvailableTo)
	}

	b := &pgx.Batch{}
	for _, each := range items {
		queueFn(b, each)
	}

	result := db.SendBatch(ctx, b)
	defer result.Close()

	for i := 0; i < b.Len(); i++ {
		_, err := result.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
	}

	return nil
}

// BulkSync inserts items into individual_study_plan table
func (m *IndividualStudyPlan) BulkSync(
	ctx context.Context,
	db database.QueryExecer,
	items []*entities.IndividualStudyPlan,
) (
	insertItems []*entities.IndividualStudyPlan,
	err error,
) {
	queueFn := func(b *pgx.Batch, item *entities.IndividualStudyPlan) {
		fieldNames := database.GetFieldNames(item)
		placeHolders := database.GeneratePlaceholders(len(fieldNames))
		query := fmt.Sprintf(insertIndividualStudyPlan, strings.Join(fieldNames, ","), placeHolders)
		scanFields := database.GetScanFields(item, fieldNames)
		b.Queue(query, scanFields...)
	}

	b := &pgx.Batch{}
	for _, item := range items {
		queueFn(b, item)
	}
	result := db.SendBatch(ctx, b)
	defer result.Close()

	for i := 0; i < b.Len(); i++ {
		var returnedID pgtype.Text
		if serr := result.QueryRow().Scan(&returnedID); serr != nil {
			err = fmt.Errorf("result.QueryRow.Scan: %w", serr)
			return
		}

		// In case of insert, the item id in DB should be the same with passed item id.
		if returnedID.String == items[i].ID.String {
			insertItems = append(insertItems, items[i])
		}
	}
	return
}

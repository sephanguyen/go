package learning_objectives

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/manabie-com/backend/features/repository/syllabus/utils"
	"github.com/manabie-com/backend/internal/bob/constants"
	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
)

func (s *Suite) insertAValidLOInDB(ctx context.Context, losType string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	stepState.LoID = idutil.ULIDNow()
	stepState.TopicID = idutil.ULIDNow()

	if err := s.genInsertTopic(ctx); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("insert topic: %w", err)
	}

	if _, err := s.DB.Exec(
		ctx,
		`INSERT INTO learning_objectives(lo_id, "name", topic_id, "type", created_at, updated_at)
		VALUES($1, $2, $3, $4, $5, $5)`,
		database.Text(stepState.LoID),
		database.Text(idutil.ULIDNow()),
		database.Text(stepState.TopicID),
		database.Text(losType),
		database.Timestamptz(time.Now()),
	); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("insert learning_objectives: %w", err)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) updateLOs(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	if _, err := s.DB.Exec(
		ctx,
		`UPDATE
			learning_objectives
		SET
			name = $1,
			instruction = $2,
			grade_to_pass = $3,
			manual_grading = $4,
			time_limit = $5,
			maximum_attempt = $6,
			approve_grading = $7,
			grade_capping = $8,
			review_option = $9
		WHERE
			lo_id = $10
		`,
		database.Text(idutil.ULIDNow()),
		database.Text(idutil.ULIDNow()),
		database.Int4(100),
		database.Bool(true),
		database.Int4(50),
		database.Int4(99),
		database.Bool(true),
		database.Bool(false),
		database.Text("EXAM_LO_REVIEW_OPTION_AFTER_DUE_DATE"),
		database.Text(stepState.LoID),
	); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("update learning_objectives: %w", err)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) thereIsARowLOsInDB(ctx context.Context, record string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	var gotID pgtype.Text
	if err := s.DB.QueryRow(
		ctx,
		`SELECT el.learning_material_id
		FROM exam_lo el
		INNER JOIN learning_objectives lo ON
			lo.lo_id = el.learning_material_id
			AND lo.topic_id = el.topic_id
			AND lo.name = el.name
			AND COALESCE(lo.instruction, lo.lo_id) = COALESCE(el.instruction, el.learning_material_id)
			AND COALESCE(lo.grade_to_pass, 0) = COALESCE(el.grade_to_pass, 0)
			AND lo.manual_grading = el.manual_grading
			AND COALESCE(lo.time_limit, 0) = COALESCE(el.time_limit, 0)
			AND COALESCE(lo.maximum_attempt, 0) = COALESCE(el.maximum_attempt, 0)
			AND lo.approve_grading = el.approve_grading
			AND lo.grade_capping = el.grade_capping
			AND lo.review_option = el.review_option
		WHERE
			el.learning_material_id = $1
		`,
		stepState.LoID,
	).Scan(&gotID); err != nil && err != pgx.ErrNoRows {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("select exam_lo: %w", err)
	}

	if record == "valid" && gotID.String != stepState.LoID {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("wrong exam_lo record, expect: %s, got: %s", stepState.LoID, gotID.String)
	}

	if record != "valid" && gotID.String != "" {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("wrong exam_lo record, expect: no record, got: %s", gotID.String)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) genInsertTopic(ctx context.Context) error {
	stepState := utils.StepStateFromContext[StepState](ctx)

	topic := &entities.Topic{}
	database.AllNullEntity(topic)
	if err := multierr.Combine(
		topic.SchoolID.Set(constants.ManabieSchool),
		topic.ID.Set(stepState.TopicID),
		topic.ChapterID.Set(idutil.ULIDNow()),
		topic.Name.Set(fmt.Sprintf("topic-%s", idutil.ULIDNow())),
		topic.Grade.Set(rand.Intn(5)+1),
		topic.Subject.Set(epb.Subject_SUBJECT_NONE),
		topic.Status.Set(epb.TopicStatus_TOPIC_STATUS_NONE),
		topic.CreatedAt.Set(time.Now()),
		topic.UpdatedAt.Set(time.Now()),
		topic.TotalLOs.Set(1),
		topic.TopicType.Set(epb.TopicType_TOPIC_TYPE_EXAM),
		topic.EssayRequired.Set(true),
	); err != nil {
		return err
	}

	topicRepo := repositories.TopicRepo{}
	return topicRepo.BulkUpsertWithoutDisplayOrder(ctx, s.DB, []*entities.Topic{topic})
}

package entities

import (
	"math/rand"
	"strconv"
	"time"

	"github.com/jackc/pgtype"
)

type FlashcardProgression struct {
	OriginalQuizSetID     pgtype.Text
	StudySetID            pgtype.Text
	OriginalStudySetID    pgtype.Text
	StudentID             pgtype.Text
	StudyPlanItemID       pgtype.Text
	LoID                  pgtype.Text
	QuizExternalIDs       pgtype.TextArray
	StudyingIndex         pgtype.Int4
	SkippedQuestionIDs    pgtype.TextArray
	RememberedQuestionIDs pgtype.TextArray
	CreatedAt             pgtype.Timestamptz
	UpdatedAt             pgtype.Timestamptz
	CompletedAt           pgtype.Timestamptz
	DeletedAt             pgtype.Timestamptz
	StudyPlanID           pgtype.Text
	LearningMaterialID    pgtype.Text
}

func (s *FlashcardProgression) FieldMap() ([]string, []interface{}) {
	return []string{
			"original_quiz_set_id",
			"study_set_id",
			"original_study_set_id",
			"student_id",
			"study_plan_item_id",
			"lo_id",
			"quiz_external_ids",
			"studying_index",
			"skipped_question_ids",
			"remembered_question_ids",
			"created_at",
			"updated_at",
			"completed_at",
			"deleted_at",
			"study_plan_id",
			"learning_material_id",
		}, []interface{}{
			&s.OriginalQuizSetID,
			&s.StudySetID,
			&s.OriginalStudySetID,
			&s.StudentID,
			&s.StudyPlanItemID,
			&s.LoID,
			&s.QuizExternalIDs,
			&s.StudyingIndex,
			&s.SkippedQuestionIDs,
			&s.RememberedQuestionIDs,
			&s.CreatedAt,
			&s.UpdatedAt,
			&s.CompletedAt,
			&s.DeletedAt,
			&s.StudyPlanID,
			&s.LearningMaterialID,
		}
}

func (s *FlashcardProgression) TableName() string {
	return "flashcard_progressions"
}

func (s *FlashcardProgression) Shuffle() {
	unixNano := time.Now().UTC().UnixNano()
	randomSeed := strconv.FormatInt(unixNano, 10)
	seed, _ := strconv.ParseInt(randomSeed, 10, 64)
	r := rand.New(rand.NewSource(seed))
	eqIDs := s.QuizExternalIDs.Elements
	r.Shuffle(len(eqIDs), func(i, j int) { eqIDs[i], eqIDs[j] = eqIDs[j], eqIDs[i] })
}

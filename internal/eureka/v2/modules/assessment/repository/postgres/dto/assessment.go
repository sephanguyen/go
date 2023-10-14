package dto

import (
	"time"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/assessment/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/constants"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type Assessment struct {
	BaseEntity
	ID                 pgtype.Text
	CourseID           pgtype.Text
	LearningMaterialID pgtype.Text
	RefTable           pgtype.Varchar
}

func (a *Assessment) FieldMap() ([]string, []interface{}) {
	return []string{
			"id",
			"course_id",
			"learning_material_id",
			"ref_table",
			"created_at",
			"updated_at",
			"deleted_at",
		}, []interface{}{
			&a.ID,
			&a.CourseID,
			&a.LearningMaterialID,
			&a.RefTable,
			&a.CreatedAt,
			&a.UpdatedAt,
			&a.DeletedAt,
		}
}

func (a *Assessment) TableName() string {
	return "assessment"
}

func (a *Assessment) ToEntity() (domain.Assessment, error) {
	assessment := domain.Assessment{
		ID:                   a.ID.String,
		CourseID:             a.CourseID.String,
		LearningMaterialID:   a.LearningMaterialID.String,
		LearningMaterialType: toLMType(a.RefTable.String),
	}
	err := assessment.Validate()
	if err != nil {
		return domain.Assessment{}, errors.NewConversionError("domain.Validate", err)
	}

	return assessment, nil
}

func (a *Assessment) FromEntity(now time.Time, d domain.Assessment) error {
	database.AllNullEntity(a)

	if err := multierr.Combine(
		a.ID.Set(d.ID),
		a.CourseID.Set(d.CourseID),
		a.LearningMaterialID.Set(d.LearningMaterialID),
		a.RefTable.Set(toRefTable(d.LearningMaterialType)),
		a.CreatedAt.Set(now),
		a.UpdatedAt.Set(now),
	); err != nil {
		return errors.NewConversionError("multierr.Combine", err)
	}

	return nil
}

func toRefTable(lmType constants.LearningMaterialType) string {
	switch lmType {
	case constants.LearningObjective:
		return "learning_objective"
	case constants.ExamLO:
		return "exam_lo"
	case constants.FlashCard:
		return "flash_card"
	case constants.GeneralAssignment:
		return "assignment"
	case constants.TaskAssignment:
		return "task_assignment"
	}
	return ""
}

func toLMType(refTable string) constants.LearningMaterialType {
	switch refTable {
	case "learning_objective":
		return constants.LearningObjective
	case "exam_lo":
		return constants.ExamLO
	case "flash_card":
		return constants.FlashCard
	case "assignment":
		return constants.GeneralAssignment
	case "task_assignment":
		return constants.TaskAssignment
	}
	return ""
}

type AssessmentExtended struct {
	BaseEntity
	ID                 pgtype.Text
	CourseID           pgtype.Text
	LearningMaterialID pgtype.Text
	RefTable           pgtype.Varchar

	// Virtual fields.
	ManualGrading pgtype.Bool
}

func (av *AssessmentExtended) FieldMap() ([]string, []interface{}) {
	return []string{
			"id",
			"course_id",
			"learning_material_id",
			"ref_table",
			"manual_grading",
			"created_at",
			"updated_at",
			"deleted_at",
		}, []interface{}{
			&av.ID,
			&av.CourseID,
			&av.LearningMaterialID,
			&av.RefTable,
			&av.ManualGrading,
			&av.CreatedAt,
			&av.UpdatedAt,
			&av.DeletedAt,
		}
}

func (av *AssessmentExtended) TableName() string {
	return "assessment"
}

func (av *AssessmentExtended) ToEntity() (domain.Assessment, error) {
	assessment := domain.Assessment{
		ID:                   av.ID.String,
		CourseID:             av.CourseID.String,
		LearningMaterialID:   av.LearningMaterialID.String,
		LearningMaterialType: toLMType(av.RefTable.String),
		ManualGrading:        av.ManualGrading.Bool,
	}
	err := assessment.Validate()
	if err != nil {
		return domain.Assessment{}, errors.NewConversionError("domain.Validate", err)
	}

	return assessment, nil
}

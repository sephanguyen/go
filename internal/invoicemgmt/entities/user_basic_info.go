package entities

import "github.com/jackc/pgtype"

type UserBasicInfo struct {
	UserID            pgtype.Text
	Name              pgtype.Text
	FirstName         pgtype.Text
	LastName          pgtype.Text
	FullNamePhonetic  pgtype.Text
	FirstNamePhonetic pgtype.Text
	LastNamePhonetic  pgtype.Text
	CurrentGrade      pgtype.Int2
	GradeID           pgtype.Text
	CreatedAt         pgtype.Timestamptz
	UpdatedAt         pgtype.Timestamptz
	DeletedAt         pgtype.Timestamptz
	ResourcePath      pgtype.Text
}

func (e *UserBasicInfo) FieldMap() ([]string, []interface{}) {
	return []string{
			"user_id",
			"name",
			"first_name",
			"last_name",
			"full_name_phonetic",
			"first_name_phonetic",
			"last_name_phonetic",
			"current_grade",
			"grade_id",
			"updated_at",
			"created_at",
			"deleted_at",
			"resource_path",
		}, []interface{}{
			&e.UserID,
			&e.Name,
			&e.FirstName,
			&e.LastName,
			&e.FullNamePhonetic,
			&e.FirstNamePhonetic,
			&e.LastNamePhonetic,
			&e.CurrentGrade,
			&e.GradeID,
			&e.UpdatedAt,
			&e.CreatedAt,
			&e.DeletedAt,
			&e.ResourcePath,
		}
}

func (e *UserBasicInfo) TableName() string {
	return "user_basic_info"
}

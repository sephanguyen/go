package entities

import (
	"github.com/jackc/pgtype"
)

type Package struct {
	ID              pgtype.Text
	Country         pgtype.Text
	Name            pgtype.Text
	Descriptions    pgtype.TextArray
	Price           pgtype.Int4
	DiscountedPrice pgtype.Int4
	StartAt         pgtype.Timestamptz
	EndAt           pgtype.Timestamptz
	Duration        pgtype.Int4 // check duration first, if zero then use start_at, end_at
	PrioritizeLevel pgtype.Int4
	Properties      pgtype.JSONB
	IsRecommended   pgtype.Bool
	IsActive        pgtype.Bool
	CreatedAt       pgtype.Timestamptz
	UpdatedAt       pgtype.Timestamptz
}

func (rcv *Package) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"package_id", "country", "name", "descriptions", "price", "discounted_price", "start_at", "end_at", "duration", "prioritize_level", "properties", "is_recommended", "is_active", "created_at", "updated_at"}
	values = []interface{}{&rcv.ID, &rcv.Country, &rcv.Name, &rcv.Descriptions, &rcv.Price, &rcv.DiscountedPrice, &rcv.StartAt, &rcv.EndAt, &rcv.Duration, &rcv.PrioritizeLevel, &rcv.Properties, &rcv.IsRecommended, &rcv.IsActive, &rcv.CreatedAt, &rcv.UpdatedAt}
	return
}

func (*Package) TableName() string {
	return "packages"
}

type PackageProperties struct {
	CanWatchVideo     []string     `json:"can_watch_video"`
	CanViewStudyGuide []string     `json:"can_view_study_guide"`
	CanDoQuiz         []string     `json:"can_do_quiz"`
	LimitOnlineLesson int          `json:"limit_online_lession"` // -1 is unlimited
	AskTutor          *AskTutorCfg `json:"ask_tutor,omitempty"`
}

type AskTutorCfg struct {
	TotalQuestionLimit int    `json:"total_question_limit"`
	LimitDuration      string `json:"limit_duration"` // THIS_DAY, THIS_WEEK, THIS_MONTH
}

func (rcv *Package) GetProperties() (*PackageProperties, error) {
	p := &PackageProperties{}
	err := rcv.Properties.AssignTo(p)
	return p, err
}

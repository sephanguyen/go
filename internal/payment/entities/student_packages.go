package entities

import (
	"github.com/manabie-com/backend/internal/golibs"

	"github.com/jackc/pgtype"
)

type StudentPackages struct {
	ID          pgtype.Text `sql:"student_package_id,pk"`
	StudentID   pgtype.Text `sql:"student_id"`
	PackageID   pgtype.Text `sql:"package_id"`
	StartAt     pgtype.Timestamptz
	EndAt       pgtype.Timestamptz
	Properties  pgtype.JSONB
	IsActive    pgtype.Bool
	LocationIDs pgtype.TextArray
	CreatedAt   pgtype.Timestamptz
	UpdatedAt   pgtype.Timestamptz
	DeletedAt   pgtype.Timestamptz
}

func (p *StudentPackages) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"student_package_id",
		"student_id",
		"package_id",
		"start_at",
		"end_at",
		"properties",
		"is_active",
		"location_ids",
		"created_at",
		"updated_at",
		"deleted_at",
	}
	values = []interface{}{
		&p.ID,
		&p.StudentID,
		&p.PackageID,
		&p.StartAt,
		&p.EndAt,
		&p.Properties,
		&p.IsActive,
		&p.LocationIDs,
		&p.CreatedAt,
		&p.UpdatedAt,
		&p.DeletedAt,
	}
	return
}

type PackageProperties struct {
	AllCourseInfo     []CourseInfo `json:"all_course_info"`
	CanWatchVideo     []string     `json:"can_watch_video"`
	CanViewStudyGuide []string     `json:"can_view_study_guide"`
	CanDoQuiz         []string     `json:"can_do_quiz"`
	LimitOnlineLesson int          `json:"limit_online_lession"` // -1 is unlimited
	AskTutor          *AskTutorCfg `json:"ask_tutor,omitempty"`
}

type CourseInfo struct {
	CourseID      string `json:"course_id"`
	Name          string `json:"name"`
	NumberOfSlots int    `json:"number_of_slots"`
	Weight        int    `json:"weight"`
}

type AskTutorCfg struct {
	TotalQuestionLimit int    `json:"total_question_limit"`
	LimitDuration      string `json:"limit_duration"` // THIS_DAY, THIS_WEEK, THIS_MONTH
}

func (p *StudentPackages) GetProperties() (*PackageProperties, error) {
	pp := &PackageProperties{}
	err := p.Properties.AssignTo(pp)
	return pp, err
}

func (p *StudentPackages) GetCourseIDs() ([]string, error) {
	prop, err := p.GetProperties()
	if err != nil {
		return nil, err
	}
	courseIDs := make([]string, 0)
	courseIDs = append(courseIDs, prop.CanDoQuiz...)
	courseIDs = append(courseIDs, prop.CanViewStudyGuide...)
	courseIDs = append(courseIDs, prop.CanWatchVideo...)

	courseIDs = golibs.Uniq(courseIDs)
	return courseIDs, nil
}

func (p *StudentPackages) GetLocationIDs() (locationIDs []string) {
	for _, locationID := range p.LocationIDs.Elements {
		if locationID.Status == pgtype.Present {
			locationIDs = append(locationIDs, locationID.String)
		}
	}
	return
}

func (p *StudentPackages) TableName() string {
	return "student_packages"
}

package entity

import (
	"time"

	"github.com/jackc/pgtype"
)

type CourseStudyPlanAttrs struct {
	CourseID    string `graphql:"course_id"`
	StudyPlanID string `graphql:"study_plan_id"`

	StudyPlan struct {
		StudyPlanAttrs    `graphql:"... on study_plans"`
		CreatedAt         time.Time `graphql:"created_at"`
		MasterStudyPlanID string    `graphql:"master_study_plan_id"`
	} `graphql:"study_plan"`
}

type StudyPlanAttrs struct {
	Name        string `graphql:"name"`
	StudyPlanID string `graphql:"study_plan_id"`
}

type ContentStructure struct {
	pgtype.JSONB
	CourseID     string `graphql:"course_id"`
	BookID       string `graphql:"book_id"`
	ChapterID    string `graphql:"chapter_id"`
	TopicID      string `graphql:"topic_id"`
	LoID         string `graphql:"lo_id"`
	AssignmentID string `graphql:"assignment_id"`
}

type LoStudyPlanItem struct {
	LoID string `graphql:"lo_id"`
}

type AssignmentStudyPlanItem struct {
	AssignmentID string `graphql:"assignment_id"`
}

func (l *LoStudyPlanItem) GetLoID() string {
	return l.LoID
}

func (l *AssignmentStudyPlanItem) GetAssignmentID() string {
	return l.AssignmentID
}

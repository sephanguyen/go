package entity

import (
	"time"

	"github.com/jackc/pgtype"
)

type GraphqlStudyPlanOneQuery struct {
	StudyPlanOne []struct {
		Name          string `graphql:"name"`
		StudyPlanID   string `graphql:"study_plan_id"`
		StudyPlanType string `graphql:"study_plan_type"`
		CourseID      string `graphql:"course_id"`

		StudyPlanItems []struct {
			StudyPlanItemID string    `graphql:"study_plan_item_id"`
			AvailableFrom   time.Time `graphql:"available_from"`
			AvailableTo     time.Time `graphql:"available_to"`

			ContentStructure struct {
				pgtype.JSONB
				CourseID     string `graphql:"course_id"`
				BookID       string `graphql:"book_id"`
				ChapterID    string `graphql:"chapter_id"`
				TopicID      string `graphql:"topic_id"`
				LoID         string `graphql:"lo_id"`
				AssignmentID string `graphql:"assignment_id"`
			} `graphql:"content_structure"`

			DisplayOder int32     `graphql:"display_order"`
			StartDate   time.Time `graphql:"start_date"`
			EndDate     time.Time `graphql:"end_date"`

			AssignmentStudyPlanItem struct {
				AssignmentID string `graphql:"assignment_id"`
			} `graphql:"assignment_study_plan_item"`

			LoStudyPlanItem struct {
				LoID string `graphql:"lo_id"`
			} `graphql:"lo_study_plan_item"`
		} `graphql:"study_plan_items"`
	} `graphql:" study_plans(where: {study_plan_id: {_eq: $study_plan_id}})"`
}

package entity

import (
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"

	"go.uber.org/multierr"
)

type GraphqlStudentStudyPlansByCourseIDQuery struct {
	StudentStudyPlans []struct {
		StudentStudyPlanAttrs `graphql:"... on student_study_plans"`
		StudyPlan             struct {
			StudyPlanAttrs `graphql:"... on study_plans"`
			StudyPlanItems []struct {
				LoStudyPlanItem         `graphql:"lo_study_plan_item"`
				AssignmentStudyPlanItem `graphql:"assignment_study_plan_item"`

				StudyPlanItemID  string           `graphql:"study_plan_item_id"`
				DisplayOrder     int              `graphql:"display_order"`
				AvailableTo      time.Time        `graphql:"available_to"`
				AvailableFrom    time.Time        `graphql:"available_from"`
				EndDate          time.Time        `graphql:"end_date"`
				StartDate        time.Time        `graphql:"start_date"`
				ContentStructure ContentStructure `graphql:"content_structure"`
			} `graphql:"study_plan_items(order_by: {display_order: asc})"`
		} `graphql:"study_plan"`
	} `graphql:"student_study_plans(order_by: {created_at: desc}, where: {course_students: {course_id: {_eq: $course_id}}, study_plan_id: {_eq: $study_plan_id}})"`
}

type StudentStudyPlanAttrs struct {
	StudentID   string `graphql:"student_id"`
	StudyPlanID string `graphql:"study_plan_id"`
	StudyPlan   struct {
		StudyPlanAttrs    `graphql:"... on study_plans"`
		CreatedAt         time.Time `graphql:"created_at"`
		MasterStudyPlanID string    `graphql:"master_study_plan_id"`
	} `graphql:"study_plan"`
}

// StudentStudyPlansManyV2
type StudentStudyPlansManyV2 struct {
	StudentStudyPlans []struct {
		StudentStudyPlanAttrsV2 `graphql:"... on student_study_plans"`
	} `graphql:"student_study_plans(order_by: {created_at: desc}, where: {student_id: {_in: $student_ids}, study_plan: {course_id: {_eq: $course_id}, status: {_eq: $status}}})"`
}

type StudentStudyPlanAttrsV2 struct {
	StudentID   string `graphql:"student_id"`
	StudyPlanID string `graphql:"study_plan_id"`
	StudyPlan   struct {
		StudyPlanAttrsV2 `graphql:"... on study_plans"`
	} `graphql:"study_plan"`
}

type StudyPlanAttrsV2 struct {
	Name              string    `graphql:"name"`
	StudyPlanID       string    `graphql:"study_plan_id"`
	CreatedAt         time.Time `graphql:"created_at"`
	MasterStudyPlanID string    `graphql:"master_study_plan_id"`
	BookID            string    `graphql:"book_id"`
	Grades            []int64   `graphql:"grades"`
	Status            string    `graphql:"status"`
}

func (q *GraphqlStudentStudyPlansByCourseIDQuery) getStudyPlanItems() []*entities.StudyPlanItem {
	retrievedSPItems := q.StudentStudyPlans[0].StudyPlan.StudyPlanItems
	n := len(retrievedSPItems)
	spItems := make([]*entities.StudyPlanItem, 0, n)
	for i := 0; i < n; i++ {
		spItem := new(entities.StudyPlanItem)
		contentStructure := entities.ContentStructure{
			CourseID:     retrievedSPItems[i].ContentStructure.CourseID,
			BookID:       retrievedSPItems[i].ContentStructure.BookID,
			ChapterID:    retrievedSPItems[i].ContentStructure.ChapterID,
			TopicID:      retrievedSPItems[i].ContentStructure.TopicID,
			LoID:         retrievedSPItems[i].ContentStructure.LoID,
			AssignmentID: retrievedSPItems[i].ContentStructure.AssignmentID,
		}
		if err := multierr.Combine(
			spItem.ID.Set(retrievedSPItems[i].StudyPlanItemID),
			spItem.DisplayOrder.Set(retrievedSPItems[i].DisplayOrder),
			spItem.AvailableTo.Set(retrievedSPItems[i].AvailableTo),
			spItem.AvailableFrom.Set(retrievedSPItems[i].AvailableFrom),
			spItem.EndDate.Set(retrievedSPItems[i].EndDate),
			spItem.StartDate.Set(retrievedSPItems[i].StartDate),
			spItem.ContentStructure.Set(contentStructure)); err != nil {
			return nil
		}
		spItems = append(spItems, spItem)
	}
	return spItems
}

func (q *GraphqlStudentStudyPlansByCourseIDQuery) getStudyPlan() *entities.StudyPlan {
	retrievedStudyPlan := q.StudentStudyPlans[0].StudentStudyPlanAttrs.StudyPlan
	sp := new(entities.StudyPlan)
	if err := multierr.Combine(
		sp.ID.Set(retrievedStudyPlan.StudyPlanID),
		sp.Name.Set(retrievedStudyPlan.Name),
		sp.MasterStudyPlan.Set(retrievedStudyPlan.MasterStudyPlanID),
		sp.CreatedAt.Set(retrievedStudyPlan.CreatedAt)); err != nil {
		return nil
	}
	return sp
}

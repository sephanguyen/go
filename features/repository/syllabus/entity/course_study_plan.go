package entity

import (
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"

	"go.uber.org/multierr"
)

// for CourseStudyPlansByCourseID
type GraphqlCourseStudyPlansByCourseIDQuery struct {
	CourseStudyPlans []struct {
		CourseStudyPlanAttrs `graphql:"... on course_study_plans"`
		StudyPlan            struct {
			StudyPlanAttrs `graphql:"... on study_plans"`
			StudyPlanItems []struct {
				LoStudyPlanItem         `graphql:"lo_study_plan_item"`
				AssignmentStudyPlanItem `graphql:"assignment_study_plan_item"`

				StudyPlanItemID  string           `graphql:"study_plan_item_id"`
				DisplayOrder     int32            `graphql:"display_order"`
				AvailableTo      time.Time        `graphql:"available_to"`
				AvailableFrom    time.Time        `graphql:"available_from"`
				EndDate          time.Time        `graphql:"end_date"`
				StartDate        time.Time        `graphql:"start_date"`
				ContentStructure ContentStructure `graphql:"content_structure"`
			} `graphql:"study_plan_items(order_by: {display_order: asc})"`
		} `graphql:"study_plan"`
	} `graphql:"course_study_plans(order_by: {created_at: desc}, where: {course_id: {_eq: $course_id}, study_plan_id: {_eq: $study_plan_id}})"`
}

// for CourseStudyPlansList
type GraphqlCourseStudyPlansListQuery struct {
	CourseStudyPlans []struct {
		CourseStudyPlanAttrs `graphql:"... on course_study_plans"`
	} `graphql:"course_study_plans(limit: $limit, offset: $offset, order_by: {created_at: desc}, where: {course_id: {_eq: $course_id}})"`
	CourseStudyPlanAggregate struct {
		Aggregate struct {
			Count int `graphql:"count"`
		} `graphql:"aggregate"`
	} `graphql:"course_study_plans_aggregate(where: {course_id: {_eq: $course_id}})"`
}

func (q *GraphqlCourseStudyPlansByCourseIDQuery) getStudyPlanItems() []*entities.StudyPlanItem {
	retrievedSPItems := q.CourseStudyPlans[0].StudyPlan.StudyPlanItems
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

func (q *GraphqlCourseStudyPlansByCourseIDQuery) getStudyPlan() *entities.StudyPlan {
	retrievedStudyPlan := q.CourseStudyPlans[0].CourseStudyPlanAttrs.StudyPlan
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

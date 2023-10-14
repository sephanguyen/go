package utils

import (
	"fmt"

	"github.com/manabie-com/backend/internal/eureka/entities"

	"github.com/jackc/pgtype"
)

func CompareContentStructure(actualContentStructure, expectedContentStructure *pgtype.JSONB) error {
	actual := new(entities.ContentStructure)
	expected := new(entities.ContentStructure)
	actualContentStructure.AssignTo(actual)
	expectedContentStructure.AssignTo(expected)
	if actual.CourseID != expected.CourseID {
		return fmt.Errorf("unexpected content structure course id: want: %v, actual: %v", expected.CourseID, actual.CourseID)
	}
	if actual.BookID != expected.BookID {
		return fmt.Errorf("unexpected content structure book id: want: %v, actual: %v", expected.BookID, actual.BookID)
	}
	if actual.ChapterID != expected.ChapterID {
		return fmt.Errorf("unexpected content structure chapter id: want: %v, actual: %v", expected.ChapterID, actual.ChapterID)
	}
	if actual.TopicID != expected.TopicID {
		return fmt.Errorf("unexpected content structure topic id: want: %v, actual: %v", expected.TopicID, actual.TopicID)
	}
	return nil
}

func CompareStudyPlanItem(expectedStudyPlanItems, actualStudyPlanItems []*entities.StudyPlanItem) error {
	n := len(actualStudyPlanItems)
	if n == 0 {
		return fmt.Errorf("study plan items is empty")
	}
	if len(actualStudyPlanItems) != len(expectedStudyPlanItems) {
		return fmt.Errorf("not equal study plan items length: expected %d, got %d", len(expectedStudyPlanItems), len(actualStudyPlanItems))
	}

	for i := 0; i < n; i++ {
		if actualStudyPlanItems[i].ID != expectedStudyPlanItems[i].ID {
			return fmt.Errorf("unexpected study plan item id: want %s, got %s", expectedStudyPlanItems[i].ID.String, actualStudyPlanItems[i].ID.String)
		}
		if actualStudyPlanItems[i].AvailableTo.Time != expectedStudyPlanItems[i].AvailableTo.Time {
			return fmt.Errorf("unexpected available to: want %v, got %v", expectedStudyPlanItems[i].AvailableTo, actualStudyPlanItems[i].AvailableTo)
		}
		if actualStudyPlanItems[i].AvailableFrom.Time != expectedStudyPlanItems[i].AvailableFrom.Time {
			return fmt.Errorf("unexpected available from: want %v, got %v", expectedStudyPlanItems[i].AvailableFrom, actualStudyPlanItems[i].AvailableFrom)
		}
		if actualStudyPlanItems[i].EndDate.Time != expectedStudyPlanItems[i].EndDate.Time {
			return fmt.Errorf("unexpected end date: want %v, got %v", expectedStudyPlanItems[i].EndDate, actualStudyPlanItems[i].EndDate)
		}
		if actualStudyPlanItems[i].StartDate.Time != expectedStudyPlanItems[i].StartDate.Time {
			return fmt.Errorf("unexpected start date: want %v, got %v", expectedStudyPlanItems[i].StartDate, actualStudyPlanItems[i].StartDate)
		}
		if actualStudyPlanItems[i].DisplayOrder.Int != expectedStudyPlanItems[i].DisplayOrder.Int {
			return fmt.Errorf("unexpected display order: want %d, got %d", expectedStudyPlanItems[i].DisplayOrder, actualStudyPlanItems[i].DisplayOrder)
		}
		if err := CompareContentStructure(&actualStudyPlanItems[i].ContentStructure, &expectedStudyPlanItems[i].ContentStructure); err != nil {
			return err
		}
	}
	return nil
}

func CompareStudyPlan(expectedStudyPlan, actualStudyPlan *entities.StudyPlan) error {
	if actualStudyPlan.ID.String != expectedStudyPlan.ID.String {
		return fmt.Errorf("unexpected study plan ids: want: %v, actual: %v", expectedStudyPlan.ID.String, actualStudyPlan.ID.String)
	}
	if actualStudyPlan.Name.String != expectedStudyPlan.Name.String {
		return fmt.Errorf("unexpected study plan names: want: %v, actual: %v", expectedStudyPlan.Name.String, actualStudyPlan.Name.String)
	}
	if actualStudyPlan.MasterStudyPlan.String != expectedStudyPlan.MasterStudyPlan.String {
		return fmt.Errorf("unexpected study plan master ids: want: %v, actual: %v", expectedStudyPlan.MasterStudyPlan.String, actualStudyPlan.MasterStudyPlan.String)
	}
	if actualStudyPlan.CreatedAt.Time.UnixMilli() != expectedStudyPlan.CreatedAt.Time.UnixMilli() {
		return fmt.Errorf("unexpected created at time: want: %v, actual: %v", expectedStudyPlan.CreatedAt.Time, actualStudyPlan.CreatedAt.Time)
	}
	return nil
}

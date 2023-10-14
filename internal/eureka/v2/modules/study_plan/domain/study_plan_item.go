package domain

import (
	"time"
)

type StudyPlanItem struct {
	StudyPlanItemID string
	StudyPlanID     string
	LmListID        string
	LmList          []string
	Name            string
	StartDate       time.Time
	EndDate         time.Time
	DisplayOrder    int
	Status          string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
}

func NewStudyPlanItem(studyPlanItemID, studyPlanID, lmListID, status string, startDate, endDate, createdAt, updatedAt time.Time, deletedAt *time.Time) StudyPlanItem {
	spi := StudyPlanItem{
		StudyPlanItemID: studyPlanItemID,
		StudyPlanID:     studyPlanID,
		LmListID:        lmListID,
		Status:          status,
		StartDate:       startDate,
		EndDate:         endDate,
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
		DeletedAt:       deletedAt,
	}
	return spi
}

package services

import (
	"fmt"
	"strconv"
	"time"

	entities "github.com/manabie-com/backend/internal/eureka/entities/learning_history_data_sync"
	"github.com/manabie-com/backend/internal/golibs/database"
)

type MappingCourseIDRow struct {
	ManabieCourseID string
	WithusCourseID  string
	LastUpdatedDate time.Time
	LastUpdatedBy   string
	IsArchived      string
}

func newMappingCourseIDRow(manabieCourseID, withusCourseID string, lastUpdatedDate time.Time, lastUpdatedBy, isArchived string) (*MappingCourseIDRow, error) {
	if manabieCourseID == "" {
		return nil, fmt.Errorf("manabieCourseID is empty")
	}

	return &MappingCourseIDRow{
		ManabieCourseID: manabieCourseID,
		WithusCourseID:  withusCourseID,
		LastUpdatedDate: lastUpdatedDate,
		LastUpdatedBy:   lastUpdatedBy,
		IsArchived:      isArchived,
	}, nil
}

func (r *MappingCourseIDRow) toEntity() (*entities.MappingCourseID, error) {
	isArchived, err := strconv.ParseBool(r.IsArchived)
	if err != nil {
		return nil, fmt.Errorf("is_archived format is invalid, must be true or false")
	}
	return &entities.MappingCourseID{
		ManabieCourseID: database.Text(r.ManabieCourseID),
		WithusCourseID:  database.Text(r.WithusCourseID),
		LastUpdatedDate: database.Timestamptz(r.LastUpdatedDate),
		LastUpdatedBy:   database.Text(r.LastUpdatedBy),
		IsArchived:      database.Bool(isArchived),
	}, nil
}

type MappingExamLoIDRow struct {
	ExamLoID        string
	MaterialCode    string
	LastUpdatedDate time.Time
	LastUpdatedBy   string
	IsArchived      string
}

func newMappingExamLoIDRow(examLoID, materialCode string, lastUpdatedDate time.Time, lastUpdatedBy, isArchived string) (*MappingExamLoIDRow, error) {
	if examLoID == "" {
		return nil, fmt.Errorf("examLoID is empty")
	}

	return &MappingExamLoIDRow{
		ExamLoID:        examLoID,
		MaterialCode:    materialCode,
		LastUpdatedDate: lastUpdatedDate,
		LastUpdatedBy:   lastUpdatedBy,
		IsArchived:      isArchived,
	}, nil
}

func (r *MappingExamLoIDRow) toEntity() (*entities.MappingExamLoID, error) {
	isArchived, err := strconv.ParseBool(r.IsArchived)
	if err != nil {
		return nil, fmt.Errorf("is_archived format is invalid, must be true or false")
	}
	return &entities.MappingExamLoID{
		ExamLoID:        database.Text(r.ExamLoID),
		MaterialCode:    database.Text(r.MaterialCode),
		LastUpdatedDate: database.Timestamptz(r.LastUpdatedDate),
		LastUpdatedBy:   database.Text(r.LastUpdatedBy),
		IsArchived:      database.Bool(isArchived),
	}, nil
}

type MappingQuestionTagRow struct {
	ManabieTagID    string
	ManabieTagName  string
	WithusTagName   string
	LastUpdatedDate time.Time
	LastUpdatedBy   string
	IsArchived      string
}

func newMappingQuestionTagRow(manabieTagID, manabieTagName, withusTagName string, lastUpdatedDate time.Time, lastUpdatedBy, isArchived string) (*MappingQuestionTagRow, error) {
	if manabieTagID == "" {
		return nil, fmt.Errorf("manabieTagID is empty")
	}
	if manabieTagName == "" {
		return nil, fmt.Errorf("manabieTagName is empty")
	}

	return &MappingQuestionTagRow{
		ManabieTagID:    manabieTagID,
		ManabieTagName:  manabieTagName,
		WithusTagName:   withusTagName,
		LastUpdatedDate: lastUpdatedDate,
		LastUpdatedBy:   lastUpdatedBy,
		IsArchived:      isArchived,
	}, nil
}

func (r *MappingQuestionTagRow) toEntity() (*entities.MappingQuestionTag, error) {
	isArchived, err := strconv.ParseBool(r.IsArchived)
	if err != nil {
		return nil, fmt.Errorf("is_archived format is invalid, must be true or false")
	}
	return &entities.MappingQuestionTag{
		ManabieTagID:    database.Text(r.ManabieTagID),
		ManabieTagName:  database.Text(r.ManabieTagName),
		WithusTagName:   database.Text(r.WithusTagName),
		LastUpdatedDate: database.Timestamptz(r.LastUpdatedDate),
		LastUpdatedBy:   database.Text(r.LastUpdatedBy),
		IsArchived:      database.Bool(isArchived),
	}, nil
}

type FailedSyncEmailRecipientRow struct {
	RecipientID     string
	EmailAddress    string
	LastUpdatedDate time.Time
	LastUpdatedBy   string
	IsArchived      string
}

func newFailedSyncEmailRecipientRow(recipientID, emailAddress string, lastUpdatedDate time.Time, lastUpdatedBy, isArchived string) (*FailedSyncEmailRecipientRow, error) {
	if recipientID == "" {
		return nil, fmt.Errorf("recipientID is empty")
	}
	if emailAddress == "" {
		return nil, fmt.Errorf("emailAddress is empty")
	}

	return &FailedSyncEmailRecipientRow{
		RecipientID:     recipientID,
		EmailAddress:    emailAddress,
		LastUpdatedDate: lastUpdatedDate,
		LastUpdatedBy:   lastUpdatedBy,
		IsArchived:      isArchived,
	}, nil
}

func (r *FailedSyncEmailRecipientRow) toEntity() (*entities.FailedSyncEmailRecipient, error) {
	isArchived, err := strconv.ParseBool(r.IsArchived)
	if err != nil {
		return nil, fmt.Errorf("is_archived format is invalid, must be true or false")
	}
	return &entities.FailedSyncEmailRecipient{
		RecipientID:     database.Text(r.RecipientID),
		EmailAddress:    database.Text(r.EmailAddress),
		LastUpdatedDate: database.Timestamptz(r.LastUpdatedDate),
		LastUpdatedBy:   database.Text(r.LastUpdatedBy),
		IsArchived:      database.Bool(isArchived),
	}, nil
}

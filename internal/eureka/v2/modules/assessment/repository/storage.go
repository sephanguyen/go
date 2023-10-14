package repository

import (
	"context"
	"time"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/assessment/domain"
	"github.com/manabie-com/backend/internal/golibs/database"
)

type AssessmentSessionRepo interface {
	GetLatestByIdentity(ctx context.Context, db database.Ext, assessmentID, userID string) (domain.Session, error)
	GetManyByAssessments(ctx context.Context, db database.Ext, assessmentID, userID string) ([]domain.Session, error)
	Insert(ctx context.Context, db database.Ext, now time.Time, session domain.Session) error
	UpdateStatus(ctx context.Context, db database.Ext, now time.Time, session domain.Session) error
	GetByID(ctx context.Context, db database.Ext, id string) (domain.Session, error)
}

type AssessmentRepo interface {
	GetManyByIDs(ctx context.Context, db database.Ext, ids []string) ([]domain.Assessment, error)
	GetManyByLMAndCourseIDs(ctx context.Context, db database.Ext, assessments []domain.Assessment) ([]domain.Assessment, error)
	GetOneByLMAndCourseID(ctx context.Context, db database.Ext, courseID, lmID string) (*domain.Assessment, error)
	Upsert(ctx context.Context, db database.Ext, now time.Time, assessment domain.Assessment) (string, error)
	GetVirtualByID(ctx context.Context, db database.Ext, id string) (domain.Assessment, error)
}

type StudentEventLogRepo interface {
	GetManyByEventTypesAndLMs(ctx context.Context, db database.Ext, courseID, userID string, eventTypes, lmIDs []string) ([]domain.StudentEventLog, error)
}

type SubmissionRepo interface {
	GetOneBySessionID(ctx context.Context, db database.Ext, sessionID string) (*domain.Submission, error)
	GetOneBySubmissionID(ctx context.Context, db database.Ext, sessionID string) (*domain.Submission, error)
	GetManyBySessionIDs(ctx context.Context, db database.Ext, sessionIDs []string) ([]domain.Submission, error)
	GetManyByAssessments(ctx context.Context, db database.Ext, studentID, asmID string) (subs []domain.Submission, err error)
	Insert(ctx context.Context, db database.Ext, now time.Time, submission domain.Submission) error
	UpdateAllocateMarkerSubmissions(ctx context.Context, db database.Ext, submissions []domain.Submission) error
}

type FeedbackSessionRepo interface {
	GetOneBySubmissionID(ctx context.Context, db database.Ext, sub string) (*domain.FeedbackSession, error)
	GetManyBySubmissionIDs(ctx context.Context, db database.Ext, subs []string) ([]domain.FeedbackSession, error)
	Insert(ctx context.Context, db database.Ext, feedback domain.FeedbackSession) error
}

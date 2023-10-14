package infrastructure

import (
	"context"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	lesson_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson_report/domain"

	"github.com/jackc/pgtype"
)

type LessonReportRepo interface {
	DeleteReportsBelongToLesson(ctx context.Context, db database.Ext, lessonIDs []string) error
	Create(ctx context.Context, db database.Ext, report *domain.LessonReport) (*domain.LessonReport, error)
	Update(ctx context.Context, db database.Ext, report *domain.LessonReport) (*domain.LessonReport, error)
	FindByID(ctx context.Context, db database.Ext, id string) (*domain.LessonReport, error)
	FindByLessonID(ctx context.Context, db database.Ext, lessonID string) (*domain.LessonReport, error)
	FindByResourcePath(ctx context.Context, db database.Ext, resourcePath string, limit int, offSet int) (domain.LessonReports, error)
}

type LessonReportDetailRepo interface {
	GetByLessonReportID(ctx context.Context, db database.Ext, lessonReportID string) (domain.LessonReportDetails, error)
	GetDetailByLessonReportID(ctx context.Context, db database.Ext, lessonReportID string) (domain.LessonReportDetails, error)
	GetReportVersionByLessonID(ctx context.Context, db database.Ext, lessonID string) (domain.LessonReportDetails, error)
	Upsert(ctx context.Context, db database.Ext, lessonReportID string, details domain.LessonReportDetails) error
	UpsertFieldValues(ctx context.Context, db database.Ext, values []*domain.PartnerDynamicFormFieldValue) error
	UpsertWithVersion(ctx context.Context, db database.Ext, lessonReportID string, details domain.LessonReportDetails) error
	UpsertOne(ctx context.Context, db database.Ext, lessonReportID string, details domain.LessonReportDetail) error
}

type PartnerFormConfigRepo interface {
	FindByPartnerAndFeatureName(ctx context.Context, db database.Ext, partnerID int, featureName string) (*domain.PartnerFormConfig, error)
	DeleteByLessonReportDetailIDs(ctx context.Context, db database.Ext, ids []string) error
	GetMapStudentFieldValuesByDetailID(ctx context.Context, db database.Ext, lessonReportDetailId string) (map[string]domain.LessonReportFields, error)
	CreatePartnerFormConfig(ctx context.Context, db database.Ext, partnerFormConfig *domain.PartnerFormConfig) error
}
type LessonRepo interface {
	FindByID(ctx context.Context, db database.QueryExecer, id string) (*lesson_domain.Lesson, error)
	GetLearnerIDsOfLesson(ctx context.Context, db database.QueryExecer, lessonID string) ([]string, error)
}

type ConfigRepo interface {
	Retrieve(ctx context.Context, db database.QueryExecer, country pgtype.Text, group pgtype.Text, keys pgtype.TextArray) ([]*entities.Config, error)
}

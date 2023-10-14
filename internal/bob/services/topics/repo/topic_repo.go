package repo

import (
	"context"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type TopicRepo interface {
	Create(ctx context.Context, db database.Ext, plans []*entities.Topic) error
	UpdateNameByLessonID(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, newName pgtype.Text) error
	SoftDeleteV2(ctx context.Context, db database.QueryExecer, topicIDs pgtype.TextArray) error
	SoftDeleteByPresetStudyPlanWeeklyIDs(ctx context.Context, db database.Ext, pspwIDs pgtype.TextArray) error
}

var _ TopicRepo = new(TopicRepoMock)

type TopicRepoMock struct {
	CreateMock                               func(ctx context.Context, db database.Ext, plans []*entities.Topic) error
	UpdateNameByLessonIDMock                 func(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, newName pgtype.Text) error
	SoftDeleteV2Mock                         func(ctx context.Context, db database.QueryExecer, topicIDs pgtype.TextArray) error
	SoftDeleteByPresetStudyPlanWeeklyIDsMock func(ctx context.Context, db database.Ext, pspwIDs pgtype.TextArray) error
}

func (t TopicRepoMock) Create(ctx context.Context, db database.Ext, plans []*entities.Topic) error {
	return t.CreateMock(ctx, db, plans)
}

func (t TopicRepoMock) UpdateNameByLessonID(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, newName pgtype.Text) error {
	return t.UpdateNameByLessonIDMock(ctx, db, lessonID, newName)
}

func (t TopicRepoMock) SoftDeleteV2(ctx context.Context, db database.QueryExecer, topicIDs pgtype.TextArray) error {
	return t.SoftDeleteV2Mock(ctx, db, topicIDs)
}

func (t TopicRepoMock) SoftDeleteByPresetStudyPlanWeeklyIDs(ctx context.Context, db database.Ext, pspwIDs pgtype.TextArray) error {
	return t.SoftDeleteByPresetStudyPlanWeeklyIDsMock(ctx, db, pspwIDs)
}

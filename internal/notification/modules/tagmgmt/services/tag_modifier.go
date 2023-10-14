package services

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/notification/modules/tagmgmt/entities"
	tagRepo "github.com/manabie-com/backend/internal/notification/modules/tagmgmt/repositories"
	notiRepo "github.com/manabie-com/backend/internal/notification/repositories"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/jackc/pgtype"
)

type TagMgmtModifierService struct {
	npb.UnimplementedTagMgmtModifierServiceServer
	DB      database.Ext
	TagRepo interface {
		DoesTagNameExist(ctx context.Context, db database.QueryExecer, name pgtype.Text) (bool, error)
		Upsert(ctx context.Context, db database.QueryExecer, tag *entities.Tag) error
		SoftDelete(ctx context.Context, db database.QueryExecer, tagIds pgtype.TextArray) error
		FindByID(ctx context.Context, db database.QueryExecer, tagID pgtype.Text) (*entities.Tag, error)
		BulkUpsert(ctx context.Context, db database.QueryExecer, items []*entities.Tag) error
		FindDuplicateTagNames(ctx context.Context, db database.QueryExecer, tags []*entities.Tag) (map[string]string, error)
		FindTagIDsNotExist(ctx context.Context, db database.QueryExecer, tagIDs pgtype.TextArray) ([]string, error)
	}
	InfoNotificationTagRepo interface {
		SoftDelete(ctx context.Context, db database.QueryExecer, filter *notiRepo.SoftDeleteNotificationTagFilter) error
	}
}

func NewTagModifierService(db database.Ext) *TagMgmtModifierService {
	return &TagMgmtModifierService{
		DB:                      db,
		TagRepo:                 &tagRepo.TagRepo{},
		InfoNotificationTagRepo: &notiRepo.InfoNotificationTagRepo{},
	}
}

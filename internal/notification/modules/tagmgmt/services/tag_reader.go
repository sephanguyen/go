package services

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/notification/modules/tagmgmt/entities"
	tagRepo "github.com/manabie-com/backend/internal/notification/modules/tagmgmt/repositories"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/jackc/pgtype"
)

type TagMgmtReaderService struct {
	npb.UnimplementedTagMgmtReaderServiceServer
	DB      database.Ext
	TagRepo interface {
		DoesTagNameExist(ctx context.Context, db database.QueryExecer, name pgtype.Text) (bool, error)
		Upsert(ctx context.Context, db database.QueryExecer, tag *entities.Tag) error
		FindByFilter(ctx context.Context, db database.QueryExecer, filter tagRepo.FindTagFilter) (entities.Tags, uint32, error)
	}
}

func NewTagReaderService(db database.Ext) *TagMgmtReaderService {
	return &TagMgmtReaderService{
		DB:      db,
		TagRepo: &tagRepo.TagRepo{},
	}
}

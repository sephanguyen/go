package services

import (
	"context"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ChapterReaderService struct {
	pb.UnimplementedChapterReaderServiceServer
	DB database.Ext

	ChapterRepo interface {
		ListChapters(context.Context, database.QueryExecer, *repositories.ListChaptersArgs) ([]*entities.Chapter, error)
	}
}

func NewChapterReaderService(
	db database.Ext,
) *ChapterReaderService {
	return &ChapterReaderService{
		DB:          db,
		ChapterRepo: &repositories.ChapterRepo{},
	}
}

// ListChapters returns list of chapter
func (c *ChapterReaderService) ListChapters(ctx context.Context, req *pb.ListChaptersRequest) (*pb.ListChaptersResponse, error) {
	args := &repositories.ListChaptersArgs{
		ChapterIDs: pgtype.TextArray{Status: pgtype.Null},
		Limit:      10,
		Offset:     pgtype.Int4{Status: pgtype.Null},
		ChapterID:  pgtype.Text{Status: pgtype.Null},
	}

	if req != nil {
		if req.Filter != nil && len(req.Filter.Ids) > 0 {
			args.ChapterIDs.Set(req.Filter.Ids)
		}
		if paging := req.Paging; paging != nil {
			if limit := paging.Limit; 1 <= limit {
				args.Limit = limit
			}
			if c := paging.GetOffsetCombined(); c != nil {
				if c.OffsetInteger > 0 {
					args.Offset.Set(c.OffsetInteger)
				}
				if c.OffsetString != "" {
					args.ChapterID.Set(c.OffsetString)
				}
			}
		}
	}
	chapters, err := c.ChapterRepo.ListChapters(ctx, c.DB, args)
	if err != nil {
		return nil, err
	}
	if len(chapters) == 0 {
		return &pb.ListChaptersResponse{}, nil
	}
	pbChapters := make([]*cpb.Chapter, 0, len(chapters))
	for _, chapter := range chapters {
		pbChapters = append(pbChapters, toChapterPb(chapter))
	}
	lastItem := chapters[len(chapters)-1]
	return &pb.ListChaptersResponse{
		Items: pbChapters,
		NextPage: &cpb.Paging{
			Limit: args.Limit,
			Offset: &cpb.Paging_OffsetCombined{
				OffsetCombined: &cpb.Paging_Combined{
					OffsetInteger: int64(lastItem.DisplayOrder.Int),
					OffsetString:  lastItem.ID.String,
				},
			},
		},
	}, nil
}

func toChapterPb(e *entities.Chapter) *cpb.Chapter {
	return &cpb.Chapter{
		Info: &cpb.ContentBasicInfo{
			Id:           e.ID.String,
			Name:         e.Name.String,
			Country:      cpb.Country(cpb.Country_value[e.Country.String]),
			Subject:      cpb.Subject(cpb.Country_value[e.Subject.String]),
			Grade:        int32(e.Grade.Int),
			SchoolId:     e.SchoolID.Int,
			DisplayOrder: int32(e.DisplayOrder.Int),
			CreatedAt:    timestamppb.New(e.CreatedAt.Time),
			UpdatedAt:    timestamppb.New(e.UpdatedAt.Time),
		},
	}
}

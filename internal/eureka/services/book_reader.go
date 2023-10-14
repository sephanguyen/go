package services

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	usermgmt_entities "github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type BookReaderService struct {
	DB database.Ext

	StudentStudyPlanRepo interface {
		GetBookIDsBelongsToStudentStudyPlan(ctx context.Context, db database.QueryExecer, studentID pgtype.Text, bookIDs pgtype.TextArray) ([]string, error)
	}
	BookRepo interface {
		ListBooks(ctx context.Context, db database.QueryExecer, args *repositories.ListBooksArgs) ([]*entities.Book, error)
	}
}

func NewBookReaderService(
	db database.Ext,
) *BookReaderService {
	return &BookReaderService{
		DB:                   db,
		StudentStudyPlanRepo: &repositories.StudentStudyPlanRepo{},
		BookRepo:             &repositories.BookRepo{},
	}
}

func (crs *BookReaderService) ListBooks(ctx context.Context, req *pb.ListBooksRequest) (*pb.ListBooksResponse, error) {
	args := &repositories.ListBooksArgs{
		BookIDs:               pgtype.TextArray{Status: pgtype.Null},
		Limit:                 10,
		Offset:                pgtype.Timestamptz{Status: pgtype.Null},
		BookID:                pgtype.Text{Status: pgtype.Null},
		StudentStudyPlanBooks: pgtype.TextArray{Status: pgtype.Null},
	}

	if interceptors.UserGroupFromContext(ctx) == usermgmt_entities.UserGroupStudent && len(req.Filter.Ids) > 0 {
		var (
			studentIDReq pgtype.Text
			bookIDsReq   pgtype.TextArray
		)

		err := multierr.Combine(
			studentIDReq.Set(interceptors.UserIDFromContext(ctx)),
			bookIDsReq.Set(req.Filter.Ids),
		)
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Errorf("ListBooks: value invalid").Error())
		}

		bookIDs, err := crs.StudentStudyPlanRepo.GetBookIDsBelongsToStudentStudyPlan(ctx, crs.DB, studentIDReq, bookIDsReq)
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Errorf("crs.StudentStudyPlanRepo.GetBookIDsBelongsToStudentStudyPlan: %w", err).Error())
		}

		if len(bookIDs) > 0 {
			args.StudentStudyPlanBooks.Set(bookIDs)
		}
	}

	if req != nil {
		if req.Filter != nil && len(req.Filter.Ids) > 0 {
			args.BookIDs.Set(req.Filter.Ids)
		}
		if paging := req.Paging; paging != nil {
			if limit := paging.Limit; 1 <= limit && limit <= 100 {
				args.Limit = limit
			}
			if c := paging.GetOffsetCombined(); c != nil {
				if c.OffsetTime != nil && c.OffsetTime.AsTime().Unix() > 0 {
					args.Offset.Set(c.OffsetTime.AsTime())
				}
				if c.OffsetString != "" {
					args.BookID.Set(c.OffsetString)
				}
			}
		}
	}

	books, err := crs.BookRepo.ListBooks(ctx, crs.DB, args)
	if err != nil {
		return nil, err
	}
	if len(books) == 0 {
		return &pb.ListBooksResponse{}, nil
	}

	pbBooks := make([]*cpb.Book, 0, len(books))
	for _, book := range books {
		pbBooks = append(pbBooks, toBookPb(book))
	}

	lastItem := books[len(books)-1]
	return &pb.ListBooksResponse{
		Items: pbBooks,
		NextPage: &cpb.Paging{
			Limit: args.Limit,
			Offset: &cpb.Paging_OffsetCombined{
				OffsetCombined: &cpb.Paging_Combined{
					OffsetTime:   timestamppb.New(lastItem.CreatedAt.Time),
					OffsetString: lastItem.ID.String,
				},
			},
		},
	}, nil
}

func toBookPb(e *entities.Book) *cpb.Book {
	return &cpb.Book{
		Info: &cpb.ContentBasicInfo{
			Id:        e.ID.String,
			Name:      e.Name.String,
			Country:   cpb.Country(cpb.Country_value[e.Country.String]),
			Subject:   cpb.Subject(cpb.Country_value[e.Subject.String]),
			Grade:     int32(e.Grade.Int),
			SchoolId:  e.SchoolID.Int,
			CreatedAt: timestamppb.New(e.CreatedAt.Time),
			UpdatedAt: timestamppb.New(e.UpdatedAt.Time),
		},
	}
}

package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/discount/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	pb "github.com/manabie-com/backend/pkg/manabuf/discount/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *DiscountService) UpsertStudentDiscountTag(ctx context.Context, req *pb.UpsertStudentDiscountTagRequest) (*pb.UpsertStudentDiscountTagResponse, error) {
	// validate request
	if strings.TrimSpace(req.StudentId) == "" {
		return nil, status.Error(codes.FailedPrecondition, "student id should be required")
	}
	// retrieve all user discount tag by student id
	userDiscountTags, err := s.DiscountTagService.RetrieveEligibleDiscountTagsOfStudent(ctx, s.DB, req.StudentId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) (err error) {
		// if discount_tag_ids are empty, mark the student user_discount_tag records as deleted
		if len(req.DiscountTagIds) == 0 || req.DiscountTagIds == nil {
			err := s.filterAndDeleteUserDiscountTags(ctx, tx, userDiscountTags, req.StudentId)
			if err != nil {
				return err
			}

			return nil
		}
		// if there are discount tag ids, check if it's not existing and do creation of user discount tag
		discountTagsUpserted, err := s.processUpsertDiscountTags(ctx, tx, req, userDiscountTags)
		if err != nil {
			return err
		}
		// delete existing user discount tag that are not existing on discount tag id request payload
		deleteUserDiscountTags := make([]*entities.UserDiscountTag, 0)
		for _, userDiscountTag := range userDiscountTags {
			var userDiscountTagExist bool
			for _, discountTagUpserted := range discountTagsUpserted {
				if userDiscountTag.DiscountTagID.String == discountTagUpserted.DiscountTagID.String {
					userDiscountTagExist = true
				}
			}
			if !userDiscountTagExist {
				deleteUserDiscountTags = append(deleteUserDiscountTags, userDiscountTag)
			}
		}

		if len(deleteUserDiscountTags) > 0 {
			err = s.filterAndDeleteUserDiscountTags(ctx, tx, deleteUserDiscountTags, req.StudentId)
			if err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.UpsertStudentDiscountTagResponse{
		Successful: true,
	}, nil
}

func (s *DiscountService) filterAndDeleteUserDiscountTags(ctx context.Context, db database.QueryExecer, userDiscountTags []*entities.UserDiscountTag, studentID string) error {
	filterUserDiscountTagsTypesToDelete := make([]string, 0)

	for _, userDiscountTag := range userDiscountTags {
		filterUserDiscountTagsTypesToDelete = append(filterUserDiscountTagsTypesToDelete, userDiscountTag.DiscountType.String)
	}

	// if there user discount tags types to delete
	if len(filterUserDiscountTagsTypesToDelete) > 0 {
		err := s.DiscountTagService.SoftDeleteUserDiscountTagsByTypesAndUserID(ctx, db, studentID, database.TextArray(filterUserDiscountTagsTypesToDelete))
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *DiscountService) processUpsertDiscountTags(ctx context.Context, db database.QueryExecer, req *pb.UpsertStudentDiscountTagRequest, existingUserDiscountTags []*entities.UserDiscountTag) ([]*entities.UserDiscountTag, error) {
	discountTagsUpserted := make([]*entities.UserDiscountTag, 0)
	for _, discountTagID := range req.DiscountTagIds {
		var discountTagExist bool
		for _, userDiscountTag := range existingUserDiscountTags {
			if userDiscountTag.DiscountTagID.String == discountTagID {
				discountTagExist = true
				discountTagsUpserted = append(discountTagsUpserted, userDiscountTag)
			}
		}

		if discountTagExist {
			// leave the existing record as it is, no values updated
			continue
		}
		// if discount not exist, create new user discount tag record
		discounts, err := s.DiscountRepo.GetByDiscountTagIDs(ctx, db, []string{discountTagID})
		if err != nil {
			return nil, err
		}

		var discountType string
		for _, discount := range discounts {
			if discountType == "" {
				discountType = discount.DiscountType.String
			} else if discountType != discount.DiscountType.String {
				return nil, status.Error(codes.Internal, fmt.Sprintf("there are different discount types on discount table with discount_tag_id: %v", discount.DiscountTagID.String))
			}
		}

		// generate entity to create
		userDiscountTagToCreate := generateUserDiscountTag(req.StudentId, discountType, discountTagID)
		err = s.DiscountTagService.CreateUserDiscountTag(ctx, db, userDiscountTagToCreate)
		if err != nil {
			return nil, err
		}

		discountTagsUpserted = append(discountTagsUpserted, userDiscountTagToCreate)
	}

	return discountTagsUpserted, nil
}

func generateUserDiscountTag(studentID, discountType, discountTagID string) *entities.UserDiscountTag {
	return &entities.UserDiscountTag{
		UserID: pgtype.Text{
			String: studentID,
			Status: pgtype.Present,
		},
		DiscountType: pgtype.Text{
			String: discountType,
			Status: pgtype.Present,
		},
		DiscountTagID: pgtype.Text{
			String: discountTagID,
			Status: pgtype.Present,
		},
		LocationID: pgtype.Text{
			Status: pgtype.Null,
		},
		ProductID: pgtype.Text{
			Status: pgtype.Null,
		},
		StartDate: pgtype.Timestamptz{
			Status: pgtype.Null,
		},
		EndDate: pgtype.Timestamptz{
			Status: pgtype.Null,
		},
		ProductGroupID: pgtype.Text{
			Status: pgtype.Null,
		},
		DeletedAt: pgtype.Timestamptz{
			Status: pgtype.Null,
		},
	}
}

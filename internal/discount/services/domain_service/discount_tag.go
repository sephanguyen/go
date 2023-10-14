package services

import (
	"context"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/discount/entities"
	"github.com/manabie-com/backend/internal/discount/repositories"
	"github.com/manabie-com/backend/internal/discount/utils"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
	"golang.org/x/exp/maps"
)

type DiscountTagService struct {
	DB                  database.Ext
	UserDiscountTagRepo interface {
		GetDiscountTagsByUserIDAndLocationID(
			ctx context.Context,
			db database.QueryExecer,
			userID string,
			locationID string,
		) (
			[]*entities.UserDiscountTag,
			error,
		)
		GetDiscountEligibilityOfStudentProduct(
			ctx context.Context,
			db database.QueryExecer,
			userID string,
			locationID string,
			productID string,
		) (
			[]*entities.UserDiscountTag,
			error,
		)
		GetDiscountTagsWithActivityOnDate(
			ctx context.Context,
			db database.QueryExecer,
			timestamp time.Time,
		) (
			[]*entities.UserDiscountTag,
			error,
		)
		GetUserIDsWithActivityOnDate(
			ctx context.Context,
			db database.QueryExecer,
			timestamp time.Time,
		) (
			[]pgtype.Text,
			error,
		)
		GetActiveDiscountTagIDsByDateAndUserID(
			ctx context.Context,
			db database.QueryExecer,
			timestamp time.Time,
			userID string,
		) (
			[]string,
			error,
		)
		SoftDeleteByTypesAndUserID(
			ctx context.Context,
			db database.QueryExecer,
			userID string,
			discountTypes pgtype.TextArray,
		) error
		Create(
			ctx context.Context,
			db database.QueryExecer,
			e *entities.UserDiscountTag,
		) error
		GetDiscountTagsByUserID(
			ctx context.Context,
			db database.QueryExecer,
			userID string,
		) (
			[]*entities.UserDiscountTag,
			error,
		)
	}
	DiscountTagRepo interface {
		GetByID(
			ctx context.Context,
			db database.QueryExecer,
			discountTagID string,
		) (
			*entities.DiscountTag,
			error,
		)
	}
	DiscountRepo interface {
		GetByDiscountType(
			ctx context.Context,
			db database.QueryExecer,
			discountType string,
		) (
			[]*entities.Discount,
			error,
		)
	}
}

func (s *DiscountTagService) RetrieveEligibleDiscountTagsOfStudentInLocation(
	ctx context.Context,
	db database.QueryExecer,
	userID string,
	locationID string,
) (
	userDiscountTags []*entities.UserDiscountTag,
	err error,
) {
	return s.UserDiscountTagRepo.GetDiscountTagsByUserIDAndLocationID(ctx, db, userID, locationID)
}

func (s *DiscountTagService) RetrieveDiscountEligibilityOfStudentProduct(
	ctx context.Context,
	db database.QueryExecer,
	userID string,
	locationID string,
	productID string,
) (
	userDiscountTags []*entities.UserDiscountTag,
	err error,
) {
	return s.UserDiscountTagRepo.GetDiscountEligibilityOfStudentProduct(ctx, db, userID, locationID, productID)
}

func (s *DiscountTagService) RetrieveDiscountTagsWithActivityOnDate(
	ctx context.Context,
	db database.QueryExecer,
	timestamp time.Time,
) (
	userDiscountTags []*entities.UserDiscountTag,
	err error,
) {
	return s.UserDiscountTagRepo.GetDiscountTagsWithActivityOnDate(ctx, db, timestamp)
}

func (s *DiscountTagService) RetrieveUserIDsWithActivityOnDate(
	ctx context.Context,
	db database.QueryExecer,
	timestamp time.Time,
) (
	userIDs []string,
	err error,
) {
	var pgIDs []pgtype.Text
	userIDs = []string{}

	pgIDs, err = s.UserDiscountTagRepo.GetUserIDsWithActivityOnDate(ctx, db, timestamp)
	if err != nil {
		return
	}

	for _, id := range pgIDs {
		userIDs = append(userIDs, id.String)
	}

	return
}

func NewDiscountTagService(db database.Ext) *DiscountTagService {
	return &DiscountTagService{
		DB:                  db,
		UserDiscountTagRepo: &repositories.UserDiscountTagRepo{},
		DiscountTagRepo:     &repositories.DiscountTagRepo{},
		DiscountRepo:        &repositories.DiscountRepo{},
	}
}

func (s *DiscountTagService) RetrieveActiveDiscountTagIDsByDateAndUserID(
	ctx context.Context,
	db database.QueryExecer,
	timestamp time.Time,
	userID string,
) (
	discountTagIDs []string,
	err error,
) {
	return s.UserDiscountTagRepo.GetActiveDiscountTagIDsByDateAndUserID(ctx, db, timestamp, userID)
}

func (s *DiscountTagService) RetrieveDiscountTagByDiscountTagID(
	ctx context.Context,
	db database.QueryExecer,
	discountTagID string,
) (
	discountTag *entities.DiscountTag,
	err error,
) {
	return s.DiscountTagRepo.GetByID(ctx, db, discountTagID)
}

func (s *DiscountTagService) RetrieveDiscountTagIDsByDiscountType(
	ctx context.Context,
	db database.QueryExecer,
	discountType string,
) (
	discountTagIDs []string,
	err error,
) {
	discounts, err := s.DiscountRepo.GetByDiscountType(ctx, db, discountType)
	if err != nil {
		return
	}

	uniqueDiscountTagIDsMap := map[string]bool{}
	for _, discount := range discounts {
		if discount.DiscountTagID.Status == pgtype.Present {
			uniqueDiscountTagIDsMap[discount.DiscountTagID.String] = true
		}
	}

	discountTagIDs = maps.Keys(uniqueDiscountTagIDsMap)
	return
}

func (s *DiscountTagService) SoftDeleteUserDiscountTagsByTypesAndUserID(
	ctx context.Context,
	db database.QueryExecer,
	userID string,
	discountTypes pgtype.TextArray,
) (
	err error,
) {
	return s.UserDiscountTagRepo.SoftDeleteByTypesAndUserID(ctx, db, userID, discountTypes)
}

func (s *DiscountTagService) CreateUserDiscountTag(
	ctx context.Context,
	db database.QueryExecer,
	userDiscountTag *entities.UserDiscountTag,
) (
	err error,
) {
	return s.UserDiscountTagRepo.Create(ctx, db, userDiscountTag)
}

func (s *DiscountTagService) RetrieveEligibleDiscountTagsOfStudent(
	ctx context.Context,
	db database.QueryExecer,
	userID string,
) (
	userDiscountTags []*entities.UserDiscountTag,
	err error,
) {
	return s.UserDiscountTagRepo.GetDiscountTagsByUserID(ctx, db, userID)
}

func (s *DiscountTagService) UpdateDiscountTagOfStudentIDWithTimeSegment(
	ctx context.Context,
	db database.QueryExecer,
	studentID string,
	discountType string,
	discountTagIDs []string,
	timestampSegments []entities.TimestampSegment,
) (
	err error,
) {
	// soft delete all old user tag data and reinsert new tags
	err = s.SoftDeleteUserDiscountTagsByTypesAndUserID(ctx, db, studentID, database.TextArray([]string{discountType}))
	if err != nil && !strings.Contains(err.Error(), "0 RowsAffected") {
		return
	}

	for _, timeSegment := range timestampSegments {
		for _, discountTagID := range discountTagIDs {
			userDiscountTag := entities.UserDiscountTag{}
			err = utils.GroupErrorFunc(
				userDiscountTag.UserID.Set(studentID),
				userDiscountTag.LocationID.Set(nil),
				userDiscountTag.ProductID.Set(nil),
				userDiscountTag.ProductGroupID.Set(nil),
				userDiscountTag.DiscountType.Set(discountType),
				userDiscountTag.StartDate.Set(timeSegment.StartDate),
				userDiscountTag.EndDate.Set(timeSegment.EndDate),
				userDiscountTag.DiscountTagID.Set(discountTagID),
			)
			if err != nil {
				return
			}

			err = s.CreateUserDiscountTag(ctx, s.DB, &userDiscountTag)
			if err != nil {
				return
			}
		}
	}
	return
}

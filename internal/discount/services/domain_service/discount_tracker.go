package services

import (
	"context"

	"github.com/manabie-com/backend/internal/discount/entities"
	"github.com/manabie-com/backend/internal/discount/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
)

type DiscountTrackerService struct {
	DB                         database.Ext
	StudentDiscountTrackerRepo interface {
		Create(ctx context.Context, db database.QueryExecer, e *entities.StudentDiscountTracker) error
		GetActiveTrackingByStudentIDs(ctx context.Context, db database.QueryExecer, studentID []string) ([]entities.StudentDiscountTracker, error)
		UpdateTrackingDurationByStudentProduct(ctx context.Context, db database.QueryExecer, studentProduct entities.StudentProduct) error
	}
}

func (s *DiscountTrackerService) TrackDiscount(ctx context.Context, db database.QueryExecer, studentDiscontTracker *entities.StudentDiscountTracker) (err error) {
	return s.StudentDiscountTrackerRepo.Create(ctx, db, studentDiscontTracker)
}

func (s *DiscountTrackerService) UpdateTrackingDurationByStudentProduct(ctx context.Context, db database.QueryExecer, studentProduct entities.StudentProduct) error {
	return s.StudentDiscountTrackerRepo.UpdateTrackingDurationByStudentProduct(ctx, db, studentProduct)
}

func (s *DiscountTrackerService) RetrieveSiblingDiscountTrackingHistoriesByStudentIDs(
	ctx context.Context,
	db database.QueryExecer,
	studentIDs []string,
) (
	studentTrackingData map[string][]entities.StudentDiscountTracker,
	siblingTrackingData map[string][]entities.StudentDiscountTracker,
	err error,
) {
	trackingData, err := s.StudentDiscountTrackerRepo.GetActiveTrackingByStudentIDs(ctx, db, studentIDs)
	if err != nil {
		return
	}

	studentTrackingData = map[string][]entities.StudentDiscountTracker{}
	siblingTrackingData = map[string][]entities.StudentDiscountTracker{}

	for _, studentID := range studentIDs {
		studentData := []entities.StudentDiscountTracker{}
		siblingData := []entities.StudentDiscountTracker{}

		for _, data := range trackingData {
			if data.StudentID.String == studentID {
				studentData = append(studentData, data)
			} else {
				siblingData = append(siblingData, data)
			}
		}

		studentTrackingData[studentID] = studentData
		siblingTrackingData[studentID] = siblingData
	}

	return
}

func NewDiscountTrackerService(db database.Ext) *DiscountTrackerService {
	return &DiscountTrackerService{
		DB:                         db,
		StudentDiscountTrackerRepo: &repositories.StudentDiscountTrackerRepo{},
	}
}

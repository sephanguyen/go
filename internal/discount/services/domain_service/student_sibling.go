package services

import (
	"context"

	"github.com/manabie-com/backend/internal/discount/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
)

type StudentSiblingService struct {
	DB                database.Ext
	StudentParentRepo interface {
		GetSiblingIDsByStudentID(ctx context.Context, db database.QueryExecer, studentID string) ([]string, error)
	}
}

func (s *StudentSiblingService) RetrieveStudentSiblingIDs(
	ctx context.Context,
	db database.QueryExecer,
	studentID string,
) (
	[]string,
	error,
) {
	return s.StudentParentRepo.GetSiblingIDsByStudentID(ctx, db, studentID)
}

func NewStudentSiblingService(db database.Ext) *StudentSiblingService {
	return &StudentSiblingService{
		DB:                db,
		StudentParentRepo: &repositories.StudentParentRepo{},
	}
}

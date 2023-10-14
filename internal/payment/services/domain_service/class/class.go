package service

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/repositories"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ClassService struct {
	ClassRepo interface {
		GetMapClassWithLocationByClassIDs(ctx context.Context, db database.QueryExecer, classIDs []string) (mapClass map[string]entities.Class, err error)
	}
}

func (s *ClassService) GetMapClassWithLocationByClassIDs(ctx context.Context, db database.QueryExecer, classIDs []string) (mapClass map[string]entities.Class, err error) {
	mapClass, err = s.ClassRepo.GetMapClassWithLocationByClassIDs(ctx, db, classIDs)
	if err != nil {
		err = status.Errorf(codes.Internal, "can't get classes from classIDs with err: %v", err.Error())
	}
	return
}

func NewClassService() *ClassService {
	return &ClassService{
		ClassRepo: &repositories.ClassRepo{},
	}
}

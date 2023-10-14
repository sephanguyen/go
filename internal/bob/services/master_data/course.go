package services

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	ys_pb "github.com/manabie-com/backend/pkg/genproto/yasuo"
	bpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MasterDataCourseService struct {
	DB            database.Ext
	CourseService interface {
		UpsertCourses(ctx context.Context, req *ys_pb.UpsertCoursesRequest) (*ys_pb.UpsertCoursesResponse, error)
	}
	CourseAccessPathRepo interface {
		Upsert(ctx context.Context, db database.Ext, cc []*entities.CourseAccessPath) error
		FindByCourseIDs(ctx context.Context, db database.QueryExecer, courseIDs []string) (mapLocationIDByCourseID map[string][]string, err error)
		Delete(ctx context.Context, db database.QueryExecer, courseIDs pgtype.TextArray) error
	}
}

func (m *MasterDataCourseService) UpsertCourses(ctx context.Context, req *bpb.UpsertCoursesRequest) (*bpb.UpsertCoursesResponse, error) {
	courseIDs := []string{}
	courses := []*ys_pb.UpsertCoursesRequest_Course{}
	for _, c := range req.Courses {
		courses = append(courses, &ys_pb.UpsertCoursesRequest_Course{
			Id:           c.Id,
			Name:         c.Name,
			Country:      pb.Country(c.Country),
			Subject:      pb.Subject(c.Subject),
			Grade:        c.Grade,
			DisplayOrder: c.DisplayOrder,
			SchoolId:     c.SchoolId,
			BookIds:      c.BookIds,
			Icon:         c.Icon,
		})
		if c.Id != "" {
			courseIDs = append(courseIDs, c.Id)
		}
	}
	res, err := m.CourseService.UpsertCourses(ctx, &ys_pb.UpsertCoursesRequest{
		Courses: courses,
	})
	if err != nil {
		return nil, err
	}
	if res.Successful {
		coursesAP := []*entities.CourseAccessPath{}
		for _, c := range req.Courses {
			for _, location := range c.LocationIds {
				enCourseAP, err := toCourseAccessPathEntity(location, c.Id)
				if err != nil {
					return nil, status.Error(codes.Internal, fmt.Errorf("toCourseAccessPathEntity: %w", err).Error())
				}
				coursesAP = append(coursesAP, enCourseAP)
			}
		}
		if len(courseIDs) > 0 {
			if err := m.CourseAccessPathRepo.Delete(ctx, m.DB, database.TextArray(courseIDs)); err != nil {
				return nil, status.Error(codes.Internal, fmt.Errorf("CourseAccessPathRepo.Delete: %w", err).Error())
			}
		}
		if err := m.CourseAccessPathRepo.Upsert(ctx, m.DB, coursesAP); err != nil {
			return nil, status.Error(codes.Internal, fmt.Errorf("CourseAccessPathRepo.Upsert: %w", err).Error())
		}
	}
	return &bpb.UpsertCoursesResponse{
		Successful: res.Successful,
	}, nil
}

func toCourseAccessPathEntity(locationID, courseID string) (*entities.CourseAccessPath, error) {
	cap := &entities.CourseAccessPath{}
	database.AllNullEntity(cap)
	err := multierr.Combine(
		cap.CourseID.Set(courseID),
		cap.LocationID.Set(locationID),
	)
	if err != nil {
		return nil, err
	}
	return cap, nil
}

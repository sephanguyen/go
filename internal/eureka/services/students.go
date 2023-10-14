package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type StudentsService struct {
	DB database.Ext

	StudentRepo interface {
		FindStudentsByCourseLocation(ctx context.Context, db database.QueryExecer, courseID pgtype.Text, locationIDs pgtype.TextArray) (*pgtype.TextArray, error)
	}

	UserRepo interface {
		GetUsersByIDsAndName(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray, name string, limit, offset uint32) ([]*entities.User, error)
		CountUsersByIDsAndName(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray, name string) (int32, error)
	}
}

func NewStudentsService(db database.Ext) *StudentsService {
	return &StudentsService{
		DB:          db,
		StudentRepo: &repositories.StudentRepo{},
		UserRepo:    &repositories.UserRepo{},
	}
}

func (s *StudentsService) GetStudentsByLocationAndCourse(ctx context.Context, req *pb.GetStudentsByLocationAndCourseRequest) (*pb.GetStudentsByLocationAndCourseResponse, error) {
	if err := s.validateRetrieveStudentByLocationAndCourseRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	studentIds, err := s.StudentRepo.FindStudentsByCourseLocation(ctx, s.DB, database.Text(req.CourseId), database.TextArray(req.LocationIds))
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, status.Errorf(codes.Internal, "StudentsService.RetrieveStudentByLocationAndCourse %v", err)
	}

	offset := req.Paging.GetOffsetInteger()
	limit := req.Paging.Limit

	listStudents, err := s.UserRepo.GetUsersByIDsAndName(ctx, s.DB, *studentIds, req.StudentName, limit, uint32(offset))
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, status.Errorf(codes.Internal, "UserRepo.GetUsersByIDsAndName %v", err)
	}

	totalStudent, err := s.UserRepo.CountUsersByIDsAndName(ctx, s.DB, *studentIds, req.StudentName)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, status.Errorf(codes.Internal, "UserRepo.CountUsersByIDsAndName %v", err)
	}

	responseListStudents := []*pb.GetStudentsByLocationAndCourseResponse_Student{}
	for _, student := range listStudents {
		temp := pb.GetStudentsByLocationAndCourseResponse_Student{
			StudentId: student.UserID.String,
			Name:      student.Name.String,
		}

		responseListStudents = append(responseListStudents, &temp)
	}

	return &pb.GetStudentsByLocationAndCourseResponse{
		NextPage: &cpb.Paging{
			Limit: req.Paging.Limit,
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: req.Paging.GetOffsetInteger() + int64(req.Paging.Limit),
			},
		},
		Students:   responseListStudents,
		TotalItems: uint32(totalStudent),
	}, nil
}

func (s *StudentsService) validateRetrieveStudentByLocationAndCourseRequest(req *pb.GetStudentsByLocationAndCourseRequest) error {
	if req.GetCourseId() == "" {
		return fmt.Errorf("req must have course id")
	}
	if req.Paging == nil {
		return fmt.Errorf("req must have paging field")
	}
	if req.Paging.GetOffsetInteger() < 0 {
		return fmt.Errorf("offset must be positive")
	}
	if req.Paging.Limit <= 0 {
		req.Paging.Limit = 100
	}

	return nil
}

package services

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgtype"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/eureka/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestStudentsService_GetStudentsByLocationAndCourse(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var (
		studentRepo             = new(mock_repositories.MockStudentRepo)
		userRepo                = new(mock_repositories.MockUserRepo)
		db                      = new(mock_database.Ext)
		responseListStudents    = []*pb.GetStudentsByLocationAndCourseResponse_Student{}
		responseListNilStudents = []*pb.GetStudentsByLocationAndCourseResponse_Student{}
	)
	var nilGetStudentsByLocationAndCourseResponse *pb.GetStudentsByLocationAndCourseResponse

	s := StudentsService{
		DB:          db,
		StudentRepo: studentRepo,
		UserRepo:    userRepo,
	}

	studentId1 := database.Text("user-1")
	studentId2 := database.Text("user-2")
	listStudents := []pgtype.Text{studentId1, studentId2}
	listStudentIds := &pgtype.TextArray{
		Elements: listStudents,
	}
	listNilStudentIds := &pgtype.TextArray{}
	UserE1 := &entities.User{
		UserID: database.Text("user-1"),
		Name:   database.Text("user 1"),
	}
	UserE2 := &entities.User{
		UserID: database.Text("user-2"),
		Name:   database.Text("user 2"),
	}
	listUserE := []*entities.User{UserE1, UserE2}
	listNilUserE := []*entities.User{}
	responseStudent1 := pb.GetStudentsByLocationAndCourseResponse_Student{
		StudentId: "user-1",
		Name:      "user 1",
	}
	responseStudent2 := pb.GetStudentsByLocationAndCourseResponse_Student{
		StudentId: "user-2",
		Name:      "user 2",
	}
	responseListStudents = append(responseListStudents, &responseStudent1)
	responseListStudents = append(responseListStudents, &responseStudent2)
	testCases := []TestCase{
		{
			name: "case not get any students by ids",
			ctx:  ctx,
			req: &pb.GetStudentsByLocationAndCourseRequest{
				CourseId:    "course-1",
				StudentName: "",
				Paging: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
			},
			expectedErr: nil,
			expectedResp: &pb.GetStudentsByLocationAndCourseResponse{
				NextPage: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 10,
					},
				},
				Students:   responseListNilStudents,
				TotalItems: 0,
			},
			setup: func(ctx context.Context) {
				studentRepo.On("FindStudentsByCourseLocation", ctx, db, mock.Anything, mock.Anything).
					Return(listStudentIds, nil).Once()
				userRepo.On("GetUsersByIDsAndName", ctx, db, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(listNilUserE, nil).Once()
				userRepo.On("CountUsersByIDsAndName", ctx, db, mock.Anything, mock.Anything).
					Return(int32(0), nil).Once()
			},
		},
		{
			name: "case not have any student in course",
			ctx:  ctx,
			req: &pb.GetStudentsByLocationAndCourseRequest{
				CourseId:    "course-1",
				StudentName: "",
				Paging: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
			},
			expectedErr: nil,
			expectedResp: &pb.GetStudentsByLocationAndCourseResponse{
				NextPage: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 10,
					},
				},
				Students:   responseListNilStudents,
				TotalItems: 0,
			},
			setup: func(ctx context.Context) {
				studentRepo.On("FindStudentsByCourseLocation", ctx, db, mock.Anything, mock.Anything).
					Return(listNilStudentIds, nil).Once()
				userRepo.On("GetUsersByIDsAndName", ctx, db, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(listNilUserE, nil).Once()
				userRepo.On("CountUsersByIDsAndName", ctx, db, mock.Anything, mock.Anything).
					Return(int32(0), nil).Once()
			},
		},
		{
			name: "error case not sent course id",
			ctx:  ctx,
			req: &pb.GetStudentsByLocationAndCourseRequest{
				StudentName: "",
				Paging: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
			},
			expectedErr:  status.Error(codes.InvalidArgument, "req must have course id"),
			expectedResp: nilGetStudentsByLocationAndCourseResponse,
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "error case not sent paging",
			ctx:  ctx,
			req: &pb.GetStudentsByLocationAndCourseRequest{
				CourseId:    "course-1",
				StudentName: "",
			},
			expectedErr:  status.Error(codes.InvalidArgument, "req must have paging field"),
			expectedResp: nilGetStudentsByLocationAndCourseResponse,
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "happy case get list locations success",
			ctx:  ctx,
			req: &pb.GetStudentsByLocationAndCourseRequest{
				CourseId:    "course-1",
				StudentName: "",
				Paging: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
			},
			expectedErr: nil,
			expectedResp: &pb.GetStudentsByLocationAndCourseResponse{
				NextPage: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 10,
					},
				},
				Students:   responseListStudents,
				TotalItems: 2,
			},
			setup: func(ctx context.Context) {
				studentRepo.On("FindStudentsByCourseLocation", ctx, db, mock.Anything, mock.Anything).
					Return(listStudentIds, nil).Once()
				userRepo.On("GetUsersByIDsAndName", ctx, db, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(listUserE, nil).Once()
				userRepo.On("CountUsersByIDsAndName", ctx, db, mock.Anything, mock.Anything).
					Return(int32(2), nil).Once()
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*pb.GetStudentsByLocationAndCourseRequest)
			resp, err := s.GetStudentsByLocationAndCourse(testCase.ctx, req)
			assert.Equal(t, testCase.expectedErr, err)
			assert.Equal(t, testCase.expectedResp, resp)
		})
	}
}

package services

import (
	"context"
	"testing"
	"time"

	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	yasuo_ser "github.com/manabie-com/backend/internal/yasuo/services"
	mock_repositories "github.com/manabie-com/backend/mock/bob/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	pb_bob "github.com/manabie-com/backend/pkg/genproto/bob"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	bpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TestCase struct {
	name         string
	ctx          context.Context
	req          interface{}
	expectedResp interface{}
	expectedErr  error
	setup        func(ctx context.Context)
}

func TestMasterDataCourseService_UpsertCourses(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	now := time.Now().UTC()
	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	courseAccessPathRepo := new(mock_repositories.MockCourseAccessPathRepo)
	courseRepo := new(mock_repositories.MockCourseRepo)
	courseBookRepo := new(mock_repositories.MockCourseBookRepo)

	s := &MasterDataCourseService{
		CourseAccessPathRepo: courseAccessPathRepo,
		DB:                   db,
		CourseService: &yasuo_ser.CourseService{
			DBTrace:        db,
			CourseRepo:     courseRepo,
			CourseBookRepo: courseBookRepo,
		},
	}
	removeCourseAPIDs := map[string][]string{}
	removeCourseAPIDs["course-1"] = []string{"location-1", "location-2"}
	testCases := []TestCase{
		{
			name: "happy case update",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &bpb.UpsertCoursesRequest{Courses: []*bpb.UpsertCoursesRequest_Course{
				{
					Id:          "course-1",
					Name:        "course-name",
					LocationIds: []string{"location-1", "location-2"},
					SchoolId:    constant.ManabieSchool,
					Grade:       "Grade 12",
					Country:     cpb.Country_COUNTRY_MASTER,
					Subject:     cpb.Subject_SUBJECT_ENGLISH,
				},
			}},
			expectedErr: nil,
			expectedResp: &bpb.UpsertCoursesResponse{
				Successful: true,
			},
			setup: func(ctx context.Context) {
				courseRepo.On("FindByIDs", ctx, mock.Anything, database.TextArray([]string{"course-1"})).Once().
					Return(
						map[pgtype.Text]*entities_bob.Course{
							database.Text("course-1"): {
								ID:                database.Text("id"),
								Name:              database.Text("current name"),
								Country:           database.Text(pb_bob.COUNTRY_NONE.String()),
								Subject:           database.Text(pb_bob.SUBJECT_NONE.String()),
								Grade:             database.Int2(3),
								DisplayOrder:      database.Int2(2),
								SchoolID:          database.Int4(2),
								TeacherIDs:        database.TextArray([]string{"teacher-1", "teacher-2"}),
								CourseType:        database.Text(pb_bob.COURSE_TYPE_NONE.String()),
								Icon:              database.Text("icon"),
								UpdatedAt:         database.Timestamptz(now),
								CreatedAt:         database.Timestamptz(now),
								DeletedAt:         pgtype.Timestamptz{Status: pgtype.Null},
								StartDate:         database.Timestamptz(now),
								EndDate:           database.Timestamptz(now.Add(2 * time.Minute)),
								PresetStudyPlanID: database.Text("preset-study-plan-id"),
								Status:            database.Text(cpb.CourseStatus_COURSE_STATUS_ACTIVE.String()),
							},
						},
						nil,
					)
				courseBookRepo.On("FindByCourseIDs", ctx, mock.Anything, mock.Anything).Once().Return(map[string][]string{
					"id": {"course-1"},
				}, nil)

				db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
				courseRepo.On("Upsert", ctx, mock.Anything, mock.Anything).Once().Return(nil)
				courseBookRepo.On("SoftDelete", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				courseBookRepo.On("Upsert", ctx, mock.Anything, mock.Anything).Once().Return(nil)
				courseAccessPathRepo.On("Delete", ctx, mock.Anything, mock.Anything).Once().Return(nil)
				courseAccessPathRepo.On("Upsert", ctx, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "happy case insert",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &bpb.UpsertCoursesRequest{Courses: []*bpb.UpsertCoursesRequest_Course{
				{
					Id:          "course-1",
					Name:        "course-name",
					LocationIds: []string{"location-1", "location-2"},
					SchoolId:    constant.ManabieSchool,
					Grade:       "Grade 12",
					Country:     cpb.Country_COUNTRY_MASTER,
					Subject:     cpb.Subject_SUBJECT_ENGLISH,
				},
			}},
			expectedErr: nil,
			expectedResp: &bpb.UpsertCoursesResponse{
				Successful: true,
			},
			setup: func(ctx context.Context) {
				courseRepo.On("FindByIDs", ctx, mock.Anything, database.TextArray([]string{"course-1"})).Once().
					Return(
						nil,
						nil,
					)
				courseBookRepo.On("FindByCourseIDs", ctx, mock.Anything, mock.Anything).Once().Return(map[string][]string{
					"id": {"course-1"},
				}, nil)

				db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
				courseRepo.On("Upsert", ctx, mock.Anything, mock.Anything).Once().Return(nil)
				courseBookRepo.On("SoftDelete", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				courseBookRepo.On("Upsert", ctx, mock.Anything, mock.Anything).Once().Return(nil)
				courseAccessPathRepo.On("Delete", ctx, mock.Anything, mock.Anything).Once().Return(nil)
				courseAccessPathRepo.On("Upsert", ctx, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "missing course name",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &bpb.UpsertCoursesRequest{Courses: []*bpb.UpsertCoursesRequest_Course{
				{
					Id:          "course-1",
					LocationIds: []string{"location-1", "location-2"},
					SchoolId:    constant.ManabieSchool,
					Grade:       "Grade 12",
					Country:     cpb.Country_COUNTRY_MASTER,
					Subject:     cpb.Subject_SUBJECT_ENGLISH,
				},
			}},
			expectedErr: status.Error(codes.InvalidArgument, "course name cannot be empty"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "missing course country",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &bpb.UpsertCoursesRequest{Courses: []*bpb.UpsertCoursesRequest_Course{
				{
					Id:          "course-1",
					Name:        "course-name",
					LocationIds: []string{"location-1", "location-2"},
					SchoolId:    constant.ManabieSchool,
					Grade:       "Grade 12",
					Subject:     cpb.Subject_SUBJECT_ENGLISH,
				},
			}},
			expectedErr: status.Error(codes.InvalidArgument, "cannot find country grade map"),
			setup: func(ctx context.Context) {
				courseRepo.On("FindByIDs", ctx, mock.Anything, database.TextArray([]string{"course-1"})).Once().
					Return(
						nil,
						nil,
					)
			},
		},
		{
			name: "missing school id of course name",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &bpb.UpsertCoursesRequest{Courses: []*bpb.UpsertCoursesRequest_Course{
				{
					Id:          "course-1",
					Name:        "name",
					LocationIds: []string{"location-1", "location-2"},
					Grade:       "Grade 12",
					Country:     cpb.Country_COUNTRY_MASTER,
					Subject:     cpb.Subject_SUBJECT_ENGLISH,
				},
			}},
			expectedErr: status.Error(codes.InvalidArgument, "missing school id of course name"),
			setup: func(ctx context.Context) {
				courseRepo.On("FindByIDs", ctx, mock.Anything, database.TextArray([]string{"course-1"})).Once().
					Return(
						nil,
						nil,
					)
			},
		},
		{
			name: "course repo upsert failed",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &bpb.UpsertCoursesRequest{Courses: []*bpb.UpsertCoursesRequest_Course{
				{
					Id:          "course-1",
					Name:        "course-name",
					LocationIds: []string{"location-1", "location-2"},
					SchoolId:    constant.ManabieSchool,
					Grade:       "Grade 12",
					Country:     cpb.Country_COUNTRY_MASTER,
					Subject:     cpb.Subject_SUBJECT_ENGLISH,
				},
			}},
			expectedErr: errors.Wrap(status.Error(codes.Internal, "upsert course failed"), "s.CourseRepo.Upsert"),
			setup: func(ctx context.Context) {
				courseRepo.On("FindByIDs", ctx, mock.Anything, database.TextArray([]string{"course-1"})).Once().
					Return(
						map[pgtype.Text]*entities_bob.Course{
							database.Text("course-1"): {
								ID:                database.Text("id"),
								Name:              database.Text("current name"),
								Country:           database.Text(pb_bob.COUNTRY_NONE.String()),
								Subject:           database.Text(pb_bob.SUBJECT_NONE.String()),
								Grade:             database.Int2(3),
								DisplayOrder:      database.Int2(2),
								SchoolID:          database.Int4(2),
								TeacherIDs:        database.TextArray([]string{"teacher-1", "teacher-2"}),
								CourseType:        database.Text(pb_bob.COURSE_TYPE_NONE.String()),
								Icon:              database.Text("icon"),
								UpdatedAt:         database.Timestamptz(now),
								CreatedAt:         database.Timestamptz(now),
								DeletedAt:         pgtype.Timestamptz{Status: pgtype.Null},
								StartDate:         database.Timestamptz(now),
								EndDate:           database.Timestamptz(now.Add(2 * time.Minute)),
								PresetStudyPlanID: database.Text("preset-study-plan-id"),
								Status:            database.Text(cpb.CourseStatus_COURSE_STATUS_ACTIVE.String()),
							},
						},
						nil,
					)
				courseRepo.On("Upsert", ctx, mock.Anything, mock.Anything).Once().Return(status.Error(codes.Internal, "upsert course failed"))
			},
		},
	}
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*bpb.UpsertCoursesRequest)
			resp, err := s.UpsertCourses(testCase.ctx, req)
			if testCase.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}

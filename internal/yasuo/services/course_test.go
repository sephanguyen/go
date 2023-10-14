package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	mock_bob_repositories "github.com/manabie-com/backend/mock/bob/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/yasuo/repositories"
	pb_bob "github.com/manabie-com/backend/pkg/genproto/bob"
	pb "github.com/manabie-com/backend/pkg/genproto/yasuo"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var ErrCanNotFindCountryGradeMap = status.Error(codes.InvalidArgument, "cannot find country grade map")
var ErrCourseRepoUpsert = status.Error(codes.Internal, "upsert course failed")
var ErrCourseBookRepoUpsert = status.Error(codes.Internal, "upsert course book failed")
var ErrCourseBookRepoSoftDelete = status.Error(codes.Internal, "soft delete course book failed")

func TestUpsertCoursesV2_Error(t *testing.T) {
	t.Parallel()
	courseBookRepo := &mock_bob_repositories.MockCourseBookRepo{}
	courseRepo := &mock_bob_repositories.MockCourseRepo{}
	now := time.Now().UTC()

	mockDB := &mock_database.Ext{}
	mockTxer := &mock_database.Tx{}

	courseService := &CourseService{
		CourseRepo:     courseRepo,
		CourseBookRepo: courseBookRepo,
		DBTrace:        &database.DBTrace{DB: mockDB},
	}

	testCases := map[string]TestCase{
		"missing course name": {
			req: &pb.UpsertCoursesRequest{
				Courses: []*pb.UpsertCoursesRequest_Course{{
					Id:       "id",
					Name:     "",
					Country:  pb_bob.COUNTRY_VN,
					Subject:  pb_bob.SUBJECT_BIOLOGY,
					SchoolId: constant.ManabieSchool,
					Grade:    "Lớp 1",
				}},
			},
			expectedErr: status.Error(codes.InvalidArgument, "course name cannot be empty"),
			setup: func(ctx context.Context) {
			},
		},
		// // make sure it work after remove
		"missing course country": {
			req: &pb.UpsertCoursesRequest{
				Courses: []*pb.UpsertCoursesRequest_Course{{
					Id:       "id",
					Name:     "name",
					Subject:  pb_bob.SUBJECT_BIOLOGY,
					SchoolId: constant.ManabieSchool,
					Grade:    "Lớp 1",
				}},
			},
			expectedErr: ErrCanNotFindCountryGradeMap,
			setup: func(ctx context.Context) {
				courseRepo.On("FindByIDs", ctx, mock.Anything, database.TextArray([]string{"id"})).Once().
					Return(
						map[pgtype.Text]*entities_bob.Course{
							database.Text("id"): {
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
			},
		},
		"missing school id of course name": {
			req: &pb.UpsertCoursesRequest{
				Courses: []*pb.UpsertCoursesRequest_Course{{
					Id:       "id",
					Name:     "name",
					Country:  pb_bob.COUNTRY_VN,
					Subject:  pb_bob.SUBJECT_BIOLOGY,
					SchoolId: 0,
					Grade:    "Lớp 1",
				}},
			},
			expectedErr: status.Error(codes.InvalidArgument, "missing school id of course name"),
			setup: func(ctx context.Context) {
			},
		},
		"course repo upsert failed": {
			req: &pb.UpsertCoursesRequest{
				Courses: []*pb.UpsertCoursesRequest_Course{{
					Id:           "id",
					Name:         "name",
					Country:      pb_bob.COUNTRY_VN,
					Subject:      pb_bob.SUBJECT_BIOLOGY,
					SchoolId:     constant.ManabieSchool,
					Grade:        "Lớp 1",
					DisplayOrder: 1,
					Icon:         "icon-1",
				}},
			},
			expectedErr: errors.Wrap(ErrCourseRepoUpsert, "s.CourseRepo.Upsert"),
			setup: func(ctx context.Context) {
				courseRepo.On("FindByIDs", ctx, mock.Anything, database.TextArray([]string{"id"})).Once().
					Return(
						map[pgtype.Text]*entities_bob.Course{
							database.Text("id"): {
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
				courseRepo.On("Upsert", ctx, mock.Anything, []*entities_bob.Course{
					{
						ID:                database.Text("id"),
						Name:              database.Text("name"),
						Country:           database.Text(pb_bob.COUNTRY_VN.String()),
						Subject:           database.Text(pb_bob.SUBJECT_BIOLOGY.String()),
						Grade:             database.Int2(1),
						DisplayOrder:      database.Int2(1),
						SchoolID:          database.Int4(constant.ManabieSchool),
						TeacherIDs:        database.TextArray([]string{"teacher-1", "teacher-2"}),
						CourseType:        database.Text(pb_bob.COURSE_TYPE_CONTENT.String()),
						Icon:              database.Text("icon-1"),
						UpdatedAt:         database.Timestamptz(now),
						CreatedAt:         database.Timestamptz(now),
						DeletedAt:         pgtype.Timestamptz{Status: pgtype.Null},
						StartDate:         database.Timestamptz(now),
						EndDate:           database.Timestamptz(now.Add(2 * time.Minute)),
						PresetStudyPlanID: database.Text("preset-study-plan-id"),
						Status:            database.Text(cpb.CourseStatus_COURSE_STATUS_NONE.String()),
					},
				}).Once().Return(ErrCourseRepoUpsert)

			},
		},
		"course repo upsert failed with handle upsert course book": {
			req: &pb.UpsertCoursesRequest{
				Courses: []*pb.UpsertCoursesRequest_Course{{
					Id:           "id",
					Name:         "name",
					Country:      pb_bob.COUNTRY_VN,
					Subject:      pb_bob.SUBJECT_BIOLOGY,
					SchoolId:     constant.ManabieSchool,
					Grade:        "Lớp 1",
					DisplayOrder: 1,
					BookIds:      []string{"book-id"},
					Icon:         "icon-1",
				}},
			},
			expectedErr: errors.Wrap(ErrCourseRepoUpsert, "s.CourseRepo.Upsert"),
			setup: func(ctx context.Context) {
				courseRepo.On("FindByIDs", ctx, mock.Anything, database.TextArray([]string{"id"})).Once().
					Return(
						map[pgtype.Text]*entities_bob.Course{
							database.Text("id"): {
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
					"id": {"book-old-id"},
				}, nil)
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTxer, nil)
				mockTxer.On("Rollback", mock.Anything).Return(nil)
				courseRepo.On("Upsert", ctx, mock.Anything, []*entities_bob.Course{
					{
						ID:                database.Text("id"),
						Name:              database.Text("name"),
						Country:           database.Text(pb_bob.COUNTRY_VN.String()),
						Subject:           database.Text(pb_bob.SUBJECT_BIOLOGY.String()),
						Grade:             database.Int2(1),
						DisplayOrder:      database.Int2(1),
						SchoolID:          database.Int4(constant.ManabieSchool),
						TeacherIDs:        database.TextArray([]string{"teacher-1", "teacher-2"}),
						CourseType:        database.Text(pb_bob.COURSE_TYPE_CONTENT.String()),
						Icon:              database.Text("icon-1"),
						UpdatedAt:         database.Timestamptz(now),
						CreatedAt:         database.Timestamptz(now),
						DeletedAt:         pgtype.Timestamptz{Status: pgtype.Null},
						StartDate:         database.Timestamptz(now),
						EndDate:           database.Timestamptz(now.Add(2 * time.Minute)),
						PresetStudyPlanID: database.Text("preset-study-plan-id"),
						Status:            database.Text(cpb.CourseStatus_COURSE_STATUS_NONE.String()),
					},
				}).Once().Return(ErrCourseRepoUpsert)
				courseBookRepo.On("SoftDelete", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(ErrCourseBookRepoSoftDelete)
				courseBookRepo.On("Upsert", ctx, mock.Anything, mock.Anything).Once().Return(ErrCourseBookRepoUpsert)
			},
		},
		"course book repo upsert failed": {
			req: &pb.UpsertCoursesRequest{
				Courses: []*pb.UpsertCoursesRequest_Course{{
					Id:           "id",
					Name:         "name",
					Country:      pb_bob.COUNTRY_VN,
					Subject:      pb_bob.SUBJECT_BIOLOGY,
					SchoolId:     constant.ManabieSchool,
					Grade:        "Lớp 1",
					BookIds:      []string{"book-id"},
					DisplayOrder: 1,
					Icon:         "icon-1",
				}},
			},
			expectedErr: errors.Wrap(ErrCourseBookRepoUpsert, "s.CourseBookRepo.Upsert"),
			setup: func(ctx context.Context) {
				courseRepo.On("FindByIDs", ctx, mock.Anything, database.TextArray([]string{"id"})).Once().
					Return(
						map[pgtype.Text]*entities_bob.Course{
							database.Text("id"): {
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
					"id": {"book-old-id"},
				}, nil)
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTxer, nil)
				mockTxer.On("Rollback", mock.Anything).Return(nil)
				courseRepo.On("Upsert", ctx, mock.Anything, []*entities_bob.Course{
					{
						ID:                database.Text("id"),
						Name:              database.Text("name"),
						Country:           database.Text(pb_bob.COUNTRY_VN.String()),
						Subject:           database.Text(pb_bob.SUBJECT_BIOLOGY.String()),
						Grade:             database.Int2(1),
						DisplayOrder:      database.Int2(1),
						SchoolID:          database.Int4(constant.ManabieSchool),
						TeacherIDs:        database.TextArray([]string{"teacher-1", "teacher-2"}),
						CourseType:        database.Text(pb_bob.COURSE_TYPE_CONTENT.String()),
						Icon:              database.Text("icon-1"),
						UpdatedAt:         database.Timestamptz(now),
						CreatedAt:         database.Timestamptz(now),
						DeletedAt:         pgtype.Timestamptz{Status: pgtype.Null},
						StartDate:         database.Timestamptz(now),
						EndDate:           database.Timestamptz(now.Add(2 * time.Minute)),
						PresetStudyPlanID: database.Text("preset-study-plan-id"),
						Status:            database.Text(cpb.CourseStatus_COURSE_STATUS_NONE.String()),
					},
				}).Once().Return(nil)
				courseBookRepo.On("SoftDelete", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(ErrCourseBookRepoSoftDelete)
				courseBookRepo.On("Upsert", ctx, mock.Anything, mock.Anything).Once().Return(ErrCourseBookRepoUpsert)
			},
		},
		"course book repo soft delete failed": {
			req: &pb.UpsertCoursesRequest{
				Courses: []*pb.UpsertCoursesRequest_Course{{
					Id:           "id",
					Name:         "name",
					Country:      pb_bob.COUNTRY_VN,
					Subject:      pb_bob.SUBJECT_BIOLOGY,
					SchoolId:     constant.ManabieSchool,
					Grade:        "Lớp 1",
					BookIds:      []string{"book-id"},
					DisplayOrder: 1,
					Icon:         "icon-1",
				}},
			},
			expectedErr: errors.Wrap(ErrCourseBookRepoSoftDelete, "s.CourseBookRepo.SoftDelete"),
			setup: func(ctx context.Context) {
				courseRepo.On("FindByIDs", ctx, mock.Anything, database.TextArray([]string{"id"})).Once().
					Return(
						map[pgtype.Text]*entities_bob.Course{
							database.Text("id"): {
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
					"id": {"book-old-id"},
				}, nil)
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTxer, nil)
				mockTxer.On("Rollback", mock.Anything).Return(nil)
				courseRepo.On("Upsert", ctx, mock.Anything, []*entities_bob.Course{
					{
						ID:                database.Text("id"),
						Name:              database.Text("name"),
						Country:           database.Text(pb_bob.COUNTRY_VN.String()),
						Subject:           database.Text(pb_bob.SUBJECT_BIOLOGY.String()),
						Grade:             database.Int2(1),
						DisplayOrder:      database.Int2(1),
						SchoolID:          database.Int4(constant.ManabieSchool),
						TeacherIDs:        database.TextArray([]string{"teacher-1", "teacher-2"}),
						CourseType:        database.Text(pb_bob.COURSE_TYPE_CONTENT.String()),
						Icon:              database.Text("icon-1"),
						UpdatedAt:         database.Timestamptz(now),
						CreatedAt:         database.Timestamptz(now),
						DeletedAt:         pgtype.Timestamptz{Status: pgtype.Null},
						StartDate:         database.Timestamptz(now),
						EndDate:           database.Timestamptz(now.Add(2 * time.Minute)),
						PresetStudyPlanID: database.Text("preset-study-plan-id"),
						Status:            database.Text(cpb.CourseStatus_COURSE_STATUS_NONE.String()),
					},
				}).Once().Return(nil)
				courseBookRepo.On("SoftDelete", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(ErrCourseBookRepoSoftDelete)
				courseBookRepo.On("Upsert", ctx, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
		"happy case with handle upsert course book": {
			req: &pb.UpsertCoursesRequest{
				Courses: []*pb.UpsertCoursesRequest_Course{{
					Id:           "id",
					Name:         "name",
					Country:      pb_bob.COUNTRY_VN,
					Subject:      pb_bob.SUBJECT_BIOLOGY,
					DisplayOrder: 1,
					SchoolId:     constant.ManabieSchool,
					BookIds:      []string{"book-id"},
					Grade:        "Lớp 1",
					Icon:         "icon-1",
				}},
			},
			setup: func(ctx context.Context) {
				courseRepo.On("FindByIDs", ctx, mock.Anything, database.TextArray([]string{"id"})).Once().
					Return(
						map[pgtype.Text]*entities_bob.Course{
							database.Text("id"): {
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
					"id": {"book-old-id"},
				}, nil)
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTxer, nil)
				mockTxer.On("Commit", mock.Anything).Return(nil)
				courseRepo.On("Upsert", ctx, mock.Anything, []*entities_bob.Course{
					{
						ID:                database.Text("id"),
						Name:              database.Text("name"),
						Country:           database.Text(pb_bob.COUNTRY_VN.String()),
						Subject:           database.Text(pb_bob.SUBJECT_BIOLOGY.String()),
						Grade:             database.Int2(1),
						DisplayOrder:      database.Int2(1),
						SchoolID:          database.Int4(constant.ManabieSchool),
						TeacherIDs:        database.TextArray([]string{"teacher-1", "teacher-2"}),
						CourseType:        database.Text(pb_bob.COURSE_TYPE_CONTENT.String()),
						Icon:              database.Text("icon-1"),
						UpdatedAt:         database.Timestamptz(now),
						CreatedAt:         database.Timestamptz(now),
						DeletedAt:         pgtype.Timestamptz{Status: pgtype.Null},
						StartDate:         database.Timestamptz(now),
						EndDate:           database.Timestamptz(now.Add(2 * time.Minute)),
						PresetStudyPlanID: database.Text("preset-study-plan-id"),
						Status:            database.Text(cpb.CourseStatus_COURSE_STATUS_NONE.String()),
					},
				}).Once().Return(nil)
				courseBookRepo.On("SoftDelete", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				courseBookRepo.On("Upsert", ctx, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
		"happy case with handle insert course book": {
			req: &pb.UpsertCoursesRequest{
				Courses: []*pb.UpsertCoursesRequest_Course{{
					Id:           "id",
					Name:         "name",
					Country:      pb_bob.COUNTRY_VN,
					Subject:      pb_bob.SUBJECT_BIOLOGY,
					DisplayOrder: 1,
					SchoolId:     constant.ManabieSchool,
					BookIds:      []string{"book-id"},
					Grade:        "Lớp 1",
					Icon:         "icon-1",
				}},
			},
			setup: func(ctx context.Context) {
				courseRepo.On("FindByIDs", ctx, mock.Anything, database.TextArray([]string{"id"})).Once().
					Return(
						nil,
						nil,
					)
				courseBookRepo.On("FindByCourseIDs", ctx, mock.Anything, mock.Anything).Once().Return(map[string][]string{
					"id": {"book-old-id"},
				}, nil)
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTxer, nil)
				mockTxer.On("Commit", mock.Anything).Return(nil)
				courseRepo.On("Upsert", ctx, mock.Anything, []*entities_bob.Course{
					{
						ID:                database.Text("id"),
						Name:              database.Text("name"),
						Country:           database.Text(pb_bob.COUNTRY_VN.String()),
						Subject:           database.Text(pb_bob.SUBJECT_BIOLOGY.String()),
						Grade:             database.Int2(1),
						DisplayOrder:      database.Int2(1),
						SchoolID:          database.Int4(constant.ManabieSchool),
						TeacherIDs:        pgtype.TextArray{Status: pgtype.Null},
						CourseType:        database.Text(pb_bob.COURSE_TYPE_CONTENT.String()),
						Icon:              database.Text("icon-1"),
						UpdatedAt:         pgtype.Timestamptz{Status: pgtype.Null},
						CreatedAt:         pgtype.Timestamptz{Status: pgtype.Null},
						DeletedAt:         pgtype.Timestamptz{Status: pgtype.Null},
						StartDate:         pgtype.Timestamptz{Status: pgtype.Null},
						EndDate:           pgtype.Timestamptz{Status: pgtype.Null},
						PresetStudyPlanID: pgtype.Text{Status: pgtype.Null},
						Status:            database.Text(cpb.CourseStatus_COURSE_STATUS_NONE.String()),
					},
				}).Once().Return(nil)
				courseBookRepo.On("SoftDelete", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				courseBookRepo.On("Upsert", ctx, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
		"happy case without handle upsert course book": {
			req: &pb.UpsertCoursesRequest{
				Courses: []*pb.UpsertCoursesRequest_Course{{
					Id:           "id",
					Name:         "name",
					Country:      pb_bob.COUNTRY_VN,
					Subject:      pb_bob.SUBJECT_BIOLOGY,
					DisplayOrder: 1,
					SchoolId:     constant.ManabieSchool,
					Grade:        "Lớp 1",
					Icon:         "icon-1",
				}},
			},
			setup: func(ctx context.Context) {
				courseRepo.On("FindByIDs", ctx, mock.Anything, database.TextArray([]string{"id"})).Once().
					Return(
						map[pgtype.Text]*entities_bob.Course{
							database.Text("id"): {
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
				courseRepo.On("Upsert", ctx, mock.Anything, []*entities_bob.Course{
					{
						ID:                database.Text("id"),
						Name:              database.Text("name"),
						Country:           database.Text(pb_bob.COUNTRY_VN.String()),
						Subject:           database.Text(pb_bob.SUBJECT_BIOLOGY.String()),
						Grade:             database.Int2(1),
						DisplayOrder:      database.Int2(1),
						SchoolID:          database.Int4(constant.ManabieSchool),
						TeacherIDs:        database.TextArray([]string{"teacher-1", "teacher-2"}),
						CourseType:        database.Text(pb_bob.COURSE_TYPE_CONTENT.String()),
						Icon:              database.Text("icon-1"),
						UpdatedAt:         database.Timestamptz(now),
						CreatedAt:         database.Timestamptz(now),
						DeletedAt:         pgtype.Timestamptz{Status: pgtype.Null},
						StartDate:         database.Timestamptz(now),
						EndDate:           database.Timestamptz(now.Add(2 * time.Minute)),
						PresetStudyPlanID: database.Text("preset-study-plan-id"),
						Status:            database.Text(cpb.CourseStatus_COURSE_STATUS_NONE.String()),
					},
				}).Once().Return(nil)
			},
		},
		"happy case without handle insert course book": {
			req: &pb.UpsertCoursesRequest{
				Courses: []*pb.UpsertCoursesRequest_Course{{
					Id:           "id",
					Name:         "name",
					Country:      pb_bob.COUNTRY_VN,
					Subject:      pb_bob.SUBJECT_BIOLOGY,
					DisplayOrder: 1,
					SchoolId:     constant.ManabieSchool,
					Grade:        "Lớp 1",
					Icon:         "icon-1",
				}},
			},
			setup: func(ctx context.Context) {
				courseRepo.On("FindByIDs", ctx, mock.Anything, database.TextArray([]string{"id"})).Once().
					Return(
						nil,
						nil,
					)
				courseRepo.On("Upsert", ctx, mock.Anything, []*entities_bob.Course{
					{
						ID:                database.Text("id"),
						Name:              database.Text("name"),
						Country:           database.Text(pb_bob.COUNTRY_VN.String()),
						Subject:           database.Text(pb_bob.SUBJECT_BIOLOGY.String()),
						Grade:             database.Int2(1),
						DisplayOrder:      database.Int2(1),
						SchoolID:          database.Int4(constant.ManabieSchool),
						TeacherIDs:        pgtype.TextArray{Status: pgtype.Null},
						CourseType:        database.Text(pb_bob.COURSE_TYPE_CONTENT.String()),
						Icon:              database.Text("icon-1"),
						UpdatedAt:         pgtype.Timestamptz{Status: pgtype.Null},
						CreatedAt:         pgtype.Timestamptz{Status: pgtype.Null},
						DeletedAt:         pgtype.Timestamptz{Status: pgtype.Null},
						StartDate:         pgtype.Timestamptz{Status: pgtype.Null},
						EndDate:           pgtype.Timestamptz{Status: pgtype.Null},
						PresetStudyPlanID: pgtype.Text{Status: pgtype.Null},
						Status:            database.Text(cpb.CourseStatus_COURSE_STATUS_NONE.String()),
					},
				}).Once().Return(nil)
			},
		},
	}

	for caseName, testCase := range testCases {
		t.Run(caseName, func(t *testing.T) {
			ctx := context.Background()
			testCase.setup(ctx)

			resp, err := courseService.UpsertCourses(ctx, testCase.req.(*pb.UpsertCoursesRequest))

			if testCase.expectedErr != nil {
				assert.Error(t, testCase.expectedErr, err)
			}
			if testCase.expectedResp == nil {
				assert.Nil(t, testCase.expectedResp, resp)
			} else {
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}

func TestUpsertChapters_Error(t *testing.T) {
}

func TestUpsertBooks_Error(t *testing.T) {
}

func TestDeleteCourse_Error(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	chapterRepo := &mock_bob_repositories.MockChapterRepo{}
	bookRepo := &mock_bob_repositories.MockBookRepo{}
	courseRepo := &mock_bob_repositories.MockCourseRepo{}
	userRepo := &mock_bob_repositories.MockUserRepo{}

	courseService := &CourseAbac{
		CourseService: &CourseService{
			ChapterRepo: chapterRepo,
			BookRepo:    bookRepo,
			CourseRepo:  courseRepo,
			UserRepo:    userRepo,
		},
	}

	userID := idutil.ULIDNow()
	ctx = interceptors.ContextWithUserID(ctx, userID)

	testCases := map[string]TestCase{
		"missing course id": {
			req: &pb.DeleteCoursesRequest{
				CourseIds: []string{},
			},
			expectedErr: status.Error(codes.InvalidArgument, "Empty request"),
			setup: func(ctx context.Context) {
			},
		},
		"not found course": {
			req: &pb.DeleteCoursesRequest{
				CourseIds: []string{"course-id"},
			},
			expectedErr: status.Error(codes.InvalidArgument, "course not found"),
			setup: func(ctx context.Context) {
				u := &entities_bob.User{}
				u.Group = database.Text("USER_GROUP_ADMIN")

				userRepo.On("Get", ctx, mock.Anything, mock.Anything).
					Once().Return(u, nil)

				courseRepo.On("FindSchoolIDsOnCourses", ctx, mock.Anything, []string{"course-id"}).
					Once().Return([]int32{1}, nil)

				courseRepo.On("FindByIDs", ctx, mock.Anything, mock.Anything).Once().Return(map[pgtype.Text]*entities_bob.Course{}, nil)
			},
		},
	}

	for caseName, testCase := range testCases {
		t.Run(caseName, func(t *testing.T) {
			testCase.setup(ctx)

			_, err := courseService.DeleteCourses(ctx, testCase.req.(*pb.DeleteCoursesRequest))
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}
}

func TestAddBook(t *testing.T) {
}

func TestCourseService_CourseIDsByClass(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	courseClassRepo := &mock_repositories.MockCourseClassRepo{}
	c := &CourseService{
		CourseClassRepo: courseClassRepo,
	}

	t.Run("err findByClassID", func(t *testing.T) {
		courseClassRepo.On("FindByClassIDs", ctx, mock.Anything, database.Int4Array([]int32{1, 2})).
			Once().Return(map[pgtype.Int4]pgtype.TextArray{}, pgx.ErrTxClosed)

		resp, err := c.CourseIDsByClass(ctx, []int32{1, 2})
		assert.Nil(t, resp)
		assert.EqualError(t, err, "err s.CourseClassRepo.FindByClassIDs: tx is closed")
	})

	t.Run("success", func(t *testing.T) {
		courseClassRepo.On("FindByClassIDs", ctx, mock.Anything, database.Int4Array([]int32{1, 2})).
			Once().Return(map[pgtype.Int4]pgtype.TextArray{
			database.Int4(1): database.TextArray([]string{"courseID1", "courseID2"}),
			database.Int4(2): database.TextArray([]string{"courseID1", "courseID3"}),
		}, nil)

		resp, err := c.CourseIDsByClass(ctx, []int32{1, 2})
		assert.Nil(t, err)
		assert.Equal(t, map[int32][]string{
			1: {"courseID1", "courseID2"},
			2: {"courseID1", "courseID3"},
		}, resp)
	})
}

func TestCourseService_SyncCourse(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	now := time.Now().UTC()

	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}

	courseBookRepo := &mock_bob_repositories.MockCourseBookRepo{}
	courseRepo := &mock_bob_repositories.MockCourseRepo{}
	courseAccessPathRepo := &mock_bob_repositories.MockCourseAccessPathRepo{}

	s := &CourseService{
		DBTrace:              db,
		CourseBookRepo:       courseBookRepo,
		CourseRepo:           courseRepo,
		CourseAccessPathRepo: courseAccessPathRepo,
	}

	t.Run("err insert course, missing course name", func(t *testing.T) {
		req := &npb.EventMasterRegistration_Course{
			ActionKind: npb.ActionKind_ACTION_KIND_UPSERTED,
			CourseId:   "courseID-1",
			CourseName: "",
		}

		err := s.SyncCourse(ctx, []*npb.EventMasterRegistration_Course{req})
		assert.EqualError(t, err, "s.UpsertCourses: rpc error: code = InvalidArgument desc = course name cannot be empty")
	})

	t.Run("err insert course", func(t *testing.T) {
		courseBookRepo.On("FindByCourseIDs", ctx, db, []string{"courseID-1"}).
			Once().Return(map[string][]string{}, nil)
		courseRepo.On("FindByIDs", ctx, mock.Anything, database.TextArray([]string{"courseID-1"})).Once().
			Return(
				map[pgtype.Text]*entities_bob.Course{
					database.Text("courseID-1"): {
						ID:                database.Text("courseID-1"),
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

		db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
		tx.On("Rollback", mock.Anything).Once().Return(nil)

		courseRepo.On("UpsertV2", mock.Anything, tx, []*entities_bob.Course{
			{
				ID:                database.Text("courseID-1"),
				Name:              database.Text("course name"),
				Country:           database.Text(pb_bob.COUNTRY_JP.String()),
				Subject:           database.Text(pb_bob.SUBJECT_ENGLISH.String()),
				Grade:             database.Int2(0),
				DisplayOrder:      database.Int2(1),
				SchoolID:          database.Int4(constants.JPREPSchool),
				TeacherIDs:        database.TextArray([]string{"teacher-1", "teacher-2"}),
				CourseType:        database.Text(pb_bob.COURSE_TYPE_CONTENT.String()),
				Icon:              database.Text(""),
				UpdatedAt:         database.Timestamptz(now),
				CreatedAt:         database.Timestamptz(now),
				DeletedAt:         pgtype.Timestamptz{Status: pgtype.Null},
				StartDate:         database.Timestamptz(now),
				EndDate:           database.Timestamptz(now.Add(2 * time.Minute)),
				PresetStudyPlanID: database.Text("preset-study-plan-id"),
				Status:            database.Text(cpb.CourseStatus_COURSE_STATUS_NONE.String()),
			},
		}).
			Once().Return(pgx.ErrTxClosed)

		req := &npb.EventMasterRegistration_Course{
			ActionKind: npb.ActionKind_ACTION_KIND_UPSERTED,
			CourseId:   "courseID-1",
			CourseName: "course name",
		}

		err := s.SyncCourse(ctx, []*npb.EventMasterRegistration_Course{req})
		assert.EqualError(t, err, "s.UpsertCourses: s.CourseRepo.UpsertV2: tx is closed")
	})

	t.Run("success insert course", func(t *testing.T) {
		courseBookRepo.On("FindByCourseIDs", ctx, db, []string{"courseID-1"}).
			Once().Return(map[string][]string{}, nil)

		courseRepo.On("FindByIDs", ctx, mock.Anything, database.TextArray([]string{"courseID-1"})).Once().
			Return(
				nil,
				nil,
			)

		db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
		tx.On("Commit", mock.Anything).Once().Return(nil)

		courseRepo.On("UpsertV2", mock.Anything, tx, []*entities_bob.Course{
			{
				ID:           database.Text("courseID-1"),
				Name:         database.Text("course name"),
				Country:      database.Text(pb_bob.COUNTRY_JP.String()),
				Subject:      database.Text(pb_bob.SUBJECT_ENGLISH.String()),
				Grade:        database.Int2(0),
				DisplayOrder: database.Int2(1),
				SchoolID:     database.Int4(constants.JPREPSchool),
				TeacherIDs:   pgtype.TextArray{Status: pgtype.Null},
				CourseType:   database.Text(pb_bob.COURSE_TYPE_CONTENT.String()),
				Icon:         database.Text(""),
				UpdatedAt:    pgtype.Timestamptz{Status: pgtype.Null},
				CreatedAt:    pgtype.Timestamptz{Status: pgtype.Null},
				DeletedAt:    pgtype.Timestamptz{Status: pgtype.Null},
				StartDate:    pgtype.Timestamptz{Status: pgtype.Null},
				EndDate:      pgtype.Timestamptz{Status: pgtype.Null},
				PresetStudyPlanID: pgtype.Text{
					Status: pgtype.Null,
				},
				Status: database.Text(cpb.CourseStatus_COURSE_STATUS_NONE.String()),
			},
		}).
			Once().Return(nil)
		courseAccessPathRepo.On("Upsert", mock.Anything, tx, mock.IsType([]*entities_bob.CourseAccessPath{})).Return(nil)

		req := &npb.EventMasterRegistration_Course{
			ActionKind: npb.ActionKind_ACTION_KIND_UPSERTED,
			CourseId:   "courseID-1",
			CourseName: "course name",
		}

		err := s.SyncCourse(ctx, []*npb.EventMasterRegistration_Course{req})
		assert.Nil(t, err)
	})

	t.Run("success update course", func(t *testing.T) {
		courseBookRepo.On("FindByCourseIDs", ctx, db, []string{"courseID-1"}).
			Once().Return(map[string][]string{}, nil)

		courseRepo.On("FindByIDs", ctx, mock.Anything, database.TextArray([]string{"courseID-1"})).Once().
			Return(
				map[pgtype.Text]*entities_bob.Course{
					database.Text("courseID-1"): {
						ID:                database.Text("courseID-1"),
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

		db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
		tx.On("Commit", mock.Anything).Once().Return(nil)

		courseRepo.On("UpsertV2", mock.Anything, tx, []*entities_bob.Course{
			{
				ID:                database.Text("courseID-1"),
				Name:              database.Text("course name"),
				Country:           database.Text(pb_bob.COUNTRY_JP.String()),
				Subject:           database.Text(pb_bob.SUBJECT_ENGLISH.String()),
				Grade:             database.Int2(0),
				DisplayOrder:      database.Int2(1),
				SchoolID:          database.Int4(constants.JPREPSchool),
				TeacherIDs:        database.TextArray([]string{"teacher-1", "teacher-2"}),
				CourseType:        database.Text(pb_bob.COURSE_TYPE_CONTENT.String()),
				Icon:              database.Text(""),
				UpdatedAt:         database.Timestamptz(now),
				CreatedAt:         database.Timestamptz(now),
				DeletedAt:         pgtype.Timestamptz{Status: pgtype.Null},
				StartDate:         database.Timestamptz(now),
				EndDate:           database.Timestamptz(now.Add(2 * time.Minute)),
				PresetStudyPlanID: database.Text("preset-study-plan-id"),
				Status:            database.Text(cpb.CourseStatus_COURSE_STATUS_NONE.String()),
			},
		}).
			Once().Return(nil)

		req := &npb.EventMasterRegistration_Course{
			ActionKind: npb.ActionKind_ACTION_KIND_UPSERTED,
			CourseId:   "courseID-1",
			CourseName: "course name",
		}

		err := s.SyncCourse(ctx, []*npb.EventMasterRegistration_Course{req})
		assert.Nil(t, err)
	})

	t.Run("err delete course", func(t *testing.T) {
		courseRepo.On("SoftDelete", ctx, db, database.TextArray([]string{"courseID-1"})).
			Once().Return(pgx.ErrTxClosed)

		req := &npb.EventMasterRegistration_Course{
			ActionKind: npb.ActionKind_ACTION_KIND_DELETED,
			CourseId:   "courseID-1",
			CourseName: "course name",
		}

		err := s.SyncCourse(ctx, []*npb.EventMasterRegistration_Course{req})
		assert.EqualError(t, err, "s.DeleteCourses: CourseRepo.SoftDelete: tx is closed")
	})

	t.Run("success delete course", func(t *testing.T) {
		courseRepo.On("SoftDelete", ctx, db, database.TextArray([]string{"courseID-1"})).
			Once().Return(nil)

		req := &npb.EventMasterRegistration_Course{
			ActionKind: npb.ActionKind_ACTION_KIND_DELETED,
			CourseId:   "courseID-1",
			CourseName: "course name",
		}

		err := s.SyncCourse(ctx, []*npb.EventMasterRegistration_Course{req})
		assert.Nil(t, err)
	})
}

func TestCourseService_SyncAcademicYear(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	academicYearRepo := &mock_bob_repositories.MockAcademicYearRepo{}

	mockDB := &mock_database.Ext{}
	mockTxer := &mock_database.Tx{}

	s := &CourseService{
		DBTrace:          mockDB,
		AcademicYearRepo: academicYearRepo,
	}

	t.Run("err create academicYear", func(t *testing.T) {
		mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTxer, nil)
		mockTxer.On("Rollback", mock.Anything).Once().Return(nil)

		academicYearRepo.On("Create", ctx, mockTxer, mock.AnythingOfType("*entities.AcademicYear")).
			Once().Return(pgx.ErrTxCommitRollback)

		academicYear := &npb.EventMasterRegistration_AcademicYear{
			ActionKind:     npb.ActionKind_ACTION_KIND_UPSERTED,
			AcademicYearId: "2021",
			Name:           "2021",
			StartYearDate:  timestamppb.Now(),
			EndYearDate:    timestamppb.Now(),
		}

		err := s.SyncAcademicYear(ctx, []*npb.EventMasterRegistration_AcademicYear{
			academicYear,
		})

		assert.EqualError(t, err, "err AcademicYearRepo.Create 2021: commit unexpectedly resulted in rollback")
	})

	t.Run("success", func(t *testing.T) {
		mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTxer, nil)
		mockTxer.On("Commit", mock.Anything).Once().Return(nil)

		academicYearRepo.On("Create", ctx, mockTxer, mock.AnythingOfType("*entities.AcademicYear")).
			Once().Run(func(args mock.Arguments) {
			academicYear := args[2].(*entities_bob.AcademicYear)
			assert.Equal(t, "2021", academicYear.ID.String)
			assert.Equal(t, "2021", academicYear.Name.String)
			assert.Equal(t, entities_bob.AcademicYearStatusActive, academicYear.Status.String)
		}).Return(nil)

		academicYear := &npb.EventMasterRegistration_AcademicYear{
			ActionKind:     npb.ActionKind_ACTION_KIND_UPSERTED,
			AcademicYearId: "2021",
			Name:           "2021",
			StartYearDate:  timestamppb.Now(),
			EndYearDate:    timestamppb.Now(),
		}

		err := s.SyncAcademicYear(ctx, []*npb.EventMasterRegistration_AcademicYear{
			academicYear,
		})

		assert.Nil(t, err)
	})
}

func TestCourseService_UpdateAcademicYear(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	courseRepo := &mock_bob_repositories.MockCourseRepo{}

	mockDB := &mock_database.Ext{}
	mockTxer := &mock_database.Tx{}

	s := &CourseService{
		DBTrace:    mockDB,
		CourseRepo: courseRepo,
	}

	t.Run("err create academicYear", func(t *testing.T) {
		mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTxer, nil)
		mockTxer.On("Rollback", mock.Anything).Once().Return(nil)

		courseRepo.On("UpdateAcademicYear", ctx, mockTxer, mock.AnythingOfType("[]*repositories.UpdateAcademicYearOpts")).
			Once().Return(pgx.ErrTxCommitRollback)

		opt := &repositories.UpdateAcademicYearOpts{
			CourseID:       "course-id",
			AcademicYearID: "2021",
		}

		err := s.UpdateAcademicYear(ctx, []*repositories.UpdateAcademicYearOpts{
			opt,
		})

		assert.EqualError(t, err, "commit unexpectedly resulted in rollback")
	})

	t.Run("success", func(t *testing.T) {
		mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTxer, nil)
		mockTxer.On("Commit", mock.Anything).Once().Return(nil)

		opt := &repositories.UpdateAcademicYearOpts{
			CourseID:       "course-id",
			AcademicYearID: "2021",
		}

		courseRepo.On("UpdateAcademicYear", ctx, mockTxer, mock.AnythingOfType("[]*repositories.UpdateAcademicYearOpts")).
			Once().Run(func(args mock.Arguments) {
			opts := args[2].([]*repositories.UpdateAcademicYearOpts)
			assert.Equal(t, 1, len(opts))
			assert.Equal(t, opt, opts[0])
		}).Return(nil)

		err := s.UpdateAcademicYear(ctx, []*repositories.UpdateAcademicYearOpts{
			opt,
		})

		assert.Nil(t, err)
	})
}

func TestValidateCourseV2(t *testing.T) {
	testCases := map[string]TestCase{
		"missing course id": {
			req: &Course{
				Name:       "name",
				Country:    pb_bob.COUNTRY_VN,
				Subject:    pb_bob.SUBJECT_MATHS,
				Grade:      "Lớp 1",
				ChapterIDs: []string{},
				SchoolID:   2,
				Icon:       "icon",
			},
			expectedErr: ErrMissingCourseID,
			setup: func(ctx context.Context) {
			},
		},

		"missing name": {
			req: &Course{
				ID:       idutil.ULIDNow(),
				Country:  pb_bob.COUNTRY_VN,
				Subject:  pb_bob.SUBJECT_MATHS,
				Grade:    "Lớp 1",
				SchoolID: 2,
				Icon:     "icon",
			},
			expectedErr: ErrMissingName,
			setup: func(ctx context.Context) {
			},
		},

		"missing school id": {
			req: &Course{
				ID:      idutil.ULIDNow(),
				Name:    "name",
				Country: pb_bob.COUNTRY_VN,
				Subject: pb_bob.SUBJECT_MATHS,
				Grade:   "Lớp 1",
				Icon:    "icon",
			},
			expectedErr: status.Error(codes.InvalidArgument, "missing school id of course name"),
			setup: func(ctx context.Context) {
			},
		},

		// make sure it work after remove
		"missing country": {
			req: &Course{
				ID:       idutil.ULIDNow(),
				Name:     "name",
				Subject:  pb_bob.SUBJECT_MATHS,
				Grade:    "Lớp 1",
				SchoolID: 2,
				Icon:     "icon",
			},
			setup: func(ctx context.Context) {
			},
		},

		// make sure it work after remove
		"missing subject": {
			req: &Course{
				ID:       idutil.ULIDNow(),
				Name:     "name",
				Country:  pb_bob.COUNTRY_VN,
				Grade:    "Lớp 1",
				SchoolID: 2,
				Icon:     "icon",
			},
			setup: func(ctx context.Context) {
			},
		},

		// make sure it work after remove
		"missing grade": {
			req: &Course{
				ID:       idutil.ULIDNow(),
				Name:     "name",
				Country:  pb_bob.COUNTRY_VN,
				Subject:  pb_bob.SUBJECT_MATHS,
				SchoolID: 2,
				Icon:     "icon",
			},
			setup: func(ctx context.Context) {
			},
		},

		"happy case": {
			req: &Course{
				ID:           idutil.ULIDNow(),
				Name:         "name",
				Country:      pb_bob.COUNTRY_VN,
				Subject:      pb_bob.SUBJECT_MATHS,
				Grade:        "Lớp 1",
				DisplayOrder: 1,
				ChapterIDs:   []string{},
				SchoolID:     2,
				BookIDs:      []string{},
				Icon:         "icon",
			},
			setup: func(ctx context.Context) {
			},
		},
	}
	for caseName, testCase := range testCases {
		t.Run(caseName, func(t *testing.T) {
			ctx := context.Background()
			testCase.setup(ctx)
			course := testCase.req.(*Course)
			if err := validateCourseV2(course); testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}
}

var ErrSomethingWentWrong = fmt.Errorf("something went wrong")

func TestDeleteChapter(t *testing.T) {
}

func TestCourseService_DeleteTopics(t *testing.T) {
}

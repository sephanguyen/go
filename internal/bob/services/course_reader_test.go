package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/manabie-com/backend/internal/bob/constants"
	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mock_repositories "github.com/manabie-com/backend/mock/bob/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"

	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestCourseModifier_ListLessonMedias(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	lessonGroupRepo := &mock_repositories.MockLessonGroupRepo{}
	lessonRepo := &mock_repositories.MockLessonRepo{}
	mediaRepo := &mock_repositories.MockMediaRepo{}
	db := &mock_database.Ext{}

	s := &CourseReaderService{
		DB:              db,
		LessonGroupRepo: lessonGroupRepo,
		LessonRepo:      lessonRepo,
		MediaRepo:       mediaRepo,
	}

	lessonID := "lessonID1"
	lessons := []*entities.Lesson{}
	lessons = append(lessons, &entities.Lesson{
		LessonID:      database.Text(lessonID),
		LessonGroupID: database.Text("lessonGroupID"),
		CourseID:      database.Text("courseID"),
	})
	medias := entities.Medias{}
	medias = append(medias, &entities.Media{})

	testCases := []TestCase{
		{
			name: "happy case attach successfully",
			ctx:  ctx,
			req: &bpb.ListLessonMediasRequest{
				LessonId: lessonID,
				Paging: &cpb.Paging{
					Limit: 2,
					Offset: &cpb.Paging_OffsetString{
						OffsetString: "",
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				lessonRepo.On("Find", mock.Anything, mock.Anything, mock.Anything).Once().Return(lessons, nil)
				lessonGroupRepo.On("GetMedias", mock.Anything, mock.Anything, lessons[0].LessonGroupID, lessons[0].CourseID, mock.Anything, mock.Anything).Once().Return(medias, nil)
			},
		},
		{
			name: "not found lesson",
			ctx:  ctx,
			req: &bpb.ListLessonMediasRequest{
				LessonId: lessonID,
				Paging: &cpb.Paging{
					Limit: 2,
					Offset: &cpb.Paging_OffsetString{
						OffsetString: "",
					},
				},
			},
			expectedErr: fmt.Errorf("ListLessonMedias.LessonRepo.Find: %v", pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				lessonRepo.On("Find", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name: "not found media",
			ctx:  ctx,
			req: &bpb.ListLessonMediasRequest{
				LessonId: lessonID,
				Paging: &cpb.Paging{
					Limit: 2,
					Offset: &cpb.Paging_OffsetString{
						OffsetString: "",
					},
				},
			},
			expectedErr: fmt.Errorf("ListLessonMedias.LessonGroupRepo.GetMedias: %v", pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				lessonRepo.On("Find", mock.Anything, mock.Anything, mock.Anything).Once().Return(lessons, nil)
				lessonGroupRepo.On("GetMedias", mock.Anything, mock.Anything, lessons[0].LessonGroupID, lessons[0].CourseID, mock.Anything, mock.Anything).Once().Return(entities.Medias{}, pgx.ErrNoRows)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			_, err := s.ListLessonMedias(testCase.ctx, testCase.req.(*bpb.ListLessonMediasRequest))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func TestCourseReaderService_ListCourses(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	teacherRepo := &mock_repositories.MockTeacherRepo{}
	courseRepo := &mock_repositories.MockCourseRepo{}
	c := &CourseReaderService{
		TeacherRepo: teacherRepo,
		CourseRepo:  courseRepo,
	}

	userID := idutil.ULIDNow()
	ctx = interceptors.ContextWithUserID(ctx, userID)
	ctx = interceptors.ContextWithUserGroup(ctx, entities.UserGroupTeacher)
	t.Run("err find teacher", func(t *testing.T) {
		teacherRepo.On("FindByID", ctx, mock.Anything, database.Text(userID)).Once().
			Return(nil, pgx.ErrNoRows)

		resp, err := c.ListCourses(ctx, &bpb.ListCoursesRequest{})
		assert.Nil(t, resp)
		assert.EqualError(t, err, pgx.ErrNoRows.Error())
	})

	t.Run("err retrieve course", func(t *testing.T) {
		courseRepo.On("RetrieveCourses", ctx, mock.Anything, &repositories.CourseQuery{
			Countries: []string{cpb.Country_COUNTRY_VN.String()},
			Subject:   "",
			Grade:     0,
			SchoolIDs: []int{constants.ManabieSchool},
			Limit:     10,
			Offset:    0,
			Type:      "",
			Status:    "",
		}).Once().Return(entities.Courses{}, pgx.ErrTxClosed)

		resp, err := c.ListCourses(ctx, &bpb.ListCoursesRequest{
			Filter: &cpb.CommonFilter{
				SchoolId: constants.ManabieSchool,
				Country:  cpb.Country_COUNTRY_VN,
			},
		})
		assert.Nil(t, resp)
		assert.EqualError(t, err, "err c.CourseRepo.RetrieveCourses: tx is closed")
	})

	t.Run("success retrieve course", func(t *testing.T) {
		courses := entities.Courses{
			{
				ID: database.Text("course-id"),
			},
		}
		courseRepo.On("RetrieveCourses", ctx, mock.Anything, &repositories.CourseQuery{
			Countries: []string{cpb.Country_COUNTRY_VN.String()},
			Subject:   cpb.Subject_SUBJECT_ENGLISH.String(),
			Grade:     12,
			SchoolIDs: []int{constants.ManabieSchool},
			Limit:     10,
			Offset:    0,
			Keyword:   "course",
		}).Once().Return(courses, nil)

		resp, err := c.ListCourses(ctx, &bpb.ListCoursesRequest{
			Filter: &cpb.CommonFilter{
				SchoolId: constants.ManabieSchool,
				Country:  cpb.Country_COUNTRY_VN,
				Grade:    12,
				Subject:  cpb.Subject_SUBJECT_ENGLISH,
			},
			Keyword: "course",
		})
		assert.Nil(t, err)

		course := courses[0]
		assert.Equal(t, &bpb.ListCoursesResponse{
			NextPage: &cpb.Paging{
				Limit: 10,
				Offset: &cpb.Paging_OffsetCombined{
					OffsetCombined: &cpb.Paging_Combined{
						OffsetInteger: 10,
					},
				},
			},
			Items: []*cpb.Course{
				{
					Info: &cpb.ContentBasicInfo{
						Id:           course.ID.String,
						Name:         course.Name.String,
						Country:      cpb.Country(cpb.Country_value[course.Country.String]),
						Subject:      cpb.Subject(cpb.Subject_value[course.Subject.String]),
						Grade:        int32(course.Grade.Int),
						SchoolId:     course.SchoolID.Int,
						DisplayOrder: int32(course.DisplayOrder.Int),
						UpdatedAt:    timestamppb.New(course.UpdatedAt.Time),
						CreatedAt:    timestamppb.New(course.CreatedAt.Time),
					},
					CourseStatus: cpb.CourseStatus_COURSE_STATUS_ACTIVE,
				},
			},
		}, resp)
	})
}

func TestCourseReaderService_ListCoursesByLocations(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	teacherRepo := &mock_repositories.MockTeacherRepo{}
	courseRepo := &mock_repositories.MockCourseRepo{}
	c := &CourseReaderService{
		TeacherRepo: teacherRepo,
		CourseRepo:  courseRepo,
	}

	userID := idutil.ULIDNow()
	ctx = interceptors.ContextWithUserID(ctx, userID)
	ctx = interceptors.ContextWithUserGroup(ctx, entities.UserGroupTeacher)
	t.Run("err find teacher", func(t *testing.T) {
		teacherRepo.On("FindByID", ctx, mock.Anything, database.Text(userID)).Once().
			Return(nil, pgx.ErrNoRows)

		resp, err := c.ListCoursesByLocations(ctx, &bpb.ListCoursesByLocationsRequest{})
		assert.Nil(t, resp)
		assert.EqualError(t, err, pgx.ErrNoRows.Error())
	})

	t.Run("err retrieve course", func(t *testing.T) {
		courseRepo.On("RetrieveCourses", ctx, mock.Anything, &repositories.CourseQuery{
			Countries: []string{cpb.Country_COUNTRY_VN.String()},
			Subject:   "",
			Grade:     0,
			SchoolIDs: []int{constants.ManabieSchool},
			Limit:     10,
			Offset:    0,
			Type:      "",
			Status:    "",
		}).Once().Return(entities.Courses{}, pgx.ErrTxClosed)

		resp, err := c.ListCoursesByLocations(ctx, &bpb.ListCoursesByLocationsRequest{
			Filter: &cpb.CommonFilter{
				SchoolId: constants.ManabieSchool,
				Country:  cpb.Country_COUNTRY_VN,
			},
		})
		assert.Nil(t, resp)
		assert.EqualError(t, err, "err c.CourseRepo.RetrieveCourses: tx is closed")
	})

	t.Run("success retrieve course", func(t *testing.T) {
		courses := entities.Courses{
			{
				ID: database.Text("course-id"),
			},
		}
		courseRepo.On("RetrieveCourses", ctx, mock.Anything, &repositories.CourseQuery{
			Countries: []string{cpb.Country_COUNTRY_VN.String()},
			Subject:   cpb.Subject_SUBJECT_ENGLISH.String(),
			Grade:     12,
			SchoolIDs: []int{constants.ManabieSchool},
			Limit:     10,
			Offset:    0,
			Keyword:   "course",
		}).Once().Return(courses, nil)

		resp, err := c.ListCoursesByLocations(ctx, &bpb.ListCoursesByLocationsRequest{
			Filter: &cpb.CommonFilter{
				SchoolId: constants.ManabieSchool,
				Country:  cpb.Country_COUNTRY_VN,
				Grade:    12,
				Subject:  cpb.Subject_SUBJECT_ENGLISH,
			},
			Keyword: "course",
		})
		assert.Nil(t, err)

		course := courses[0]
		assert.Equal(t, &bpb.ListCoursesByLocationsResponse{
			NextPage: &cpb.Paging{
				Limit: 10,
				Offset: &cpb.Paging_OffsetCombined{
					OffsetCombined: &cpb.Paging_Combined{
						OffsetInteger: 10,
					},
				},
			},
			Items: []*cpb.Course{
				{
					Info: &cpb.ContentBasicInfo{
						Id:           course.ID.String,
						Name:         course.Name.String,
						Country:      cpb.Country(cpb.Country_value[course.Country.String]),
						Subject:      cpb.Subject(cpb.Subject_value[course.Subject.String]),
						Grade:        int32(course.Grade.Int),
						SchoolId:     course.SchoolID.Int,
						DisplayOrder: int32(course.DisplayOrder.Int),
						UpdatedAt:    timestamppb.New(course.UpdatedAt.Time),
						CreatedAt:    timestamppb.New(course.CreatedAt.Time),
					},
					CourseStatus: cpb.CourseStatus_COURSE_STATUS_ACTIVE,
				},
			},
		}, resp)
	})
}

func TestCourseReaderService_ToMediaPbV1(t *testing.T) {
	t.Parallel()
	media := generateMediaEnt()

	mediaPb, err := toMediaPbV1(&media)
	assert.Nil(t, err)

	assert.Equal(t, media.MediaID.String, mediaPb.MediaId)
	assert.Equal(t, media.Name.String, mediaPb.Name)
	assert.Equal(t, media.Resource.String, mediaPb.Resource)
	assert.Equal(t, media.Type.String, bpb.MediaType_name[int32(mediaPb.Type)])
	var comments []entities.Comment
	err = media.Comments.AssignTo(&comments)
	assert.Nil(t, err)
	for i := range mediaPb.Comments {
		assert.Equal(t, comments[i].Comment, mediaPb.Comments[i].Comment)
		assert.Equal(t, time.Duration(comments[i].Duration), mediaPb.Comments[i].Duration.AsDuration())
	}

	var convertedImages []*entities.ConvertedImage
	err = media.ConvertedImages.AssignTo(&convertedImages)
	assert.Nil(t, err)
	for i := range mediaPb.Images {
		assert.Equal(t, convertedImages[i].Width, mediaPb.Images[i].Width)
		assert.Equal(t, convertedImages[i].Height, mediaPb.Images[i].Height)
		assert.Equal(t, convertedImages[i].ImageURL, mediaPb.Images[i].ImageUrl)
	}

	assert.True(t, media.CreatedAt.Time.Equal(mediaPb.CreatedAt.AsTime()))
	assert.True(t, media.UpdatedAt.Time.Equal(mediaPb.UpdatedAt.AsTime()))
}

func generateMediaEnt() entities.Media {
	media := entities.Media{}
	media.MediaID.Set(idutil.ULIDNow())
	media.Name.Set("media-name")
	media.Resource.Set("media Resource")
	media.Type.Set("MEDIA_TYPE_IMAGE")
	media.Comments.Set(`[{"comment": "1222", "duration": 0}, {"comment": "www", "duration": 0}]`)
	media.CreatedAt.Set(time.Now())
	media.UpdatedAt.Set(time.Now())
	media.DeletedAt.Set(nil)
	media.ConvertedImages.Set(`[{"width": 100, "height": 100, "image_url":"http://src.img"},{"width": 200, "height": 200, "image_url":"http://src2.img"}]`)
	return media
}

var ErrSomethingWentWrong = fmt.Errorf("something went wrong")

func TestCourseReaderService_RetrieveBookTreeByTopicIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	topicRepo := &mock_repositories.MockTopicRepo{}
	bookRepo := &mock_repositories.MockBookRepo{}

	s := &CourseReaderService{
		DB:            db,
		EurekaDBTrace: db,
		TopicRepo:     topicRepo,
		BookRepo:      bookRepo,
	}
	req := &bpb.RetrieveBookTreeByTopicIDsRequest{
		TopicIds: []string{"topic-1"},
	}
	testCases := []TestCase{
		{
			name: "happy case",
			ctx:  ctx,
			req:  req,
			setup: func(ctx context.Context) {
				topicRepo.On("RetrieveByIDs", ctx, db, mock.Anything).Once().Return([]*entities.Topic{
					{
						ID: database.Text("topic-1"),
					},
				}, nil)
				bookRepo.On("RetrieveBookTreeByTopicIDs", ctx, db, mock.Anything).Once().Return([]*repositories.BookTreeInfo{}, nil)
			},
		},
		{
			name: "topic id not found",
			ctx:  ctx,
			req: &bpb.RetrieveBookTreeByTopicIDsRequest{
				TopicIds: []string{"topic-1", "invalid-topic-id"},
			},
			expectedErr: status.Errorf(codes.NotFound, "topic invalid-topic-id not exist"),
			setup: func(ctx context.Context) {
				topicRepo.On("RetrieveByIDs", ctx, db, mock.Anything).Once().Return([]*entities.Topic{
					{
						ID: database.Text("topic-1"),
					},
				}, nil)
			},
		},
		{
			name:        "err topicRepo.RetrieveByIDs",
			ctx:         ctx,
			req:         req,
			expectedErr: status.Errorf(codes.Internal, fmt.Errorf("unable to retrieve topic by ids: %w", ErrSomethingWentWrong).Error()),
			setup: func(ctx context.Context) {
				topicRepo.On("RetrieveByIDs", ctx, db, mock.Anything).Once().Return(nil, ErrSomethingWentWrong)
			},
		},
		{
			name:        " err bookRepo.RetrieveBookTreeByTopicIDs",
			ctx:         ctx,
			req:         req,
			expectedErr: status.Errorf(codes.Internal, fmt.Errorf("unable to retrieve los: %w", ErrSomethingWentWrong).Error()),
			setup: func(ctx context.Context) {
				topicRepo.On("RetrieveByIDs", ctx, db, mock.Anything).Once().Return([]*entities.Topic{
					{
						ID: database.Text("topic-1"),
					},
				}, nil)
				bookRepo.On("RetrieveBookTreeByTopicIDs", ctx, db, mock.Anything).Once().Return(nil, ErrSomethingWentWrong)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			_, err := s.RetrieveBookTreeByTopicIDs(testCase.ctx, testCase.req.(*bpb.RetrieveBookTreeByTopicIDsRequest))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

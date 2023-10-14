package services

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/bob/configurations"
	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	coursesRepo "github.com/manabie-com/backend/internal/bob/services/courses/repo"
	"github.com/manabie-com/backend/internal/bob/services/log"
	mediaRepo "github.com/manabie-com/backend/internal/bob/services/media/repo"
	topicsRepo "github.com/manabie-com/backend/internal/bob/services/topics/repo"
	cconstants "github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	mock_repositories "github.com/manabie-com/backend/mock/bob/repositories"
	mock_repositories_lessonmgmt "github.com/manabie-com/backend/mock/lessonmgmt/lesson/repositories"

	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	yasuo_mock_repositories "github.com/manabie-com/backend/mock/yasuo/repositories"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestLessonModifierService_Unpublish(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	lessonRepo := new(mock_repositories.MockLessonRepo)
	activityRepo := new(mock_repositories.MockActivityLogRepo)
	mockBobCfg := configurations.Config{Agora: configurations.AgoraConfig{MaximumLearnerStreamings: 13}}
	s := &LessonModifierServices{
		LessonRepo:      lessonRepo,
		DB:              db,
		Cfg:             mockBobCfg,
		ActivityLogRepo: activityRepo,
	}

	testCases := []TestCase{
		{
			name:         "happy case",
			ctx:          interceptors.ContextWithUserID(ctx, "id"),
			req:          &bpb.UnpublishRequest{LessonId: "0", LearnerId: "0"},
			expectedErr:  nil,
			expectedResp: &bpb.UnpublishResponse{},
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
				lessonRepo.On("DecreaseNumberOfStreaming", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				activityRepo.On("Create", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
		{
			name:         "unpublish before",
			ctx:          interceptors.ContextWithUserID(ctx, "id"),
			req:          &bpb.UnpublishRequest{LessonId: "0", LearnerId: "0"},
			expectedErr:  nil,
			expectedResp: &bpb.UnpublishResponse{Status: bpb.UnpublishStatus_UNPUBLISH_STATUS_UNPUBLISHED_BEFORE},
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Rollback", mock.Anything).Return(nil)
				lessonRepo.On("DecreaseNumberOfStreaming", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(repositories.ErrUnAffected)
				activityRepo.On("Create", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
		{
			name:         "err tx closed when update",
			ctx:          interceptors.ContextWithUserID(ctx, "id"),
			req:          &bpb.UnpublishRequest{LessonId: "0", LearnerId: "0"},
			expectedErr:  fmt.Errorf("ExecInTx: s.LessonStreamRepo.DecreaseNumberOfStreaming: %w", pgx.ErrTxClosed),
			expectedResp: &bpb.UnpublishResponse{Status: bpb.UnpublishStatus_UNPUBLISH_STATUS_UNPUBLISHED_BEFORE},
			setup: func(ctx context.Context) {
				tx := &mock_database.Tx{}
				db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Rollback", mock.Anything).Return(nil)
				lessonRepo.On("DecreaseNumberOfStreaming", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(pgx.ErrTxClosed)
				activityRepo.On("Create", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*bpb.UnpublishRequest)
			resp, err := s.Unpublish(testCase.ctx, req)
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

func TestLessonModifierService_Publish(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	lessonRepo := new(mock_repositories.MockLessonRepo)
	activityRepo := new(mock_repositories.MockActivityLogRepo)
	mockBobCfg := configurations.Config{Agora: configurations.AgoraConfig{MaximumLearnerStreamings: 13}}
	s := &LessonModifierServices{
		LessonRepo:      lessonRepo,
		DB:              db,
		Cfg:             mockBobCfg,
		ActivityLogRepo: activityRepo,
	}

	testCases := []TestCase{
		{
			name:         "happy case",
			ctx:          interceptors.ContextWithUserID(ctx, "id"),
			req:          &bpb.PreparePublishRequest{LessonId: "0", LearnerId: "3"},
			expectedErr:  nil,
			expectedResp: &bpb.PreparePublishResponse{},
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
				lessonRepo.On("GetStreamingLearners", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return([]string{"0", "1"}, nil)
				lessonRepo.On("IncreaseNumberOfStreaming", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				activityRepo.On("Create", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
		{
			name:         "prepared before",
			ctx:          interceptors.ContextWithUserID(ctx, "id"),
			req:          &bpb.PreparePublishRequest{LessonId: "0", LearnerId: "0"},
			expectedErr:  nil,
			expectedResp: &bpb.PreparePublishResponse{Status: bpb.PrepareToPublishStatus_PREPARE_TO_PUBLISH_STATUS_PREPARED_BEFORE},
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Rollback", mock.Anything).Return(nil)
				lessonRepo.On("GetStreamingLearners", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return([]string{"0", "1"}, nil)
				lessonRepo.On("IncreaseNumberOfStreaming", ctx, mock.Anything, mock.Anything, mock.Anything, mockBobCfg.Agora.MaximumLearnerStreamings).Once().Return(repositories.ErrUnAffected)
				activityRepo.On("Create", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
		{
			name:         "reached maximum limit",
			ctx:          interceptors.ContextWithUserID(ctx, "id"),
			req:          &bpb.PreparePublishRequest{LessonId: "0", LearnerId: "14"},
			expectedErr:  nil,
			expectedResp: &bpb.PreparePublishResponse{Status: bpb.PrepareToPublishStatus_PREPARE_TO_PUBLISH_STATUS_REACHED_MAX_UPSTREAM_LIMIT},
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Rollback", mock.Anything).Return(nil)
				mockIds := make([]string, 0, mockBobCfg.Agora.MaximumLearnerStreamings)
				for i := 0; i < mockBobCfg.Agora.MaximumLearnerStreamings; i++ {
					mockIds = append(mockIds, fmt.Sprintf("%v", i))
				}
				lessonRepo.On("GetStreamingLearners", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(mockIds, nil)
				lessonRepo.On("IncreaseNumberOfStreaming", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(repositories.ErrUnAffected)
				activityRepo.On("Create", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
		{
			name:         "err tx closed when update",
			ctx:          interceptors.ContextWithUserID(ctx, "id"),
			req:          &bpb.PreparePublishRequest{LessonId: "0", LearnerId: "14"},
			expectedErr:  fmt.Errorf("ExecInTx: s.LessonStreamRepo.IncreaseNumberOfStreaming: %w", pgx.ErrTxClosed),
			expectedResp: &bpb.PreparePublishResponse{},
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Rollback", mock.Anything).Return(nil)
				lessonRepo.On("GetStreamingLearners", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return([]string{"0", "1"}, nil)
				lessonRepo.On("IncreaseNumberOfStreaming", ctx, mock.Anything, mock.Anything, mock.Anything, mockBobCfg.Agora.MaximumLearnerStreamings).Once().Return(pgx.ErrTxClosed)
				activityRepo.On("Create", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*bpb.PreparePublishRequest)
			resp, err := s.PreparePublish(testCase.ctx, req)
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

func TestMaterialsToMedias(t *testing.T) {
	t.Parallel()
	t.Run("materials to medias successfully", func(t *testing.T) {
		inp := []*bpb.Material{
			{
				Resource: &bpb.Material_BrightcoveVideo_{
					BrightcoveVideo: &bpb.Material_BrightcoveVideo{
						Name: "video 1",
						Url:  "https://brightcove.com/account/2/video?videoId=abc123",
					},
				},
			},
			{
				Resource: &bpb.Material_MediaId{
					MediaId: "media-id-2",
				},
			},
		}
		medias, err := materialsToMedias(inp)
		require.NoError(t, err)
		assert.Len(t, medias, 2)

		assert.Empty(t, medias[0].MediaID.String)
		assert.Equal(t, "video 1", medias[0].Name.String)
		assert.Equal(t, "abc123", medias[0].Resource.String)
		assert.EqualValues(t, entities.MediaTypeVideo, medias[0].Type.String)

		assert.Equal(t, "media-id-2", medias[1].MediaID.String)
	})

	t.Run("materials to medias failed", func(t *testing.T) {
		inp := []*bpb.Material{
			{
				Resource: &bpb.Material_BrightcoveVideo_{
					BrightcoveVideo: &bpb.Material_BrightcoveVideo{
						Name: "video 1",
						Url:  "https://brightcove.com/account/2/video?videoID=abc123",
					},
				},
			},
			{
				Resource: &bpb.Material_MediaId{
					MediaId: "media-id-2",
				},
			},
		}
		_, err := materialsToMedias(inp)
		require.Error(t, err)
	})
}

type UserRepoMock struct {
	get       func(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*entities.User, error)
	userGroup func(context.Context, database.QueryExecer, pgtype.Text) (string, error)
}

func (u UserRepoMock) Get(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*entities.User, error) {
	return u.get(ctx, db, id)
}

func (u UserRepoMock) UserGroup(ctx context.Context, db database.QueryExecer, id pgtype.Text) (string, error) {
	return u.userGroup(ctx, db, id)
}

type SchoolAdminRepoMock struct {
	get func(ctx context.Context, db database.QueryExecer, schoolAdminID pgtype.Text) (*entities.SchoolAdmin, error)
}

func (s SchoolAdminRepoMock) Get(ctx context.Context, db database.QueryExecer, schoolAdminID pgtype.Text) (*entities.SchoolAdmin, error) {
	return s.get(ctx, db, schoolAdminID)
}

type TeacherRepoMock struct {
	retrieve func(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray, fields ...string) ([]entities.Teacher, error)
}

func (t TeacherRepoMock) Retrieve(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray, fields ...string) ([]entities.Teacher, error) {
	return t.retrieve(ctx, db, ids, fields...)
}

type StudentRepoMock struct {
	retrieve func(context.Context, database.QueryExecer, pgtype.TextArray) ([]repositories.StudentProfile, error)
}

func (s StudentRepoMock) Retrieve(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]repositories.StudentProfile, error) {
	return s.retrieve(ctx, db, ids)
}

type LessonGroupRepoMock struct {
	create       func(ctx context.Context, db database.QueryExecer, e *entities.LessonGroup) error
	get          func(ctx context.Context, db database.QueryExecer, lessonGroupID, courseID pgtype.Text) (*entities.LessonGroup, error)
	updateMedias func(ctx context.Context, db database.QueryExecer, e *entities.LessonGroup) error
}

func (l LessonGroupRepoMock) Create(ctx context.Context, db database.QueryExecer, e *entities.LessonGroup) error {
	return l.create(ctx, db, e)
}

func (l LessonGroupRepoMock) Get(ctx context.Context, db database.QueryExecer, lessonGroupID, courseID pgtype.Text) (*entities.LessonGroup, error) {
	return l.get(ctx, db, lessonGroupID, courseID)
}

func (l LessonGroupRepoMock) UpdateMedias(ctx context.Context, db database.QueryExecer, e *entities.LessonGroup) error {
	return l.updateMedias(ctx, db, e)
}

func isTx(db database.QueryExecer) bool {
	switch db.(type) {
	case pgx.Tx:
		return true
	}
	return false
}

func TestCreateLiveLesson(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	isUpsertedCourseByTestCase := make(map[string]bool)
	jsm := new(mock_nats.JetStreamManagement)
	tcs := []struct {
		name          string
		req           *bpb.CreateLiveLessonRequest
		caller        string // user id who call this api
		mediaRp       mediaRepo.MediaRepoMock
		lessonGroupRp LessonGroupRepoMock
		lessonRepo    LessonRepoMock
		courseRepo    coursesRepo.CourseRepoMock
		topicRepo     topicsRepo.TopicRepoMock
		pSPWRepo      coursesRepo.PresetStudyPlanWeeklyRepo
		pSPRepo       coursesRepo.PresetStudyPlanRepo
		userRp        UserRepo
		studentRp     StudentRepo
		schoolAdminRp SchoolAdminRepo
		teacherRp     TeacherRepo
		setup         func(context.Context)
		hasError      bool
	}{
		{
			name:   "create a live lesson successfully",
			caller: "school-admin-1",
			req: &bpb.CreateLiveLessonRequest{
				Name:       "Lesson name 1",
				StartTime:  timestamppb.New(time.Date(2020, 2, 3, 4, 5, 6, 7, time.UTC)),
				EndTime:    timestamppb.New(time.Date(2021, 2, 3, 4, 5, 6, 7, time.UTC)),
				TeacherIds: []string{"teacher-id-1", "teacher-id-2"},
				CourseIds:  []string{"course-id-1", "course-id-4", "course-id-4", "course-id-5"},
				LearnerIds: []string{"learner-id-2", "learner-id-3", "learner-id-4", "learner-id-5", "learner-id-6"},
				Materials: []*bpb.Material{
					{
						Resource: &bpb.Material_BrightcoveVideo_{
							BrightcoveVideo: &bpb.Material_BrightcoveVideo{
								Name: "video 1",
								Url:  "https://brightcove.com/account/2/video?videoId=abc123",
							},
						},
					},
					{
						Resource: &bpb.Material_MediaId{
							MediaId: "media-id-2",
						},
					},
					{
						Resource: &bpb.Material_MediaId{
							MediaId: "media-id-2",
						},
					},
					{
						Resource: &bpb.Material_MediaId{
							MediaId: "media-id-3",
						},
					},
				},
			},
			mediaRp: mediaRepo.MediaRepoMock{
				RetrieveByIDsMock: func(ctx context.Context, db database.QueryExecer, mediaIDs pgtype.TextArray) ([]*entities.Media, error) {
					assert.True(t, isTx(db), "retrieve media by IDs is not processing in transaction")
					actualMediaIDs := database.FromTextArray(mediaIDs)
					assert.ElementsMatch(t, actualMediaIDs, []string{"media-id-2", "media-id-3"})
					return []*entities.Media{
						{
							MediaID: database.Text("media-id-2"),
						},
						{
							MediaID: database.Text("media-id-3"),
						},
					}, nil
				},
				UpsertMediaBatchMock: func(ctx context.Context, db database.QueryExecer, medias entities.Medias) error {
					assert.Len(t, medias, 1)
					assert.EqualValues(t, entities.MediaTypeVideo, medias[0].Type.String)
					assert.Equal(t, "video 1", medias[0].Name.String)
					assert.Equal(t, "abc123", medias[0].Resource.String)
					err := medias.PreInsert()
					require.NoError(t, err)

					return nil
				},
			},
			lessonGroupRp: LessonGroupRepoMock{
				create: func(ctx context.Context, db database.QueryExecer, e *entities.LessonGroup) error {
					assert.True(t, isTx(db), "create lesson group is not processing in transaction")
					assert.Equal(t, "course-id-1", e.CourseID.String)

					// validate mediaIDs
					actualMedias := database.FromTextArray(e.MediaIDs)
					mediaIDsMap := make(map[string]bool)
					for _, mediaID := range actualMedias {
						mediaIDsMap[mediaID] = true
					}
					assert.Len(t, mediaIDsMap, 3)
					err := e.PreInsert()
					require.NoError(t, err)

					return nil
				},
			},
			lessonRepo: LessonRepoMock{
				create: func(ctx context.Context, db database.Ext, lesson *entities.Lesson) (*entities.Lesson, error) {
					assert.True(t, isTx(db), "create lesson group is not processing in transaction")

					assert.NotEmpty(t, lesson.LessonID.String)
					assert.Equal(t, "Lesson name 1", lesson.Name.String)
					assert.Equal(t, "teacher-id-1", lesson.TeacherID.String)
					assert.Equal(t, "course-id-1", lesson.CourseID.String)
					assert.Equal(t, time.Date(2020, 2, 3, 4, 5, 6, 7, time.UTC).Unix(), lesson.StartTime.Time.Unix())
					assert.Equal(t, time.Date(2021, 2, 3, 4, 5, 6, 7, time.UTC).Unix(), lesson.EndTime.Time.Unix())
					assert.NotEmpty(t, lesson.LessonGroupID.String)
					assert.EqualValues(t, entities.LessonTypeOnline, lesson.LessonType.String)
					assert.EqualValues(t, entities.LessonTeachingMediumOnline, lesson.TeachingMedium.String)
					assert.EqualValues(t, entities.LessonStatusDraft, lesson.Status.String)
					assert.EqualValues(t, 0, lesson.StreamLearnerCounter.Int)

					return lesson, nil
				},
				upsertLessonCourses: func(ctx context.Context, db database.Ext, lessonID pgtype.Text, courseIDs pgtype.TextArray) error {
					assert.True(t, isTx(db), "upsert lesson courses is not processing in transaction")
					assert.NotEmpty(t, lessonID.String)
					assert.Equal(t, []string{"course-id-1", "course-id-4", "course-id-5"}, database.FromTextArray(courseIDs))
					return nil
				},
				upsertLessonMembers: func(ctx context.Context, db database.Ext, lessonID pgtype.Text, userIDs pgtype.TextArray) error {
					assert.True(t, isTx(db), "upsert lesson members is not processing in transaction")
					assert.Equal(t, []string{"learner-id-2", "learner-id-3", "learner-id-4", "learner-id-5", "learner-id-6"}, database.FromTextArray(userIDs))
					return nil
				},
				upsertLessonTeachers: func(ctx context.Context, db database.Ext, lessonID pgtype.Text, teacherIDs pgtype.TextArray) error {
					assert.True(t, isTx(db), "upsert lesson teachers is not processing in transaction")
					assert.Equal(t, []string{"teacher-id-1", "teacher-id-2"}, database.FromTextArray(teacherIDs))
					return nil
				},
				findEarliestAndLatestTimeLessonByCourses: func(ctx context.Context, db database.Ext, courseIDs pgtype.TextArray) (*entities.CourseAvailableRanges, error) {
					assert.ElementsMatch(t, []string{"course-id-1", "course-id-4", "course-id-5"}, database.FromTextArray(courseIDs))

					res := &entities.CourseAvailableRanges{}
					res.Add(
						&entities.CourseAvailableRange{
							ID:        database.Text("course-id-1"),
							StartDate: database.Timestamptz(time.Date(2019, 2, 3, 4, 5, 6, 7, time.UTC)),
							EndDate:   database.Timestamptz(time.Date(2020, 2, 3, 4, 5, 6, 7, time.UTC)),
						},
						&entities.CourseAvailableRange{
							ID:        database.Text("course-id-4"),
							StartDate: database.Timestamptz(time.Date(2020, 2, 3, 4, 5, 6, 7, time.UTC)),
							EndDate:   database.Timestamptz(time.Date(2021, 2, 3, 4, 5, 6, 7, time.UTC)),
						},
						&entities.CourseAvailableRange{
							ID:        database.Text("course-id-5"),
							StartDate: database.Timestamptz(time.Date(2018, 2, 3, 4, 5, 6, 7, time.UTC)),
							EndDate:   database.Timestamptz(time.Date(2021, 2, 3, 4, 5, 6, 7, time.UTC)),
						},
					)
					return res, nil
				},
			},
			courseRepo: coursesRepo.CourseRepoMock{
				FindByIDsMock: func(ctx context.Context, db database.Ext, courseIDs pgtype.TextArray) (map[pgtype.Text]*entities.Course, error) {
					presetStudyPlanIDs := make(map[string]pgtype.Text)
					if _, ok := isUpsertedCourseByTestCase["create a live lesson successfully"]; ok {
						presetStudyPlanIDs["course-id-1"] = database.Text("preset-study-plan-id-1")
						presetStudyPlanIDs["course-id-4"] = database.Text("preset-study-plan-id-2")
						presetStudyPlanIDs["course-id-5"] = database.Text("preset-study-plan-id-3")
					}

					assert.ElementsMatch(t, []string{"course-id-1", "course-id-4", "course-id-5"}, database.FromTextArray(courseIDs))
					return map[pgtype.Text]*entities.Course{
						database.Text("course-id-1"): {
							ID:                database.Text("course-id-1"),
							Name:              database.Text("math level 1"),
							PresetStudyPlanID: presetStudyPlanIDs["course-id-1"],
							Country:           database.Text("vietnam"),
							Grade:             database.Int2(2),
							Subject:           database.Text("math"),
							SchoolID:          database.Int4(1),
							StartDate:         database.Timestamptz(time.Date(2021, 2, 3, 4, 5, 6, 7, time.UTC)),
						},
						database.Text("course-id-4"): {
							ID:                database.Text("course-id-4"),
							Name:              database.Text("physics level 2"),
							PresetStudyPlanID: presetStudyPlanIDs["course-id-4"],
							Country:           database.Text("vietnam"),
							Grade:             database.Int2(5),
							Subject:           database.Text("physics"),
							SchoolID:          database.Int4(1),
							StartDate:         database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.UTC)),
						},
						database.Text("course-id-5"): {
							ID:                database.Text("course-id-5"),
							Name:              database.Text("chemistry level 1"),
							PresetStudyPlanID: presetStudyPlanIDs["course-id-5"],
							Country:           database.Text("vietnam"),
							Grade:             database.Int2(7),
							Subject:           database.Text("chemistry"),
							SchoolID:          database.Int4(1),
							StartDate:         database.Timestamptz(time.Date(2023, 2, 3, 4, 5, 6, 7, time.UTC)),
						},
					}, nil
				},
				UpsertMock: func(ctx context.Context, db database.Ext, cc []*entities.Course) error {
					assert.Len(t, cc, 3)
					if isUpsertedCourseByTestCase["create a live lesson successfully"] {
						expected := map[string]*entities.CourseAvailableRange{
							"course-id-1": {
								ID:        database.Text("course-id-1"),
								StartDate: database.Timestamptz(time.Date(2019, 2, 3, 4, 5, 6, 7, time.UTC)),
								EndDate:   database.Timestamptz(time.Date(2020, 2, 3, 4, 5, 6, 7, time.UTC)),
							},
							"course-id-4": {
								ID:        database.Text("course-id-4"),
								StartDate: database.Timestamptz(time.Date(2020, 2, 3, 4, 5, 6, 7, time.UTC)),
								EndDate:   database.Timestamptz(time.Date(2021, 2, 3, 4, 5, 6, 7, time.UTC)),
							},
							"course-id-5": {
								ID:        database.Text("course-id-5"),
								StartDate: database.Timestamptz(time.Date(2018, 2, 3, 4, 5, 6, 7, time.UTC)),
								EndDate:   database.Timestamptz(time.Date(2021, 2, 3, 4, 5, 6, 7, time.UTC)),
							},
						}

						// test start data and end data
						for _, c := range cc {
							if v, ok := expected[c.ID.String]; ok {
								assert.Equal(t, v.StartDate, c.StartDate)
								assert.Equal(t, v.EndDate, c.EndDate)
							}
						}
					}

					courseIDs := make([]string, 0, len(cc))
					for _, c := range cc {
						assert.NotEmpty(t, c.ID.String)
						assert.NotEmpty(t, c.PresetStudyPlanID.String)
						courseIDs = append(courseIDs, c.ID.String)
					}
					assert.ElementsMatch(t, []string{"course-id-1", "course-id-4", "course-id-5"}, courseIDs)
					isUpsertedCourseByTestCase["create a live lesson successfully"] = true

					return nil
				},
			},
			topicRepo: topicsRepo.TopicRepoMock{
				CreateMock: func(ctx context.Context, db database.Ext, plans []*entities.Topic) error {
					assert.Len(t, plans, 3)
					assert.NotEmpty(t, plans[0].ID.String)
					assert.NotZero(t, plans[0].PublishedAt.Time)
					assert.NotEmpty(t, plans[1].ID.String)
					assert.NotZero(t, plans[1].PublishedAt.Time)
					assert.NotEmpty(t, plans[2].ID.String)
					assert.NotZero(t, plans[2].PublishedAt.Time)

					plansBySubject := make(map[string]*entities.Topic)
					for i := range plans {
						plansBySubject[plans[i].Subject.String] = plans[i]
					}

					e := &entities.Topic{}
					database.AllNullEntity(e)
					err := multierr.Combine(
						e.ID.Set(plansBySubject["math"].ID),
						e.Name.Set("Lesson name 1"),
						e.Country.Set("vietnam"),
						e.Grade.Set(2),
						e.Subject.Set("math"),
						e.SchoolID.Set(1),
						e.TopicType.Set(entities.TopicTypeLiveLesson),
						e.Status.Set(entities.TopicStatusPublished),
						e.DisplayOrder.Set(1),
						e.PublishedAt.Set(plansBySubject["math"].PublishedAt),
						e.TotalLOs.Set(0),
						e.ChapterID.Set(nil),
						e.IconURL.Set(nil),
						e.DeletedAt.Set(nil),
						e.EssayRequired.Set(false),
					)
					require.NoError(t, err)
					assert.Equal(t, e, plansBySubject["math"])

					database.AllNullEntity(e)
					err = multierr.Combine(
						e.ID.Set(plansBySubject["physics"].ID),
						e.Name.Set("Lesson name 1"),
						e.Country.Set("vietnam"),
						e.Grade.Set(5),
						e.Subject.Set("physics"),
						e.SchoolID.Set(1),
						e.TopicType.Set(entities.TopicTypeLiveLesson),
						e.Status.Set(entities.TopicStatusPublished),
						e.DisplayOrder.Set(1),
						e.PublishedAt.Set(plansBySubject["physics"].PublishedAt),
						e.TotalLOs.Set(0),
						e.ChapterID.Set(nil),
						e.IconURL.Set(nil),
						e.DeletedAt.Set(nil),
						e.EssayRequired.Set(false),
					)
					require.NoError(t, err)
					assert.Equal(t, e, plansBySubject["physics"])

					database.AllNullEntity(e)
					err = multierr.Combine(
						e.ID.Set(plansBySubject["chemistry"].ID),
						e.Name.Set("Lesson name 1"),
						e.Country.Set("vietnam"),
						e.Grade.Set(7),
						e.Subject.Set("chemistry"),
						e.SchoolID.Set(1),
						e.TopicType.Set(entities.TopicTypeLiveLesson),
						e.Status.Set(entities.TopicStatusPublished),
						e.DisplayOrder.Set(1),
						e.PublishedAt.Set(plansBySubject["chemistry"].PublishedAt),
						e.TotalLOs.Set(0),
						e.ChapterID.Set(nil),
						e.IconURL.Set(nil),
						e.DeletedAt.Set(nil),
						e.EssayRequired.Set(false),
					)
					require.NoError(t, err)
					assert.Equal(t, e, plansBySubject["chemistry"])

					return nil
				},
			},
			pSPRepo: coursesRepo.PresetStudyPlanRepoMock{
				CreatePresetStudyPlanMock: func(ctx context.Context, db database.Ext, presetStudyPlans []*entities.PresetStudyPlan) error {
					assert.Len(t, presetStudyPlans, 3)
					assert.NotEmpty(t, presetStudyPlans[0].ID.String)
					assert.NotEmpty(t, presetStudyPlans[1].ID.String)

					pspBySubject := make(map[string]*entities.PresetStudyPlan)
					for i := range presetStudyPlans {
						pspBySubject[presetStudyPlans[i].Subject.String] = presetStudyPlans[i]
					}

					expected := &entities.PresetStudyPlan{}
					database.AllNullEntity(expected)
					err := multierr.Combine(
						expected.ID.Set(pspBySubject["math"].ID),
						expected.Name.Set("math level 1"),
						expected.Country.Set("vietnam"),
						expected.Grade.Set(2),
						expected.Subject.Set("math"),
						expected.StartDate.Set(time.Date(2021, 2, 3, 4, 5, 6, 7, time.UTC)),
					)
					require.NoError(t, err)
					assert.Equal(t, expected, pspBySubject["math"])

					database.AllNullEntity(expected)
					err = multierr.Combine(
						expected.ID.Set(pspBySubject["physics"].ID),
						expected.Name.Set("physics level 2"),
						expected.Country.Set("vietnam"),
						expected.Grade.Set(5),
						expected.Subject.Set("physics"),
						expected.StartDate.Set(time.Date(2022, 2, 3, 4, 5, 6, 7, time.UTC)),
					)
					require.NoError(t, err)
					assert.Equal(t, expected, pspBySubject["physics"])

					database.AllNullEntity(expected)
					err = multierr.Combine(
						expected.ID.Set(pspBySubject["chemistry"].ID),
						expected.Name.Set("chemistry level 1"),
						expected.Country.Set("vietnam"),
						expected.Grade.Set(7),
						expected.Subject.Set("chemistry"),
						expected.StartDate.Set(time.Date(2023, 2, 3, 4, 5, 6, 7, time.UTC)),
					)
					require.NoError(t, err)
					assert.Equal(t, expected, pspBySubject["chemistry"])

					return nil
				},
			},
			pSPWRepo: coursesRepo.PresetStudyPlanWeeklyRepoMock{
				CreateMock: func(ctx context.Context, db database.Ext, plans []*entities.PresetStudyPlanWeekly) error {
					assert.Len(t, plans, 3)

					assert.NotEmpty(t, plans[0].ID.String)
					assert.NotEmpty(t, plans[0].TopicID.String)
					assert.Equal(t, int16(0), plans[0].Week.Int)
					assert.NotEmpty(t, plans[0].LessonID.String)
					assert.Equal(t, time.Date(2020, 2, 3, 4, 5, 6, 7, time.UTC), plans[0].StartDate.Time.UTC())
					assert.Equal(t, time.Date(2021, 2, 3, 4, 5, 6, 7, time.UTC), plans[0].EndDate.Time.UTC())

					assert.NotEmpty(t, plans[1].ID.String)
					assert.NotEmpty(t, plans[1].TopicID.String)
					assert.Equal(t, int16(0), plans[1].Week.Int)
					assert.NotEmpty(t, plans[1].LessonID.String)
					assert.Equal(t, time.Date(2020, 2, 3, 4, 5, 6, 7, time.UTC), plans[1].StartDate.Time.UTC())
					assert.Equal(t, time.Date(2021, 2, 3, 4, 5, 6, 7, time.UTC), plans[1].EndDate.Time.UTC())

					assert.NotEmpty(t, plans[2].ID.String)
					assert.NotEmpty(t, plans[2].TopicID.String)
					assert.Equal(t, int16(0), plans[2].Week.Int)
					assert.NotEmpty(t, plans[2].LessonID.String)
					assert.Equal(t, time.Date(2020, 2, 3, 4, 5, 6, 7, time.UTC), plans[1].StartDate.Time.UTC())
					assert.Equal(t, time.Date(2021, 2, 3, 4, 5, 6, 7, time.UTC), plans[1].EndDate.Time.UTC())

					actualPresetStudyPlanID := []string{plans[0].PresetStudyPlanID.String, plans[1].PresetStudyPlanID.String, plans[2].PresetStudyPlanID.String}
					expectedPresetStudyPlanID := []string{"preset-study-plan-id-1", "preset-study-plan-id-2", "preset-study-plan-id-3"}
					assert.ElementsMatch(t, expectedPresetStudyPlanID, actualPresetStudyPlanID)

					return nil
				},
			},
			userRp: UserRepoMock{
				get: func(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*entities.User, error) {
					assert.Equal(t, "school-admin-1", id.String)
					return &entities.User{
						ID:      database.Text("school-admin-1"),
						Country: database.Text("vietnam"),
					}, nil
				},
			},
			schoolAdminRp: SchoolAdminRepoMock{
				get: func(ctx context.Context, db database.QueryExecer, schoolAdminID pgtype.Text) (*entities.SchoolAdmin, error) {
					assert.Equal(t, "school-admin-1", schoolAdminID.String)
					return &entities.SchoolAdmin{
						SchoolAdminID: database.Text("school-admin-1"),
						SchoolID:      database.Int4(1),
					}, nil
				},
			},
			teacherRp: TeacherRepoMock{
				retrieve: func(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray, fields ...string) ([]entities.Teacher, error) {
					assert.ElementsMatch(t, []string{"teacher-id-1", "teacher-id-2"}, database.FromTextArray(ids))
					return []entities.Teacher{
						{
							ID:        database.Text("teacher-id-1"),
							SchoolIDs: database.Int4Array([]int32{3, 1}),
							User: entities.User{
								Country: database.Text("vietnam"),
							},
						},
						{
							ID:        database.Text("teacher-id-2"),
							SchoolIDs: database.Int4Array([]int32{1}),
							User: entities.User{
								Country: database.Text("vietnam"),
							},
						},
					}, nil
				},
			},
			studentRp: StudentRepoMock{
				retrieve: func(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]repositories.StudentProfile, error) {
					assert.ElementsMatch(t, []string{"learner-id-2", "learner-id-3", "learner-id-4", "learner-id-5", "learner-id-6"}, database.FromTextArray(ids))
					return []repositories.StudentProfile{
						{
							Student: entities.Student{
								ID:       database.Text("learner-id-2"),
								SchoolID: database.Int4(1),
								User: entities.User{
									ID:      database.Text("learner-id-2"),
									Country: database.Text("vietnam"),
								},
							},
							School: entities.School{
								ID: database.Int4(1),
							},
						},
						{
							Student: entities.Student{
								ID:       database.Text("learner-id-3"),
								SchoolID: database.Int4(1),
								User: entities.User{
									ID:      database.Text("learner-id-3"),
									Country: database.Text("vietnam"),
								},
							},
							School: entities.School{
								ID: database.Int4(1),
							},
						},
						{
							Student: entities.Student{
								ID:       database.Text("learner-id-4"),
								SchoolID: database.Int4(1),
								User: entities.User{
									ID:      database.Text("learner-id-4"),
									Country: database.Text("vietnam"),
								},
							},
							School: entities.School{
								ID: database.Int4(1),
							},
						},
						{
							Student: entities.Student{
								ID:       database.Text("learner-id-5"),
								SchoolID: database.Int4(1),
								User: entities.User{
									ID:      database.Text("learner-id-5"),
									Country: database.Text("vietnam"),
								},
							},
							School: entities.School{
								ID: database.Int4(1),
							},
						},
						{
							Student: entities.Student{
								ID:       database.Text("learner-id-6"),
								SchoolID: database.Int4(1),
								User: entities.User{
									ID:      database.Text("learner-id-6"),
									Country: database.Text("vietnam"),
								},
							},
							School: entities.School{
								ID: database.Int4(1),
							},
						},
					}, nil
				},
			},
			setup: func(ctx context.Context) {
				ctx = interceptors.ContextWithUserID(ctx, "school-admin-1")
				jsm.On("PublishAsyncContext", ctx, cconstants.SubjectLessonCreated, mock.Anything).Run(func(args mock.Arguments) {
					data := args.Get(2).([]byte)
					msg := &pb.EvtLesson{}
					err := msg.Unmarshal(data)
					require.NoError(t, err)

					message := msg.Message.(*pb.EvtLesson_CreateLessons_)
					assert.NotEmpty(t, message.CreateLessons.Lessons[0].LessonId)
					assert.Equal(t, "Lesson name 1", message.CreateLessons.Lessons[0].Name)
				}).Return("", nil).Once()
			},
		},
		{
			name:   "create a live lesson with missing materials successfully",
			caller: "school-admin-1",
			req: &bpb.CreateLiveLessonRequest{
				Name:       "Lesson name 1",
				StartTime:  timestamppb.New(time.Date(2020, 2, 3, 4, 5, 6, 7, time.UTC)),
				EndTime:    timestamppb.New(time.Date(2021, 2, 3, 4, 5, 6, 7, time.UTC)),
				TeacherIds: []string{"teacher-id-1", "teacher-id-2"},
				CourseIds:  []string{"course-id-1", "course-id-4", "course-id-4", "course-id-5"},
				LearnerIds: []string{"learner-id-2", "learner-id-3", "learner-id-4", "learner-id-5", "learner-id-6"},
			},
			mediaRp: mediaRepo.MediaRepoMock{
				RetrieveByIDsMock: func(ctx context.Context, db database.QueryExecer, mediaIDs pgtype.TextArray) ([]*entities.Media, error) {
					require.Fail(t, "expected retrieveByIDs not be called")
					return nil, nil
				},
				UpsertMediaBatchMock: func(ctx context.Context, db database.QueryExecer, medias entities.Medias) error {
					require.Fail(t, "expected upsertMediaBatch not be called")
					return nil
				},
			},
			lessonGroupRp: LessonGroupRepoMock{
				create: func(ctx context.Context, db database.QueryExecer, e *entities.LessonGroup) error {
					assert.True(t, isTx(db), "create lesson group is not processing in transaction")
					assert.Equal(t, "course-id-1", e.CourseID.String)

					// validate mediaIDs
					assert.Len(t, e.MediaIDs.Elements, 0)
					err := e.PreInsert()
					require.NoError(t, err)

					return nil
				},
			},
			lessonRepo: LessonRepoMock{
				create: func(ctx context.Context, db database.Ext, lesson *entities.Lesson) (*entities.Lesson, error) {
					assert.True(t, isTx(db), "create lesson group is not processing in transaction")

					assert.NotEmpty(t, lesson.LessonID.String)
					assert.Equal(t, "Lesson name 1", lesson.Name.String)
					assert.Equal(t, "teacher-id-1", lesson.TeacherID.String)
					assert.Equal(t, "course-id-1", lesson.CourseID.String)
					assert.Equal(t, time.Date(2020, 2, 3, 4, 5, 6, 7, time.UTC).Unix(), lesson.StartTime.Time.Unix())
					assert.Equal(t, time.Date(2021, 2, 3, 4, 5, 6, 7, time.UTC).Unix(), lesson.EndTime.Time.Unix())
					assert.NotEmpty(t, lesson.LessonGroupID.String)
					assert.EqualValues(t, entities.LessonTypeOnline, lesson.LessonType.String)
					assert.EqualValues(t, entities.LessonTeachingMediumOnline, lesson.TeachingMedium.String)
					assert.EqualValues(t, entities.LessonStatusDraft, lesson.Status.String)
					assert.EqualValues(t, 0, lesson.StreamLearnerCounter.Int)

					return lesson, nil
				},
				upsertLessonCourses: func(ctx context.Context, db database.Ext, lessonID pgtype.Text, courseIDs pgtype.TextArray) error {
					assert.True(t, isTx(db), "upsert lesson courses is not processing in transaction")
					assert.NotEmpty(t, lessonID.String)
					assert.Equal(t, []string{"course-id-1", "course-id-4", "course-id-5"}, database.FromTextArray(courseIDs))
					return nil
				},
				upsertLessonMembers: func(ctx context.Context, db database.Ext, lessonID pgtype.Text, userIDs pgtype.TextArray) error {
					assert.True(t, isTx(db), "upsert lesson members is not processing in transaction")
					assert.Equal(t, []string{"learner-id-2", "learner-id-3", "learner-id-4", "learner-id-5", "learner-id-6"}, database.FromTextArray(userIDs))
					return nil
				},
				upsertLessonTeachers: func(ctx context.Context, db database.Ext, lessonID pgtype.Text, teacherIDs pgtype.TextArray) error {
					assert.True(t, isTx(db), "upsert lesson teachers is not processing in transaction")
					assert.Equal(t, []string{"teacher-id-1", "teacher-id-2"}, database.FromTextArray(teacherIDs))
					return nil
				},
				findEarliestAndLatestTimeLessonByCourses: func(ctx context.Context, db database.Ext, courseIDs pgtype.TextArray) (*entities.CourseAvailableRanges, error) {
					assert.ElementsMatch(t, []string{"course-id-1", "course-id-4", "course-id-5"}, database.FromTextArray(courseIDs))

					res := &entities.CourseAvailableRanges{}
					res.Add(
						&entities.CourseAvailableRange{
							ID:        database.Text("course-id-1"),
							StartDate: database.Timestamptz(time.Date(2019, 2, 3, 4, 5, 6, 7, time.UTC)),
							EndDate:   database.Timestamptz(time.Date(2020, 2, 3, 4, 5, 6, 7, time.UTC)),
						},
						&entities.CourseAvailableRange{
							ID:        database.Text("course-id-4"),
							StartDate: database.Timestamptz(time.Date(2020, 2, 3, 4, 5, 6, 7, time.UTC)),
							EndDate:   database.Timestamptz(time.Date(2021, 2, 3, 4, 5, 6, 7, time.UTC)),
						},
						&entities.CourseAvailableRange{
							ID:        database.Text("course-id-5"),
							StartDate: database.Timestamptz(time.Date(2018, 2, 3, 4, 5, 6, 7, time.UTC)),
							EndDate:   database.Timestamptz(time.Date(2021, 2, 3, 4, 5, 6, 7, time.UTC)),
						},
					)
					return res, nil
				},
			},
			courseRepo: coursesRepo.CourseRepoMock{
				FindByIDsMock: func(ctx context.Context, db database.Ext, courseIDs pgtype.TextArray) (map[pgtype.Text]*entities.Course, error) {
					assert.ElementsMatch(t, []string{"course-id-1", "course-id-4", "course-id-5"}, database.FromTextArray(courseIDs))
					presetStudyPlanIDs := make(map[string]pgtype.Text)
					if _, ok := isUpsertedCourseByTestCase["create a live lesson with missing materials successfully"]; ok {
						presetStudyPlanIDs["course-id-1"] = database.Text("preset-study-plan-id-1")
						presetStudyPlanIDs["course-id-4"] = database.Text("preset-study-plan-id-2")
						presetStudyPlanIDs["course-id-5"] = database.Text("preset-study-plan-id-3")
					}
					assert.ElementsMatch(t, []string{"course-id-1", "course-id-4", "course-id-5"}, database.FromTextArray(courseIDs))

					return map[pgtype.Text]*entities.Course{
						database.Text("course-id-1"): {
							ID:                database.Text("course-id-1"),
							Name:              database.Text("math level 1"),
							PresetStudyPlanID: presetStudyPlanIDs["course-id-1"],
							Country:           database.Text("vietnam"),
							Grade:             database.Int2(2),
							Subject:           database.Text("math"),
							SchoolID:          database.Int4(1),
							StartDate:         database.Timestamptz(time.Date(2021, 2, 3, 4, 5, 6, 7, time.UTC)),
						},
						database.Text("course-id-4"): {
							ID:                database.Text("course-id-4"),
							Name:              database.Text("physics level 2"),
							PresetStudyPlanID: presetStudyPlanIDs["course-id-4"],
							Country:           database.Text("vietnam"),
							Grade:             database.Int2(5),
							Subject:           database.Text("physics"),
							SchoolID:          database.Int4(1),
							StartDate:         database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.UTC)),
						},
						database.Text("course-id-5"): {
							ID:                database.Text("course-id-5"),
							Name:              database.Text("chemistry level 1"),
							PresetStudyPlanID: presetStudyPlanIDs["course-id-5"],
							Country:           database.Text("vietnam"),
							Grade:             database.Int2(7),
							Subject:           database.Text("chemistry"),
							SchoolID:          database.Int4(1),
							StartDate:         database.Timestamptz(time.Date(2023, 2, 3, 4, 5, 6, 7, time.UTC)),
						},
					}, nil
				},
				UpsertMock: func(ctx context.Context, db database.Ext, cc []*entities.Course) error {
					assert.Len(t, cc, 3)
					if isUpsertedCourseByTestCase["create a live lesson with missing materials successfully"] {
						expected := map[string]*entities.CourseAvailableRange{
							"course-id-1": {
								ID:        database.Text("course-id-1"),
								StartDate: database.Timestamptz(time.Date(2019, 2, 3, 4, 5, 6, 7, time.UTC)),
								EndDate:   database.Timestamptz(time.Date(2020, 2, 3, 4, 5, 6, 7, time.UTC)),
							},
							"course-id-4": {
								ID:        database.Text("course-id-4"),
								StartDate: database.Timestamptz(time.Date(2020, 2, 3, 4, 5, 6, 7, time.UTC)),
								EndDate:   database.Timestamptz(time.Date(2021, 2, 3, 4, 5, 6, 7, time.UTC)),
							},
							"course-id-5": {
								ID:        database.Text("course-id-5"),
								StartDate: database.Timestamptz(time.Date(2018, 2, 3, 4, 5, 6, 7, time.UTC)),
								EndDate:   database.Timestamptz(time.Date(2021, 2, 3, 4, 5, 6, 7, time.UTC)),
							},
						}

						// test start data and end data
						for _, c := range cc {
							if v, ok := expected[c.ID.String]; ok {
								assert.Equal(t, v.StartDate, c.StartDate)
								assert.Equal(t, v.EndDate, c.EndDate)
							}
						}
					}

					courseIDs := make([]string, 0, len(cc))
					for _, c := range cc {
						assert.NotEmpty(t, c.ID.String)
						assert.NotEmpty(t, c.PresetStudyPlanID.String)
						courseIDs = append(courseIDs, c.ID.String)
					}
					assert.ElementsMatch(t, []string{"course-id-1", "course-id-4", "course-id-5"}, courseIDs)
					isUpsertedCourseByTestCase["create a live lesson with missing materials successfully"] = true

					return nil
				},
			},
			topicRepo: topicsRepo.TopicRepoMock{
				CreateMock: func(ctx context.Context, db database.Ext, plans []*entities.Topic) error {
					assert.Len(t, plans, 3)
					assert.NotEmpty(t, plans[0].ID.String)
					assert.NotZero(t, plans[0].PublishedAt.Time)
					assert.NotEmpty(t, plans[1].ID.String)
					assert.NotZero(t, plans[1].PublishedAt.Time)
					assert.NotEmpty(t, plans[2].ID.String)
					assert.NotZero(t, plans[2].PublishedAt.Time)

					plansBySubject := make(map[string]*entities.Topic)
					for i := range plans {
						plansBySubject[plans[i].Subject.String] = plans[i]
					}

					e := &entities.Topic{}
					database.AllNullEntity(e)
					err := multierr.Combine(
						e.ID.Set(plansBySubject["math"].ID),
						e.Name.Set("Lesson name 1"),
						e.Country.Set("vietnam"),
						e.Grade.Set(2),
						e.Subject.Set("math"),
						e.SchoolID.Set(1),
						e.TopicType.Set(entities.TopicTypeLiveLesson),
						e.Status.Set(entities.TopicStatusPublished),
						e.DisplayOrder.Set(1),
						e.PublishedAt.Set(plansBySubject["math"].PublishedAt),
						e.TotalLOs.Set(0),
						e.ChapterID.Set(nil),
						e.IconURL.Set(nil),
						e.DeletedAt.Set(nil),
						e.EssayRequired.Set(false),
					)
					require.NoError(t, err)
					assert.Equal(t, e, plansBySubject["math"])

					database.AllNullEntity(e)
					err = multierr.Combine(
						e.ID.Set(plansBySubject["physics"].ID),
						e.Name.Set("Lesson name 1"),
						e.Country.Set("vietnam"),
						e.Grade.Set(5),
						e.Subject.Set("physics"),
						e.SchoolID.Set(1),
						e.TopicType.Set(entities.TopicTypeLiveLesson),
						e.Status.Set(entities.TopicStatusPublished),
						e.DisplayOrder.Set(1),
						e.PublishedAt.Set(plansBySubject["physics"].PublishedAt),
						e.TotalLOs.Set(0),
						e.ChapterID.Set(nil),
						e.IconURL.Set(nil),
						e.DeletedAt.Set(nil),
						e.EssayRequired.Set(false),
					)
					require.NoError(t, err)
					assert.Equal(t, e, plansBySubject["physics"])

					database.AllNullEntity(e)
					err = multierr.Combine(
						e.ID.Set(plansBySubject["chemistry"].ID),
						e.Name.Set("Lesson name 1"),
						e.Country.Set("vietnam"),
						e.Grade.Set(7),
						e.Subject.Set("chemistry"),
						e.SchoolID.Set(1),
						e.TopicType.Set(entities.TopicTypeLiveLesson),
						e.Status.Set(entities.TopicStatusPublished),
						e.DisplayOrder.Set(1),
						e.PublishedAt.Set(plansBySubject["chemistry"].PublishedAt),
						e.TotalLOs.Set(0),
						e.ChapterID.Set(nil),
						e.IconURL.Set(nil),
						e.DeletedAt.Set(nil),
						e.EssayRequired.Set(false),
					)
					require.NoError(t, err)
					assert.Equal(t, e, plansBySubject["chemistry"])

					return nil
				},
			},
			pSPRepo: coursesRepo.PresetStudyPlanRepoMock{
				CreatePresetStudyPlanMock: func(ctx context.Context, db database.Ext, presetStudyPlans []*entities.PresetStudyPlan) error {
					assert.Len(t, presetStudyPlans, 3)
					assert.NotEmpty(t, presetStudyPlans[0].ID.String)
					assert.NotEmpty(t, presetStudyPlans[1].ID.String)

					pspBySubject := make(map[string]*entities.PresetStudyPlan)
					for i := range presetStudyPlans {
						pspBySubject[presetStudyPlans[i].Subject.String] = presetStudyPlans[i]
					}

					expected := &entities.PresetStudyPlan{}
					database.AllNullEntity(expected)
					err := multierr.Combine(
						expected.ID.Set(pspBySubject["math"].ID),
						expected.Name.Set("math level 1"),
						expected.Country.Set("vietnam"),
						expected.Grade.Set(2),
						expected.Subject.Set("math"),
						expected.StartDate.Set(time.Date(2021, 2, 3, 4, 5, 6, 7, time.UTC)),
					)
					require.NoError(t, err)
					assert.Equal(t, expected, pspBySubject["math"])

					database.AllNullEntity(expected)
					err = multierr.Combine(
						expected.ID.Set(pspBySubject["physics"].ID),
						expected.Name.Set("physics level 2"),
						expected.Country.Set("vietnam"),
						expected.Grade.Set(5),
						expected.Subject.Set("physics"),
						expected.StartDate.Set(time.Date(2022, 2, 3, 4, 5, 6, 7, time.UTC)),
					)
					require.NoError(t, err)
					assert.Equal(t, expected, pspBySubject["physics"])

					database.AllNullEntity(expected)
					err = multierr.Combine(
						expected.ID.Set(pspBySubject["chemistry"].ID),
						expected.Name.Set("chemistry level 1"),
						expected.Country.Set("vietnam"),
						expected.Grade.Set(7),
						expected.Subject.Set("chemistry"),
						expected.StartDate.Set(time.Date(2023, 2, 3, 4, 5, 6, 7, time.UTC)),
					)
					require.NoError(t, err)
					assert.Equal(t, expected, pspBySubject["chemistry"])

					return nil
				},
			},
			pSPWRepo: coursesRepo.PresetStudyPlanWeeklyRepoMock{
				CreateMock: func(ctx context.Context, db database.Ext, plans []*entities.PresetStudyPlanWeekly) error {
					assert.Len(t, plans, 3)

					assert.NotEmpty(t, plans[0].ID.String)
					assert.NotEmpty(t, plans[0].TopicID.String)
					assert.Equal(t, int16(0), plans[0].Week.Int)
					assert.NotEmpty(t, plans[0].LessonID.String)
					assert.Equal(t, time.Date(2020, 2, 3, 4, 5, 6, 7, time.UTC), plans[0].StartDate.Time.UTC())
					assert.Equal(t, time.Date(2021, 2, 3, 4, 5, 6, 7, time.UTC), plans[0].EndDate.Time.UTC())

					assert.NotEmpty(t, plans[1].ID.String)
					assert.NotEmpty(t, plans[1].TopicID.String)
					assert.Equal(t, int16(0), plans[1].Week.Int)
					assert.NotEmpty(t, plans[1].LessonID.String)
					assert.Equal(t, time.Date(2020, 2, 3, 4, 5, 6, 7, time.UTC), plans[1].StartDate.Time.UTC())
					assert.Equal(t, time.Date(2021, 2, 3, 4, 5, 6, 7, time.UTC), plans[1].EndDate.Time.UTC())

					assert.NotEmpty(t, plans[2].ID.String)
					assert.NotEmpty(t, plans[2].TopicID.String)
					assert.Equal(t, int16(0), plans[2].Week.Int)
					assert.NotEmpty(t, plans[2].LessonID.String)
					assert.Equal(t, time.Date(2020, 2, 3, 4, 5, 6, 7, time.UTC), plans[1].StartDate.Time.UTC())
					assert.Equal(t, time.Date(2021, 2, 3, 4, 5, 6, 7, time.UTC), plans[1].EndDate.Time.UTC())

					actualPresetStudyPlanID := []string{plans[0].PresetStudyPlanID.String, plans[1].PresetStudyPlanID.String, plans[2].PresetStudyPlanID.String}
					expectedPresetStudyPlanID := []string{"preset-study-plan-id-1", "preset-study-plan-id-2", "preset-study-plan-id-3"}
					assert.ElementsMatch(t, expectedPresetStudyPlanID, actualPresetStudyPlanID)

					return nil
				},
			},
			userRp: UserRepoMock{
				get: func(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*entities.User, error) {
					assert.Equal(t, "school-admin-1", id.String)
					return &entities.User{
						ID:      database.Text("school-admin-1"),
						Country: database.Text("vietnam"),
					}, nil
				},
			},
			schoolAdminRp: SchoolAdminRepoMock{
				get: func(ctx context.Context, db database.QueryExecer, schoolAdminID pgtype.Text) (*entities.SchoolAdmin, error) {
					assert.Equal(t, "school-admin-1", schoolAdminID.String)
					return &entities.SchoolAdmin{
						SchoolAdminID: database.Text("school-admin-1"),
						SchoolID:      database.Int4(1),
					}, nil
				},
			},
			teacherRp: TeacherRepoMock{
				retrieve: func(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray, fields ...string) ([]entities.Teacher, error) {
					assert.ElementsMatch(t, []string{"teacher-id-1", "teacher-id-2"}, database.FromTextArray(ids))
					return []entities.Teacher{
						{
							ID:        database.Text("teacher-id-1"),
							SchoolIDs: database.Int4Array([]int32{3, 1}),
							User: entities.User{
								Country: database.Text("vietnam"),
							},
						},
						{
							ID:        database.Text("teacher-id-2"),
							SchoolIDs: database.Int4Array([]int32{1}),
							User: entities.User{
								Country: database.Text("vietnam"),
							},
						},
					}, nil
				},
			},
			studentRp: StudentRepoMock{
				retrieve: func(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]repositories.StudentProfile, error) {
					assert.ElementsMatch(t, []string{"learner-id-2", "learner-id-3", "learner-id-4", "learner-id-5", "learner-id-6"}, database.FromTextArray(ids))
					return []repositories.StudentProfile{
						{
							Student: entities.Student{
								ID:       database.Text("learner-id-2"),
								SchoolID: database.Int4(1),
								User: entities.User{
									ID:      database.Text("learner-id-2"),
									Country: database.Text("vietnam"),
								},
							},
							School: entities.School{
								ID: database.Int4(1),
							},
						},
						{
							Student: entities.Student{
								ID:       database.Text("learner-id-3"),
								SchoolID: database.Int4(1),
								User: entities.User{
									ID:      database.Text("learner-id-3"),
									Country: database.Text("vietnam"),
								},
							},
							School: entities.School{
								ID: database.Int4(1),
							},
						},
						{
							Student: entities.Student{
								ID:       database.Text("learner-id-4"),
								SchoolID: database.Int4(1),
								User: entities.User{
									ID:      database.Text("learner-id-4"),
									Country: database.Text("vietnam"),
								},
							},
							School: entities.School{
								ID: database.Int4(1),
							},
						},
						{
							Student: entities.Student{
								ID:       database.Text("learner-id-5"),
								SchoolID: database.Int4(1),
								User: entities.User{
									ID:      database.Text("learner-id-5"),
									Country: database.Text("vietnam"),
								},
							},
							School: entities.School{
								ID: database.Int4(1),
							},
						},
						{
							Student: entities.Student{
								ID:       database.Text("learner-id-6"),
								SchoolID: database.Int4(1),
								User: entities.User{
									ID:      database.Text("learner-id-6"),
									Country: database.Text("vietnam"),
								},
							},
							School: entities.School{
								ID: database.Int4(1),
							},
						},
					}, nil
				},
			},
			setup: func(ctx context.Context) {
				ctx = interceptors.ContextWithUserID(ctx, "school-admin-1")
				jsm.On("PublishAsyncContext", ctx, cconstants.SubjectLessonCreated, mock.Anything).Run(func(args mock.Arguments) {
					data := args.Get(2).([]byte)
					msg := &pb.EvtLesson{}
					err := msg.Unmarshal(data)
					require.NoError(t, err)

					message := msg.Message.(*pb.EvtLesson_CreateLessons_)
					assert.NotEmpty(t, message.CreateLessons.Lessons[0].LessonId)
					assert.Equal(t, "Lesson name 1", message.CreateLessons.Lessons[0].Name)
				}).Return("", nil).Once()
			},
		},
		{
			name:   "create a live lesson with missing name",
			caller: "school-admin-1",
			req: &bpb.CreateLiveLessonRequest{
				StartTime:  timestamppb.New(time.Date(2020, 2, 3, 4, 5, 6, 7, time.UTC)),
				EndTime:    timestamppb.New(time.Date(2021, 2, 3, 4, 5, 6, 7, time.UTC)),
				TeacherIds: []string{"teacher-id-1", "teacher-id-2"},
				CourseIds:  []string{"course-id-1", "course-id-4", "course-id-4", "course-id-5"},
				LearnerIds: []string{"learner-id-2", "learner-id-3", "learner-id-4", "learner-id-5", "learner-id-6"},
				Materials: []*bpb.Material{
					{
						Resource: &bpb.Material_BrightcoveVideo_{
							BrightcoveVideo: &bpb.Material_BrightcoveVideo{
								Name: "video 1",
								Url:  "https://brightcove.com/account/2/video?videoId=abc123",
							},
						},
					},
					{
						Resource: &bpb.Material_MediaId{
							MediaId: "media-id-2",
						},
					},
					{
						Resource: &bpb.Material_MediaId{
							MediaId: "media-id-2",
						},
					},
					{
						Resource: &bpb.Material_MediaId{
							MediaId: "media-id-3",
						},
					},
				},
			},
			userRp: UserRepoMock{
				get: func(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*entities.User, error) {
					assert.Equal(t, "school-admin-1", id.String)
					return &entities.User{
						ID:      database.Text("school-admin-1"),
						Country: database.Text("vietnam"),
					}, nil
				},
			},
			schoolAdminRp: SchoolAdminRepoMock{
				get: func(ctx context.Context, db database.QueryExecer, schoolAdminID pgtype.Text) (*entities.SchoolAdmin, error) {
					assert.Equal(t, "school-admin-1", schoolAdminID.String)
					return &entities.SchoolAdmin{
						SchoolAdminID: database.Text("school-admin-1"),
						SchoolID:      database.Int4(1),
					}, nil
				},
			},
			hasError: true,
		},
		{
			name:   "create a live lesson with missing start time",
			caller: "school-admin-1",
			req: &bpb.CreateLiveLessonRequest{
				Name:       "Lesson name 1",
				EndTime:    timestamppb.New(time.Date(2021, 2, 3, 4, 5, 6, 7, time.UTC)),
				TeacherIds: []string{"teacher-id-1", "teacher-id-2"},
				CourseIds:  []string{"course-id-1", "course-id-4", "course-id-4", "course-id-5"},
				LearnerIds: []string{"learner-id-2", "learner-id-3", "learner-id-4", "learner-id-5", "learner-id-6"},
				Materials: []*bpb.Material{
					{
						Resource: &bpb.Material_BrightcoveVideo_{
							BrightcoveVideo: &bpb.Material_BrightcoveVideo{
								Name: "video 1",
								Url:  "https://brightcove.com/account/2/video?videoId=abc123",
							},
						},
					},
					{
						Resource: &bpb.Material_MediaId{
							MediaId: "media-id-2",
						},
					},
					{
						Resource: &bpb.Material_MediaId{
							MediaId: "media-id-2",
						},
					},
					{
						Resource: &bpb.Material_MediaId{
							MediaId: "media-id-3",
						},
					},
				},
			},
			userRp: UserRepoMock{
				get: func(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*entities.User, error) {
					assert.Equal(t, "school-admin-1", id.String)
					return &entities.User{
						ID:      database.Text("school-admin-1"),
						Country: database.Text("vietnam"),
					}, nil
				},
			},
			schoolAdminRp: SchoolAdminRepoMock{
				get: func(ctx context.Context, db database.QueryExecer, schoolAdminID pgtype.Text) (*entities.SchoolAdmin, error) {
					assert.Equal(t, "school-admin-1", schoolAdminID.String)
					return &entities.SchoolAdmin{
						SchoolAdminID: database.Text("school-admin-1"),
						SchoolID:      database.Int4(1),
					}, nil
				},
			},
			hasError: true,
		},
		{
			name:   "create a live lesson with missing teacher ids",
			caller: "school-admin-1",
			req: &bpb.CreateLiveLessonRequest{
				Name:       "Lesson name 1",
				StartTime:  timestamppb.New(time.Date(2020, 2, 3, 4, 5, 6, 7, time.UTC)),
				EndTime:    timestamppb.New(time.Date(2021, 2, 3, 4, 5, 6, 7, time.UTC)),
				TeacherIds: []string{},
				CourseIds:  []string{"course-id-1", "course-id-4", "course-id-4", "course-id-5"},
				LearnerIds: []string{"learner-id-2", "learner-id-3", "learner-id-4", "learner-id-5", "learner-id-6"},
				Materials: []*bpb.Material{
					{
						Resource: &bpb.Material_BrightcoveVideo_{
							BrightcoveVideo: &bpb.Material_BrightcoveVideo{
								Name: "video 1",
								Url:  "https://brightcove.com/account/2/video?videoId=abc123",
							},
						},
					},
					{
						Resource: &bpb.Material_MediaId{
							MediaId: "media-id-2",
						},
					},
					{
						Resource: &bpb.Material_MediaId{
							MediaId: "media-id-2",
						},
					},
					{
						Resource: &bpb.Material_MediaId{
							MediaId: "media-id-3",
						},
					},
				},
			},
			userRp: UserRepoMock{
				get: func(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*entities.User, error) {
					assert.Equal(t, "school-admin-1", id.String)
					return &entities.User{
						ID:      database.Text("school-admin-1"),
						Country: database.Text("vietnam"),
					}, nil
				},
			},
			schoolAdminRp: SchoolAdminRepoMock{
				get: func(ctx context.Context, db database.QueryExecer, schoolAdminID pgtype.Text) (*entities.SchoolAdmin, error) {
					assert.Equal(t, "school-admin-1", schoolAdminID.String)
					return &entities.SchoolAdmin{
						SchoolAdminID: database.Text("school-admin-1"),
						SchoolID:      database.Int4(1),
					}, nil
				},
			},
			hasError: true,
		},
		{
			name:   "create a live lesson with missing course ids",
			caller: "school-admin-1",
			req: &bpb.CreateLiveLessonRequest{
				Name:       "Lesson name 1",
				StartTime:  timestamppb.New(time.Date(2020, 2, 3, 4, 5, 6, 7, time.UTC)),
				EndTime:    timestamppb.New(time.Date(2021, 2, 3, 4, 5, 6, 7, time.UTC)),
				TeacherIds: []string{"teacher-id-1", "teacher-id-2"},
				CourseIds:  []string{},
				LearnerIds: []string{"learner-id-2", "learner-id-3", "learner-id-4", "learner-id-5", "learner-id-6"},
				Materials: []*bpb.Material{
					{
						Resource: &bpb.Material_BrightcoveVideo_{
							BrightcoveVideo: &bpb.Material_BrightcoveVideo{
								Name: "video 1",
								Url:  "https://brightcove.com/account/2/video?videoId=abc123",
							},
						},
					},
					{
						Resource: &bpb.Material_MediaId{
							MediaId: "media-id-2",
						},
					},
					{
						Resource: &bpb.Material_MediaId{
							MediaId: "media-id-2",
						},
					},
					{
						Resource: &bpb.Material_MediaId{
							MediaId: "media-id-3",
						},
					},
				},
			},
			userRp: UserRepoMock{
				get: func(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*entities.User, error) {
					assert.Equal(t, "school-admin-1", id.String)
					return &entities.User{
						ID:      database.Text("school-admin-1"),
						Country: database.Text("vietnam"),
					}, nil
				},
			},
			schoolAdminRp: SchoolAdminRepoMock{
				get: func(ctx context.Context, db database.QueryExecer, schoolAdminID pgtype.Text) (*entities.SchoolAdmin, error) {
					assert.Equal(t, "school-admin-1", schoolAdminID.String)
					return &entities.SchoolAdmin{
						SchoolAdminID: database.Text("school-admin-1"),
						SchoolID:      database.Int4(1),
					}, nil
				},
			},
			hasError: true,
		},
		{
			name:   "create a live lesson with missing learner ids",
			caller: "school-admin-1",
			req: &bpb.CreateLiveLessonRequest{
				Name:       "Lesson name 1",
				StartTime:  timestamppb.New(time.Date(2020, 2, 3, 4, 5, 6, 7, time.UTC)),
				EndTime:    timestamppb.New(time.Date(2021, 2, 3, 4, 5, 6, 7, time.UTC)),
				TeacherIds: []string{"teacher-id-1", "teacher-id-2"},
				CourseIds:  []string{"course-id-1", "course-id-4", "course-id-4", "course-id-5"},
				Materials: []*bpb.Material{
					{
						Resource: &bpb.Material_BrightcoveVideo_{
							BrightcoveVideo: &bpb.Material_BrightcoveVideo{
								Name: "video 1",
								Url:  "https://brightcove.com/account/2/video?videoId=abc123",
							},
						},
					},
					{
						Resource: &bpb.Material_MediaId{
							MediaId: "media-id-2",
						},
					},
					{
						Resource: &bpb.Material_MediaId{
							MediaId: "media-id-2",
						},
					},
					{
						Resource: &bpb.Material_MediaId{
							MediaId: "media-id-3",
						},
					},
				},
			},
			userRp: UserRepoMock{
				get: func(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*entities.User, error) {
					assert.Equal(t, "school-admin-1", id.String)
					return &entities.User{
						ID:      database.Text("school-admin-1"),
						Country: database.Text("vietnam"),
					}, nil
				},
			},
			schoolAdminRp: SchoolAdminRepoMock{
				get: func(ctx context.Context, db database.QueryExecer, schoolAdminID pgtype.Text) (*entities.SchoolAdmin, error) {
					assert.Equal(t, "school-admin-1", schoolAdminID.String)
					return &entities.SchoolAdmin{
						SchoolAdminID: database.Text("school-admin-1"),
						SchoolID:      database.Int4(1),
					}, nil
				},
			},
			hasError: true,
		},
		{
			name: "create a live lesson with invalid brightcove video url",
			req: &bpb.CreateLiveLessonRequest{
				Name:       "Lesson name 1",
				StartTime:  timestamppb.New(time.Date(2020, 2, 3, 4, 5, 6, 7, time.UTC)),
				EndTime:    timestamppb.New(time.Date(2021, 2, 3, 4, 5, 6, 7, time.UTC)),
				TeacherIds: []string{"teacher-id-1", "teacher-id-2"},
				CourseIds:  []string{"course-id-1", "course-id-4", "course-id-4", "course-id-5"},
				LearnerIds: []string{"learner-id-2", "learner-id-3", "learner-id-4", "learner-id-5", "learner-id-6"},
				Materials: []*bpb.Material{
					{
						Resource: &bpb.Material_BrightcoveVideo_{
							BrightcoveVideo: &bpb.Material_BrightcoveVideo{
								Name: "video 1",
								Url:  "https://brightcove.com/account/2/video?videoID=abc123",
							},
						},
					},
					{
						Resource: &bpb.Material_MediaId{
							MediaId: "media-id-2",
						},
					},
					{
						Resource: &bpb.Material_MediaId{
							MediaId: "media-id-2",
						},
					},
					{
						Resource: &bpb.Material_MediaId{
							MediaId: "media-id-3",
						},
					},
				},
			},
			hasError: true,
		},
		{
			name:   "create a live lesson with not exist media id",
			caller: "school-admin-1",
			req: &bpb.CreateLiveLessonRequest{
				Name:       "Lesson name 1",
				StartTime:  timestamppb.New(time.Date(2020, 2, 3, 4, 5, 6, 7, time.UTC)),
				EndTime:    timestamppb.New(time.Date(2021, 2, 3, 4, 5, 6, 7, time.UTC)),
				TeacherIds: []string{"teacher-id-1", "teacher-id-2"},
				CourseIds:  []string{"course-id-1", "course-id-4", "course-id-4", "course-id-5"},
				LearnerIds: []string{"learner-id-2", "learner-id-3", "learner-id-4", "learner-id-5", "learner-id-6"},
				Materials: []*bpb.Material{
					{
						Resource: &bpb.Material_BrightcoveVideo_{
							BrightcoveVideo: &bpb.Material_BrightcoveVideo{
								Name: "video 1",
								Url:  "https://brightcove.com/account/2/video?videoId=abc123",
							},
						},
					},
					{
						Resource: &bpb.Material_MediaId{
							MediaId: "media-id-2",
						},
					},
					{
						Resource: &bpb.Material_MediaId{
							MediaId: "media-id-3",
						},
					},
					{
						Resource: &bpb.Material_MediaId{
							MediaId: "media-id-non-exist",
						},
					},
				},
			},
			mediaRp: mediaRepo.MediaRepoMock{
				RetrieveByIDsMock: func(ctx context.Context, db database.QueryExecer, mediaIDs pgtype.TextArray) ([]*entities.Media, error) {
					assert.True(t, isTx(db), "retrieve media by IDs is not processing in transaction")
					actualMediaIDs := database.FromTextArray(mediaIDs)
					assert.Contains(t, actualMediaIDs, "media-id-2")
					assert.Contains(t, actualMediaIDs, "media-id-3")
					return []*entities.Media{
						{
							MediaID: database.Text("media-id-2"),
						},
						{
							MediaID: database.Text("media-id-3"),
						},
					}, nil
				},
			},
			courseRepo: coursesRepo.CourseRepoMock{
				FindByIDsMock: func(ctx context.Context, db database.Ext, courseIDs pgtype.TextArray) (map[pgtype.Text]*entities.Course, error) {
					assert.ElementsMatch(t, []string{"course-id-1", "course-id-4", "course-id-5"}, database.FromTextArray(courseIDs))
					return map[pgtype.Text]*entities.Course{
						database.Text("course-id-1"): {
							ID:        database.Text("course-id-1"),
							Name:      database.Text("math level 1"),
							Country:   database.Text("vietnam"),
							Grade:     database.Int2(2),
							Subject:   database.Text("math"),
							SchoolID:  database.Int4(1),
							StartDate: database.Timestamptz(time.Date(2021, 2, 3, 4, 5, 6, 7, time.UTC)),
						},
						database.Text("course-id-4"): {
							ID:        database.Text("course-id-4"),
							Name:      database.Text("physics level 2"),
							Country:   database.Text("vietnam"),
							Grade:     database.Int2(5),
							Subject:   database.Text("physics"),
							SchoolID:  database.Int4(1),
							StartDate: database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.UTC)),
						},
						database.Text("course-id-5"): {
							ID:        database.Text("course-id-5"),
							Name:      database.Text("chemistry level 1"),
							Country:   database.Text("vietnam"),
							Grade:     database.Int2(7),
							Subject:   database.Text("chemistry"),
							SchoolID:  database.Int4(1),
							StartDate: database.Timestamptz(time.Date(2023, 2, 3, 4, 5, 6, 7, time.UTC)),
						},
					}, nil
				},
			},
			userRp: UserRepoMock{
				get: func(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*entities.User, error) {
					assert.Equal(t, "school-admin-1", id.String)
					return &entities.User{
						ID:      database.Text("school-admin-1"),
						Country: database.Text("vietnam"),
					}, nil
				},
			},
			schoolAdminRp: SchoolAdminRepoMock{
				get: func(ctx context.Context, db database.QueryExecer, schoolAdminID pgtype.Text) (*entities.SchoolAdmin, error) {
					assert.Equal(t, "school-admin-1", schoolAdminID.String)
					return &entities.SchoolAdmin{
						SchoolAdminID: database.Text("school-admin-1"),
						SchoolID:      database.Int4(1),
					}, nil
				},
			},
			teacherRp: TeacherRepoMock{
				retrieve: func(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray, fields ...string) ([]entities.Teacher, error) {
					assert.ElementsMatch(t, []string{"teacher-id-1", "teacher-id-2"}, database.FromTextArray(ids))
					return []entities.Teacher{
						{
							ID:        database.Text("teacher-id-1"),
							SchoolIDs: database.Int4Array([]int32{3, 1}),
							User: entities.User{
								Country: database.Text("vietnam"),
							},
						},
						{
							ID:        database.Text("teacher-id-2"),
							SchoolIDs: database.Int4Array([]int32{1}),
							User: entities.User{
								Country: database.Text("vietnam"),
							},
						},
					}, nil
				},
			},
			studentRp: StudentRepoMock{
				retrieve: func(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]repositories.StudentProfile, error) {
					assert.ElementsMatch(t, []string{"learner-id-2", "learner-id-3", "learner-id-4", "learner-id-5", "learner-id-6"}, database.FromTextArray(ids))
					return []repositories.StudentProfile{
						{
							Student: entities.Student{
								ID:       database.Text("learner-id-2"),
								SchoolID: database.Int4(1),
								User: entities.User{
									ID:      database.Text("learner-id-2"),
									Country: database.Text("vietnam"),
								},
							},
							School: entities.School{
								ID: database.Int4(1),
							},
						},
						{
							Student: entities.Student{
								ID:       database.Text("learner-id-3"),
								SchoolID: database.Int4(1),
								User: entities.User{
									ID:      database.Text("learner-id-3"),
									Country: database.Text("vietnam"),
								},
							},
							School: entities.School{
								ID: database.Int4(1),
							},
						},
						{
							Student: entities.Student{
								ID:       database.Text("learner-id-4"),
								SchoolID: database.Int4(1),
								User: entities.User{
									ID:      database.Text("learner-id-4"),
									Country: database.Text("vietnam"),
								},
							},
							School: entities.School{
								ID: database.Int4(1),
							},
						},
						{
							Student: entities.Student{
								ID:       database.Text("learner-id-5"),
								SchoolID: database.Int4(1),
								User: entities.User{
									ID:      database.Text("learner-id-5"),
									Country: database.Text("vietnam"),
								},
							},
							School: entities.School{
								ID: database.Int4(1),
							},
						},
						{
							Student: entities.Student{
								ID:       database.Text("learner-id-6"),
								SchoolID: database.Int4(1),
								User: entities.User{
									ID:      database.Text("learner-id-6"),
									Country: database.Text("vietnam"),
								},
							},
							School: entities.School{
								ID: database.Int4(1),
							},
						},
					}, nil
				},
			},
			hasError: true,
		},
		{
			name:   "create a live lesson with admin successfully",
			caller: "admin-1",
			req: &bpb.CreateLiveLessonRequest{
				Name:       "Lesson name 1",
				StartTime:  timestamppb.New(time.Date(2020, 2, 3, 4, 5, 6, 7, time.UTC)),
				EndTime:    timestamppb.New(time.Date(2021, 2, 3, 4, 5, 6, 7, time.UTC)),
				TeacherIds: []string{"teacher-id-1", "teacher-id-2"},
				CourseIds:  []string{"course-id-1", "course-id-4", "course-id-4", "course-id-5"},
				LearnerIds: []string{"learner-id-2", "learner-id-3", "learner-id-4", "learner-id-5", "learner-id-6"},
				Materials: []*bpb.Material{
					{
						Resource: &bpb.Material_BrightcoveVideo_{
							BrightcoveVideo: &bpb.Material_BrightcoveVideo{
								Name: "video 1",
								Url:  "https://brightcove.com/account/2/video?videoId=abc123",
							},
						},
					},
					{
						Resource: &bpb.Material_MediaId{
							MediaId: "media-id-2",
						},
					},
					{
						Resource: &bpb.Material_MediaId{
							MediaId: "media-id-2",
						},
					},
					{
						Resource: &bpb.Material_MediaId{
							MediaId: "media-id-3",
						},
					},
				},
			},
			mediaRp: mediaRepo.MediaRepoMock{
				RetrieveByIDsMock: func(ctx context.Context, db database.QueryExecer, mediaIDs pgtype.TextArray) ([]*entities.Media, error) {
					assert.True(t, isTx(db), "retrieve media by IDs is not processing in transaction")
					actualMediaIDs := database.FromTextArray(mediaIDs)
					assert.ElementsMatch(t, actualMediaIDs, []string{"media-id-2", "media-id-3"})
					return []*entities.Media{
						{
							MediaID: database.Text("media-id-2"),
						},
						{
							MediaID: database.Text("media-id-3"),
						},
					}, nil
				},
				UpsertMediaBatchMock: func(ctx context.Context, db database.QueryExecer, medias entities.Medias) error {
					assert.Len(t, medias, 1)
					assert.EqualValues(t, entities.MediaTypeVideo, medias[0].Type.String)
					assert.Equal(t, "video 1", medias[0].Name.String)
					assert.Equal(t, "abc123", medias[0].Resource.String)
					err := medias.PreInsert()
					require.NoError(t, err)

					return nil
				},
			},
			lessonGroupRp: LessonGroupRepoMock{
				create: func(ctx context.Context, db database.QueryExecer, e *entities.LessonGroup) error {
					assert.True(t, isTx(db), "create lesson group is not processing in transaction")
					assert.Equal(t, "course-id-1", e.CourseID.String)

					// validate mediaIDs
					actualMedias := database.FromTextArray(e.MediaIDs)
					mediaIDsMap := make(map[string]bool)
					for _, mediaID := range actualMedias {
						mediaIDsMap[mediaID] = true
					}
					assert.Len(t, mediaIDsMap, 3)
					err := e.PreInsert()
					require.NoError(t, err)

					return nil
				},
			},
			lessonRepo: LessonRepoMock{
				create: func(ctx context.Context, db database.Ext, lesson *entities.Lesson) (*entities.Lesson, error) {
					assert.True(t, isTx(db), "create lesson group is not processing in transaction")

					assert.NotEmpty(t, lesson.LessonID.String)
					assert.Equal(t, "Lesson name 1", lesson.Name.String)
					assert.Equal(t, "teacher-id-1", lesson.TeacherID.String)
					assert.Equal(t, "course-id-1", lesson.CourseID.String)
					assert.Equal(t, time.Date(2020, 2, 3, 4, 5, 6, 7, time.UTC).Unix(), lesson.StartTime.Time.Unix())
					assert.Equal(t, time.Date(2021, 2, 3, 4, 5, 6, 7, time.UTC).Unix(), lesson.EndTime.Time.Unix())
					assert.NotEmpty(t, lesson.LessonGroupID.String)
					assert.EqualValues(t, entities.LessonTypeOnline, lesson.LessonType.String)
					assert.EqualValues(t, entities.LessonTeachingMediumOnline, lesson.TeachingMedium.String)
					assert.EqualValues(t, entities.LessonStatusDraft, lesson.Status.String)
					assert.EqualValues(t, 0, lesson.StreamLearnerCounter.Int)

					return lesson, nil
				},
				upsertLessonCourses: func(ctx context.Context, db database.Ext, lessonID pgtype.Text, courseIDs pgtype.TextArray) error {
					assert.True(t, isTx(db), "upsert lesson courses is not processing in transaction")
					assert.NotEmpty(t, lessonID.String)
					assert.Equal(t, []string{"course-id-1", "course-id-4", "course-id-5"}, database.FromTextArray(courseIDs))
					return nil
				},
				upsertLessonMembers: func(ctx context.Context, db database.Ext, lessonID pgtype.Text, userIDs pgtype.TextArray) error {
					assert.True(t, isTx(db), "upsert lesson members is not processing in transaction")
					assert.Equal(t, []string{"learner-id-2", "learner-id-3", "learner-id-4", "learner-id-5", "learner-id-6"}, database.FromTextArray(userIDs))
					return nil
				},
				upsertLessonTeachers: func(ctx context.Context, db database.Ext, lessonID pgtype.Text, teacherIDs pgtype.TextArray) error {
					assert.True(t, isTx(db), "upsert lesson teachers is not processing in transaction")
					assert.Equal(t, []string{"teacher-id-1", "teacher-id-2"}, database.FromTextArray(teacherIDs))
					return nil
				},
				findEarliestAndLatestTimeLessonByCourses: func(ctx context.Context, db database.Ext, courseIDs pgtype.TextArray) (*entities.CourseAvailableRanges, error) {
					assert.ElementsMatch(t, []string{"course-id-1", "course-id-4", "course-id-5"}, database.FromTextArray(courseIDs))

					res := &entities.CourseAvailableRanges{}
					res.Add(
						&entities.CourseAvailableRange{
							ID:        database.Text("course-id-1"),
							StartDate: database.Timestamptz(time.Date(2019, 2, 3, 4, 5, 6, 7, time.UTC)),
							EndDate:   database.Timestamptz(time.Date(2020, 2, 3, 4, 5, 6, 7, time.UTC)),
						},
						&entities.CourseAvailableRange{
							ID:        database.Text("course-id-4"),
							StartDate: database.Timestamptz(time.Date(2020, 2, 3, 4, 5, 6, 7, time.UTC)),
							EndDate:   database.Timestamptz(time.Date(2021, 2, 3, 4, 5, 6, 7, time.UTC)),
						},
						&entities.CourseAvailableRange{
							ID:        database.Text("course-id-5"),
							StartDate: database.Timestamptz(time.Date(2018, 2, 3, 4, 5, 6, 7, time.UTC)),
							EndDate:   database.Timestamptz(time.Date(2021, 2, 3, 4, 5, 6, 7, time.UTC)),
						},
					)
					return res, nil
				},
			},
			courseRepo: coursesRepo.CourseRepoMock{
				FindByIDsMock: func(ctx context.Context, db database.Ext, courseIDs pgtype.TextArray) (map[pgtype.Text]*entities.Course, error) {
					presetStudyPlanIDs := make(map[string]pgtype.Text)
					if _, ok := isUpsertedCourseByTestCase["create a live lesson successfully"]; ok {
						presetStudyPlanIDs["course-id-1"] = database.Text("preset-study-plan-id-1")
						presetStudyPlanIDs["course-id-4"] = database.Text("preset-study-plan-id-2")
						presetStudyPlanIDs["course-id-5"] = database.Text("preset-study-plan-id-3")
					}

					assert.ElementsMatch(t, []string{"course-id-1", "course-id-4", "course-id-5"}, database.FromTextArray(courseIDs))
					return map[pgtype.Text]*entities.Course{
						database.Text("course-id-1"): {
							ID:                database.Text("course-id-1"),
							Name:              database.Text("math level 1"),
							PresetStudyPlanID: presetStudyPlanIDs["course-id-1"],
							Country:           database.Text("vietnam"),
							Grade:             database.Int2(2),
							Subject:           database.Text("math"),
							SchoolID:          database.Int4(1),
							StartDate:         database.Timestamptz(time.Date(2021, 2, 3, 4, 5, 6, 7, time.UTC)),
						},
						database.Text("course-id-4"): {
							ID:                database.Text("course-id-4"),
							Name:              database.Text("physics level 2"),
							PresetStudyPlanID: presetStudyPlanIDs["course-id-4"],
							Country:           database.Text("vietnam"),
							Grade:             database.Int2(5),
							Subject:           database.Text("physics"),
							SchoolID:          database.Int4(1),
							StartDate:         database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.UTC)),
						},
						database.Text("course-id-5"): {
							ID:                database.Text("course-id-5"),
							Name:              database.Text("chemistry level 1"),
							PresetStudyPlanID: presetStudyPlanIDs["course-id-5"],
							Country:           database.Text("vietnam"),
							Grade:             database.Int2(7),
							Subject:           database.Text("chemistry"),
							SchoolID:          database.Int4(1),
							StartDate:         database.Timestamptz(time.Date(2023, 2, 3, 4, 5, 6, 7, time.UTC)),
						},
					}, nil
				},
				UpsertMock: func(ctx context.Context, db database.Ext, cc []*entities.Course) error {
					assert.Len(t, cc, 3)
					if isUpsertedCourseByTestCase["create a live lesson successfully"] {
						expected := map[string]*entities.CourseAvailableRange{
							"course-id-1": {
								ID:        database.Text("course-id-1"),
								StartDate: database.Timestamptz(time.Date(2019, 2, 3, 4, 5, 6, 7, time.UTC)),
								EndDate:   database.Timestamptz(time.Date(2020, 2, 3, 4, 5, 6, 7, time.UTC)),
							},
							"course-id-4": {
								ID:        database.Text("course-id-4"),
								StartDate: database.Timestamptz(time.Date(2020, 2, 3, 4, 5, 6, 7, time.UTC)),
								EndDate:   database.Timestamptz(time.Date(2021, 2, 3, 4, 5, 6, 7, time.UTC)),
							},
							"course-id-5": {
								ID:        database.Text("course-id-5"),
								StartDate: database.Timestamptz(time.Date(2018, 2, 3, 4, 5, 6, 7, time.UTC)),
								EndDate:   database.Timestamptz(time.Date(2021, 2, 3, 4, 5, 6, 7, time.UTC)),
							},
						}

						// test start data and end data
						for _, c := range cc {
							if v, ok := expected[c.ID.String]; ok {
								assert.Equal(t, v.StartDate, c.StartDate)
								assert.Equal(t, v.EndDate, c.EndDate)
							}
						}
					}

					courseIDs := make([]string, 0, len(cc))
					for _, c := range cc {
						assert.NotEmpty(t, c.ID.String)
						assert.NotEmpty(t, c.PresetStudyPlanID.String)
						courseIDs = append(courseIDs, c.ID.String)
					}
					assert.ElementsMatch(t, []string{"course-id-1", "course-id-4", "course-id-5"}, courseIDs)
					isUpsertedCourseByTestCase["create a live lesson successfully"] = true

					return nil
				},
			},
			topicRepo: topicsRepo.TopicRepoMock{
				CreateMock: func(ctx context.Context, db database.Ext, plans []*entities.Topic) error {
					assert.Len(t, plans, 3)
					assert.NotEmpty(t, plans[0].ID.String)
					assert.NotZero(t, plans[0].PublishedAt.Time)
					assert.NotEmpty(t, plans[1].ID.String)
					assert.NotZero(t, plans[1].PublishedAt.Time)
					assert.NotEmpty(t, plans[2].ID.String)
					assert.NotZero(t, plans[2].PublishedAt.Time)

					plansBySubject := make(map[string]*entities.Topic)
					for i := range plans {
						plansBySubject[plans[i].Subject.String] = plans[i]
					}

					e := &entities.Topic{}
					database.AllNullEntity(e)
					err := multierr.Combine(
						e.ID.Set(plansBySubject["math"].ID),
						e.Name.Set("Lesson name 1"),
						e.Country.Set("vietnam"),
						e.Grade.Set(2),
						e.Subject.Set("math"),
						e.SchoolID.Set(1),
						e.TopicType.Set(entities.TopicTypeLiveLesson),
						e.Status.Set(entities.TopicStatusPublished),
						e.DisplayOrder.Set(1),
						e.PublishedAt.Set(plansBySubject["math"].PublishedAt),
						e.TotalLOs.Set(0),
						e.ChapterID.Set(nil),
						e.IconURL.Set(nil),
						e.DeletedAt.Set(nil),
						e.EssayRequired.Set(false),
					)
					require.NoError(t, err)
					assert.Equal(t, e, plansBySubject["math"])

					database.AllNullEntity(e)
					err = multierr.Combine(
						e.ID.Set(plansBySubject["physics"].ID),
						e.Name.Set("Lesson name 1"),
						e.Country.Set("vietnam"),
						e.Grade.Set(5),
						e.Subject.Set("physics"),
						e.SchoolID.Set(1),
						e.TopicType.Set(entities.TopicTypeLiveLesson),
						e.Status.Set(entities.TopicStatusPublished),
						e.DisplayOrder.Set(1),
						e.PublishedAt.Set(plansBySubject["physics"].PublishedAt),
						e.TotalLOs.Set(0),
						e.ChapterID.Set(nil),
						e.IconURL.Set(nil),
						e.DeletedAt.Set(nil),
						e.EssayRequired.Set(false),
					)
					require.NoError(t, err)
					assert.Equal(t, e, plansBySubject["physics"])

					database.AllNullEntity(e)
					err = multierr.Combine(
						e.ID.Set(plansBySubject["chemistry"].ID),
						e.Name.Set("Lesson name 1"),
						e.Country.Set("vietnam"),
						e.Grade.Set(7),
						e.Subject.Set("chemistry"),
						e.SchoolID.Set(1),
						e.TopicType.Set(entities.TopicTypeLiveLesson),
						e.Status.Set(entities.TopicStatusPublished),
						e.DisplayOrder.Set(1),
						e.PublishedAt.Set(plansBySubject["chemistry"].PublishedAt),
						e.TotalLOs.Set(0),
						e.ChapterID.Set(nil),
						e.IconURL.Set(nil),
						e.DeletedAt.Set(nil),
						e.EssayRequired.Set(false),
					)
					require.NoError(t, err)
					assert.Equal(t, e, plansBySubject["chemistry"])

					return nil
				},
			},
			pSPRepo: coursesRepo.PresetStudyPlanRepoMock{
				CreatePresetStudyPlanMock: func(ctx context.Context, db database.Ext, presetStudyPlans []*entities.PresetStudyPlan) error {
					assert.Len(t, presetStudyPlans, 3)
					assert.NotEmpty(t, presetStudyPlans[0].ID.String)
					assert.NotEmpty(t, presetStudyPlans[1].ID.String)

					pspBySubject := make(map[string]*entities.PresetStudyPlan)
					for i := range presetStudyPlans {
						pspBySubject[presetStudyPlans[i].Subject.String] = presetStudyPlans[i]
					}

					expected := &entities.PresetStudyPlan{}
					database.AllNullEntity(expected)
					err := multierr.Combine(
						expected.ID.Set(pspBySubject["math"].ID),
						expected.Name.Set("math level 1"),
						expected.Country.Set("vietnam"),
						expected.Grade.Set(2),
						expected.Subject.Set("math"),
						expected.StartDate.Set(time.Date(2021, 2, 3, 4, 5, 6, 7, time.UTC)),
					)
					require.NoError(t, err)
					assert.Equal(t, expected, pspBySubject["math"])

					database.AllNullEntity(expected)
					err = multierr.Combine(
						expected.ID.Set(pspBySubject["physics"].ID),
						expected.Name.Set("physics level 2"),
						expected.Country.Set("vietnam"),
						expected.Grade.Set(5),
						expected.Subject.Set("physics"),
						expected.StartDate.Set(time.Date(2022, 2, 3, 4, 5, 6, 7, time.UTC)),
					)
					require.NoError(t, err)
					assert.Equal(t, expected, pspBySubject["physics"])

					database.AllNullEntity(expected)
					err = multierr.Combine(
						expected.ID.Set(pspBySubject["chemistry"].ID),
						expected.Name.Set("chemistry level 1"),
						expected.Country.Set("vietnam"),
						expected.Grade.Set(7),
						expected.Subject.Set("chemistry"),
						expected.StartDate.Set(time.Date(2023, 2, 3, 4, 5, 6, 7, time.UTC)),
					)
					require.NoError(t, err)
					assert.Equal(t, expected, pspBySubject["chemistry"])

					return nil
				},
			},
			pSPWRepo: coursesRepo.PresetStudyPlanWeeklyRepoMock{
				CreateMock: func(ctx context.Context, db database.Ext, plans []*entities.PresetStudyPlanWeekly) error {
					assert.Len(t, plans, 3)

					assert.NotEmpty(t, plans[0].ID.String)
					assert.NotEmpty(t, plans[0].TopicID.String)
					assert.Equal(t, int16(0), plans[0].Week.Int)
					assert.NotEmpty(t, plans[0].LessonID.String)
					assert.Equal(t, time.Date(2020, 2, 3, 4, 5, 6, 7, time.UTC), plans[0].StartDate.Time.UTC())
					assert.Equal(t, time.Date(2021, 2, 3, 4, 5, 6, 7, time.UTC), plans[0].EndDate.Time.UTC())

					assert.NotEmpty(t, plans[1].ID.String)
					assert.NotEmpty(t, plans[1].TopicID.String)
					assert.Equal(t, int16(0), plans[1].Week.Int)
					assert.NotEmpty(t, plans[1].LessonID.String)
					assert.Equal(t, time.Date(2020, 2, 3, 4, 5, 6, 7, time.UTC), plans[1].StartDate.Time.UTC())
					assert.Equal(t, time.Date(2021, 2, 3, 4, 5, 6, 7, time.UTC), plans[1].EndDate.Time.UTC())

					assert.NotEmpty(t, plans[2].ID.String)
					assert.NotEmpty(t, plans[2].TopicID.String)
					assert.Equal(t, int16(0), plans[2].Week.Int)
					assert.NotEmpty(t, plans[2].LessonID.String)
					assert.Equal(t, time.Date(2020, 2, 3, 4, 5, 6, 7, time.UTC), plans[1].StartDate.Time.UTC())
					assert.Equal(t, time.Date(2021, 2, 3, 4, 5, 6, 7, time.UTC), plans[1].EndDate.Time.UTC())

					actualPresetStudyPlanID := []string{plans[0].PresetStudyPlanID.String, plans[1].PresetStudyPlanID.String, plans[2].PresetStudyPlanID.String}
					expectedPresetStudyPlanID := []string{"preset-study-plan-id-1", "preset-study-plan-id-2", "preset-study-plan-id-3"}
					assert.ElementsMatch(t, expectedPresetStudyPlanID, actualPresetStudyPlanID)

					return nil
				},
			},
			userRp: UserRepoMock{
				get: func(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*entities.User, error) {
					assert.Equal(t, "admin-1", id.String)
					return &entities.User{
						ID:      database.Text("admin-1"),
						Country: database.Text("vietnam"),
					}, nil
				},
			},
			schoolAdminRp: SchoolAdminRepoMock{
				get: func(ctx context.Context, db database.QueryExecer, schoolAdminID pgtype.Text) (*entities.SchoolAdmin, error) {
					assert.Equal(t, "admin-1", schoolAdminID.String)
					return nil, pgx.ErrNoRows
				},
			},
			teacherRp: TeacherRepoMock{
				retrieve: func(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray, fields ...string) ([]entities.Teacher, error) {
					assert.ElementsMatch(t, []string{"teacher-id-1", "teacher-id-2"}, database.FromTextArray(ids))
					return []entities.Teacher{
						{
							ID:        database.Text("teacher-id-1"),
							SchoolIDs: database.Int4Array([]int32{1}),
							User: entities.User{
								Country: database.Text("vietnam"),
							},
						},
						{
							ID:        database.Text("teacher-id-2"),
							SchoolIDs: database.Int4Array([]int32{3, 1}),
							User: entities.User{
								Country: database.Text("vietnam"),
							},
						},
					}, nil
				},
			},
			studentRp: StudentRepoMock{
				retrieve: func(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]repositories.StudentProfile, error) {
					assert.ElementsMatch(t, []string{"learner-id-2", "learner-id-3", "learner-id-4", "learner-id-5", "learner-id-6"}, database.FromTextArray(ids))
					return []repositories.StudentProfile{
						{
							Student: entities.Student{
								ID:       database.Text("learner-id-2"),
								SchoolID: database.Int4(1),
								User: entities.User{
									ID:      database.Text("learner-id-2"),
									Country: database.Text("vietnam"),
								},
							},
							School: entities.School{
								ID: database.Int4(1),
							},
						},
						{
							Student: entities.Student{
								ID:       database.Text("learner-id-3"),
								SchoolID: database.Int4(1),
								User: entities.User{
									ID:      database.Text("learner-id-3"),
									Country: database.Text("vietnam"),
								},
							},
							School: entities.School{
								ID: database.Int4(1),
							},
						},
						{
							Student: entities.Student{
								ID:       database.Text("learner-id-4"),
								SchoolID: database.Int4(1),
								User: entities.User{
									ID:      database.Text("learner-id-4"),
									Country: database.Text("vietnam"),
								},
							},
							School: entities.School{
								ID: database.Int4(1),
							},
						},
						{
							Student: entities.Student{
								ID:       database.Text("learner-id-5"),
								SchoolID: database.Int4(1),
								User: entities.User{
									ID:      database.Text("learner-id-5"),
									Country: database.Text("vietnam"),
								},
							},
							School: entities.School{
								ID: database.Int4(1),
							},
						},
						{
							Student: entities.Student{
								ID:       database.Text("learner-id-6"),
								SchoolID: database.Int4(1),
								User: entities.User{
									ID:      database.Text("learner-id-6"),
									Country: database.Text("vietnam"),
								},
							},
							School: entities.School{
								ID: database.Int4(1),
							},
						},
					}, nil
				},
			},
			setup: func(ctx context.Context) {
				ctx = interceptors.ContextWithUserID(ctx, "admin-1")
				jsm.On("PublishAsyncContext", ctx, cconstants.SubjectLessonCreated, mock.Anything).Run(func(args mock.Arguments) {
					data := args.Get(2).([]byte)
					msg := &pb.EvtLesson{}
					err := msg.Unmarshal(data)
					require.NoError(t, err)

					message := msg.Message.(*pb.EvtLesson_CreateLessons_)
					assert.NotEmpty(t, message.CreateLessons.Lessons[0].LessonId)
					assert.Equal(t, "Lesson name 1", message.CreateLessons.Lessons[0].Name)
				}).Return("", nil).Once()
			},
		},
		{
			name:   "create a live lesson with teacher not same school id",
			caller: "school-admin-1",
			req: &bpb.CreateLiveLessonRequest{
				Name:       "Lesson name 1",
				StartTime:  timestamppb.New(time.Date(2020, 2, 3, 4, 5, 6, 7, time.UTC)),
				EndTime:    timestamppb.New(time.Date(2021, 2, 3, 4, 5, 6, 7, time.UTC)),
				TeacherIds: []string{"teacher-id-1", "teacher-id-2"},
				CourseIds:  []string{"course-id-1", "course-id-4", "course-id-4", "course-id-5"},
				LearnerIds: []string{"learner-id-2", "learner-id-3", "learner-id-4", "learner-id-5", "learner-id-6"},
				Materials: []*bpb.Material{
					{
						Resource: &bpb.Material_BrightcoveVideo_{
							BrightcoveVideo: &bpb.Material_BrightcoveVideo{
								Name: "video 1",
								Url:  "https://brightcove.com/account/2/video?videoId=abc123",
							},
						},
					},
					{
						Resource: &bpb.Material_MediaId{
							MediaId: "media-id-2",
						},
					},
					{
						Resource: &bpb.Material_MediaId{
							MediaId: "media-id-2",
						},
					},
					{
						Resource: &bpb.Material_MediaId{
							MediaId: "media-id-3",
						},
					},
				},
			},
			courseRepo: coursesRepo.CourseRepoMock{
				FindByIDsMock: func(ctx context.Context, db database.Ext, courseIDs pgtype.TextArray) (map[pgtype.Text]*entities.Course, error) {
					assert.ElementsMatch(t, []string{"course-id-1", "course-id-4", "course-id-5"}, database.FromTextArray(courseIDs))
					return map[pgtype.Text]*entities.Course{
						database.Text("course-id-1"): {
							ID:        database.Text("course-id-1"),
							Name:      database.Text("math level 1"),
							Country:   database.Text("vietnam"),
							Grade:     database.Int2(2),
							Subject:   database.Text("math"),
							SchoolID:  database.Int4(1),
							StartDate: database.Timestamptz(time.Date(2021, 2, 3, 4, 5, 6, 7, time.UTC)),
						},
						database.Text("course-id-4"): {
							ID:        database.Text("course-id-4"),
							Name:      database.Text("physics level 2"),
							Country:   database.Text("vietnam"),
							Grade:     database.Int2(5),
							Subject:   database.Text("physics"),
							SchoolID:  database.Int4(1),
							StartDate: database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.UTC)),
						},
						database.Text("course-id-5"): {
							ID:        database.Text("course-id-5"),
							Name:      database.Text("chemistry level 1"),
							Country:   database.Text("vietnam"),
							Grade:     database.Int2(7),
							Subject:   database.Text("chemistry"),
							SchoolID:  database.Int4(1),
							StartDate: database.Timestamptz(time.Date(2023, 2, 3, 4, 5, 6, 7, time.UTC)),
						},
					}, nil
				},
			},
			userRp: UserRepoMock{
				get: func(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*entities.User, error) {
					assert.Equal(t, "school-admin-1", id.String)
					return &entities.User{
						ID:      database.Text("school-admin-1"),
						Country: database.Text("vietnam"),
					}, nil
				},
			},
			schoolAdminRp: SchoolAdminRepoMock{
				get: func(ctx context.Context, db database.QueryExecer, schoolAdminID pgtype.Text) (*entities.SchoolAdmin, error) {
					assert.Equal(t, "school-admin-1", schoolAdminID.String)
					return &entities.SchoolAdmin{
						SchoolAdminID: database.Text("school-admin-1"),
						SchoolID:      database.Int4(1),
					}, nil
				},
			},
			teacherRp: TeacherRepoMock{
				retrieve: func(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray, fields ...string) ([]entities.Teacher, error) {
					assert.ElementsMatch(t, []string{"teacher-id-1", "teacher-id-2"}, database.FromTextArray(ids))
					return []entities.Teacher{
						{
							ID:        database.Text("teacher-id-1"),
							SchoolIDs: database.Int4Array([]int32{3}),
							User: entities.User{
								Country: database.Text("vietnam"),
							},
						},
						{
							ID:        database.Text("teacher-id-2"),
							SchoolIDs: database.Int4Array([]int32{1}),
							User: entities.User{
								Country: database.Text("vietnam"),
							},
						},
					}, nil
				},
			},
			studentRp: StudentRepoMock{
				retrieve: func(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]repositories.StudentProfile, error) {
					assert.ElementsMatch(t, []string{"learner-id-2", "learner-id-3", "learner-id-4", "learner-id-5", "learner-id-6"}, database.FromTextArray(ids))
					return []repositories.StudentProfile{
						{
							Student: entities.Student{
								ID:       database.Text("learner-id-2"),
								SchoolID: database.Int4(1),
								User: entities.User{
									ID:      database.Text("learner-id-2"),
									Country: database.Text("vietnam"),
								},
							},
							School: entities.School{
								ID: database.Int4(1),
							},
						},
						{
							Student: entities.Student{
								ID:       database.Text("learner-id-3"),
								SchoolID: database.Int4(1),
								User: entities.User{
									ID:      database.Text("learner-id-3"),
									Country: database.Text("vietnam"),
								},
							},
							School: entities.School{
								ID: database.Int4(1),
							},
						},
						{
							Student: entities.Student{
								ID:       database.Text("learner-id-4"),
								SchoolID: database.Int4(1),
								User: entities.User{
									ID:      database.Text("learner-id-4"),
									Country: database.Text("vietnam"),
								},
							},
							School: entities.School{
								ID: database.Int4(1),
							},
						},
						{
							Student: entities.Student{
								ID:       database.Text("learner-id-5"),
								SchoolID: database.Int4(1),
								User: entities.User{
									ID:      database.Text("learner-id-5"),
									Country: database.Text("vietnam"),
								},
							},
							School: entities.School{
								ID: database.Int4(1),
							},
						},
						{
							Student: entities.Student{
								ID:       database.Text("learner-id-6"),
								SchoolID: database.Int4(1),
								User: entities.User{
									ID:      database.Text("learner-id-6"),
									Country: database.Text("vietnam"),
								},
							},
							School: entities.School{
								ID: database.Int4(1),
							},
						},
					}, nil
				},
			},
			hasError: true,
		},
		{
			name:   "create a live lesson with course not same school id",
			caller: "school-admin-1",
			req: &bpb.CreateLiveLessonRequest{
				Name:       "Lesson name 1",
				StartTime:  timestamppb.New(time.Date(2020, 2, 3, 4, 5, 6, 7, time.UTC)),
				EndTime:    timestamppb.New(time.Date(2021, 2, 3, 4, 5, 6, 7, time.UTC)),
				TeacherIds: []string{"teacher-id-1", "teacher-id-2"},
				CourseIds:  []string{"course-id-1", "course-id-4", "course-id-4", "course-id-5"},
				LearnerIds: []string{"learner-id-2", "learner-id-3", "learner-id-4", "learner-id-5", "learner-id-6"},
				Materials: []*bpb.Material{
					{
						Resource: &bpb.Material_BrightcoveVideo_{
							BrightcoveVideo: &bpb.Material_BrightcoveVideo{
								Name: "video 1",
								Url:  "https://brightcove.com/account/2/video?videoId=abc123",
							},
						},
					},
					{
						Resource: &bpb.Material_MediaId{
							MediaId: "media-id-2",
						},
					},
					{
						Resource: &bpb.Material_MediaId{
							MediaId: "media-id-2",
						},
					},
					{
						Resource: &bpb.Material_MediaId{
							MediaId: "media-id-3",
						},
					},
				},
			},
			courseRepo: coursesRepo.CourseRepoMock{
				FindByIDsMock: func(ctx context.Context, db database.Ext, courseIDs pgtype.TextArray) (map[pgtype.Text]*entities.Course, error) {
					assert.ElementsMatch(t, []string{"course-id-1", "course-id-4", "course-id-5"}, database.FromTextArray(courseIDs))
					return map[pgtype.Text]*entities.Course{
						database.Text("course-id-1"): {
							ID:        database.Text("course-id-1"),
							Name:      database.Text("math level 1"),
							Country:   database.Text("vietnam"),
							Grade:     database.Int2(2),
							Subject:   database.Text("math"),
							SchoolID:  database.Int4(1),
							StartDate: database.Timestamptz(time.Date(2021, 2, 3, 4, 5, 6, 7, time.UTC)),
						},
						database.Text("course-id-4"): {
							ID:        database.Text("course-id-4"),
							Name:      database.Text("physics level 2"),
							Country:   database.Text("vietnam"),
							Grade:     database.Int2(5),
							Subject:   database.Text("physics"),
							SchoolID:  database.Int4(1),
							StartDate: database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.UTC)),
						},
						database.Text("course-id-5"): {
							ID:        database.Text("course-id-5"),
							Name:      database.Text("chemistry level 1"),
							Country:   database.Text("vietnam"),
							Grade:     database.Int2(7),
							Subject:   database.Text("chemistry"),
							SchoolID:  database.Int4(2),
							StartDate: database.Timestamptz(time.Date(2023, 2, 3, 4, 5, 6, 7, time.UTC)),
						},
					}, nil
				},
			},
			userRp: UserRepoMock{
				get: func(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*entities.User, error) {
					assert.Equal(t, "school-admin-1", id.String)
					return &entities.User{
						ID:      database.Text("school-admin-1"),
						Country: database.Text("vietnam"),
					}, nil
				},
			},
			schoolAdminRp: SchoolAdminRepoMock{
				get: func(ctx context.Context, db database.QueryExecer, schoolAdminID pgtype.Text) (*entities.SchoolAdmin, error) {
					assert.Equal(t, "school-admin-1", schoolAdminID.String)
					return &entities.SchoolAdmin{
						SchoolAdminID: database.Text("school-admin-1"),
						SchoolID:      database.Int4(1),
					}, nil
				},
			},
			teacherRp: TeacherRepoMock{
				retrieve: func(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray, fields ...string) ([]entities.Teacher, error) {
					assert.ElementsMatch(t, []string{"teacher-id-1", "teacher-id-2"}, database.FromTextArray(ids))
					return []entities.Teacher{
						{
							ID:        database.Text("teacher-id-1"),
							SchoolIDs: database.Int4Array([]int32{1, 3}),
							User: entities.User{
								Country: database.Text("vietnam"),
							},
						},
						{
							ID:        database.Text("teacher-id-2"),
							SchoolIDs: database.Int4Array([]int32{1}),
							User: entities.User{
								Country: database.Text("vietnam"),
							},
						},
					}, nil
				},
			},
			studentRp: StudentRepoMock{
				retrieve: func(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]repositories.StudentProfile, error) {
					assert.ElementsMatch(t, []string{"learner-id-2", "learner-id-3", "learner-id-4", "learner-id-5", "learner-id-6"}, database.FromTextArray(ids))
					return []repositories.StudentProfile{
						{
							Student: entities.Student{
								ID:       database.Text("learner-id-2"),
								SchoolID: database.Int4(1),
								User: entities.User{
									ID:      database.Text("learner-id-2"),
									Country: database.Text("vietnam"),
								},
							},
							School: entities.School{
								ID: database.Int4(1),
							},
						},
						{
							Student: entities.Student{
								ID:       database.Text("learner-id-3"),
								SchoolID: database.Int4(1),
								User: entities.User{
									ID:      database.Text("learner-id-3"),
									Country: database.Text("vietnam"),
								},
							},
							School: entities.School{
								ID: database.Int4(1),
							},
						},
						{
							Student: entities.Student{
								ID:       database.Text("learner-id-4"),
								SchoolID: database.Int4(1),
								User: entities.User{
									ID:      database.Text("learner-id-4"),
									Country: database.Text("vietnam"),
								},
							},
							School: entities.School{
								ID: database.Int4(1),
							},
						},
						{
							Student: entities.Student{
								ID:       database.Text("learner-id-5"),
								SchoolID: database.Int4(1),
								User: entities.User{
									ID:      database.Text("learner-id-5"),
									Country: database.Text("vietnam"),
								},
							},
							School: entities.School{
								ID: database.Int4(1),
							},
						},
						{
							Student: entities.Student{
								ID:       database.Text("learner-id-6"),
								SchoolID: database.Int4(1),
								User: entities.User{
									ID:      database.Text("learner-id-6"),
									Country: database.Text("vietnam"),
								},
							},
							School: entities.School{
								ID: database.Int4(1),
							},
						},
					}, nil
				},
			},
			hasError: true,
		},
		{
			name:   "create a live lesson with learner not same school id",
			caller: "school-admin-1",
			req: &bpb.CreateLiveLessonRequest{
				Name:       "Lesson name 1",
				StartTime:  timestamppb.New(time.Date(2020, 2, 3, 4, 5, 6, 7, time.UTC)),
				EndTime:    timestamppb.New(time.Date(2021, 2, 3, 4, 5, 6, 7, time.UTC)),
				TeacherIds: []string{"teacher-id-1", "teacher-id-2"},
				CourseIds:  []string{"course-id-1", "course-id-4", "course-id-4", "course-id-5"},
				LearnerIds: []string{"learner-id-2", "learner-id-3", "learner-id-4", "learner-id-5", "learner-id-6"},
				Materials: []*bpb.Material{
					{
						Resource: &bpb.Material_BrightcoveVideo_{
							BrightcoveVideo: &bpb.Material_BrightcoveVideo{
								Name: "video 1",
								Url:  "https://brightcove.com/account/2/video?videoId=abc123",
							},
						},
					},
					{
						Resource: &bpb.Material_MediaId{
							MediaId: "media-id-2",
						},
					},
					{
						Resource: &bpb.Material_MediaId{
							MediaId: "media-id-2",
						},
					},
					{
						Resource: &bpb.Material_MediaId{
							MediaId: "media-id-3",
						},
					},
				},
			},
			courseRepo: coursesRepo.CourseRepoMock{
				FindByIDsMock: func(ctx context.Context, db database.Ext, courseIDs pgtype.TextArray) (map[pgtype.Text]*entities.Course, error) {
					assert.ElementsMatch(t, []string{"course-id-1", "course-id-4", "course-id-5"}, database.FromTextArray(courseIDs))
					return map[pgtype.Text]*entities.Course{
						database.Text("course-id-1"): {
							ID:        database.Text("course-id-1"),
							Name:      database.Text("math level 1"),
							Country:   database.Text("vietnam"),
							Grade:     database.Int2(2),
							Subject:   database.Text("math"),
							SchoolID:  database.Int4(1),
							StartDate: database.Timestamptz(time.Date(2021, 2, 3, 4, 5, 6, 7, time.UTC)),
						},
						database.Text("course-id-4"): {
							ID:        database.Text("course-id-4"),
							Name:      database.Text("physics level 2"),
							Country:   database.Text("vietnam"),
							Grade:     database.Int2(5),
							Subject:   database.Text("physics"),
							SchoolID:  database.Int4(1),
							StartDate: database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.UTC)),
						},
						database.Text("course-id-5"): {
							ID:        database.Text("course-id-5"),
							Name:      database.Text("chemistry level 1"),
							Country:   database.Text("vietnam"),
							Grade:     database.Int2(7),
							Subject:   database.Text("chemistry"),
							SchoolID:  database.Int4(1),
							StartDate: database.Timestamptz(time.Date(2023, 2, 3, 4, 5, 6, 7, time.UTC)),
						},
					}, nil
				},
			},
			userRp: UserRepoMock{
				get: func(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*entities.User, error) {
					assert.Equal(t, "school-admin-1", id.String)
					return &entities.User{
						ID:      database.Text("school-admin-1"),
						Country: database.Text("vietnam"),
					}, nil
				},
			},
			schoolAdminRp: SchoolAdminRepoMock{
				get: func(ctx context.Context, db database.QueryExecer, schoolAdminID pgtype.Text) (*entities.SchoolAdmin, error) {
					assert.Equal(t, "school-admin-1", schoolAdminID.String)
					return &entities.SchoolAdmin{
						SchoolAdminID: database.Text("school-admin-1"),
						SchoolID:      database.Int4(1),
					}, nil
				},
			},
			teacherRp: TeacherRepoMock{
				retrieve: func(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray, fields ...string) ([]entities.Teacher, error) {
					assert.ElementsMatch(t, []string{"teacher-id-1", "teacher-id-2"}, database.FromTextArray(ids))
					return []entities.Teacher{
						{
							ID:        database.Text("teacher-id-1"),
							SchoolIDs: database.Int4Array([]int32{1, 3}),
							User: entities.User{
								Country: database.Text("vietnam"),
							},
						},
						{
							ID:        database.Text("teacher-id-2"),
							SchoolIDs: database.Int4Array([]int32{1}),
							User: entities.User{
								Country: database.Text("vietnam"),
							},
						},
					}, nil
				},
			},
			studentRp: StudentRepoMock{
				retrieve: func(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]repositories.StudentProfile, error) {
					assert.ElementsMatch(t, []string{"learner-id-2", "learner-id-3", "learner-id-4", "learner-id-5", "learner-id-6"}, database.FromTextArray(ids))
					return []repositories.StudentProfile{
						{
							Student: entities.Student{
								ID:       database.Text("learner-id-2"),
								SchoolID: database.Int4(1),
								User: entities.User{
									ID:      database.Text("learner-id-2"),
									Country: database.Text("vietnam"),
								},
							},
							School: entities.School{
								ID: database.Int4(1),
							},
						},
						{
							Student: entities.Student{
								ID:       database.Text("learner-id-3"),
								SchoolID: database.Int4(1),
								User: entities.User{
									ID:      database.Text("learner-id-3"),
									Country: database.Text("vietnam"),
								},
							},
							School: entities.School{
								ID: database.Int4(1),
							},
						},
						{
							Student: entities.Student{
								ID:       database.Text("learner-id-4"),
								SchoolID: database.Int4(1),
								User: entities.User{
									ID:      database.Text("learner-id-4"),
									Country: database.Text("vietnam"),
								},
							},
							School: entities.School{
								ID: database.Int4(1),
							},
						},
						{
							Student: entities.Student{
								ID:       database.Text("learner-id-5"),
								SchoolID: database.Int4(1),
								User: entities.User{
									ID:      database.Text("learner-id-5"),
									Country: database.Text("vietnam"),
								},
							},
							School: entities.School{
								ID: database.Int4(1),
							},
						},
						{
							Student: entities.Student{
								ID:       database.Text("learner-id-6"),
								SchoolID: database.Int4(6),
								User: entities.User{
									ID:      database.Text("learner-id-6"),
									Country: database.Text("vietnam"),
								},
							},
							School: entities.School{
								ID: database.Int4(6),
							},
						},
					}, nil
				},
			},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setup != nil {
				tc.setup(ctx)
			}
			db := &database.ExtMock{
				TxStarterMock: *database.NewTxStarterMock(
					func(ctx context.Context) (pgx.Tx, error) {
						return database.NewTxMock(
							nil,
							nil,
							func(ctx context.Context) error { return nil },
							func(ctx context.Context) error { return nil },
							nil,
							nil,
							nil,
							nil,
							nil,
							nil,
							nil,
							nil,
						), nil
					},
				),
			}
			srv := NewLessonModifierServices(
				configurations.Config{},
				db,
				tc.mediaRp,
				nil,
				tc.lessonRepo,
				tc.lessonGroupRp,
				tc.courseRepo,
				tc.pSPRepo,
				tc.pSPWRepo,
				tc.topicRepo,
				tc.userRp,
				tc.schoolAdminRp,
				tc.teacherRp,
				tc.studentRp,
				nil,
				nil,
				nil,
				jsm,
				nil,
			)
			jsm.On("PublishAsyncContext", mock.Anything, mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {}).Return("", nil)
			ctx := context.Background()
			ctx = interceptors.ContextWithUserID(ctx, tc.caller)
			res, err := srv.CreateLiveLesson(ctx, tc.req)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, res.Id)
			}
		})
	}
}

func TestLessonModifierServices_UpdateLiveLesson(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// mock structs
	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	mediaRepo := &mock_repositories.MockMediaRepo{}
	actLogRepo := &mock_repositories.MockActivityLogRepo{}
	lessonRepo := &mock_repositories.MockLessonRepo{}
	lessonGroupRepo := &mock_repositories.MockLessonGroupRepo{}
	courseRepo := &mock_repositories.MockCourseRepo{}
	pspRepo := &mock_repositories.MockPresetStudyPlanRepo{}
	pspwRepo := &yasuo_mock_repositories.MockPresetStudyPlanWeeklyRepo{}
	topicRepo := &yasuo_mock_repositories.MockTopicRepo{}
	teacherRepo := &mock_repositories.MockTeacherRepo{}
	studentRepo := &mock_repositories.MockStudentRepo{}
	jsm := new(mock_nats.JetStreamManagement)

	// test cases
	testcases := []struct {
		name             string
		request          *bpb.UpdateLiveLessonRequest
		setup            func(context.Context)
		expectedResponse *bpb.UpdateLiveLessonResponse
		expectedError    error
	}{
		{
			name: "update successfully",
			request: &bpb.UpdateLiveLessonRequest{
				Id:         "lessonid1",
				Name:       "lessonname1_updated",
				StartTime:  timestamppb.New(time.Date(2020, 2, 3, 4, 5, 6, 7, time.UTC)),
				EndTime:    timestamppb.New(time.Date(2020, 3, 3, 4, 5, 6, 7, time.UTC)),
				TeacherIds: []string{"teacherid1", "teacherid2"},
				CourseIds:  []string{"courseid2", "courseid1"}, // adding courseid1 and removing courseid3
				LearnerIds: []string{"learnerid1", "learnerid2"},
				Materials: []*bpb.Material{
					{
						Resource: &bpb.Material_BrightcoveVideo_{
							BrightcoveVideo: &bpb.Material_BrightcoveVideo{
								Name: "video 1",
								Url:  "https://brightcove.com/account/2/video?videoId=abc123",
							},
						},
					},
					{Resource: &bpb.Material_MediaId{MediaId: "mediaid2"}},
					{Resource: &bpb.Material_MediaId{MediaId: "mediaid3"}},
				},
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lessonid1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lessonid1"),
						CourseID:      database.Text("courseid2"),
						Name:          database.Text("lessonname1"),
						StartTime:     database.Timestamptz(time.Date(2020, 4, 3, 4, 5, 6, 7, time.UTC)),
						EndTime:       database.Timestamptz(time.Date(2020, 5, 3, 4, 5, 6, 7, time.UTC)),
						LessonGroupID: database.Text("lessongroupid1"),
					}, nil).Once()
				lessonRepo.
					On("GetCourseIDsOfLesson", ctx, db, database.Text("lessonid1")).
					Return(database.TextArrayVariadic("courseid2", "courseid3"), nil).Once()
				courseRepo.
					On("FindByID", ctx, db, database.Text("courseid2")).
					Return(
						&entities.Course{
							ID:       database.Text("courseid2"),
							SchoolID: database.Int4(1),
							Country:  database.Text("vietnam"),
						},
						nil,
					).Once()
				teacherRepo.
					On("Retrieve", ctx, db, database.TextArray([]string{"teacherid1", "teacherid2"}), mock.Anything).
					Return(
						[]entities.Teacher{
							{
								ID:        database.Text("teacherid1"),
								SchoolIDs: database.Int4Array([]int32{3, 1}),
								User: entities.User{
									Country: database.Text("vietnam"),
								},
							},
							{
								ID:        database.Text("teacherid2"),
								SchoolIDs: database.Int4Array([]int32{1}),
								User: entities.User{
									Country: database.Text("vietnam"),
								},
							},
						},
						nil,
					).
					Once()
				courseRepo.
					On("FindByIDs", ctx, db, database.TextArrayVariadic("courseid2", "courseid1")).
					Return(map[pgtype.Text]*entities.Course{
						database.Text("courseid2"): {
							ID:                database.Text("courseid2"),
							PresetStudyPlanID: database.Text("pspid2"),
							SchoolID:          database.Int4(1),
							Country:           database.Text("vietnam"),
						},
						database.Text("courseid1"): {
							ID:                database.Text("courseid1"),
							PresetStudyPlanID: pgtype.Text{Status: pgtype.Null},
							SchoolID:          database.Int4(1),
							Country:           database.Text("vietnam"),
						},
					}, nil).Once()
				studentRepo.
					On("Retrieve", ctx, db, database.TextArray([]string{"learnerid1", "learnerid2"})).
					Return([]repositories.StudentProfile{
						{
							Student: entities.Student{
								ID:       database.Text("learnerid1"),
								SchoolID: database.Int4(1),
								User: entities.User{
									ID:      database.Text("learnerid1"),
									Country: database.Text("vietnam"),
								},
							},
							School: entities.School{
								ID: database.Int4(1),
							},
						},
						{
							Student: entities.Student{
								ID:       database.Text("learnerid2"),
								SchoolID: database.Int4(1),
								User: entities.User{
									ID:      database.Text("learnerid2"),
									Country: database.Text("vietnam"),
								},
							},
							School: entities.School{
								ID: database.Int4(1),
							},
						},
					}, nil).
					Once()
				mediaRepo.
					On("RetrieveByIDs", ctx, tx, database.TextArrayVariadic("mediaid2", "mediaid3")).
					Return([]*entities.Media{{MediaID: database.Text("mediaid2")}, {MediaID: database.Text("mediaid3")}}, nil).Once()
				mediaRepo.
					On("UpsertMediaBatch", ctx, tx, entities.Medias{{
						MediaID:         pgtype.Text{Status: pgtype.Null},
						Name:            database.Text("video 1"),
						Resource:        database.Text("abc123"),
						Type:            database.Text(string(entities.MediaTypeVideo)),
						Comments:        pgtype.JSONB{Status: pgtype.Null},
						CreatedAt:       pgtype.Timestamptz{Status: pgtype.Null},
						UpdatedAt:       pgtype.Timestamptz{Status: pgtype.Null},
						DeletedAt:       pgtype.Timestamptz{Status: pgtype.Null},
						ConvertedImages: pgtype.JSONB{Status: pgtype.Null},
					}}).
					Run(func(args mock.Arguments) {
						medias := args.Get(2).(entities.Medias)
						medias[0].MediaID = database.Text("newmediaid1")
					}).
					Return(nil).Once()
				lessonGroupRepo.
					On("UpdateMedias", ctx, tx, &entities.LessonGroup{
						LessonGroupID: database.Text("lessongroupid1"),
						CourseID:      database.Text("courseid2"),
						MediaIDs:      database.TextArrayVariadic("mediaid2", "mediaid3", "newmediaid1"),
						CreatedAt:     pgtype.Timestamptz{Status: pgtype.Null},
						UpdatedAt:     pgtype.Timestamptz{Status: pgtype.Null},
					}).
					Return(nil).Once()
				courseRepo.
					On("FindByIDs", ctx, tx, database.TextArrayVariadic("courseid2", "courseid1")).
					Return(map[pgtype.Text]*entities.Course{
						database.Text("courseid2"): {
							ID:                database.Text("courseid2"),
							PresetStudyPlanID: database.Text("pspid2"),
						},
						database.Text("courseid1"): {
							ID:                database.Text("courseid1"),
							PresetStudyPlanID: pgtype.Text{Status: pgtype.Null},
						},
					}, nil).Once()
				pspRepo.
					On("CreatePresetStudyPlan", ctx, tx, mock.Anything). // can't predict ID of PSP, so use check in .Run()
					Run(func(args mock.Arguments) {
						psp := args.Get(2).([]*entities.PresetStudyPlan)
						require.Len(t, psp, 1) // 1 course in request doesnt have PSP, so must create for one
					}).
					Return(nil).Once()
				courseRepo.
					On("Upsert", ctx, tx, mock.Anything). // can't predict ID of PSP here too
					Run(func(args mock.Arguments) {
						psp := args.Get(2).([]*entities.Course)
						require.Len(t, psp, 1) // must update PSP for 1 above course
					}).
					Return(nil).Once()
				courseRepo.
					On("FindByIDs", ctx, tx, database.TextArrayVariadic("courseid1")).
					Return(map[pgtype.Text]*entities.Course{
						database.Text("courseid1"): {
							ID:                database.Text("courseid1"),
							PresetStudyPlanID: database.Text("pspid1"),
						},
					}, nil).Once()
				topicRepo.
					On("Create", ctx, tx, mock.Anything).
					Return(nil).Once()
				pspwRepo.
					On("Create", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						pspw := args.Get(2).([]*entities.PresetStudyPlanWeekly)
						assert.Len(t, pspw, 1) // one courses added, so must create one PSPW
					}).
					Return(nil).Once()
				courseRepo.
					On("GetPresetStudyPlanIDsByCourseIDs", ctx, tx, database.TextArrayVariadic("courseid3")).
					Return([]string{"pspid3"}, nil).Once() // may not actually pspid3
				pspwRepo.
					On("GetIDsByLessonIDAndPresetStudyPlanIDs", ctx, tx, database.Text("lessonid1"), database.TextArrayVariadic("pspid3")).
					Return([]string{"pspwid1&3"}, nil).Once()
				topicRepo.
					On("SoftDeleteByPresetStudyPlanWeeklyIDs", ctx, tx, database.TextArrayVariadic("pspwid1&3")).
					Return(nil).Once()
				pspwRepo.
					On("SoftDelete", ctx, tx, database.TextArrayVariadic("pspwid1&3")).
					Return(nil).Once()
				topicRepo.
					On("UpdateNameByLessonID", ctx, tx, database.Text("lessonid1"), database.Text("lessonname1_updated")).
					Return(nil).Once()
				pspwRepo.
					On("UpdateTimeByLessonAndCourses", ctx, tx,
						database.Text("lessonid1"),
						database.TextArrayVariadic("courseid2"),
						database.Timestamptz(time.Date(2020, 2, 3, 4, 5, 6, 7, time.UTC)),
						database.Timestamptz(time.Date(2020, 3, 3, 4, 5, 6, 7, time.UTC)),
					).
					Return(nil).Once()
				lessonRepo.
					On("Update", ctx, tx, &entities.Lesson{
						LessonID:             database.Text("lessonid1"),
						Name:                 database.Text("lessonname1_updated"),
						TeacherID:            database.Text("teacherid1"),
						CourseID:             database.Text("courseid2"),
						ControlSettings:      pgtype.JSONB{Status: pgtype.Null},
						CreatedAt:            pgtype.Timestamptz{Status: pgtype.Null},
						UpdatedAt:            pgtype.Timestamptz{Status: pgtype.Null},
						DeletedAt:            pgtype.Timestamptz{Status: pgtype.Null},
						EndAt:                pgtype.Timestamptz{Status: pgtype.Null},
						StartTime:            database.Timestamptz(time.Date(2020, 2, 3, 4, 5, 6, 7, time.UTC)),
						EndTime:              database.Timestamptz(time.Date(2020, 3, 3, 4, 5, 6, 7, time.UTC)),
						LessonGroupID:        database.Text("lessongroupid1"),
						RoomID:               pgtype.Text{Status: pgtype.Null},
						LessonType:           database.Text(string(entities.LessonTypeOnline)),
						Status:               database.Text(string(entities.LessonStatusNone)),
						StreamLearnerCounter: database.Int4(0),
						LearnerIds:           database.TextArray([]string{}),
						TeacherIDs:           entities.TeacherIDs{TeacherIDs: database.TextArrayVariadic("teacherid1", "teacherid2")},
						CourseIDs:            entities.CourseIDs{CourseIDs: database.TextArrayVariadic("courseid2", "courseid1")},
						LearnerIDs:           entities.LearnerIDs{LearnerIDs: database.TextArrayVariadic("learnerid1", "learnerid2")},
						RoomState:            pgtype.JSONB{Status: pgtype.Null},
						TeachingModel:        pgtype.Text{Status: pgtype.Null},
						ClassID:              pgtype.Text{Status: pgtype.Null},
						CenterID:             pgtype.Text{Status: pgtype.Null},
						TeachingMethod:       pgtype.Text{Status: pgtype.Null},
						TeachingMedium:       database.Text(string(entities.LessonTeachingMediumOnline)),
						SchedulingStatus:     database.Text("LESSON_SCHEDULING_STATUS_PUBLISHED"),
						IsLocked:             database.Bool(false),
						ZoomLink:             pgtype.Text{Status: pgtype.Null},
					}).
					Return(nil).Once()
				lessonRepo.
					On("UpsertLessonCourses", ctx, tx, database.Text("lessonid1"), database.TextArrayVariadic("courseid2", "courseid1")).
					Return(nil).Once()
				lessonRepo.
					On("UpsertLessonTeachers", ctx, tx, database.Text("lessonid1"), database.TextArrayVariadic("teacherid1", "teacherid2")).
					Return(nil).Once()
				lessonRepo.
					On("UpsertLessonMembers", ctx, tx, database.Text("lessonid1"), database.TextArrayVariadic("learnerid1", "learnerid2")).
					Return(nil).Once()
				courseRepo.
					On("UpdateStartAndEndDate", ctx, tx, mock.Anything). // since we use golibs.Uniq, check the courseIDs with Run() instead
					Run(func(args mock.Arguments) {
						courseIDs := args.Get(2).(pgtype.TextArray)
						require.ElementsMatch(t, database.FromTextArray(courseIDs), []string{"courseid1", "courseid2", "courseid3"})
					}).
					Return(nil).Once()

				jsm.On("PublishAsyncContext", ctx, cconstants.SubjectLessonUpdated, mock.Anything).Run(func(args mock.Arguments) {
					data := args.Get(2).([]byte)
					msg := &pb.EvtLesson{}
					err := msg.Unmarshal(data)
					require.NoError(t, err)

					message := msg.Message.(*pb.EvtLesson_UpdateLesson_)

					assert.Equal(t, "lessonname1_updated", message.UpdateLesson.ClassName)
					assert.Equal(t, "lessonid1", message.UpdateLesson.LessonId)
					assert.Equal(t, []string{"learnerid1", "learnerid2"}, message.UpdateLesson.LearnerIds)
				}).Return("", nil).Once()
			},
			expectedResponse: &bpb.UpdateLiveLessonResponse{},
			expectedError:    nil,
		},
		{
			name: "update successfully without materials",
			request: &bpb.UpdateLiveLessonRequest{
				Id:         "lessonid1",
				Name:       "lessonname1_updated",
				StartTime:  timestamppb.New(time.Date(2020, 2, 3, 4, 5, 6, 7, time.UTC)),
				EndTime:    timestamppb.New(time.Date(2020, 3, 3, 4, 5, 6, 7, time.UTC)),
				TeacherIds: []string{"teacherid1", "teacherid2"},
				CourseIds:  []string{"courseid2", "courseid1"}, // adding courseid1 and removing courseid3
				LearnerIds: []string{"learnerid1", "learnerid2"},
				Materials:  nil,
			},
			setup: func(context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lessonid1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lessonid1"),
						CourseID:      database.Text("courseid2"),
						Name:          database.Text("lessonname1"),
						StartTime:     database.Timestamptz(time.Date(2020, 4, 3, 4, 5, 6, 7, time.UTC)),
						EndTime:       database.Timestamptz(time.Date(2020, 5, 3, 4, 5, 6, 7, time.UTC)),
						LessonGroupID: database.Text("lessongroupid1"),
					}, nil).Once()
				lessonRepo.
					On("GetCourseIDsOfLesson", ctx, db, database.Text("lessonid1")).
					Return(database.TextArrayVariadic("courseid2", "courseid3"), nil).Once()
				courseRepo.
					On("FindByID", ctx, db, database.Text("courseid2")).
					Return(
						&entities.Course{
							ID:       database.Text("courseid2"),
							SchoolID: database.Int4(1),
							Country:  database.Text("vietnam"),
						},
						nil,
					).Once()
				teacherRepo.
					On("Retrieve", ctx, db, database.TextArray([]string{"teacherid1", "teacherid2"}), mock.Anything).
					Return(
						[]entities.Teacher{
							{
								ID:        database.Text("teacherid1"),
								SchoolIDs: database.Int4Array([]int32{3, 1}),
								User: entities.User{
									Country: database.Text("vietnam"),
								},
							},
							{
								ID:        database.Text("teacherid2"),
								SchoolIDs: database.Int4Array([]int32{1}),
								User: entities.User{
									Country: database.Text("vietnam"),
								},
							},
						},
						nil,
					).
					Once()
				courseRepo.
					On("FindByIDs", ctx, db, database.TextArrayVariadic("courseid2", "courseid1")).
					Return(map[pgtype.Text]*entities.Course{
						database.Text("courseid2"): {
							ID:                database.Text("courseid2"),
							PresetStudyPlanID: database.Text("pspid2"),
							SchoolID:          database.Int4(1),
							Country:           database.Text("vietnam"),
						},
						database.Text("courseid1"): {
							ID:                database.Text("courseid1"),
							PresetStudyPlanID: pgtype.Text{Status: pgtype.Null},
							SchoolID:          database.Int4(1),
							Country:           database.Text("vietnam"),
						},
					}, nil).Once()
				studentRepo.
					On("Retrieve", ctx, db, database.TextArray([]string{"learnerid1", "learnerid2"})).
					Return([]repositories.StudentProfile{
						{
							Student: entities.Student{
								ID:       database.Text("learnerid1"),
								SchoolID: database.Int4(1),
								User: entities.User{
									ID:      database.Text("learnerid1"),
									Country: database.Text("vietnam"),
								},
							},
							School: entities.School{
								ID: database.Int4(1),
							},
						},
						{
							Student: entities.Student{
								ID:       database.Text("learnerid2"),
								SchoolID: database.Int4(1),
								User: entities.User{
									ID:      database.Text("learnerid2"),
									Country: database.Text("vietnam"),
								},
							},
							School: entities.School{
								ID: database.Int4(1),
							},
						},
					}, nil).
					Once()
				lessonGroupRepo.
					On("UpdateMedias", ctx, tx, &entities.LessonGroup{
						LessonGroupID: database.Text("lessongroupid1"),
						CourseID:      database.Text("courseid2"),
						MediaIDs:      pgtype.TextArray{Status: pgtype.Null},
						CreatedAt:     pgtype.Timestamptz{Status: pgtype.Null},
						UpdatedAt:     pgtype.Timestamptz{Status: pgtype.Null},
					}).
					Return(nil).Once()
				courseRepo.
					On("FindByIDs", ctx, tx, database.TextArrayVariadic("courseid2", "courseid1")).
					Return(map[pgtype.Text]*entities.Course{
						database.Text("courseid2"): {
							ID:                database.Text("courseid2"),
							PresetStudyPlanID: database.Text("pspid2"),
						},
						database.Text("courseid1"): {
							ID:                database.Text("courseid1"),
							PresetStudyPlanID: pgtype.Text{Status: pgtype.Null},
						},
					}, nil).Once()
				pspRepo.
					On("CreatePresetStudyPlan", ctx, tx, mock.Anything). // can't predict ID of PSP, so use check in .Run()
					Run(func(args mock.Arguments) {
						psp := args.Get(2).([]*entities.PresetStudyPlan)
						require.Len(t, psp, 1) // 1 course in request doesnt have PSP, so must create for one
					}).
					Return(nil).Once()
				courseRepo.
					On("Upsert", ctx, tx, mock.Anything). // can't predict ID of PSP here too
					Run(func(args mock.Arguments) {
						psp := args.Get(2).([]*entities.Course)
						require.Len(t, psp, 1) // must update PSP for 1 above course
					}).
					Return(nil).Once()
				courseRepo.
					On("FindByIDs", ctx, tx, database.TextArrayVariadic("courseid1")).
					Return(map[pgtype.Text]*entities.Course{
						database.Text("courseid1"): {
							ID:                database.Text("courseid1"),
							PresetStudyPlanID: database.Text("pspid1"),
						},
					}, nil).Once()
				topicRepo.
					On("Create", ctx, tx, mock.Anything).
					Return(nil).Once()
				pspwRepo.
					On("Create", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						pspw := args.Get(2).([]*entities.PresetStudyPlanWeekly)
						assert.Len(t, pspw, 1) // one courses added, so must create one PSPW
					}).
					Return(nil).Once()
				courseRepo.
					On("GetPresetStudyPlanIDsByCourseIDs", ctx, tx, database.TextArrayVariadic("courseid3")).
					Return([]string{"pspid3"}, nil).Once() // may not actually pspid3
				pspwRepo.
					On("GetIDsByLessonIDAndPresetStudyPlanIDs", ctx, tx, database.Text("lessonid1"), database.TextArrayVariadic("pspid3")).
					Return([]string{"pspwid1&3"}, nil).Once()
				topicRepo.
					On("SoftDeleteByPresetStudyPlanWeeklyIDs", ctx, tx, database.TextArrayVariadic("pspwid1&3")).
					Return(nil).Once()
				pspwRepo.
					On("SoftDelete", ctx, tx, database.TextArrayVariadic("pspwid1&3")).
					Return(nil).Once()
				topicRepo.
					On("UpdateNameByLessonID", ctx, tx, database.Text("lessonid1"), database.Text("lessonname1_updated")).
					Return(nil).Once()
				pspwRepo.
					On("UpdateTimeByLessonAndCourses", ctx, tx,
						database.Text("lessonid1"),
						database.TextArrayVariadic("courseid2"),
						database.Timestamptz(time.Date(2020, 2, 3, 4, 5, 6, 7, time.UTC)),
						database.Timestamptz(time.Date(2020, 3, 3, 4, 5, 6, 7, time.UTC)),
					).
					Return(nil).Once()
				lessonRepo.
					On("Update", ctx, tx, &entities.Lesson{
						LessonID:             database.Text("lessonid1"),
						Name:                 database.Text("lessonname1_updated"),
						TeacherID:            database.Text("teacherid1"),
						CourseID:             database.Text("courseid2"),
						ControlSettings:      pgtype.JSONB{Status: pgtype.Null},
						CreatedAt:            pgtype.Timestamptz{Status: pgtype.Null},
						UpdatedAt:            pgtype.Timestamptz{Status: pgtype.Null},
						DeletedAt:            pgtype.Timestamptz{Status: pgtype.Null},
						EndAt:                pgtype.Timestamptz{Status: pgtype.Null},
						StartTime:            database.Timestamptz(time.Date(2020, 2, 3, 4, 5, 6, 7, time.UTC)),
						EndTime:              database.Timestamptz(time.Date(2020, 3, 3, 4, 5, 6, 7, time.UTC)),
						LessonGroupID:        database.Text("lessongroupid1"),
						RoomID:               pgtype.Text{Status: pgtype.Null},
						LessonType:           database.Text(string(entities.LessonTypeOnline)),
						Status:               database.Text(string(entities.LessonStatusNone)),
						StreamLearnerCounter: database.Int4(0),
						LearnerIds:           database.TextArray([]string{}),
						TeacherIDs:           entities.TeacherIDs{TeacherIDs: database.TextArrayVariadic("teacherid1", "teacherid2")},
						CourseIDs:            entities.CourseIDs{CourseIDs: database.TextArrayVariadic("courseid2", "courseid1")},
						LearnerIDs:           entities.LearnerIDs{LearnerIDs: database.TextArrayVariadic("learnerid1", "learnerid2")},
						RoomState:            pgtype.JSONB{Status: pgtype.Null},
						TeachingModel:        pgtype.Text{Status: pgtype.Null},
						ClassID:              pgtype.Text{Status: pgtype.Null},
						CenterID:             pgtype.Text{Status: pgtype.Null},
						TeachingMethod:       pgtype.Text{Status: pgtype.Null},
						TeachingMedium:       database.Text(string(entities.LessonTeachingMediumOnline)),
						SchedulingStatus:     database.Text("LESSON_SCHEDULING_STATUS_PUBLISHED"),
						IsLocked:             database.Bool(false),
						ZoomLink:             pgtype.Text{Status: pgtype.Null},
					}).
					Return(nil).Once()
				lessonRepo.
					On("UpsertLessonCourses", ctx, tx, database.Text("lessonid1"), database.TextArrayVariadic("courseid2", "courseid1")).
					Return(nil).Once()
				lessonRepo.
					On("UpsertLessonTeachers", ctx, tx, database.Text("lessonid1"), database.TextArrayVariadic("teacherid1", "teacherid2")).
					Return(nil).Once()
				lessonRepo.
					On("UpsertLessonMembers", ctx, tx, database.Text("lessonid1"), database.TextArrayVariadic("learnerid1", "learnerid2")).
					Return(nil).Once()
				courseRepo.
					On("UpdateStartAndEndDate", ctx, tx, mock.Anything). // since we use golibs.Uniq, check the courseIDs with Run() instead
					Run(func(args mock.Arguments) {
						courseIDs := args.Get(2).(pgtype.TextArray)
						require.ElementsMatch(t, database.FromTextArray(courseIDs), []string{"courseid1", "courseid2", "courseid3"})
					}).
					Return(nil).Once()
				jsm.On("PublishAsyncContext", ctx, cconstants.SubjectLessonUpdated, mock.Anything).Run(func(args mock.Arguments) {
					data := args.Get(2).([]byte)
					msg := &pb.EvtLesson{}
					err := msg.Unmarshal(data)
					require.NoError(t, err)

					message := msg.Message.(*pb.EvtLesson_UpdateLesson_)

					assert.Equal(t, "lessonname1_updated", message.UpdateLesson.ClassName)
					assert.Equal(t, "lessonid1", message.UpdateLesson.LessonId)
					assert.Equal(t, []string{"learnerid1", "learnerid2"}, message.UpdateLesson.LearnerIds)
				}).Return("", nil).Once()
			},
			expectedResponse: &bpb.UpdateLiveLessonResponse{},
			expectedError:    nil,
		},
		{
			name: "update courses and start/end time, with all courses replaced",
			request: &bpb.UpdateLiveLessonRequest{
				Id:         "lessonid1",
				Name:       "lessonname1",
				StartTime:  timestamppb.New(time.Date(2020, 2, 3, 4, 5, 6, 7, time.UTC)),
				EndTime:    timestamppb.New(time.Date(2020, 3, 3, 4, 5, 6, 7, time.UTC)),
				TeacherIds: []string{"teacherid1", "teacherid2"},
				CourseIds:  []string{"courseid1"}, // adding courseid1 and removing courseid3
				LearnerIds: []string{"learnerid1", "learnerid2"},
				Materials: []*bpb.Material{
					{
						Resource: &bpb.Material_BrightcoveVideo_{
							BrightcoveVideo: &bpb.Material_BrightcoveVideo{
								Name: "video 1",
								Url:  "https://brightcove.com/account/2/video?videoId=abc123",
							},
						},
					},
					{Resource: &bpb.Material_MediaId{MediaId: "mediaid2"}},
					{Resource: &bpb.Material_MediaId{MediaId: "mediaid3"}},
				},
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lessonid1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lessonid1"),
						Name:          database.Text("lessonname1"),
						StartTime:     database.Timestamptz(time.Date(2020, 4, 3, 4, 5, 6, 7, time.UTC)),
						EndTime:       database.Timestamptz(time.Date(2020, 5, 3, 4, 5, 6, 7, time.UTC)),
						LessonGroupID: database.Text("lessongroupid1"),
					}, nil).Once()
				lessonRepo.
					On("GetCourseIDsOfLesson", ctx, db, database.Text("lessonid1")).
					Return(database.TextArrayVariadic("courseid3"), nil).Once()
				courseRepo.
					On("FindByID", ctx, db, database.Text("courseid3")).
					Return(
						&entities.Course{
							ID:       database.Text("courseid3"),
							SchoolID: database.Int4(1),
							Country:  database.Text("vietnam"),
						},
						nil,
					).Once()
				teacherRepo.
					On("Retrieve", ctx, db, database.TextArray([]string{"teacherid1", "teacherid2"}), mock.Anything).
					Return(
						[]entities.Teacher{
							{
								ID:        database.Text("teacherid1"),
								SchoolIDs: database.Int4Array([]int32{3, 1}),
								User: entities.User{
									Country: database.Text("vietnam"),
								},
							},
							{
								ID:        database.Text("teacherid2"),
								SchoolIDs: database.Int4Array([]int32{1}),
								User: entities.User{
									Country: database.Text("vietnam"),
								},
							},
						},
						nil,
					).
					Once()
				courseRepo.
					On("FindByIDs", ctx, db, database.TextArrayVariadic("courseid1")).
					Return(map[pgtype.Text]*entities.Course{
						database.Text("courseid1"): {
							ID:                database.Text("courseid1"),
							PresetStudyPlanID: pgtype.Text{Status: pgtype.Null},
							SchoolID:          database.Int4(1),
							Country:           database.Text("vietnam"),
						},
					}, nil).Once()
				studentRepo.
					On("Retrieve", ctx, db, database.TextArray([]string{"learnerid1", "learnerid2"})).
					Return([]repositories.StudentProfile{
						{
							Student: entities.Student{
								ID:       database.Text("learnerid1"),
								SchoolID: database.Int4(1),
								User: entities.User{
									ID:      database.Text("learnerid1"),
									Country: database.Text("vietnam"),
								},
							},
							School: entities.School{
								ID: database.Int4(1),
							},
						},
						{
							Student: entities.Student{
								ID:       database.Text("learnerid2"),
								SchoolID: database.Int4(1),
								User: entities.User{
									ID:      database.Text("learnerid2"),
									Country: database.Text("vietnam"),
								},
							},
							School: entities.School{
								ID: database.Int4(1),
							},
						},
					}, nil).
					Once()
				mediaRepo.
					On("RetrieveByIDs", ctx, tx, database.TextArrayVariadic("mediaid2", "mediaid3")).
					Return([]*entities.Media{{MediaID: database.Text("mediaid2")}, {MediaID: database.Text("mediaid3")}}, nil).Once()
				mediaRepo.
					On("UpsertMediaBatch", ctx, tx, entities.Medias{{
						MediaID:         pgtype.Text{Status: pgtype.Null},
						Name:            database.Text("video 1"),
						Resource:        database.Text("abc123"),
						Type:            database.Text(string(entities.MediaTypeVideo)),
						Comments:        pgtype.JSONB{Status: pgtype.Null},
						CreatedAt:       pgtype.Timestamptz{Status: pgtype.Null},
						UpdatedAt:       pgtype.Timestamptz{Status: pgtype.Null},
						DeletedAt:       pgtype.Timestamptz{Status: pgtype.Null},
						ConvertedImages: pgtype.JSONB{Status: pgtype.Null},
					}}).
					Run(func(args mock.Arguments) {
						medias := args.Get(2).(entities.Medias)
						medias[0].MediaID = database.Text("newmediaid1")
					}).
					Return(nil).Once()
				lessonGroupRepo.
					On("Create", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						lessonGr := args.Get(2).(*entities.LessonGroup)
						assert.NotEqual(t, "lessongroupid1", lessonGr.LessonGroupID.String)
						if len(lessonGr.LessonGroupID.String) == 0 {
							lessonGr.LessonGroupID = database.Text("lessongroupid2")
						}
						expected := entities.LessonGroup{
							LessonGroupID: lessonGr.LessonGroupID,
							CourseID:      database.Text("courseid1"),
							MediaIDs:      database.TextArrayVariadic("mediaid2", "mediaid3", "newmediaid1"),
							CreatedAt:     pgtype.Timestamptz{Status: pgtype.Null},
							UpdatedAt:     pgtype.Timestamptz{Status: pgtype.Null},
						}
						assert.EqualValues(t, expected, *lessonGr)
					}).
					Return(nil).Once()
				courseRepo.
					On("FindByIDs", ctx, tx, database.TextArrayVariadic("courseid1")).
					Return(map[pgtype.Text]*entities.Course{
						database.Text("courseid1"): {
							ID:                database.Text("courseid1"),
							PresetStudyPlanID: pgtype.Text{Status: pgtype.Null},
						},
					}, nil).Once()
				pspRepo.
					On("CreatePresetStudyPlan", ctx, tx, mock.Anything). // can't predict ID of PSP, so use check in .Run()
					Run(func(args mock.Arguments) {
						psp := args.Get(2).([]*entities.PresetStudyPlan)
						require.Len(t, psp, 1) // 1 course in request doesnt have PSP, so must create for one
					}).
					Return(nil).Once()
				courseRepo.
					On("Upsert", ctx, tx, mock.Anything). // can't predict ID of PSP here too
					Run(func(args mock.Arguments) {
						psp := args.Get(2).([]*entities.Course)
						require.Len(t, psp, 1) // must update PSP for 1 above course
					}).
					Return(nil).Once()
				courseRepo.
					On("FindByIDs", ctx, tx, database.TextArrayVariadic("courseid1")).
					Return(map[pgtype.Text]*entities.Course{
						database.Text("courseid1"): {
							ID:                database.Text("courseid1"),
							PresetStudyPlanID: database.Text("pspid1"),
						},
					}, nil).Once()
				topicRepo.
					On("Create", ctx, tx, mock.Anything).
					Return(nil).Once()
				pspwRepo.
					On("Create", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						pspw := args.Get(2).([]*entities.PresetStudyPlanWeekly)
						assert.Len(t, pspw, 1) // one courses added, so must create one PSPW
					}).
					Return(nil).Once()
				courseRepo.
					On("GetPresetStudyPlanIDsByCourseIDs", ctx, tx, database.TextArrayVariadic("courseid3")).
					Return([]string{"pspid3"}, nil).Once() // may not actually pspid3
				pspwRepo.
					On("GetIDsByLessonIDAndPresetStudyPlanIDs", ctx, tx, database.Text("lessonid1"), database.TextArrayVariadic("pspid3")).
					Return([]string{"pspwid1&3"}, nil).Once()
				topicRepo.
					On("SoftDeleteByPresetStudyPlanWeeklyIDs", ctx, tx, database.TextArrayVariadic("pspwid1&3")).
					Return(nil).Once()
				pspwRepo.
					On("SoftDelete", ctx, tx, database.TextArrayVariadic("pspwid1&3")).
					Return(nil).Once()
				lessonRepo.
					On("Update", ctx, tx, &entities.Lesson{
						LessonID:             database.Text("lessonid1"),
						Name:                 database.Text("lessonname1"),
						TeacherID:            database.Text("teacherid1"),
						CourseID:             database.Text("courseid1"),
						ControlSettings:      pgtype.JSONB{Status: pgtype.Null},
						CreatedAt:            pgtype.Timestamptz{Status: pgtype.Null},
						UpdatedAt:            pgtype.Timestamptz{Status: pgtype.Null},
						DeletedAt:            pgtype.Timestamptz{Status: pgtype.Null},
						EndAt:                pgtype.Timestamptz{Status: pgtype.Null},
						StartTime:            database.Timestamptz(time.Date(2020, 2, 3, 4, 5, 6, 7, time.UTC)),
						EndTime:              database.Timestamptz(time.Date(2020, 3, 3, 4, 5, 6, 7, time.UTC)),
						LessonGroupID:        database.Text("lessongroupid2"),
						RoomID:               pgtype.Text{Status: pgtype.Null},
						LessonType:           database.Text(string(entities.LessonTypeOnline)),
						Status:               database.Text(string(entities.LessonStatusNone)),
						StreamLearnerCounter: database.Int4(0),
						LearnerIds:           database.TextArray([]string{}),
						TeacherIDs:           entities.TeacherIDs{TeacherIDs: database.TextArrayVariadic("teacherid1", "teacherid2")},
						CourseIDs:            entities.CourseIDs{CourseIDs: database.TextArrayVariadic("courseid1")},
						LearnerIDs:           entities.LearnerIDs{LearnerIDs: database.TextArrayVariadic("learnerid1", "learnerid2")},
						RoomState:            pgtype.JSONB{Status: pgtype.Null},
						TeachingModel:        pgtype.Text{Status: pgtype.Null},
						ClassID:              pgtype.Text{Status: pgtype.Null},
						CenterID:             pgtype.Text{Status: pgtype.Null},
						TeachingMethod:       pgtype.Text{Status: pgtype.Null},
						TeachingMedium:       database.Text(string(entities.LessonTeachingMediumOnline)),
						SchedulingStatus:     database.Text("LESSON_SCHEDULING_STATUS_PUBLISHED"),
						IsLocked:             database.Bool(false),
						ZoomLink:             pgtype.Text{Status: pgtype.Null},
					}).
					Return(nil).Once()
				lessonRepo.
					On("UpsertLessonCourses", ctx, tx, database.Text("lessonid1"), database.TextArrayVariadic("courseid1")).
					Return(nil).Once()
				lessonRepo.
					On("UpsertLessonTeachers", ctx, tx, database.Text("lessonid1"), database.TextArrayVariadic("teacherid1", "teacherid2")).
					Return(nil).Once()
				lessonRepo.
					On("UpsertLessonMembers", ctx, tx, database.Text("lessonid1"), database.TextArrayVariadic("learnerid1", "learnerid2")).
					Return(nil).Once()
				courseRepo.
					On("UpdateStartAndEndDate", ctx, tx, mock.Anything). // since we use golibs.Uniq, check the courseIDs with Run() instead
					Run(func(args mock.Arguments) {
						courseIDs := args.Get(2).(pgtype.TextArray)
						require.ElementsMatch(t, database.FromTextArray(courseIDs), []string{"courseid1", "courseid3"})
					}).
					Return(nil).Once()
				jsm.On("PublishAsyncContext", ctx, cconstants.SubjectLessonUpdated, mock.Anything).Run(func(args mock.Arguments) {
					data := args.Get(2).([]byte)
					msg := &pb.EvtLesson{}
					err := msg.Unmarshal(data)
					require.NoError(t, err)

					message := msg.Message.(*pb.EvtLesson_UpdateLesson_)

					assert.Equal(t, "lessonname1", message.UpdateLesson.ClassName)
					assert.Equal(t, "lessonid1", message.UpdateLesson.LessonId)
					assert.Equal(t, []string{"learnerid1", "learnerid2"}, message.UpdateLesson.LearnerIds)
				}).Return("", nil).Once()
			},
			expectedResponse: &bpb.UpdateLiveLessonResponse{},
			expectedError:    nil,
		},
		{
			name: "error from missing name",
			request: &bpb.UpdateLiveLessonRequest{
				Id:         "lessonid1",
				Name:       "",
				StartTime:  timestamppb.New(time.Date(2020, 2, 3, 4, 5, 6, 7, time.UTC)),
				EndTime:    timestamppb.New(time.Date(2020, 3, 3, 4, 5, 6, 7, time.UTC)),
				TeacherIds: []string{"teacherid1", "teacherid2"},
				CourseIds:  []string{"courseid1", "courseid2"}, // adding courseid1 and removing courseid3
				LearnerIds: []string{"learnerid1", "learnerid2"},
				Materials: []*bpb.Material{
					{
						Resource: &bpb.Material_BrightcoveVideo_{
							BrightcoveVideo: &bpb.Material_BrightcoveVideo{
								Name: "video 1",
								Url:  "https://brightcove.com/account/2/video?videoId=abc123",
							},
						},
					},
					{
						Resource: &bpb.Material_MediaId{
							MediaId: "mediaid2",
						},
					},
					{
						Resource: &bpb.Material_MediaId{
							MediaId: "mediaid3",
						},
					},
				},
			},
			setup: func(context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lessonid1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lessonid1"),
						Name:          database.Text("lessonname1"),
						StartTime:     database.Timestamptz(time.Date(2020, 4, 3, 4, 5, 6, 7, time.UTC)),
						EndTime:       database.Timestamptz(time.Date(2020, 5, 3, 4, 5, 6, 7, time.UTC)),
						LessonGroupID: database.Text("lessongroupid1"),
					}, nil).Once()
				lessonRepo.
					On("GetCourseIDsOfLesson", ctx, db, database.Text("lessonid1")).
					Return(database.TextArrayVariadic("courseid2", "courseid3"), nil).Once()
				courseRepo.
					On("FindByID", ctx, db, database.Text("courseid2")).
					Return(
						&entities.Course{
							ID:       database.Text("courseid2"),
							SchoolID: database.Int4(1),
							Country:  database.Text("vietnam"),
						},
						nil,
					).Once()
			},
			expectedResponse: nil,
			expectedError:    status.Error(codes.Internal, "builder.UpdateWithMedias: Lesson.IsValid: Lesson.Name cannot be empty"),
		},
		{
			name: "error from missing start time",
			request: &bpb.UpdateLiveLessonRequest{
				Id:         "lessonid1",
				Name:       "lessonname1",
				StartTime:  nil,
				EndTime:    timestamppb.New(time.Date(2020, 3, 3, 4, 5, 6, 7, time.UTC)),
				TeacherIds: []string{"teacherid1", "teacherid2"},
				CourseIds:  []string{"courseid1", "courseid2"}, // adding courseid1 and removing courseid3
				LearnerIds: []string{"learnerid1", "learnerid2"},
				Materials: []*bpb.Material{
					{
						Resource: &bpb.Material_BrightcoveVideo_{
							BrightcoveVideo: &bpb.Material_BrightcoveVideo{
								Name: "video 1",
								Url:  "https://brightcove.com/account/2/video?videoId=abc123",
							},
						},
					},
					{Resource: &bpb.Material_MediaId{MediaId: "mediaid2"}},
					{Resource: &bpb.Material_MediaId{MediaId: "mediaid3"}},
				},
			},
			setup: func(context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lessonid1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lessonid1"),
						Name:          database.Text("lessonname1"),
						StartTime:     database.Timestamptz(time.Date(2020, 4, 3, 4, 5, 6, 7, time.UTC)),
						EndTime:       database.Timestamptz(time.Date(2020, 5, 3, 4, 5, 6, 7, time.UTC)),
						LessonGroupID: database.Text("lessongroupid1"),
					}, nil).Once()
				lessonRepo.
					On("GetCourseIDsOfLesson", ctx, db, database.Text("lessonid1")).
					Return(database.TextArrayVariadic("courseid2", "courseid3"), nil).Once()
				courseRepo.
					On("FindByID", ctx, db, database.Text("courseid2")).
					Return(
						&entities.Course{
							ID:       database.Text("courseid2"),
							SchoolID: database.Int4(1),
							Country:  database.Text("vietnam"),
						},
						nil,
					).Once()
			},
			expectedResponse: nil,
			expectedError:    status.Error(codes.Internal, "builder.UpdateWithMedias: Lesson.IsValid: Lesson.StartTime and Lesson.EndTime cannot be empty"),
		},
		{
			name: "error from not having one or more teacher IDs",
			request: &bpb.UpdateLiveLessonRequest{
				Id:         "lessonid1",
				Name:       "lessonname1",
				StartTime:  timestamppb.New(time.Date(2020, 2, 3, 4, 5, 6, 7, time.UTC)),
				EndTime:    timestamppb.New(time.Date(2020, 3, 3, 4, 5, 6, 7, time.UTC)),
				TeacherIds: nil,
				CourseIds:  []string{"courseid1", "courseid2"}, // adding courseid1 and removing courseid3
				LearnerIds: []string{"learnerid1", "learnerid2"},
				Materials: []*bpb.Material{
					{
						Resource: &bpb.Material_BrightcoveVideo_{
							BrightcoveVideo: &bpb.Material_BrightcoveVideo{
								Name: "video 1",
								Url:  "https://brightcove.com/account/2/video?videoId=abc123",
							},
						},
					},
					{Resource: &bpb.Material_MediaId{MediaId: "mediaid2"}},
					{Resource: &bpb.Material_MediaId{MediaId: "mediaid3"}},
				},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lessonid1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lessonid1"),
						Name:          database.Text("lessonname1"),
						StartTime:     database.Timestamptz(time.Date(2020, 4, 3, 4, 5, 6, 7, time.UTC)),
						EndTime:       database.Timestamptz(time.Date(2020, 5, 3, 4, 5, 6, 7, time.UTC)),
						LessonGroupID: database.Text("lessongroupid1"),
					}, nil).Once()
				lessonRepo.
					On("GetCourseIDsOfLesson", ctx, db, database.Text("lessonid1")).
					Return(database.TextArrayVariadic("courseid2", "courseid3"), nil).Once()
				courseRepo.
					On("FindByID", ctx, db, database.Text("courseid2")).
					Return(
						&entities.Course{
							ID:       database.Text("courseid2"),
							SchoolID: database.Int4(1),
							Country:  database.Text("vietnam"),
						},
						nil,
					).Once()
			},
			expectedResponse: nil,
			expectedError:    status.Error(codes.Internal, "builder.UpdateWithMedias: Lesson.IsValid: Lesson.TeacherID cannot be empty"),
		},
		{
			name: "error from not having one or more course IDs",
			request: &bpb.UpdateLiveLessonRequest{
				Id:         "lessonid1",
				Name:       "lessonname1",
				StartTime:  timestamppb.New(time.Date(2020, 2, 3, 4, 5, 6, 7, time.UTC)),
				EndTime:    timestamppb.New(time.Date(2020, 3, 3, 4, 5, 6, 7, time.UTC)),
				TeacherIds: []string{"teacherid1", "teacherid2"},
				CourseIds:  nil,
				LearnerIds: []string{"learnerid1", "learnerid2"},
				Materials: []*bpb.Material{
					{
						Resource: &bpb.Material_BrightcoveVideo_{
							BrightcoveVideo: &bpb.Material_BrightcoveVideo{
								Name: "video 1",
								Url:  "https://brightcove.com/account/2/video?videoId=abc123",
							},
						},
					},
					{Resource: &bpb.Material_MediaId{MediaId: "mediaid2"}},
					{Resource: &bpb.Material_MediaId{MediaId: "mediaid3"}},
				},
			},
			setup: func(context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lessonid1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lessonid1"),
						Name:          database.Text("lessonname1"),
						StartTime:     database.Timestamptz(time.Date(2020, 4, 3, 4, 5, 6, 7, time.UTC)),
						EndTime:       database.Timestamptz(time.Date(2020, 5, 3, 4, 5, 6, 7, time.UTC)),
						LessonGroupID: database.Text("lessongroupid1"),
					}, nil).Once()
				lessonRepo.
					On("GetCourseIDsOfLesson", ctx, db, database.Text("lessonid1")).
					Return(database.TextArrayVariadic("courseid2", "courseid3"), nil).Once()
				courseRepo.
					On("FindByID", ctx, db, database.Text("courseid2")).
					Return(
						&entities.Course{
							ID:       database.Text("courseid2"),
							SchoolID: database.Int4(1),
							Country:  database.Text("vietnam"),
						},
						nil,
					).Once()
			},
			expectedResponse: nil,
			expectedError:    status.Error(codes.Internal, "builder.UpdateWithMedias: Lesson.IsValid: Lesson.CourseID cannot be empty"),
		},
		{
			name: "error from not having one or more learner IDs",
			request: &bpb.UpdateLiveLessonRequest{
				Id:         "lessonid1",
				Name:       "lessonname1",
				StartTime:  timestamppb.New(time.Date(2020, 2, 3, 4, 5, 6, 7, time.UTC)),
				EndTime:    timestamppb.New(time.Date(2020, 3, 3, 4, 5, 6, 7, time.UTC)),
				TeacherIds: []string{"teacherid1", "teacherid2"},
				CourseIds:  []string{"courseid1", "courseid2"}, // adding courseid1 and removing courseid3
				LearnerIds: nil,
				Materials: []*bpb.Material{
					{
						Resource: &bpb.Material_BrightcoveVideo_{
							BrightcoveVideo: &bpb.Material_BrightcoveVideo{
								Name: "video 1",
								Url:  "https://brightcove.com/account/2/video?videoId=abc123",
							},
						},
					},
					{Resource: &bpb.Material_MediaId{MediaId: "mediaid2"}},
					{Resource: &bpb.Material_MediaId{MediaId: "mediaid3"}},
				},
			},
			setup: func(context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lessonid1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lessonid1"),
						Name:          database.Text("lessonname1"),
						StartTime:     database.Timestamptz(time.Date(2020, 4, 3, 4, 5, 6, 7, time.UTC)),
						EndTime:       database.Timestamptz(time.Date(2020, 5, 3, 4, 5, 6, 7, time.UTC)),
						LessonGroupID: database.Text("lessongroupid1"),
					}, nil).Once()
				lessonRepo.
					On("GetCourseIDsOfLesson", ctx, db, database.Text("lessonid1")).
					Return(database.TextArrayVariadic("courseid2", "courseid3"), nil).Once()
				courseRepo.
					On("FindByID", ctx, db, database.Text("courseid2")).
					Return(
						&entities.Course{
							ID:       database.Text("courseid2"),
							SchoolID: database.Int4(1),
							Country:  database.Text("vietnam"),
						},
						nil,
					).Once()
			},
			expectedResponse: nil,
			expectedError:    status.Error(codes.Internal, "builder.UpdateWithMedias: Lesson.IsValid: Lesson.LearnerIDs cannot be empty"),
		},
		{
			name: "error from having start time after end time",
			request: &bpb.UpdateLiveLessonRequest{
				Id:         "lessonid1",
				Name:       "lessonname1",
				StartTime:  timestamppb.New(time.Date(2021, 2, 3, 4, 5, 6, 7, time.UTC)),
				EndTime:    timestamppb.New(time.Date(2020, 3, 3, 4, 5, 6, 7, time.UTC)),
				TeacherIds: []string{"teacherid1", "teacherid2"},
				CourseIds:  []string{"courseid1", "courseid2"}, // adding courseid1 and removing courseid3
				LearnerIds: []string{"learnerid1", "learnerid2"},
				Materials: []*bpb.Material{
					{
						Resource: &bpb.Material_BrightcoveVideo_{
							BrightcoveVideo: &bpb.Material_BrightcoveVideo{
								Name: "video 1",
								Url:  "https://brightcove.com/account/2/video?videoId=abc123",
							},
						},
					},
					{Resource: &bpb.Material_MediaId{MediaId: "mediaid2"}},
					{Resource: &bpb.Material_MediaId{MediaId: "mediaid3"}},
				},
			},
			setup: func(context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lessonid1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lessonid1"),
						Name:          database.Text("lessonname1"),
						StartTime:     database.Timestamptz(time.Date(2020, 4, 3, 4, 5, 6, 7, time.UTC)),
						EndTime:       database.Timestamptz(time.Date(2020, 5, 3, 4, 5, 6, 7, time.UTC)),
						LessonGroupID: database.Text("lessongroupid1"),
					}, nil).Once()
				lessonRepo.
					On("GetCourseIDsOfLesson", ctx, db, database.Text("lessonid1")).
					Return(database.TextArrayVariadic("courseid2", "courseid3"), nil).Once()
				courseRepo.
					On("FindByID", ctx, db, database.Text("courseid2")).
					Return(
						&entities.Course{
							ID:       database.Text("courseid2"),
							SchoolID: database.Int4(1),
							Country:  database.Text("vietnam"),
						},
						nil,
					).Once()
			},
			expectedResponse: nil,
			expectedError:    status.Error(codes.Internal, "builder.UpdateWithMedias: Lesson.IsValid: Lesson.StartTime cannot be after Lesson.EndTime"),
		},
		{
			name: "error from invalid brightcove url",
			request: &bpb.UpdateLiveLessonRequest{
				Id:         "lessonid1",
				Name:       "lessonname1",
				StartTime:  timestamppb.New(time.Date(2020, 2, 3, 4, 5, 6, 7, time.UTC)),
				EndTime:    timestamppb.New(time.Date(2020, 3, 3, 4, 5, 6, 7, time.UTC)),
				TeacherIds: []string{"teacherid1", "teacherid2"},
				CourseIds:  []string{"courseid1", "courseid2"}, // adding courseid1 and removing courseid3
				LearnerIds: []string{"learnerid1", "learnerid2"},
				Materials: []*bpb.Material{
					{
						Resource: &bpb.Material_BrightcoveVideo_{
							BrightcoveVideo: &bpb.Material_BrightcoveVideo{
								Name: "video 1",
								Url:  "https://brightcove.com/account/2/video?videoID=abc123",
							},
						},
					},
					{Resource: &bpb.Material_MediaId{MediaId: "mediaid2"}},
					{Resource: &bpb.Material_MediaId{MediaId: "mediaid3"}},
				},
			},
			setup:            func(context.Context) {},
			expectedResponse: nil,
			expectedError:    status.Error(codes.Internal, "materialsToMedias: could not extract video id from brightcove video url https://brightcove.com/account/2/video?videoID=abc123"),
		},
		{
			name: "update with student not same school id",
			request: &bpb.UpdateLiveLessonRequest{
				Id:         "lessonid1",
				Name:       "lessonname1_updated",
				StartTime:  timestamppb.New(time.Date(2020, 2, 3, 4, 5, 6, 7, time.UTC)),
				EndTime:    timestamppb.New(time.Date(2020, 3, 3, 4, 5, 6, 7, time.UTC)),
				TeacherIds: []string{"teacherid1", "teacherid2"},
				CourseIds:  []string{"courseid1", "courseid2"}, // adding courseid1 and removing courseid3
				LearnerIds: []string{"learnerid1", "learnerid2"},
				Materials: []*bpb.Material{
					{
						Resource: &bpb.Material_BrightcoveVideo_{
							BrightcoveVideo: &bpb.Material_BrightcoveVideo{
								Name: "video 1",
								Url:  "https://brightcove.com/account/2/video?videoId=abc123",
							},
						},
					},
					{Resource: &bpb.Material_MediaId{MediaId: "mediaid2"}},
					{Resource: &bpb.Material_MediaId{MediaId: "mediaid3"}},
				},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lessonid1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lessonid1"),
						Name:          database.Text("lessonname1"),
						StartTime:     database.Timestamptz(time.Date(2020, 4, 3, 4, 5, 6, 7, time.UTC)),
						EndTime:       database.Timestamptz(time.Date(2020, 5, 3, 4, 5, 6, 7, time.UTC)),
						LessonGroupID: database.Text("lessongroupid1"),
					}, nil).Once()
				lessonRepo.
					On("GetCourseIDsOfLesson", ctx, db, database.Text("lessonid1")).
					Return(database.TextArrayVariadic("courseid2", "courseid3"), nil).Once()
				courseRepo.
					On("FindByID", ctx, db, database.Text("courseid2")).
					Return(
						&entities.Course{
							ID:       database.Text("courseid2"),
							SchoolID: database.Int4(1),
							Country:  database.Text("vietnam"),
						},
						nil,
					).Once()
				teacherRepo.
					On("Retrieve", ctx, db, database.TextArray([]string{"teacherid1", "teacherid2"}), mock.Anything).
					Return(
						[]entities.Teacher{
							{
								ID:        database.Text("teacherid1"),
								SchoolIDs: database.Int4Array([]int32{3, 1}),
								User: entities.User{
									Country: database.Text("vietnam"),
								},
							},
							{
								ID:        database.Text("teacherid2"),
								SchoolIDs: database.Int4Array([]int32{1}),
								User: entities.User{
									Country: database.Text("vietnam"),
								},
							},
						},
						nil,
					).
					Once()
				courseRepo.
					On("FindByIDs", ctx, db, database.TextArrayVariadic("courseid1", "courseid2")).
					Return(map[pgtype.Text]*entities.Course{
						database.Text("courseid1"): {
							ID:                database.Text("courseid1"),
							PresetStudyPlanID: pgtype.Text{Status: pgtype.Null},
							SchoolID:          database.Int4(1),
							Country:           database.Text("vietnam"),
						},
						database.Text("courseid2"): {
							ID:                database.Text("courseid2"),
							PresetStudyPlanID: database.Text("pspid2"),
							SchoolID:          database.Int4(1),
							Country:           database.Text("vietnam"),
						},
					}, nil).Once()
				studentRepo.
					On("Retrieve", ctx, db, database.TextArray([]string{"learnerid1", "learnerid2"})).
					Return([]repositories.StudentProfile{
						{
							Student: entities.Student{
								ID:       database.Text("learnerid1"),
								SchoolID: database.Int4(1),
								User: entities.User{
									ID:      database.Text("learnerid1"),
									Country: database.Text("vietnam"),
								},
							},
							School: entities.School{
								ID: database.Int4(1),
							},
						},
						{
							Student: entities.Student{
								ID:       database.Text("learnerid2"),
								SchoolID: database.Int4(2),
								User: entities.User{
									ID:      database.Text("learnerid2"),
									Country: database.Text("vietnam"),
								},
							},
							School: entities.School{
								ID: database.Int4(2),
							},
						},
					}, nil).
					Once()
			},
			expectedResponse: nil,
			expectedError:    status.Error(codes.Internal, "builder.UpdateWithMedias: Lesson.IsValid: student learnerid2 is not belong to school 1"),
		},
		{
			name: "update with teacher not same school id",
			request: &bpb.UpdateLiveLessonRequest{
				Id:         "lessonid1",
				Name:       "lessonname1_updated",
				StartTime:  timestamppb.New(time.Date(2020, 2, 3, 4, 5, 6, 7, time.UTC)),
				EndTime:    timestamppb.New(time.Date(2020, 3, 3, 4, 5, 6, 7, time.UTC)),
				TeacherIds: []string{"teacherid1", "teacherid2"},
				CourseIds:  []string{"courseid1", "courseid2"}, // adding courseid1 and removing courseid3
				LearnerIds: []string{"learnerid1", "learnerid2"},
				Materials: []*bpb.Material{
					{
						Resource: &bpb.Material_BrightcoveVideo_{
							BrightcoveVideo: &bpb.Material_BrightcoveVideo{
								Name: "video 1",
								Url:  "https://brightcove.com/account/2/video?videoId=abc123",
							},
						},
					},
					{Resource: &bpb.Material_MediaId{MediaId: "mediaid2"}},
					{Resource: &bpb.Material_MediaId{MediaId: "mediaid3"}},
				},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lessonid1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lessonid1"),
						Name:          database.Text("lessonname1"),
						StartTime:     database.Timestamptz(time.Date(2020, 4, 3, 4, 5, 6, 7, time.UTC)),
						EndTime:       database.Timestamptz(time.Date(2020, 5, 3, 4, 5, 6, 7, time.UTC)),
						LessonGroupID: database.Text("lessongroupid1"),
					}, nil).Once()
				lessonRepo.
					On("GetCourseIDsOfLesson", ctx, db, database.Text("lessonid1")).
					Return(database.TextArrayVariadic("courseid2", "courseid3"), nil).Once()
				courseRepo.
					On("FindByID", ctx, db, database.Text("courseid2")).
					Return(
						&entities.Course{
							ID:       database.Text("courseid2"),
							SchoolID: database.Int4(1),
							Country:  database.Text("vietnam"),
						},
						nil,
					).Once()
				teacherRepo.
					On("Retrieve", ctx, db, database.TextArray([]string{"teacherid1", "teacherid2"}), mock.Anything).
					Return(
						[]entities.Teacher{
							{
								ID:        database.Text("teacherid1"),
								SchoolIDs: database.Int4Array([]int32{3, 1}),
								User: entities.User{
									Country: database.Text("vietnam"),
								},
							},
							{
								ID:        database.Text("teacherid2"),
								SchoolIDs: database.Int4Array([]int32{2}),
								User: entities.User{
									Country: database.Text("vietnam"),
								},
							},
						},
						nil,
					).
					Once()
			},
			expectedResponse: nil,
			expectedError:    status.Error(codes.Internal, "builder.UpdateWithMedias: Lesson.IsValid: teacher teacherid2 is not belong to school 1"),
		},
		{
			name: "update with course not same school id",
			request: &bpb.UpdateLiveLessonRequest{
				Id:         "lessonid1",
				Name:       "lessonname1_updated",
				StartTime:  timestamppb.New(time.Date(2020, 2, 3, 4, 5, 6, 7, time.UTC)),
				EndTime:    timestamppb.New(time.Date(2020, 3, 3, 4, 5, 6, 7, time.UTC)),
				TeacherIds: []string{"teacherid1", "teacherid2"},
				CourseIds:  []string{"courseid1", "courseid2"}, // adding courseid1 and removing courseid3
				LearnerIds: []string{"learnerid1", "learnerid2"},
				Materials: []*bpb.Material{
					{
						Resource: &bpb.Material_BrightcoveVideo_{
							BrightcoveVideo: &bpb.Material_BrightcoveVideo{
								Name: "video 1",
								Url:  "https://brightcove.com/account/2/video?videoId=abc123",
							},
						},
					},
					{Resource: &bpb.Material_MediaId{MediaId: "mediaid2"}},
					{Resource: &bpb.Material_MediaId{MediaId: "mediaid3"}},
				},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lessonid1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lessonid1"),
						Name:          database.Text("lessonname1"),
						StartTime:     database.Timestamptz(time.Date(2020, 4, 3, 4, 5, 6, 7, time.UTC)),
						EndTime:       database.Timestamptz(time.Date(2020, 5, 3, 4, 5, 6, 7, time.UTC)),
						LessonGroupID: database.Text("lessongroupid1"),
					}, nil).Once()
				lessonRepo.
					On("GetCourseIDsOfLesson", ctx, db, database.Text("lessonid1")).
					Return(database.TextArrayVariadic("courseid2", "courseid3"), nil).Once()
				courseRepo.
					On("FindByID", ctx, db, database.Text("courseid2")).
					Return(
						&entities.Course{
							ID:       database.Text("courseid2"),
							SchoolID: database.Int4(1),
							Country:  database.Text("vietnam"),
						},
						nil,
					).Once()
				teacherRepo.
					On("Retrieve", ctx, db, database.TextArray([]string{"teacherid1", "teacherid2"}), mock.Anything).
					Return(
						[]entities.Teacher{
							{
								ID:        database.Text("teacherid1"),
								SchoolIDs: database.Int4Array([]int32{3, 1}),
								User: entities.User{
									Country: database.Text("vietnam"),
								},
							},
							{
								ID:        database.Text("teacherid2"),
								SchoolIDs: database.Int4Array([]int32{1}),
								User: entities.User{
									Country: database.Text("vietnam"),
								},
							},
						},
						nil,
					).
					Once()
				courseRepo.
					On("FindByIDs", ctx, db, database.TextArrayVariadic("courseid1", "courseid2")).
					Return(map[pgtype.Text]*entities.Course{
						database.Text("courseid1"): {
							ID:                database.Text("courseid1"),
							PresetStudyPlanID: pgtype.Text{Status: pgtype.Null},
							SchoolID:          database.Int4(2),
							Country:           database.Text("vietnam"),
						},
						database.Text("courseid2"): {
							ID:                database.Text("courseid2"),
							PresetStudyPlanID: database.Text("pspid2"),
							SchoolID:          database.Int4(1),
							Country:           database.Text("vietnam"),
						},
					}, nil).Once()
			},
			expectedResponse: nil,
			expectedError:    status.Error(codes.Internal, "builder.UpdateWithMedias: Lesson.IsValid: course courseid1 is not belong to school 1"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			srv := NewLessonModifierServices(
				configurations.Config{},
				db,
				mediaRepo,
				actLogRepo,
				lessonRepo,
				lessonGroupRepo,
				courseRepo,
				pspRepo,
				pspwRepo,
				topicRepo,
				nil,
				nil,
				teacherRepo,
				studentRepo,
				nil,
				nil,
				nil,
				jsm,
				nil,
			)
			actualResponse, actualError := srv.UpdateLiveLesson(ctx, tc.request)
			assert.Equal(t, tc.expectedResponse, actualResponse)
			assert.Equal(t, tc.expectedError, actualError, `actual error: "%v"`, actualError)
			mock.AssertExpectationsForObjects(
				t, db, tx, mediaRepo, actLogRepo, lessonRepo, lessonGroupRepo, courseRepo,
				pspRepo, pspwRepo, topicRepo, teacherRepo, studentRepo,
			)
		})
	}
}

func TestDeleteLesson(t *testing.T) {
	t.Parallel()

	jsm := &mock_nats.JetStreamManagement{}
	jsm.On("PublishAsyncContext", mock.Anything, mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {}).Return("", nil)

	tcs := []struct {
		name       string
		req        *bpb.DeleteLiveLessonRequest
		lessonRepo LessonRepoMock
		pSPWRepo   coursesRepo.PresetStudyPlanWeeklyRepoMock
		topicRepo  topicsRepo.TopicRepo
		crsRepo    coursesRepo.CourseRepo
		hasError   bool
	}{
		{
			name: "delete a live lesson successfully",
			req: &bpb.DeleteLiveLessonRequest{
				Id: "lesson-id-1",
			},
			lessonRepo: LessonRepoMock{
				findByID: func(ctx context.Context, db database.Ext, id pgtype.Text) (*entities.Lesson, error) {
					assert.Equal(t, "lesson-id-1", id.String)

					return &entities.Lesson{
						LessonID:  id,
						StartTime: database.Timestamptz(time.Now().Add(time.Hour)),
						EndTime:   database.Timestamptz(time.Now().Add(2 * time.Hour)),
					}, nil
				},
				deleteLessonCourses: func(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text) error {
					assert.Equal(t, "lesson-id-1", lessonID.String)
					return nil
				},
				deleteLessonTeachers: func(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text) error {
					assert.Equal(t, "lesson-id-1", lessonID.String)
					return nil
				},
				deleteLessonMembers: func(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text) error {
					assert.Equal(t, "lesson-id-1", lessonID.String)
					return nil
				},
				delete: func(ctx context.Context, db database.QueryExecer, lessonIDs pgtype.TextArray) error {
					assert.ElementsMatch(t, []string{"lesson-id-1"}, database.FromTextArray(lessonIDs))
					return nil
				},
				findEarliestAndLatestTimeLessonByCourses: func(ctx context.Context, db database.Ext, courseIDs pgtype.TextArray) (*entities.CourseAvailableRanges, error) {
					assert.ElementsMatch(t, []string{"course-id-1", "course-id-2"}, database.FromTextArray(courseIDs))

					res := &entities.CourseAvailableRanges{}
					res.Add(
						&entities.CourseAvailableRange{
							ID:        database.Text("course-id-1"),
							StartDate: database.Timestamptz(time.Date(2019, 2, 3, 4, 5, 6, 7, time.UTC)),
							EndDate:   database.Timestamptz(time.Date(2020, 2, 3, 4, 5, 6, 7, time.UTC)),
						},
						&entities.CourseAvailableRange{
							ID:        database.Text("course-id-2"),
							StartDate: database.Timestamptz(time.Date(2020, 2, 3, 4, 5, 6, 7, time.UTC)),
							EndDate:   database.Timestamptz(time.Date(2021, 2, 3, 4, 5, 6, 7, time.UTC)),
						},
					)
					return res, nil
				},
			},
			pSPWRepo: coursesRepo.PresetStudyPlanWeeklyRepoMock{
				FindByLessonIDsMock: func(ctx context.Context, db database.QueryExecer, IDs pgtype.TextArray, isAll bool) (map[pgtype.Text]*entities.PresetStudyPlanWeekly, error) {
					assert.ElementsMatch(t, []string{"lesson-id-1"}, database.FromTextArray(IDs))
					return map[pgtype.Text]*entities.PresetStudyPlanWeekly{
						database.Text("preset-study-plan-weekly-id-1"): {
							ID:       database.Text("preset-study-plan-weekly-id-1"),
							TopicID:  database.Text("topic-id-1"),
							LessonID: database.Text("lesson-id-1"),
						},
						database.Text("preset-study-plan-weekly-id-2"): {
							ID:       database.Text("preset-study-plan-weekly-id-2"),
							TopicID:  database.Text("topic-id-2"),
							LessonID: database.Text("lesson-id-1"),
						},
					}, nil
				},
				SoftDeleteMock: func(ctx context.Context, db database.QueryExecer, presetStudyPlanWeeklyIDs pgtype.TextArray) error {
					assert.ElementsMatch(t, []string{"preset-study-plan-weekly-id-1", "preset-study-plan-weekly-id-2"}, database.FromTextArray(presetStudyPlanWeeklyIDs))
					return nil
				},
			},
			topicRepo: topicsRepo.TopicRepoMock{
				SoftDeleteV2Mock: func(ctx context.Context, db database.QueryExecer, topicIDs pgtype.TextArray) error {
					assert.ElementsMatch(t, []string{"topic-id-1", "topic-id-2"}, database.FromTextArray(topicIDs))
					return nil
				},
			},
			crsRepo: coursesRepo.CourseRepoMock{
				FindByIDsMock: func(ctx context.Context, db database.Ext, courseIDs pgtype.TextArray) (map[pgtype.Text]*entities.Course, error) {
					assert.ElementsMatch(t, []string{"course-id-1", "course-id-2"}, database.FromTextArray(courseIDs))
					return map[pgtype.Text]*entities.Course{
						database.Text("course-id-1"): {
							ID: database.Text("course-id-1"),
						},
						database.Text("course-id-2"): {
							ID: database.Text("course-id-2"),
						},
					}, nil
				},
				UpsertMock: func(ctx context.Context, db database.Ext, cc []*entities.Course) error {
					assert.Len(t, cc, 2)
					for _, c := range cc {
						if c.ID.String == "course-id-1" {
							assert.True(t, time.Date(2019, 2, 3, 4, 5, 6, 7, time.UTC).Equal(c.StartDate.Time))
							assert.True(t, time.Date(2020, 2, 3, 4, 5, 6, 7, time.UTC).Equal(c.EndDate.Time))
							assert.Equal(t, c.DeletedAt.Status, pgtype.Null)
						} else {
							assert.True(t, time.Date(2020, 2, 3, 4, 5, 6, 7, time.UTC).Equal(c.StartDate.Time))
							assert.True(t, time.Date(2021, 2, 3, 4, 5, 6, 7, time.UTC).Equal(c.EndDate.Time))
							assert.Equal(t, c.DeletedAt.Status, pgtype.Null)
						}
					}

					return nil
				},
				FindByLessonIDMock: func(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text) (entities.Courses, error) {
					assert.Equal(t, "lesson-id-1", lessonID.String)
					return []*entities.Course{
						{
							ID: database.Text("course-id-1"),
						},
						{
							ID: database.Text("course-id-2"),
						},
					}, nil
				},
			},
		},
		{
			name: "could not delete a processing live lesson",
			req: &bpb.DeleteLiveLessonRequest{
				Id: "lesson-id-1",
			},
			lessonRepo: LessonRepoMock{
				findByID: func(ctx context.Context, db database.Ext, id pgtype.Text) (*entities.Lesson, error) {
					assert.Equal(t, "lesson-id-1", id.String)

					return &entities.Lesson{
						LessonID: id,
						CourseIDs: entities.CourseIDs{
							CourseIDs: database.TextArray([]string{"course-id-1", "course-id-2"}),
						},
						StartTime: database.Timestamptz(time.Now().Add(-time.Hour)),
						EndTime:   database.Timestamptz(time.Now().Add(time.Hour)),
					}, nil
				},
			},
			hasError: true,
		},
		{
			name: "could not delete a completed live lesson",
			req: &bpb.DeleteLiveLessonRequest{
				Id: "lesson-id-1",
			},
			lessonRepo: LessonRepoMock{
				findByID: func(ctx context.Context, db database.Ext, id pgtype.Text) (*entities.Lesson, error) {
					assert.Equal(t, "lesson-id-1", id.String)

					return &entities.Lesson{
						LessonID: id,
						CourseIDs: entities.CourseIDs{
							CourseIDs: database.TextArray([]string{"course-id-1", "course-id-2"}),
						},
						StartTime: database.Timestamptz(time.Now().Add(-2 * time.Hour)),
						EndTime:   database.Timestamptz(time.Now().Add(-time.Hour)),
					}, nil
				},
			},
			hasError: true,
		},
	}

	for i := range tcs {
		t.Run(tcs[i].name, func(t *testing.T) {
			db := &database.ExtMock{
				TxStarterMock: *database.NewTxStarterMock(
					func(ctx context.Context) (pgx.Tx, error) {
						return database.NewTxMock(
							nil,
							nil,
							func(ctx context.Context) error { return nil },
							func(ctx context.Context) error { return nil },
							nil,
							nil,
							nil,
							nil,
							nil,
							nil,
							nil,
							nil,
						), nil
					},
				),
			}

			srv := NewLessonModifierServices(
				configurations.Config{},
				db,
				nil,
				nil,
				tcs[i].lessonRepo,
				nil,
				tcs[i].crsRepo,
				nil,
				tcs[i].pSPWRepo,
				tcs[i].topicRepo,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				jsm,
				nil,
			)
			_, err := srv.DeleteLiveLesson(context.Background(), tcs[i].req)
			if tcs[i].hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestModifyLiveLessonState(t *testing.T) {
	lessonRepo := &mock_repositories.MockLessonRepo{}
	lessonGroupRepo := &mock_repositories.MockLessonGroupRepo{}
	userRepo := &mock_repositories.MockUserRepo{}
	lessonMemberRepo := &mock_repositories.MockLessonMemberRepo{}
	lessonPollingRepo := &mock_repositories.MockLessonPollingRepo{}
	lessonRoomStateRepo := &mock_repositories_lessonmgmt.MockLessonRoomStateRepo{}
	logRepo := new(mock_repositories.MockVirtualClassroomLogRepo)
	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	now := time.Now().UTC()
	nowString, err := now.MarshalText()
	require.NoError(t, err)

	jsm := &mock_nats.JetStreamManagement{}

	tcs := []struct {
		name      string
		reqUserID string
		req       *bpb.ModifyLiveLessonStateRequest
		setup     func(ctx context.Context)
		hasError  bool
	}{
		{
			name:      "teacher command to share a material (video) in live lesson room",
			reqUserID: "teacher-1",
			req: &bpb.ModifyLiveLessonStateRequest{
				Id: "lesson-1",
				Command: &bpb.ModifyLiveLessonStateRequest_ShareAMaterial{
					ShareAMaterial: &bpb.ModifyLiveLessonStateRequest_CurrentMaterialCommand{
						MediaId: "media-2",
						State: &bpb.ModifyLiveLessonStateRequest_CurrentMaterialCommand_VideoState{
							VideoState: &bpb.LiveLessonState_CurrentMaterial_VideoState{
								CurrentTime: durationpb.New(12 * time.Second),
								PlayerState: bpb.PlayerState_PLAYER_STATE_PAUSE,
							},
						},
					},
				},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lesson-1"),
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						RoomState: database.JSONB(`
			{
				"current_material": {
					"media_id": "media-1",
					"updated_at": "` + string(nowString) + `",
					"video_state": {
						"current_time": "23m",
						"player_state": "PLAYER_STATE_PLAYING"
					}
				}
			}`),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"teacher-1", "teacher-2"}), nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"learner-1", "learner-2", "learner-3"}), nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				lessonRepo.
					On("FindByID", ctx, tx, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lesson-1"),
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						RoomState: database.JSONB(`
						{
							"current_material": {
								"media_id": "media-1",
								"updated_at": "` + string(nowString) + `",
								"video_state": {
									"current_time": "23m",
									"player_state": "PLAYER_STATE_PLAYING"
								}
							}
						}`),
					},
						nil,
					).Once()
				lessonRoomStateRepo.
					On("UpsertCurrentMaterialState", ctx, tx, database.Text("lesson-1"), mock.Anything).
					Run(func(args mock.Arguments) {
						stateJsonb := args.Get(3).(pgtype.JSONB)
						state := &CurrentMaterial{}
						err := stateJsonb.AssignTo(state)
						require.NoError(t, err)
						assert.Equal(t, "media-2", state.MediaID)
						assert.Equal(t, 12*time.Second, state.VideoState.CurrentTime.Duration())
						assert.Equal(t, PlayerStatePause, state.VideoState.PlayerState)
						assert.False(t, state.UpdatedAt.IsZero())
						assert.False(t, now.Equal(state.UpdatedAt))
					}).
					Return(nil).Once()
				lessonGroupRepo.
					On("Get", ctx, tx, database.Text("lesson-group-1"), database.Text("course-1")).
					Return(&entities.LessonGroup{
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						MediaIDs:      database.TextArray([]string{"media-1", "media-2", "media-3"}),
					}, nil).Once()
				logRepo.On(
					"IncreaseTotalTimesByLessonID",
					ctx,
					db,
					database.Text("lesson-1"),
					entities.TotalTimesUpdatingRoomState,
				).
					Return(nil).
					Once()
			},
		},
		{
			name:      "teacher command to share a material (pdf) in live lesson room",
			reqUserID: "teacher-1",
			req: &bpb.ModifyLiveLessonStateRequest{
				Id: "lesson-1",
				Command: &bpb.ModifyLiveLessonStateRequest_ShareAMaterial{
					ShareAMaterial: &bpb.ModifyLiveLessonStateRequest_CurrentMaterialCommand{
						MediaId: "media-2",
					},
				},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lesson-1"),
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						RoomState: database.JSONB(`
			{
				"current_material": {
					"media_id": "media-1",
					"updated_at": "` + string(nowString) + `",
					"video_state": {
						"current_time": "23m",
						"player_state": "PLAYER_STATE_PLAYING"
					}
				}
			}`),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"teacher-1", "teacher-2"}), nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"learner-1", "learner-2", "learner-3"}), nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				lessonRepo.
					On("FindByID", ctx, tx, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lesson-1"),
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						RoomState: database.JSONB(`
			{
				"current_material": {
					"media_id": "media-1",
					"updated_at": "` + string(nowString) + `",
					"video_state": {
						"current_time": "23m",
						"player_state": "PLAYER_STATE_PLAYING"
					}
				}
			}`),
					},
						nil,
					).Once()
				lessonRoomStateRepo.
					On("UpsertCurrentMaterialState", ctx, tx, database.Text("lesson-1"), mock.Anything).
					Run(func(args mock.Arguments) {
						stateJsonb := args.Get(3).(pgtype.JSONB)
						state := &CurrentMaterial{}
						err := stateJsonb.AssignTo(state)
						require.NoError(t, err)
						assert.Equal(t, "media-2", state.MediaID)
						assert.Nil(t, state.VideoState)
						assert.False(t, state.UpdatedAt.IsZero())
						assert.False(t, now.Equal(state.UpdatedAt))
					}).
					Return(nil).Once()
				lessonGroupRepo.
					On("Get", ctx, tx, database.Text("lesson-group-1"), database.Text("course-1")).
					Return(&entities.LessonGroup{
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						MediaIDs:      database.TextArray([]string{"media-1", "media-2", "media-3"}),
					}, nil).Once()
				logRepo.On(
					"IncreaseTotalTimesByLessonID",
					ctx,
					db,
					database.Text("lesson-1"),
					entities.TotalTimesUpdatingRoomState,
				).
					Return(nil).
					Once()
			},
		},
		{
			name:      "teacher command to share a material (audio) in live lesson room",
			reqUserID: "teacher-1",
			req: &bpb.ModifyLiveLessonStateRequest{
				Id: "lesson-1",
				Command: &bpb.ModifyLiveLessonStateRequest_ShareAMaterial{
					ShareAMaterial: &bpb.ModifyLiveLessonStateRequest_CurrentMaterialCommand{
						MediaId: "media-3",
						State: &bpb.ModifyLiveLessonStateRequest_CurrentMaterialCommand_AudioState{
							AudioState: &bpb.LiveLessonState_CurrentMaterial_AudioState{
								CurrentTime: durationpb.New(13 * time.Second),
								PlayerState: bpb.PlayerState_PLAYER_STATE_PAUSE,
							},
						},
					},
				},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lesson-1"),
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						RoomState: database.JSONB(`
							{
								"current_material": {
									"media_id": "media-1",
									"updated_at": "` + string(nowString) + `",
									"video_state": {
										"current_time": "23m",
										"player_state": "PLAYER_STATE_PLAYING"
									}
								}
							}`),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"teacher-1", "teacher-2"}), nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"learner-1", "learner-2", "learner-3"}), nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				lessonRepo.
					On("FindByID", ctx, tx, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lesson-1"),
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						RoomState: database.JSONB(`
						{
							"current_material": {
								"media_id": "media-1",
								"updated_at": "` + string(nowString) + `",
								"video_state": {
									"current_time": "23m",
									"player_state": "PLAYER_STATE_PLAYING"
								}
							}
						}`),
					},
						nil,
					).Once()
				lessonRoomStateRepo.
					On("UpsertCurrentMaterialState", ctx, tx, database.Text("lesson-1"), mock.Anything).
					Run(func(args mock.Arguments) {
						stateJsonb := args.Get(3).(pgtype.JSONB)
						state := &CurrentMaterial{}
						err := stateJsonb.AssignTo(state)
						require.NoError(t, err)
						assert.Equal(t, "media-3", state.MediaID)
						assert.Equal(t, 13*time.Second, state.AudioState.CurrentTime.Duration())
						assert.Equal(t, PlayerStatePause, state.AudioState.PlayerState)
						assert.False(t, state.UpdatedAt.IsZero())
						assert.False(t, now.Equal(state.UpdatedAt))
					}).
					Return(nil).Once()
				lessonGroupRepo.
					On("Get", ctx, tx, database.Text("lesson-group-1"), database.Text("course-1")).
					Return(&entities.LessonGroup{
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						MediaIDs:      database.TextArray([]string{"media-1", "media-2", "media-3"}),
					}, nil).Once()
				logRepo.On(
					"IncreaseTotalTimesByLessonID",
					ctx,
					db,
					database.Text("lesson-1"),
					entities.TotalTimesUpdatingRoomState,
				).
					Return(nil).
					Once()
			},
		},
		{
			name:      "learner command to share a material (pdf) in live lesson room",
			reqUserID: "learner-1",
			req: &bpb.ModifyLiveLessonStateRequest{
				Id: "lesson-1",
				Command: &bpb.ModifyLiveLessonStateRequest_ShareAMaterial{
					ShareAMaterial: &bpb.ModifyLiveLessonStateRequest_CurrentMaterialCommand{
						MediaId: "media-2",
					},
				},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lesson-1"),
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"teacher-1", "teacher-2"}), nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"learner-1", "learner-2", "learner-3"}), nil).Once()
				userRepo.
					On("UserGroup", ctx, db, database.Text("learner-1")).
					Return(entities.UserGroupStudent, nil).
					Once()
			},
			hasError: true,
		},
		{
			name:      "teacher command to stop sharing current material in live lesson room",
			reqUserID: "teacher-1",
			req: &bpb.ModifyLiveLessonStateRequest{
				Id:      "lesson-1",
				Command: &bpb.ModifyLiveLessonStateRequest_StopSharingMaterial{},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lesson-1"),
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						RoomState: database.JSONB(`
			{
				"current_material": {
					"media_id": "media-1",
					"updated_at": "` + string(nowString) + `",
					"video_state": {
						"current_time": "23m",
						"player_state": "PLAYER_STATE_PLAYING"
					}
				}
			}`),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"teacher-1", "teacher-2"}), nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"learner-1", "learner-2", "learner-3"}), nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				lessonRepo.
					On("FindByID", ctx, tx, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lesson-1"),
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						RoomState: database.JSONB(`
			{
				"current_material": {
					"media_id": "media-1",
					"updated_at": "` + string(nowString) + `",
					"video_state": {
						"current_time": "23m",
						"player_state": "PLAYER_STATE_PLAYING"
					}
				}
			}`),
					},
						nil,
					).Once()
				lessonRoomStateRepo.
					On("UpsertCurrentMaterialState", ctx, tx, database.Text("lesson-1"), mock.Anything).
					Run(func(args mock.Arguments) {
						stateJsonb := args.Get(3).(pgtype.JSONB)
						state := &LessonRoomState{}
						err := stateJsonb.AssignTo(state)
						require.NoError(t, err)
						assert.Nil(t, state.CurrentMaterial)
					}).
					Return(nil).Once()
				logRepo.On(
					"IncreaseTotalTimesByLessonID",
					ctx,
					db,
					database.Text("lesson-1"),
					entities.TotalTimesUpdatingRoomState,
				).
					Return(nil).
					Once()
			},
		},
		{
			name:      "learner command to stop sharing current material in live lesson room",
			reqUserID: "learner-1",
			req: &bpb.ModifyLiveLessonStateRequest{
				Id:      "lesson-1",
				Command: &bpb.ModifyLiveLessonStateRequest_StopSharingMaterial{},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lesson-1"),
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						RoomState: database.JSONB(`
			{
				"current_material": {
					"media_id": "media-1",
					"updated_at": "` + string(nowString) + `",
					"video_state": {
						"current_time": "23m",
						"player_state": "PLAYER_STATE_PLAYING"
					}
				}
			}`),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"teacher-1", "teacher-2"}), nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"learner-1", "learner-2", "learner-3"}), nil).Once()
				userRepo.
					On("UserGroup", ctx, db, database.Text("learner-1")).
					Return(entities.UserGroupStudent, nil).
					Once()
			},
			hasError: true,
		},
		{
			name:      "teacher command to fold hand all learner in live lesson room",
			reqUserID: "teacher-1",
			req: &bpb.ModifyLiveLessonStateRequest{
				Id:      "lesson-1",
				Command: &bpb.ModifyLiveLessonStateRequest_FoldHandAll{},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lesson-1"),
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"teacher-1", "teacher-2"}), nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"learner-1", "learner-2", "learner-3"}), nil).Once()
				lessonMemberRepo.
					On(
						"UpsertAllLessonMemberStateByStateType",
						ctx,
						db,
						database.Text("lesson-1"),
						database.Text(string(LearnerStateTypeHandsUp)),
						&entities.StateValue{
							BoolValue:        database.Bool(false),
							StringArrayValue: database.TextArray([]string{}),
						},
					).
					Return(nil).
					Once()
				logRepo.On(
					"IncreaseTotalTimesByLessonID",
					ctx,
					db,
					database.Text("lesson-1"),
					entities.TotalTimesUpdatingRoomState,
				).
					Return(nil).
					Once()
			},
		},
		{
			name:      "learner command to fold hand all learner in live lesson room",
			reqUserID: "learner-1",
			req: &bpb.ModifyLiveLessonStateRequest{
				Id:      "lesson-1",
				Command: &bpb.ModifyLiveLessonStateRequest_FoldHandAll{},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lesson-1"),
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"teacher-1", "teacher-2"}), nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"learner-1", "learner-2", "learner-3"}), nil).Once()
				userRepo.
					On("UserGroup", ctx, db, database.Text("learner-1")).
					Return(entities.UserGroupStudent, nil).
					Once()
			},
			hasError: true,
		},
		{
			name:      "teacher command to fold user's hand in live lesson room",
			reqUserID: "teacher-1",
			req: &bpb.ModifyLiveLessonStateRequest{
				Id: "lesson-1",
				Command: &bpb.ModifyLiveLessonStateRequest_FoldUserHand{
					FoldUserHand: "learner-2",
				},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lesson-1"),
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"teacher-1", "teacher-2"}), nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"learner-1", "learner-2", "learner-3"}), nil).Once()
				lessonMemberRepo.
					On(
						"UpsertLessonMemberState",
						ctx,
						db,
						mock.Anything,
					).
					Run(func(args mock.Arguments) {
						state := args.Get(2).(*entities.LessonMemberState)
						assert.Equal(t, "lesson-1", state.LessonID.String)
						assert.Equal(t, "learner-2", state.UserID.String)
						assert.Equal(t, string(LearnerStateTypeHandsUp), state.StateType.String)
						assert.Equal(t, false, state.BoolValue.Bool)
					}).
					Return(nil).
					Once()
				logRepo.On(
					"IncreaseTotalTimesByLessonID",
					ctx,
					db,
					database.Text("lesson-1"),
					entities.TotalTimesUpdatingRoomState,
				).
					Return(nil).
					Once()
			},
		},
		{
			name:      "learner command to fold self-hands state in live lesson room",
			reqUserID: "learner-2",
			req: &bpb.ModifyLiveLessonStateRequest{
				Id: "lesson-1",
				Command: &bpb.ModifyLiveLessonStateRequest_FoldUserHand{
					FoldUserHand: "learner-2",
				},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lesson-1"),
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"teacher-1", "teacher-2"}), nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"learner-1", "learner-2", "learner-3"}), nil).Once()
				lessonMemberRepo.
					On(
						"UpsertLessonMemberState",
						ctx,
						db,
						mock.Anything,
					).
					Run(func(args mock.Arguments) {
						state := args.Get(2).(*entities.LessonMemberState)
						assert.Equal(t, "lesson-1", state.LessonID.String)
						assert.Equal(t, "learner-2", state.UserID.String)
						assert.Equal(t, string(LearnerStateTypeHandsUp), state.StateType.String)
						assert.Equal(t, false, state.BoolValue.Bool)
					}).
					Return(nil).
					Once()
				logRepo.On(
					"IncreaseTotalTimesByLessonID",
					ctx,
					db,
					database.Text("lesson-1"),
					entities.TotalTimesUpdatingRoomState,
				).
					Return(nil).
					Once()
			},
		},
		{
			name:      "learner command to fold other learner's hands state in live lesson room",
			reqUserID: "learner-2",
			req: &bpb.ModifyLiveLessonStateRequest{
				Id: "lesson-1",
				Command: &bpb.ModifyLiveLessonStateRequest_FoldUserHand{
					FoldUserHand: "learner-3",
				},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lesson-1"),
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"teacher-1", "teacher-2"}), nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"learner-1", "learner-2", "learner-3"}), nil).Once()
			},
			hasError: true,
		},
		{
			name:      "teacher command to raise hand in live lesson room",
			reqUserID: "teacher-1",
			req: &bpb.ModifyLiveLessonStateRequest{
				Id:      "lesson-1",
				Command: &bpb.ModifyLiveLessonStateRequest_RaiseHand{},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lesson-1"),
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"teacher-1", "teacher-2"}), nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"learner-1", "learner-2", "learner-3"}), nil).Once()
				lessonMemberRepo.
					On(
						"UpsertLessonMemberState",
						ctx,
						db,
						mock.Anything,
					).
					Run(func(args mock.Arguments) {
						state := args.Get(2).(*entities.LessonMemberState)
						assert.Equal(t, "lesson-1", state.LessonID.String)
						assert.Equal(t, "teacher-1", state.UserID.String)
						assert.Equal(t, string(LearnerStateTypeHandsUp), state.StateType.String)
						assert.Equal(t, true, state.BoolValue.Bool)
					}).
					Return(fmt.Errorf("got 1 error")).
					Once()
			},
			hasError: true,
		},
		{
			name:      "learner command to raise hand in live lesson room",
			reqUserID: "learner-2",
			req: &bpb.ModifyLiveLessonStateRequest{
				Id:      "lesson-1",
				Command: &bpb.ModifyLiveLessonStateRequest_RaiseHand{},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lesson-1"),
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"teacher-1", "teacher-2"}), nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"learner-1", "learner-2", "learner-3"}), nil).Once()
				lessonMemberRepo.
					On(
						"UpsertLessonMemberState",
						ctx,
						db,
						mock.Anything,
					).
					Run(func(args mock.Arguments) {
						state := args.Get(2).(*entities.LessonMemberState)
						assert.Equal(t, "lesson-1", state.LessonID.String)
						assert.Equal(t, "learner-2", state.UserID.String)
						assert.Equal(t, string(LearnerStateTypeHandsUp), state.StateType.String)
						assert.Equal(t, true, state.BoolValue.Bool)
					}).
					Return(nil).
					Once()
				logRepo.On(
					"IncreaseTotalTimesByLessonID",
					ctx,
					db,
					database.Text("lesson-1"),
					entities.TotalTimesUpdatingRoomState,
				).
					Return(nil).
					Once()
			},
		},
		{
			name:      "learner who not belong to lesson command to raise hand in live lesson room",
			reqUserID: "learner-5",
			req: &bpb.ModifyLiveLessonStateRequest{
				Id:      "lesson-1",
				Command: &bpb.ModifyLiveLessonStateRequest_RaiseHand{},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lesson-1"),
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"teacher-1", "teacher-2"}), nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"learner-1", "learner-2", "learner-3"}), nil).Once()
				userRepo.
					On("UserGroup", ctx, db, database.Text("learner-5")).
					Return(entities.UserGroupStudent, nil).
					Once()
			},
			hasError: true,
		},
		{
			name:      "learner command to hand off in live lesson room",
			reqUserID: "learner-2",
			req: &bpb.ModifyLiveLessonStateRequest{
				Id:      "lesson-1",
				Command: &bpb.ModifyLiveLessonStateRequest_HandOff{},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lesson-1"),
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"teacher-1", "teacher-2"}), nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"learner-1", "learner-2", "learner-3"}), nil).Once()
				lessonMemberRepo.
					On(
						"UpsertLessonMemberState",
						ctx,
						db,
						mock.Anything,
					).
					Run(func(args mock.Arguments) {
						state := args.Get(2).(*entities.LessonMemberState)
						assert.Equal(t, "lesson-1", state.LessonID.String)
						assert.Equal(t, "learner-2", state.UserID.String)
						assert.Equal(t, string(LearnerStateTypeHandsUp), state.StateType.String)
						assert.Equal(t, false, state.BoolValue.Bool)
					}).
					Return(nil).
					Once()
				logRepo.On(
					"IncreaseTotalTimesByLessonID",
					ctx,
					db,
					database.Text("lesson-1"),
					entities.TotalTimesUpdatingRoomState,
				).
					Return(nil).
					Once()
			},
		},
		{
			name:      "teacher command to enables annotation in live lesson room",
			reqUserID: "teacher-1",
			req: &bpb.ModifyLiveLessonStateRequest{
				Id: "lesson-1",
				Command: &bpb.ModifyLiveLessonStateRequest_AnnotationEnable{
					AnnotationEnable: &bpb.ModifyLiveLessonStateRequest_Learners{
						Learners: []string{"learner-1", "learner-2"},
					},
				},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lesson-1"),
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						RoomState: database.JSONB(`
					{
						"current_material": {
							"media_id": "media-1",
							"updated_at": "` + string(nowString) + `"
						}
					}`),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"teacher-1", "teacher-2"}), nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"learner-1", "learner-2", "learner-3"}), nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				lessonMemberRepo.
					On(
						"UpsertMultiLessonMemberStateByState",
						ctx,
						tx,
						database.Text("lesson-1"),
						database.Text(string(LearnerStateTypeAnnotation)),
						database.TextArray([]string{"learner-1", "learner-2"}),
						&entities.StateValue{
							BoolValue:        database.Bool(true),
							StringArrayValue: database.TextArray([]string{}),
						},
					).
					Return(nil).
					Once()
				logRepo.On(
					"IncreaseTotalTimesByLessonID",
					ctx,
					db,
					database.Text("lesson-1"),
					entities.TotalTimesUpdatingRoomState,
				).
					Return(nil).
					Once()
			},
		},
		{
			name:      "learner command to enables annotation in live lesson room",
			reqUserID: "learner-1",
			req: &bpb.ModifyLiveLessonStateRequest{
				Id: "lesson-1",
				Command: &bpb.ModifyLiveLessonStateRequest_AnnotationEnable{
					AnnotationEnable: &bpb.ModifyLiveLessonStateRequest_Learners{
						Learners: []string{"learner-1", "learner-2"},
					},
				},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lesson-1"),
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						RoomState: database.JSONB(`
						{
							"current_material": {
								"media_id": "media-1",
								"updated_at": "` + string(nowString) + `"
							}
						}`),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"teacher-1", "teacher-2"}), nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"learner-1", "learner-2", "learner-3"}), nil).Once()
				userRepo.
					On("UserGroup", ctx, db, database.Text("learner-1")).
					Return(entities.UserGroupStudent, nil).
					Once()
			},
			hasError: true,
		},
		{
			name:      "teacher command to disable annotation in live lesson room",
			reqUserID: "teacher-1",
			req: &bpb.ModifyLiveLessonStateRequest{
				Id: "lesson-1",
				Command: &bpb.ModifyLiveLessonStateRequest_AnnotationDisable{
					AnnotationDisable: &bpb.ModifyLiveLessonStateRequest_Learners{
						Learners: []string{"learner-1", "learner-2"},
					},
				},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lesson-1"),
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						RoomState: database.JSONB(`
					{
						"current_material": {
							"media_id": "media-1",
							"updated_at": "` + string(nowString) + `"
						}
					}`),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"teacher-1", "teacher-2"}), nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"learner-1", "learner-2", "learner-3"}), nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				lessonMemberRepo.
					On(
						"UpsertMultiLessonMemberStateByState",
						ctx,
						tx,
						database.Text("lesson-1"),
						database.Text(string(LearnerStateTypeAnnotation)),
						database.TextArray([]string{"learner-1", "learner-2"}),
						&entities.StateValue{
							BoolValue:        database.Bool(false),
							StringArrayValue: database.TextArray([]string{}),
						},
					).
					Return(nil).
					Once()
				logRepo.On(
					"IncreaseTotalTimesByLessonID",
					ctx,
					db,
					database.Text("lesson-1"),
					entities.TotalTimesUpdatingRoomState,
				).
					Return(nil).
					Once()
			},
		},
		{
			name:      "learner command to disables annotation in live lesson room",
			reqUserID: "learner-1",
			req: &bpb.ModifyLiveLessonStateRequest{
				Id: "lesson-1",
				Command: &bpb.ModifyLiveLessonStateRequest_AnnotationDisable{
					AnnotationDisable: &bpb.ModifyLiveLessonStateRequest_Learners{
						Learners: []string{"learner-1", "learner-2"},
					},
				},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lesson-1"),
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						RoomState: database.JSONB(`
						{
							"current_material": {
								"media_id": "media-1",
								"updated_at": "` + string(nowString) + `"
							}
						}`),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"teacher-1", "teacher-2"}), nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"learner-1", "learner-2", "learner-3"}), nil).Once()
				userRepo.
					On("UserGroup", ctx, db, database.Text("learner-1")).
					Return(entities.UserGroupStudent, nil).
					Once()
			},
			hasError: true,
		},
		{
			name:      "teacher command to disable all annotation in live lesson room",
			reqUserID: "teacher-1",
			req: &bpb.ModifyLiveLessonStateRequest{
				Id:      "lesson-1",
				Command: &bpb.ModifyLiveLessonStateRequest_AnnotationDisableAll{},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lesson-1"),
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						RoomState: database.JSONB(`
					{
						"current_material": {
							"media_id": "media-1",
							"updated_at": "` + string(nowString) + `"
						}
					}`),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"teacher-1", "teacher-2"}), nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"learner-1", "learner-2", "learner-3"}), nil).Once()
				lessonMemberRepo.
					On(
						"UpsertAllLessonMemberStateByStateType",
						ctx,
						db,
						database.Text("lesson-1"),
						database.Text(string(LearnerStateTypeAnnotation)),
						&entities.StateValue{
							BoolValue:        database.Bool(false),
							StringArrayValue: database.TextArray([]string{}),
						},
					).
					Return(nil).
					Once()
				logRepo.On(
					"IncreaseTotalTimesByLessonID",
					ctx,
					db,
					database.Text("lesson-1"),
					entities.TotalTimesUpdatingRoomState,
				).
					Return(nil).
					Once()
			},
		},
		{
			name:      "learner command to disables all annotation in live lesson room",
			reqUserID: "learner-1",
			req: &bpb.ModifyLiveLessonStateRequest{
				Id:      "lesson-1",
				Command: &bpb.ModifyLiveLessonStateRequest_AnnotationDisableAll{},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lesson-1"),
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						RoomState: database.JSONB(`
						{
							"current_material": {
								"media_id": "media-1",
								"updated_at": "` + string(nowString) + `"
							}
						}`),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"teacher-1", "teacher-2"}), nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"learner-1", "learner-2", "learner-3"}), nil).Once()
				userRepo.
					On("UserGroup", ctx, db, database.Text("learner-1")).
					Return(entities.UserGroupStudent, nil).
					Once()
			},
			hasError: true,
		},
		{
			name:      "teacher command to start polling in live lesson room",
			reqUserID: "teacher-1",
			req: &bpb.ModifyLiveLessonStateRequest{
				Id: "lesson-1",
				Command: &bpb.ModifyLiveLessonStateRequest_StartPolling{
					StartPolling: &bpb.ModifyLiveLessonStateRequest_PollingOptions{
						Options: []*bpb.ModifyLiveLessonStateRequest_PollingOption{
							{
								Answer:    "A",
								IsCorrect: true,
							},
							{
								Answer:    "B",
								IsCorrect: false,
							},
							{
								Answer:    "C",
								IsCorrect: false,
							},
						},
					},
				},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lesson-1"),
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						RoomState: database.JSONB(`
					{
						"current_material": {
							"media_id": "media-1",
							"updated_at": "` + string(nowString) + `"
						}
					}`),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"teacher-1", "teacher-2"}), nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"learner-1", "learner-2", "learner-3"}), nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				lessonRepo.
					On("FindByID", ctx, tx, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID: database.Text("lesson-1"),
					}, nil).Once()
				lessonRepo.
					On("UpdateLessonRoomState", ctx, tx, database.Text("lesson-1"), mock.Anything).
					Run(func(args mock.Arguments) {
						stateJsonb := args.Get(3).(pgtype.JSONB)
						state := &LessonRoomState{}
						err := stateJsonb.AssignTo(state)
						require.NoError(t, err)
						assert.Equal(t, PollingStateStarted, state.CurrentPolling.Status)
						assert.False(t, state.CurrentPolling.CreatedAt.IsZero())
					}).
					Return(nil).Once()
				logRepo.On(
					"IncreaseTotalTimesByLessonID",
					ctx,
					db,
					database.Text("lesson-1"),
					entities.TotalTimesUpdatingRoomState,
				).
					Return(nil).
					Once()
			},
		},
		{
			name:      "teacher command to start polling in live lesson room when exists",
			reqUserID: "teacher-1",
			req: &bpb.ModifyLiveLessonStateRequest{
				Id: "lesson-1",
				Command: &bpb.ModifyLiveLessonStateRequest_StartPolling{
					StartPolling: &bpb.ModifyLiveLessonStateRequest_PollingOptions{
						Options: []*bpb.ModifyLiveLessonStateRequest_PollingOption{
							{
								Answer:    "A",
								IsCorrect: true,
							},
							{
								Answer:    "B",
								IsCorrect: false,
							},
							{
								Answer:    "C",
								IsCorrect: false,
							},
						},
					},
				},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lesson-1"),
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						RoomState: database.JSONB(`
					{
						"current_material": {
							"media_id": "media-1",
							"updated_at": "` + string(nowString) + `"
						},
					}`),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"teacher-1", "teacher-2"}), nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"learner-1", "learner-2", "learner-3"}), nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()
				lessonRepo.
					On("FindByID", ctx, tx, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID: database.Text("lesson-1"),
						RoomState: database.JSONB(`
						{
							"current_material": {
								"media_id": "media-1",
								"updated_at": "` + string(nowString) + `"
							},
							"current_polling": {
								"options": [
									{
										"answer": "A",
										"is_correct": true
									},
									{
										"answer": "B",
										"is_correct": false
									},
									{
										"answer": "C",
										"is_correct": false
									}
								],
								"status": "POLLING_STOPPED",
								"created_at": "` + string(nowString) + `"
							}
						}`),
					}, nil).Once()
			},
			hasError: true,
		},
		{
			name:      "learner command to start polling in live lesson room",
			reqUserID: "learner-1",
			req: &bpb.ModifyLiveLessonStateRequest{
				Id: "lesson-1",
				Command: &bpb.ModifyLiveLessonStateRequest_StartPolling{
					StartPolling: &bpb.ModifyLiveLessonStateRequest_PollingOptions{
						Options: []*bpb.ModifyLiveLessonStateRequest_PollingOption{
							{
								Answer:    "A",
								IsCorrect: true,
							},
							{
								Answer:    "B",
								IsCorrect: false,
							},
							{
								Answer:    "C",
								IsCorrect: false,
							},
						},
					},
				},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lesson-1"),
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						RoomState: database.JSONB(`
					{
						"current_material": {
							"media_id": "media-1",
							"updated_at": "` + string(nowString) + `"
						},
					}`),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"teacher-1", "teacher-2"}), nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"learner-1", "learner-2", "learner-3"}), nil).Once()
				userRepo.
					On("UserGroup", ctx, db, database.Text("learner-1")).
					Return(entities.UserGroupStudent, nil).
					Once()
			},
			hasError: true,
		},
		{
			name:      "teacher command to stop polling in live lesson room",
			reqUserID: "teacher-1",
			req: &bpb.ModifyLiveLessonStateRequest{
				Id:      "lesson-1",
				Command: &bpb.ModifyLiveLessonStateRequest_StopPolling{},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lesson-1"),
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						RoomState: database.JSONB(`
					{
						"current_material": {
							"media_id": "media-1",
							"updated_at": "` + string(nowString) + `"
						}
					}`),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"teacher-1", "teacher-2"}), nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"learner-1", "learner-2", "learner-3"}), nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				lessonRepo.
					On("FindByID", ctx, tx, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID: database.Text("lesson-1"),
						RoomState: database.JSONB(`
					{
						"current_polling": {
							"options": [
								{
									"answer": "A",
									"is_correct": true
								},
								{
									"answer": "B",
									"is_correct": false
								},
								{
									"answer": "C",
									"is_correct": false
								}
							],
							"status": "POLLING_STATE_STARTED",
							"created_at": "` + string(nowString) + `"
						}
					}`),
					}, nil).Once()
				lessonRepo.
					On("UpdateLessonRoomState", ctx, tx, database.Text("lesson-1"), mock.Anything).
					Run(func(args mock.Arguments) {
						stateJsonb := args.Get(3).(pgtype.JSONB)
						state := &LessonRoomState{}
						err := stateJsonb.AssignTo(state)
						require.NoError(t, err)
						assert.Equal(t, PollingStateStopped, state.CurrentPolling.Status)
						assert.False(t, state.CurrentPolling.CreatedAt.IsZero())
						assert.False(t, now.Equal(state.CurrentPolling.StoppedAt))
					}).
					Return(nil).Once()
				logRepo.On(
					"IncreaseTotalTimesByLessonID",
					ctx,
					db,
					database.Text("lesson-1"),
					entities.TotalTimesUpdatingRoomState,
				).
					Return(nil).
					Once()
			},
		},
		{
			name:      "teacher command to stop polling in live lesson room when polling stopped",
			reqUserID: "teacher-1",
			req: &bpb.ModifyLiveLessonStateRequest{
				Id:      "lesson-1",
				Command: &bpb.ModifyLiveLessonStateRequest_StopPolling{},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lesson-1"),
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						RoomState: database.JSONB(`
					{
						"current_material": {
							"media_id": "media-1",
							"updated_at": "` + string(nowString) + `"
						}
					}`),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"teacher-1", "teacher-2"}), nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"learner-1", "learner-2", "learner-3"}), nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()
				lessonRepo.
					On("FindByID", ctx, tx, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID: database.Text("lesson-1"),
						RoomState: database.JSONB(`
					{
						"current_polling": {
							"options": [
								{
									"answer": "A",
									"is_correct": true
								},
								{
									"answer": "B",
									"is_correct": false
								},
								{
									"answer": "C",
									"is_correct": false
								}
							],
							"status": "POLLING_STOPPED",
							"created_at": "` + string(nowString) + `"
						}
					}`),
					}, nil).Once()
			},
			hasError: true,
		},
		{
			name:      "teacher command to stop polling in live lesson room when nothing polling",
			reqUserID: "teacher-1",
			req: &bpb.ModifyLiveLessonStateRequest{
				Id:      "lesson-1",
				Command: &bpb.ModifyLiveLessonStateRequest_StopPolling{},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lesson-1"),
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						RoomState: database.JSONB(`
					{
						"current_material": {
							"media_id": "media-1",
							"updated_at": "` + string(nowString) + `"
						}
					}`),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"teacher-1", "teacher-2"}), nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"learner-1", "learner-2", "learner-3"}), nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()
				lessonRepo.
					On("FindByID", ctx, tx, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID: database.Text("lesson-1"),
						RoomState: database.JSONB(`
					{
					}`),
					}, nil).Once()
			},
			hasError: true,
		},
		{
			name:      "learner command to stop polling in live lesson room",
			reqUserID: "learner-1",
			req: &bpb.ModifyLiveLessonStateRequest{
				Id:      "lesson-1",
				Command: &bpb.ModifyLiveLessonStateRequest_StopPolling{},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lesson-1"),
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						RoomState: database.JSONB(`
					{
						"current_material": {
							"media_id": "media-1",
							"updated_at": "` + string(nowString) + `"
						}
					}`),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"teacher-1", "teacher-2"}), nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"learner-1", "learner-2", "learner-3"}), nil).Once()
				userRepo.
					On("UserGroup", ctx, db, database.Text("learner-1")).
					Return(entities.UserGroupStudent, nil).
					Once()
			},
			hasError: true,
		},
		{
			name:      "teacher command to end polling in live lesson room",
			reqUserID: "teacher-1",
			req: &bpb.ModifyLiveLessonStateRequest{
				Id:      "lesson-1",
				Command: &bpb.ModifyLiveLessonStateRequest_EndPolling{},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lesson-1"),
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						RoomState: database.JSONB(`
					{
						"current_material": {
							"media_id": "media-1",
							"updated_at": "` + string(nowString) + `"
						}
					}`),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"teacher-1", "teacher-2"}), nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"learner-1", "learner-2", "learner-3"}), nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				lessonRepo.
					On("FindByID", ctx, tx, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID: database.Text("lesson-1"),
						RoomState: database.JSONB(`
						{
							"current_polling": {
								"options": [
									{
										"answer": "A",
										"is_correct": true
									},
									{
										"answer": "B",
										"is_correct": false
									},
									{
										"answer": "C",
										"is_correct": false
									}
								],
								"status": "POLLING_STATE_STOPPED",
								"created_at": "` + string(nowString) + `",
								"stopped_at": "` + string(nowString) + `"
							}
						}`),
					}, nil).Once()
				lessonMemberRepo.
					On(
						"GetLessonMemberStatesWithParams",
						ctx,
						tx,
						mock.Anything,
					).
					Return(
						entities.LessonMemberStates{
							{
								LessonID:         database.Text("lesson-1"),
								UserID:           database.Text("user-1"),
								StateType:        database.Text("LEARNER_STATE_TYPE_POLLING_ANSWER"),
								CreatedAt:        database.Timestamptz(now.Add(-20 * time.Minute)),
								UpdatedAt:        database.Timestamptz(now.Add(-2 * time.Minute)),
								BoolValue:        database.Bool(false),
								StringArrayValue: database.TextArray([]string{"A"}),
								DeleteAt:         database.Timestamptz(now),
							},
						},
						nil,
					).
					Once()
				lessonPollingRepo.
					On("Create", ctx, tx, mock.Anything).
					Return(&entities.LessonPolling{
						PollID: database.Text("poll-1"),
					}, nil).Once()
				lessonRepo.
					On("UpdateLessonRoomState", ctx, tx, database.Text("lesson-1"), mock.Anything).
					Run(func(args mock.Arguments) {
						stateJsonb := args.Get(3).(pgtype.JSONB)
						state := &LessonRoomState{}
						err := stateJsonb.AssignTo(state)
						require.NoError(t, err)
						assert.Nil(t, state.CurrentPolling)
					}).
					Return(nil).Once()
				lessonMemberRepo.
					On(
						"UpsertAllLessonMemberStateByStateType",
						ctx,
						tx,
						database.Text("lesson-1"),
						database.Text(string(LearnerStateTypePollingAnswer)),
						&entities.StateValue{
							BoolValue:        database.Bool(false),
							StringArrayValue: database.TextArray([]string{}),
						},
					).
					Return(nil).
					Once()
				logRepo.On(
					"IncreaseTotalTimesByLessonID",
					ctx,
					db,
					database.Text("lesson-1"),
					entities.TotalTimesUpdatingRoomState,
				).
					Return(nil).
					Once()
			},
		},
		{
			name:      "teacher command to end polling in live lesson room when polling started",
			reqUserID: "teacher-1",
			req: &bpb.ModifyLiveLessonStateRequest{
				Id:      "lesson-1",
				Command: &bpb.ModifyLiveLessonStateRequest_EndPolling{},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lesson-1"),
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						RoomState: database.JSONB(`
					{
						"current_material": {
							"media_id": "media-1",
							"updated_at": "` + string(nowString) + `"
						}
					}`),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"teacher-1", "teacher-2"}), nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"learner-1", "learner-2", "learner-3"}), nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()
				lessonRepo.
					On("FindByID", ctx, tx, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID: database.Text("lesson-1"),
						RoomState: database.JSONB(`
						{
							"current_polling": {
								"options": [
									{
										"answer": "A",
										"is_correct": true
									},
									{
										"answer": "B",
										"is_correct": false
									},
									{
										"answer": "C",
										"is_correct": false
									}
								],
								"status": "POLLING_STATE_STARTED",
								"created_at": "` + string(nowString) + `"
							}
						}`),
					}, nil).Once()
			},
			hasError: true,
		},
		{
			name:      "teacher command to end polling in live lesson room when nothing polling",
			reqUserID: "teacher-1",
			req: &bpb.ModifyLiveLessonStateRequest{
				Id:      "lesson-1",
				Command: &bpb.ModifyLiveLessonStateRequest_EndPolling{},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lesson-1"),
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						RoomState: database.JSONB(`
					{
						"current_material": {
							"media_id": "media-1",
							"updated_at": "` + string(nowString) + `"
						}
					}`),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"teacher-1", "teacher-2"}), nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"learner-1", "learner-2", "learner-3"}), nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()
				lessonRepo.
					On("FindByID", ctx, tx, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID: database.Text("lesson-1"),
						RoomState: database.JSONB(`
						{
						}`),
					}, nil).Once()
			},
			hasError: true,
		},
		{
			name:      "learner command to end polling in live lesson room",
			reqUserID: "learner-1",
			req: &bpb.ModifyLiveLessonStateRequest{
				Id:      "lesson-1",
				Command: &bpb.ModifyLiveLessonStateRequest_EndPolling{},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lesson-1"),
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						RoomState: database.JSONB(`
					{
						"current_material": {
							"media_id": "media-1",
							"updated_at": "` + string(nowString) + `"
						}
					}`),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"teacher-1", "teacher-2"}), nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"learner-1", "learner-2", "learner-3"}), nil).Once()
				userRepo.
					On("UserGroup", ctx, db, database.Text("learner-1")).
					Return(entities.UserGroupStudent, nil).
					Once()
			},
			hasError: true,
		},
		{
			name:      "learner command to submit polling answer in live lesson room",
			reqUserID: "learner-1",
			req: &bpb.ModifyLiveLessonStateRequest{
				Id: "lesson-1",
				Command: &bpb.ModifyLiveLessonStateRequest_SubmitPollingAnswer{
					SubmitPollingAnswer: &bpb.ModifyLiveLessonStateRequest_PollingAnswer{
						StringArrayValue: []string{"A"},
					},
				},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lesson-1"),
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						RoomState: database.JSONB(`
					{
						"current_material": {
							"media_id": "media-1",
							"updated_at": "` + string(nowString) + `"
						}
					}`),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"teacher-1", "teacher-2"}), nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"learner-1", "learner-2", "learner-3"}), nil).Once()
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID: database.Text("lesson-1"),
						RoomState: database.JSONB(`
						{
							"current_polling": {
								"options": [
									{
										"answer": "A",
										"is_correct": true
									},
									{
										"answer": "B",
										"is_correct": false
									},
									{
										"answer": "C",
										"is_correct": false
									}
								],
								"status": "POLLING_STATE_STARTED",
								"created_at": "` + string(nowString) + `"
							}
						}`),
					}, nil).Once()
				lessonMemberRepo.
					On(
						"GetLessonMemberStatesWithParams",
						ctx,
						db,
						mock.Anything,
					).
					Return(
						entities.LessonMemberStates{},
						nil,
					).
					Once()
				lessonMemberRepo.
					On(
						"UpsertLessonMemberState",
						ctx,
						db,
						mock.Anything,
					).
					Run(func(args mock.Arguments) {
						state := args.Get(2).(*entities.LessonMemberState)
						assert.Equal(t, "lesson-1", state.LessonID.String)
						assert.Equal(t, "learner-1", state.UserID.String)
						assert.Equal(t, string(LearnerStateTypePollingAnswer), state.StateType.String)
					}).
					Return(nil).
					Once()
				logRepo.On(
					"IncreaseTotalTimesByLessonID",
					ctx,
					db,
					database.Text("lesson-1"),
					entities.TotalTimesUpdatingRoomState,
				).
					Return(nil).
					Once()
			},
		},
		{
			name:      "teacher try command to submit polling answer in live lesson room",
			reqUserID: "teacher-1",
			req: &bpb.ModifyLiveLessonStateRequest{
				Id: "lesson-1",
				Command: &bpb.ModifyLiveLessonStateRequest_SubmitPollingAnswer{
					SubmitPollingAnswer: &bpb.ModifyLiveLessonStateRequest_PollingAnswer{
						StringArrayValue: []string{"A"},
					},
				},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lesson-1"),
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						RoomState: database.JSONB(`
					{
						"current_material": {
							"media_id": "media-1",
							"updated_at": "` + string(nowString) + `"
						}
					}`),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"teacher-1", "teacher-2"}), nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"learner-1", "learner-2", "learner-3"}), nil).Once()
			},
			hasError: true,
		},
		{
			name:      "learner command to submit polling answer in live lesson room when polling stopped",
			reqUserID: "learner-1",
			req: &bpb.ModifyLiveLessonStateRequest{
				Id: "lesson-1",
				Command: &bpb.ModifyLiveLessonStateRequest_SubmitPollingAnswer{
					SubmitPollingAnswer: &bpb.ModifyLiveLessonStateRequest_PollingAnswer{
						StringArrayValue: []string{"A"},
					},
				},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lesson-1"),
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						RoomState: database.JSONB(`
					{
						"current_material": {
							"media_id": "media-1",
							"updated_at": "` + string(nowString) + `"
						}
					}`),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"teacher-1", "teacher-2"}), nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"learner-1", "learner-2", "learner-3"}), nil).Once()
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID: database.Text("lesson-1"),
						RoomState: database.JSONB(`
						{
							"current_polling": {
								"options": [
									{
										"answer": "A",
										"is_correct": true
									},
									{
										"answer": "B",
										"is_correct": false
									},
									{
										"answer": "C",
										"is_correct": false
									}
								],
								"status": "POLLING_STOPPED",
								"created_at": "` + string(nowString) + `"
							}
						}`),
					}, nil).Once()
			},
			hasError: true,
		},
		{
			name:      "learner command to submit polling answer in live lesson room when nothing polling",
			reqUserID: "learner-1",
			req: &bpb.ModifyLiveLessonStateRequest{
				Id: "lesson-1",
				Command: &bpb.ModifyLiveLessonStateRequest_SubmitPollingAnswer{
					SubmitPollingAnswer: &bpb.ModifyLiveLessonStateRequest_PollingAnswer{
						StringArrayValue: []string{"A"},
					},
				},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lesson-1"),
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						RoomState: database.JSONB(`
					{
						"current_material": {
							"media_id": "media-1",
							"updated_at": "` + string(nowString) + `"
						}
					}`),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"teacher-1", "teacher-2"}), nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"learner-1", "learner-2", "learner-3"}), nil).Once()
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID:  database.Text("lesson-1"),
						RoomState: database.JSONB(`{}`),
					}, nil).Once()
			},
			hasError: true,
		},
		{
			name:      "request recording successfully",
			reqUserID: "teacher-1",
			req: &bpb.ModifyLiveLessonStateRequest{
				Id: "lesson-1",
				Command: &bpb.ModifyLiveLessonStateRequest_RequestRecording{
					RequestRecording: true,
				},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lesson-1"),
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						RoomState: database.JSONB(`
					{
						"current_material": {
							"media_id": "media-1",
							"updated_at": "` + string(nowString) + `"
						}
					}`),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"teacher-1", "teacher-2"}), nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"learner-1", "learner-2", "learner-3"}), nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				lessonRepo.On("GrantRecordingPermission", ctx, tx, database.Text("lesson-1"), mock.Anything).
					Run(func(args mock.Arguments) {
						stateJsonb := args.Get(3).(pgtype.JSONB)
						state := &LessonRoomState{}
						err := stateJsonb.AssignTo(state)
						require.NoError(t, err)
						assert.True(t, state.Recording.IsRecording)
						assert.Equal(t, "teacher-1", *state.Recording.Creator)
					}).
					Return(nil).Once()
				logRepo.On(
					"IncreaseTotalTimesByLessonID",
					ctx,
					db,
					database.Text("lesson-1"),
					entities.TotalTimesUpdatingRoomState,
				).
					Return(nil).
					Once()
			},
		},
		{
			name:      "request recording failed",
			reqUserID: "teacher-1",
			req: &bpb.ModifyLiveLessonStateRequest{
				Id: "lesson-1",
				Command: &bpb.ModifyLiveLessonStateRequest_RequestRecording{
					RequestRecording: true,
				},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lesson-1"),
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						RoomState: database.JSONB(`
					{
						"current_material": {
							"media_id": "media-1",
							"updated_at": "` + string(nowString) + `"
						}
					}`),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"teacher-1", "teacher-2"}), nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"learner-1", "learner-2", "learner-3"}), nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()
				lessonRepo.On("GrantRecordingPermission", ctx, tx, database.Text("lesson-1"), mock.Anything).
					Run(func(args mock.Arguments) {
						stateJsonb := args.Get(3).(pgtype.JSONB)
						state := &LessonRoomState{}
						err := stateJsonb.AssignTo(state)
						require.NoError(t, err)
						assert.True(t, state.Recording.IsRecording)
						assert.Equal(t, "teacher-1", *state.Recording.Creator)
					}).
					Return(errors.New("error")).Once()
			},
			hasError: true,
		},
		{
			name:      "stop recording successfully",
			reqUserID: "teacher-1",
			req: &bpb.ModifyLiveLessonStateRequest{
				Id: "lesson-1",
				Command: &bpb.ModifyLiveLessonStateRequest_StopRecording{
					StopRecording: true,
				},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lesson-1"),
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						RoomState: database.JSONB(`
					{
						"current_material": {
							"media_id": "media-1",
							"updated_at": "` + string(nowString) + `"
						}
					}`),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"teacher-1", "teacher-2"}), nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"learner-1", "learner-2", "learner-3"}), nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				lessonRepo.On("StopRecording", ctx, tx, database.Text("lesson-1"), database.Text("teacher-1"), mock.Anything).
					Run(func(args mock.Arguments) {
						stateJsonb := args.Get(4).(pgtype.JSONB)
						state := &LessonRoomState{}
						err := stateJsonb.AssignTo(state)
						require.NoError(t, err)
						assert.False(t, state.Recording.IsRecording)
						assert.Nil(t, state.Recording.Creator)
					}).
					Return(nil).Once()
				logRepo.On(
					"IncreaseTotalTimesByLessonID",
					ctx,
					db,
					database.Text("lesson-1"),
					entities.TotalTimesUpdatingRoomState,
				).
					Return(nil).
					Once()
			},
		},
		{
			name:      "stop recording failed",
			reqUserID: "teacher-1",
			req: &bpb.ModifyLiveLessonStateRequest{
				Id: "lesson-1",
				Command: &bpb.ModifyLiveLessonStateRequest_StopRecording{
					StopRecording: true,
				},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lesson-1"),
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						RoomState: database.JSONB(`
					{
						"current_material": {
							"media_id": "media-1",
							"updated_at": "` + string(nowString) + `"
						}
					}`),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"teacher-1", "teacher-2"}), nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"learner-1", "learner-2", "learner-3"}), nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()
				lessonRepo.On("StopRecording", ctx, tx, database.Text("lesson-1"), database.Text("teacher-1"), mock.Anything).
					Run(func(args mock.Arguments) {
						stateJsonb := args.Get(4).(pgtype.JSONB)
						state := &LessonRoomState{}
						err := stateJsonb.AssignTo(state)
						require.NoError(t, err)
						assert.False(t, state.Recording.IsRecording)
						assert.Nil(t, state.Recording.Creator)
					}).
					Return(errors.New("error")).Once()
			},
			hasError: true,
		},
		{
			name:      "spotlight user",
			reqUserID: "teacher-1",
			req: &bpb.ModifyLiveLessonStateRequest{
				Id: "lesson-1",
				Command: &bpb.ModifyLiveLessonStateRequest_Spotlight_{
					Spotlight: &bpb.ModifyLiveLessonStateRequest_Spotlight{
						UserId:      "user-id-1",
						IsSpotlight: true,
					},
				},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lesson-1"),
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						RoomState: database.JSONB(`
				{
					"current_material": {
						"media_id": "media-1",
						"updated_at": "` + string(nowString) + `"
					}
				}`),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"teacher-1", "teacher-2"}), nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"learner-1", "learner-2", "learner-3"}), nil).Once()
				lessonRoomStateRepo.
					On("Spotlight", ctx, tx, database.Text("lesson-1"), database.Text("user-id-1")).
					Return(nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				logRepo.On(
					"IncreaseTotalTimesByLessonID",
					ctx,
					db,
					database.Text("lesson-1"),
					entities.TotalTimesUpdatingRoomState,
				).
					Return(nil).
					Once()
			},
		},
		{
			name:      "unspotlight user",
			reqUserID: "teacher-1",
			req: &bpb.ModifyLiveLessonStateRequest{
				Id: "lesson-1",
				Command: &bpb.ModifyLiveLessonStateRequest_Spotlight_{
					Spotlight: &bpb.ModifyLiveLessonStateRequest_Spotlight{
						UserId:      "user-id-1",
						IsSpotlight: false,
					},
				},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lesson-1"),
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						RoomState: database.JSONB(`
				{
					"current_material": {
						"media_id": "media-1",
						"updated_at": "` + string(nowString) + `"
					}
				}`),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"teacher-1", "teacher-2"}), nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"learner-1", "learner-2", "learner-3"}), nil).Once()
				lessonRoomStateRepo.
					On("UnSpotlight", ctx, tx, database.Text("lesson-1")).
					Return(nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				logRepo.On(
					"IncreaseTotalTimesByLessonID",
					ctx,
					db,
					database.Text("lesson-1"),
					entities.TotalTimesUpdatingRoomState,
				).
					Return(nil).
					Once()
			},
		},
		{
			name:      "user zoom whiteboard state",
			reqUserID: "teacher-1",
			req: &bpb.ModifyLiveLessonStateRequest{
				Id: "lesson-1",
				Command: &bpb.ModifyLiveLessonStateRequest_WhiteboardZoomState_{
					WhiteboardZoomState: &bpb.ModifyLiveLessonStateRequest_WhiteboardZoomState{
						PdfScaleRatio: 23.32,
						CenterX:       243.5,
						CenterY:       -432.034,
						PdfWidth:      234.43,
						PdfHeight:     -0.33424,
					},
				},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lesson-1"),
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						RoomState: database.JSONB(`
				{
					"current_material": {
						"media_id": "media-1",
						"updated_at": "` + string(nowString) + `"
					}
				}`),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"teacher-1", "teacher-2"}), nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"learner-1", "learner-2", "learner-3"}), nil).Once()
				lessonRoomStateRepo.
					On("UpsertWhiteboardZoomState", ctx, tx, "lesson-1", &domain.WhiteboardZoomState{
						PdfScaleRatio: 23.32,
						CenterX:       243.5,
						CenterY:       -432.034,
						PdfWidth:      234.43,
						PdfHeight:     -0.33424,
					}).
					Return(nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				logRepo.On(
					"IncreaseTotalTimesByLessonID",
					ctx,
					db,
					database.Text("lesson-1"),
					entities.TotalTimesUpdatingRoomState,
				).
					Return(nil).
					Once()
			},
		},
		{
			name:      "teacher command to enables chat in live lesson room",
			reqUserID: "teacher-1",
			req: &bpb.ModifyLiveLessonStateRequest{
				Id: "lesson-1",
				Command: &bpb.ModifyLiveLessonStateRequest_ChatEnable{
					ChatEnable: &bpb.ModifyLiveLessonStateRequest_Learners{
						Learners: []string{"learner-1", "learner-2"},
					},
				},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lesson-1"),
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						RoomState: database.JSONB(`
					{
						"current_material": {
							"media_id": "media-1",
							"updated_at": "` + string(nowString) + `"
						}
					}`),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"teacher-1", "teacher-2"}), nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"learner-1", "learner-2", "learner-3"}), nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				lessonMemberRepo.
					On(
						"UpsertMultiLessonMemberStateByState",
						ctx,
						tx,
						database.Text("lesson-1"),
						database.Text(string(LearnerStateTypeChat)),
						database.TextArray([]string{"learner-1", "learner-2"}),
						&entities.StateValue{
							BoolValue:        database.Bool(true),
							StringArrayValue: database.TextArray([]string{}),
						},
					).
					Return(nil).
					Once()
				logRepo.On(
					"IncreaseTotalTimesByLessonID",
					ctx,
					db,
					database.Text("lesson-1"),
					entities.TotalTimesUpdatingRoomState,
				).
					Return(nil).
					Once()
			},
		},
		{
			name:      "learner command to enables chat in live lesson room",
			reqUserID: "learner-1",
			req: &bpb.ModifyLiveLessonStateRequest{
				Id: "lesson-1",
				Command: &bpb.ModifyLiveLessonStateRequest_ChatEnable{
					ChatEnable: &bpb.ModifyLiveLessonStateRequest_Learners{
						Learners: []string{"learner-1", "learner-2"},
					},
				},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lesson-1"),
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						RoomState: database.JSONB(`
						{
							"current_material": {
								"media_id": "media-1",
								"updated_at": "` + string(nowString) + `"
							}
						}`),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"teacher-1", "teacher-2"}), nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"learner-1", "learner-2", "learner-3"}), nil).Once()
				userRepo.
					On("UserGroup", ctx, db, database.Text("learner-1")).
					Return(entities.UserGroupStudent, nil).
					Once()
			},
			hasError: true,
		},
		{
			name:      "teacher command to disable chat in live lesson room",
			reqUserID: "teacher-1",
			req: &bpb.ModifyLiveLessonStateRequest{
				Id: "lesson-1",
				Command: &bpb.ModifyLiveLessonStateRequest_ChatDisable{
					ChatDisable: &bpb.ModifyLiveLessonStateRequest_Learners{
						Learners: []string{"learner-1", "learner-2"},
					},
				},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lesson-1"),
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						RoomState: database.JSONB(`
					{
						"current_material": {
							"media_id": "media-1",
							"updated_at": "` + string(nowString) + `"
						}
					}`),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"teacher-1", "teacher-2"}), nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"learner-1", "learner-2", "learner-3"}), nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				lessonMemberRepo.
					On(
						"UpsertMultiLessonMemberStateByState",
						ctx,
						tx,
						database.Text("lesson-1"),
						database.Text(string(LearnerStateTypeChat)),
						database.TextArray([]string{"learner-1", "learner-2"}),
						&entities.StateValue{
							BoolValue:        database.Bool(false),
							StringArrayValue: database.TextArray([]string{}),
						},
					).
					Return(nil).
					Once()
				logRepo.On(
					"IncreaseTotalTimesByLessonID",
					ctx,
					db,
					database.Text("lesson-1"),
					entities.TotalTimesUpdatingRoomState,
				).
					Return(nil).
					Once()
			},
		},
		{
			name:      "learner command to disables chat in live lesson room",
			reqUserID: "learner-1",
			req: &bpb.ModifyLiveLessonStateRequest{
				Id: "lesson-1",
				Command: &bpb.ModifyLiveLessonStateRequest_ChatDisable{
					ChatDisable: &bpb.ModifyLiveLessonStateRequest_Learners{
						Learners: []string{"learner-1", "learner-2"},
					},
				},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lesson-1"),
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						RoomState: database.JSONB(`
						{
							"current_material": {
								"media_id": "media-1",
								"updated_at": "` + string(nowString) + `"
							}
						}`),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"teacher-1", "teacher-2"}), nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"learner-1", "learner-2", "learner-3"}), nil).Once()
				userRepo.
					On("UserGroup", ctx, db, database.Text("learner-1")).
					Return(entities.UserGroupStudent, nil).
					Once()
			},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			ctx = interceptors.ContextWithUserID(ctx, tc.reqUserID)
			tc.setup(ctx)

			jsm.On("PublishAsyncContext", mock.Anything, mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {}).Return("", nil)

			srv := &LessonModifierServices{
				DB:                         db,
				JSM:                        jsm,
				VirtualClassRoomLogService: &log.VirtualClassRoomLogService{DB: db, Repo: logRepo},
				LessonRepo:                 lessonRepo,
				LessonGroupRepo:            lessonGroupRepo,
				UserRepo:                   userRepo,
				LessonMemberRepo:           lessonMemberRepo,
				LessonPollingRepo:          lessonPollingRepo,
				LessonRoomStateRepo:        lessonRoomStateRepo,
			}
			_, err := srv.ModifyLiveLessonState(ctx, tc.req)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, db, tx, lessonRepo, lessonGroupRepo, userRepo, lessonMemberRepo, lessonPollingRepo, logRepo)
		})
	}
}

func TestResetAllLiveLessonStatesInternal(t *testing.T) {
	lessonRepo := &mock_repositories.MockLessonRepo{}
	userRepo := &mock_repositories.MockUserRepo{}
	lessonMemberRepo := &mock_repositories.MockLessonMemberRepo{}
	lessonRoomStateRepo := &mock_repositories_lessonmgmt.MockLessonRoomStateRepo{}
	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	now := time.Now().UTC()
	nowString, err := now.MarshalText()
	require.NoError(t, err)

	tcs := []struct {
		name      string
		reqUserID string
		lessonID  string
		setup     func(ctx context.Context)
		hasError  bool
	}{
		{
			name:      "teacher reset all live lesson's state",
			reqUserID: "teacher-1",
			lessonID:  "lesson-1",
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lesson-1"),
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						RoomState: database.JSONB(`
			{
				"current_material": {
					"media_id": "media-1",
					"updated_at": "` + string(nowString) + `",
					"video_state": {
						"current_time": "23m",
						"player_state": "PLAYER_STATE_PLAYING"
					}
				}
			}`),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"teacher-1", "teacher-2"}), nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"learner-1", "learner-2", "learner-3"}), nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				lessonRepo.
					On("FindByID", ctx, tx, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lesson-1"),
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						RoomState: database.JSONB(`
					{
						"current_material": {
							"media_id": "media-1",
							"updated_at": "` + string(nowString) + `",
							"video_state": {
								"current_time": "23m",
								"player_state": "PLAYER_STATE_PLAYING"
							}
						},
						"current_polling": {
							"options": [
								{
									"answer": "A",
									"is_correct": true
								},
								{
									"answer": "B",
									"is_correct": false
								},
								{
									"answer": "C",
									"is_correct": false
								}
							],
							"status": "POLLING_STATE_STARTED",
							"created_at": "` + string(nowString) + `"
						}
					}`),
					},
						nil,
					).Once()
				lessonRoomStateRepo.
					On("UpsertCurrentMaterialState", ctx, tx, database.Text("lesson-1"), mock.Anything).
					Run(func(args mock.Arguments) {
						stateJsonb := args.Get(3).(pgtype.JSONB)
						state := &LessonRoomState{}
						err := stateJsonb.AssignTo(state)
						require.NoError(t, err)
						assert.Nil(t, state.CurrentMaterial)
					}).
					Return(nil).Once()
				lessonMemberRepo.
					On(
						"UpsertAllLessonMemberStateByStateType",
						ctx,
						tx,
						database.Text("lesson-1"),
						database.Text(string(LearnerStateTypeHandsUp)),
						&entities.StateValue{
							BoolValue:        database.Bool(false),
							StringArrayValue: database.TextArray([]string{}),
						},
					).
					Return(nil).
					Once()
				lessonMemberRepo.
					On(
						"UpsertAllLessonMemberStateByStateType",
						ctx,
						tx,
						database.Text("lesson-1"),
						database.Text(string(LearnerStateTypeAnnotation)),
						&entities.StateValue{
							BoolValue:        database.Bool(false),
							StringArrayValue: database.TextArray([]string{}),
						},
					).
					Return(nil).
					Once()
				lessonRepo.
					On("FindByID", ctx, tx, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lesson-1"),
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						RoomState: database.JSONB(`
							{
								"current_material": {
									"media_id": "media-1",
									"updated_at": "` + string(nowString) + `",
									"video_state": {
										"current_time": "23m",
										"player_state": "PLAYER_STATE_PLAYING"
									}
								},
								"current_polling": {
									"options": [
										{
											"answer": "A",
											"is_correct": true
										},
										{
											"answer": "B",
											"is_correct": false
										},
										{
											"answer": "C",
											"is_correct": false
										}
									],
									"status": "POLLING_STATE_STARTED",
									"created_at": "` + string(nowString) + `"
								}
							}`),
					},
						nil,
					).Once()
				lessonMemberRepo.
					On(
						"UpsertAllLessonMemberStateByStateType",
						ctx,
						tx,
						database.Text("lesson-1"),
						database.Text(string(LearnerStateTypePollingAnswer)),
						&entities.StateValue{
							BoolValue:        database.Bool(false),
							StringArrayValue: database.TextArray([]string{}),
						},
					).
					Return(nil).
					Once()
				lessonRepo.
					On("UpdateLessonRoomState", ctx, tx, database.Text("lesson-1"), mock.Anything).
					Run(func(args mock.Arguments) {
						stateJsonb := args.Get(3).(pgtype.JSONB)
						state := &LessonRoomState{}
						err := stateJsonb.AssignTo(state)
						require.NoError(t, err)
						assert.Nil(t, state.CurrentPolling)
					}).
					Return(nil).Once()
				lessonMemberRepo.
					On(
						"UpsertAllLessonMemberStateByStateType",
						ctx,
						tx,
						database.Text("lesson-1"),
						database.Text(string(LearnerStateTypeChat)),
						&entities.StateValue{
							BoolValue:        database.Bool(true),
							StringArrayValue: database.TextArray([]string{}),
						},
					).
					Return(nil).
					Once()
				// reset recording
				lessonRepo.
					On("FindByID", ctx, tx, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID: database.Text("lesson-1"),
						RoomState: database.JSONB(`
			{
				"recording": {
					"is_recording": true,
					"creator": "user-id-1"
				}
			}`),
					}, nil).Once()
				lessonRepo.
					On("UpdateLessonRoomState", ctx, tx, database.Text("lesson-1"), mock.Anything).
					Run(func(args mock.Arguments) {
						stateJsonb := args.Get(3).(pgtype.JSONB)
						state := &LessonRoomState{}
						err := stateJsonb.AssignTo(state)
						require.NoError(t, err)
						assert.False(t, state.Recording.IsRecording)
						assert.Nil(t, state.Recording.Creator)
					}).
					Return(nil).Once()
				lessonRoomStateRepo.On("UnSpotlight", ctx, tx, database.Text("lesson-1")).Return(nil).Once()
				lessonRoomStateRepo.On("UpsertWhiteboardZoomState", ctx, tx, "lesson-1", new(domain.WhiteboardZoomState).SetDefault()).Return(nil).Once()
			},
		},
		{
			name:      "learner reset all live lesson's state",
			reqUserID: "learner-1",
			lessonID:  "lesson-1",
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lesson-1"),
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"teacher-1", "teacher-2"}), nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(database.TextArray([]string{"learner-1", "learner-2", "learner-3"}), nil).Once()
				userRepo.
					On("UserGroup", ctx, db, database.Text("learner-1")).
					Return(entities.UserGroupStudent, nil).
					Once()
			},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			ctx = interceptors.ContextWithUserID(ctx, tc.reqUserID)
			tc.setup(ctx)

			srv := &LessonModifierServices{
				DB:                  db,
				LessonRepo:          lessonRepo,
				UserRepo:            userRepo,
				LessonMemberRepo:    lessonMemberRepo,
				LessonRoomStateRepo: lessonRoomStateRepo,
			}
			err := srv.ResetAllLiveLessonStatesInternal(ctx, tc.lessonID)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, db, tx, lessonRepo, userRepo, lessonMemberRepo)
		})
	}
}

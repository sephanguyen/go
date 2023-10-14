package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type LessonRepoMock struct {
	increaseNumberOfStreaming                func(ctx context.Context, db database.QueryExecer, lessonID, learnerID pgtype.Text, MaximumLearnerStreamings int) error
	decreaseNumberOfStreaming                func(ctx context.Context, db database.QueryExecer, lessonID, learnerID pgtype.Text) error
	getStreamingLeaners                      func(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, queryEnhancers ...repositories.QueryEnhancer) ([]string, error)
	create                                   func(ctx context.Context, db database.Ext, lesson *entities.Lesson) (*entities.Lesson, error)
	findByID                                 func(ctx context.Context, db database.Ext, id pgtype.Text) (*entities.Lesson, error)
	findEarliestAndLatestTimeLessonByCourses func(ctx context.Context, db database.Ext, courseIDs pgtype.TextArray) (*entities.CourseAvailableRanges, error)
	getCourseIDsOfLesson                     func(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text) (pgtype.TextArray, error)
	update                                   func(ctx context.Context, db database.Ext, lesson *entities.Lesson) error
	upsertLessonTeachers                     func(ctx context.Context, db database.Ext, lessonID pgtype.Text, teacherIDs pgtype.TextArray) error
	upsertLessonMembers                      func(ctx context.Context, db database.Ext, lessonID pgtype.Text, userID pgtype.TextArray) error
	upsertLessonCourses                      func(ctx context.Context, db database.Ext, lessonID pgtype.Text, courseIDs pgtype.TextArray) error
	delete                                   func(ctx context.Context, db database.QueryExecer, lessonIDs pgtype.TextArray) error
	deleteLessonMembers                      func(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text) error
	deleteLessonTeachers                     func(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text) error
	deleteLessonCourses                      func(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text) error

	retrieve                 func(ctx context.Context, db database.QueryExecer, args *repositories.ListLessonArgs) ([]*entities.Lesson, uint32, string, uint32, error)
	findPreviousPageOffset   func(ctx context.Context, db database.QueryExecer, args *repositories.ListLessonArgs) (string, error)
	countLesson              func(ctx context.Context, db database.QueryExecer, args *repositories.ListLessonArgs) (int64, error)
	getTeacherIDsOfLesson    func(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text) (pgtype.TextArray, error)
	getLearnerIDsOfLesson    func(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text) (pgtype.TextArray, error)
	updateLessonRoomState    func(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, state pgtype.JSONB) error
	grantRecordingPermission func(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, recordingState pgtype.JSONB) error
	stopRecording            func(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, creator pgtype.Text, recordingState pgtype.JSONB) error

	//Course's find lesson endpoint
	findLessonWithTime             func(ctx context.Context, db database.QueryExecer, courseIDs *pgtype.TextArray, startDate *pgtype.Timestamptz, endDate *pgtype.Timestamptz, limit int32, page int32, schedulingStatus pgtype.Text) ([]*repositories.LessonWithTime, pgtype.Int8, error)
	findLessonJoined               func(ctx context.Context, db database.QueryExecer, userID pgtype.Text, courseIDs *pgtype.TextArray, startDate *pgtype.Timestamptz, endDate *pgtype.Timestamptz, limit int32, page int32, schedulingStatus pgtype.Text) ([]*repositories.LessonWithTime, pgtype.Int8, error)
	findLessonWithTimeAndLocations func(ctx context.Context, db database.QueryExecer, courseIDs *pgtype.TextArray, startDate *pgtype.Timestamptz, endDate *pgtype.Timestamptz, locationIDs *pgtype.TextArray, limit int32, page int32, schedulingStatus pgtype.Text) ([]*repositories.LessonWithTime, pgtype.Int8, error)
	findLessonJoinedWithLocations  func(ctx context.Context, db database.QueryExecer, userID pgtype.Text, courseIDs *pgtype.TextArray, startDate *pgtype.Timestamptz, endDate *pgtype.Timestamptz, locationIDs *pgtype.TextArray, limit int32, page int32, schedulingStatus pgtype.Text) ([]*repositories.LessonWithTime, pgtype.Int8, error)
}

func (l LessonRepoMock) IncreaseNumberOfStreaming(ctx context.Context, db database.QueryExecer, lessonID, learnerID pgtype.Text, MaximumLearnerStreamings int) error {
	return l.increaseNumberOfStreaming(ctx, db, lessonID, learnerID, MaximumLearnerStreamings)
}

func (l LessonRepoMock) DecreaseNumberOfStreaming(ctx context.Context, db database.QueryExecer, lessonID, learnerID pgtype.Text) error {
	return l.decreaseNumberOfStreaming(ctx, db, lessonID, learnerID)
}

func (l LessonRepoMock) GetStreamingLearners(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, queryEnhancers ...repositories.QueryEnhancer) ([]string, error) {
	return l.getStreamingLeaners(ctx, db, lessonID, queryEnhancers...)
}

func (l LessonRepoMock) Create(ctx context.Context, db database.Ext, lesson *entities.Lesson) (*entities.Lesson, error) {
	return l.create(ctx, db, lesson)
}

func (l LessonRepoMock) FindByID(ctx context.Context, db database.Ext, id pgtype.Text) (*entities.Lesson, error) {
	return l.findByID(ctx, db, id)
}

func (l LessonRepoMock) FindEarliestAndLatestTimeLessonByCourses(ctx context.Context, db database.Ext, courseIDs pgtype.TextArray) (*entities.CourseAvailableRanges, error) {
	return l.findEarliestAndLatestTimeLessonByCourses(ctx, db, courseIDs)
}

func (l LessonRepoMock) GetCourseIDsOfLesson(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text) (pgtype.TextArray, error) {
	return l.getCourseIDsOfLesson(ctx, db, lessonID)
}

func (l LessonRepoMock) Update(ctx context.Context, db database.Ext, lesson *entities.Lesson) error {
	return l.update(ctx, db, lesson)
}

func (l LessonRepoMock) UpsertLessonTeachers(ctx context.Context, db database.Ext, lessonID pgtype.Text, teacherIDs pgtype.TextArray) error {
	return l.upsertLessonTeachers(ctx, db, lessonID, teacherIDs)
}

func (l LessonRepoMock) UpsertLessonMembers(ctx context.Context, db database.Ext, lessonID pgtype.Text, userID pgtype.TextArray) error {
	return l.upsertLessonMembers(ctx, db, lessonID, userID)
}

func (l LessonRepoMock) UpsertLessonCourses(ctx context.Context, db database.Ext, lessonID pgtype.Text, courseIDs pgtype.TextArray) error {
	return l.upsertLessonCourses(ctx, db, lessonID, courseIDs)
}

func (l LessonRepoMock) Delete(ctx context.Context, db database.QueryExecer, lessonIDs pgtype.TextArray) error {
	return l.delete(ctx, db, lessonIDs)
}

func (l LessonRepoMock) DeleteLessonMembers(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text) error {
	return l.deleteLessonMembers(ctx, db, lessonID)
}

func (l LessonRepoMock) DeleteLessonTeachers(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text) error {
	return l.deleteLessonTeachers(ctx, db, lessonID)
}

func (l LessonRepoMock) DeleteLessonCourses(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text) error {
	return l.deleteLessonCourses(ctx, db, lessonID)
}

func (l LessonRepoMock) Retrieve(ctx context.Context, db database.QueryExecer, args *repositories.ListLessonArgs) ([]*entities.Lesson, uint32, string, uint32, error) {
	return l.retrieve(ctx, db, args)
}

func (l LessonRepoMock) FindPreviousPageOffset(ctx context.Context, db database.QueryExecer, args *repositories.ListLessonArgs) (string, error) {
	return l.findPreviousPageOffset(ctx, db, args)
}

func (l LessonRepoMock) CountLesson(ctx context.Context, db database.QueryExecer, args *repositories.ListLessonArgs) (int64, error) {
	return l.countLesson(ctx, db, args)
}

func (l LessonRepoMock) GetTeacherIDsOfLesson(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text) (pgtype.TextArray, error) {
	return l.getTeacherIDsOfLesson(ctx, db, lessonID)
}

func (l LessonRepoMock) GetLearnerIDsOfLesson(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text) (pgtype.TextArray, error) {
	return l.getLearnerIDsOfLesson(ctx, db, lessonID)
}

func (l LessonRepoMock) UpdateLessonRoomState(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, state pgtype.JSONB) error {
	return l.updateLessonRoomState(ctx, db, lessonID, state)
}

func (l LessonRepoMock) GrantRecordingPermission(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, recordingState pgtype.JSONB) error {
	return l.grantRecordingPermission(ctx, db, lessonID, recordingState)
}

func (l LessonRepoMock) StopRecording(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, creator pgtype.Text, recordingState pgtype.JSONB) error {
	return l.stopRecording(ctx, db, lessonID, creator, recordingState)
}

func (l LessonRepoMock) FindLessonWithTime(ctx context.Context, db database.QueryExecer, courseIDs *pgtype.TextArray, startDate *pgtype.Timestamptz, endDate *pgtype.Timestamptz, limit int32, page int32, schedulingStatus pgtype.Text) ([]*repositories.LessonWithTime, pgtype.Int8, error) {
	return l.findLessonWithTime(ctx, db, courseIDs, startDate, endDate, limit, page, schedulingStatus)
}
func (l LessonRepoMock) FindLessonWithTimeAndLocations(ctx context.Context, db database.QueryExecer, courseIDs *pgtype.TextArray, startDate *pgtype.Timestamptz, endDate *pgtype.Timestamptz, locationIDs *pgtype.TextArray, limit int32, page int32, schedulingStatus pgtype.Text) ([]*repositories.LessonWithTime, pgtype.Int8, error) {
	return l.findLessonWithTimeAndLocations(ctx, db, courseIDs, startDate, endDate, locationIDs, limit, page, schedulingStatus)
}
func (l LessonRepoMock) FindLessonJoined(ctx context.Context, db database.QueryExecer, userID pgtype.Text, courseIDs *pgtype.TextArray, startDate *pgtype.Timestamptz, endDate *pgtype.Timestamptz, limit int32, page int32, schedulingStatus pgtype.Text) ([]*repositories.LessonWithTime, pgtype.Int8, error) {
	return l.findLessonJoined(ctx, db, userID, courseIDs, startDate, endDate, limit, page, schedulingStatus)
}
func (l LessonRepoMock) FindLessonJoinedWithLocations(ctx context.Context, db database.QueryExecer, userID pgtype.Text, courseIDs *pgtype.TextArray, startDate *pgtype.Timestamptz, endDate *pgtype.Timestamptz, locationIDs *pgtype.TextArray, limit int32, page int32, schedulingStatus pgtype.Text) ([]*repositories.LessonWithTime, pgtype.Int8, error) {
	return l.findLessonJoinedWithLocations(ctx, db, userID, courseIDs, startDate, endDate, locationIDs, limit, page, schedulingStatus)
}
func TestInsertLessonMethod(t *testing.T) {
	t.Parallel()
	classID := idutil.ULIDNow()
	tcs := []struct {
		name     string
		lesson   *entities.Lesson
		mockRepo LessonRepoMock
		hasError bool
	}{
		{
			name: "create a live lesson successfully",
			lesson: &entities.Lesson{
				Name:                 database.Text("lesson 1"),
				StartTime:            database.Timestamptz(time.Date(2020, 2, 3, 4, 5, 6, 7, time.UTC)),
				EndTime:              database.Timestamptz(time.Date(2021, 2, 3, 4, 5, 6, 7, time.UTC)),
				LessonGroupID:        database.Text("lesson group 1"),
				LessonType:           database.Text(string(entities.LessonTypeOnline)),
				TeachingMedium:       database.Text(string(entities.LessonTeachingMediumOnline)),
				Status:               database.Text(string(entities.LessonStatusDraft)),
				StreamLearnerCounter: database.Int4(2),
				TeacherIDs:           entities.TeacherIDs{TeacherIDs: database.TextArray([]string{"id-2", "id-5"})},
				CourseIDs:            entities.CourseIDs{CourseIDs: database.TextArray([]string{"id-1", "id-3"})},
				LearnerIDs: entities.LearnerIDs{
					LearnerIDs: database.TextArray([]string{"id-4", "id-6"}),
				},
				TeachingMethod: database.Text(string(entities.LessonTeachingMethodIndividual)),
				ClassID:        database.Text(classID),
			},
			mockRepo: LessonRepoMock{
				create: func(ctx context.Context, db database.Ext, lesson *entities.Lesson) (*entities.Lesson, error) {
					assert.Equal(t, "lesson 1", lesson.Name.String)
					assert.Equal(t, time.Date(2020, 2, 3, 4, 5, 6, 7, time.UTC).UTC(), lesson.StartTime.Time.UTC())
					assert.Equal(t, time.Date(2021, 2, 3, 4, 5, 6, 7, time.UTC).UTC(), lesson.EndTime.Time.UTC())
					assert.Equal(t, "lesson group 1", lesson.LessonGroupID.String)
					assert.EqualValues(t, entities.LessonTypeOnline, lesson.LessonType.String)
					assert.EqualValues(t, entities.LessonTeachingMediumOnline, lesson.TeachingMedium.String)
					assert.EqualValues(t, entities.LessonStatusDraft, lesson.Status.String)
					assert.EqualValues(t, 2, lesson.StreamLearnerCounter.Int)
					assert.Equal(t, []string{"id-2", "id-5"}, database.FromTextArray(lesson.TeacherIDs.TeacherIDs))
					assert.Equal(t, []string{"id-1", "id-3"}, database.FromTextArray(lesson.CourseIDs.CourseIDs))
					assert.Equal(t, []string{"id-4", "id-6"}, database.FromTextArray(lesson.LearnerIDs.LearnerIDs))
					assert.EqualValues(t, entities.LessonTeachingMethodIndividual, lesson.TeachingMethod.String)
					assert.EqualValues(t, classID, lesson.ClassID.String)
					return nil, nil
				},
				upsertLessonCourses: func(ctx context.Context, db database.Ext, lessonID pgtype.Text, courseIDs pgtype.TextArray) error {
					assert.Equal(t, []string{"id-1", "id-3"}, database.FromTextArray(courseIDs))
					return nil
				},
				upsertLessonTeachers: func(ctx context.Context, db database.Ext, lessonID pgtype.Text, teacherIDs pgtype.TextArray) error {
					assert.Equal(t, []string{"id-2", "id-5"}, database.FromTextArray(teacherIDs))
					return nil
				},
				upsertLessonMembers: func(ctx context.Context, db database.Ext, lessonID pgtype.Text, userIDs pgtype.TextArray) error {
					assert.Equal(t, []string{"id-4", "id-6"}, database.FromTextArray(userIDs))
					return nil
				},
			},
		},
		{
			name: "create a live lesson missing status, type and LessonGroupID successfully",
			lesson: &entities.Lesson{
				Name:                 database.Text("lesson 1"),
				StartTime:            database.Timestamptz(time.Date(2020, 2, 3, 4, 5, 6, 7, time.UTC)),
				EndTime:              database.Timestamptz(time.Date(2021, 2, 3, 4, 5, 6, 7, time.UTC)),
				StreamLearnerCounter: database.Int4(2),
				TeacherIDs:           entities.TeacherIDs{TeacherIDs: database.TextArray([]string{"id-2", "id-5"})},
				CourseIDs:            entities.CourseIDs{CourseIDs: database.TextArray([]string{"id-1", "id-3"})},
				LearnerIDs: entities.LearnerIDs{
					LearnerIDs: database.TextArray([]string{"id-4", "id-6"}),
				},
			},
			mockRepo: LessonRepoMock{
				create: func(ctx context.Context, db database.Ext, lesson *entities.Lesson) (*entities.Lesson, error) {
					assert.Equal(t, "lesson 1", lesson.Name.String)
					assert.Equal(t, time.Date(2020, 2, 3, 4, 5, 6, 7, time.UTC).UTC(), lesson.StartTime.Time.UTC())
					assert.Equal(t, time.Date(2021, 2, 3, 4, 5, 6, 7, time.UTC).UTC(), lesson.EndTime.Time.UTC())
					assert.Empty(t, lesson.LessonGroupID.String)
					assert.EqualValues(t, 2, lesson.StreamLearnerCounter.Int)
					assert.Equal(t, []string{"id-2", "id-5"}, database.FromTextArray(lesson.TeacherIDs.TeacherIDs))
					assert.Equal(t, []string{"id-1", "id-3"}, database.FromTextArray(lesson.CourseIDs.CourseIDs))
					assert.Equal(t, []string{"id-4", "id-6"}, database.FromTextArray(lesson.LearnerIDs.LearnerIDs))

					return nil, nil
				},
				upsertLessonCourses: func(ctx context.Context, db database.Ext, lessonID pgtype.Text, courseIDs pgtype.TextArray) error {
					assert.Equal(t, []string{"id-1", "id-3"}, database.FromTextArray(courseIDs))
					return nil
				},
				upsertLessonTeachers: func(ctx context.Context, db database.Ext, lessonID pgtype.Text, teacherIDs pgtype.TextArray) error {
					assert.Equal(t, []string{"id-2", "id-5"}, database.FromTextArray(teacherIDs))
					return nil
				},
				upsertLessonMembers: func(ctx context.Context, db database.Ext, lessonID pgtype.Text, userIDs pgtype.TextArray) error {
					assert.Equal(t, []string{"id-4", "id-6"}, database.FromTextArray(userIDs))
					return nil
				},
			},
		},
		{
			name: "create a live lesson failed",
			lesson: &entities.Lesson{
				Name:                 database.Text("lesson 1"),
				StartTime:            database.Timestamptz(time.Date(2020, 2, 3, 4, 5, 6, 7, time.UTC)),
				EndTime:              database.Timestamptz(time.Date(2021, 2, 3, 4, 5, 6, 7, time.UTC)),
				LessonGroupID:        database.Text("lesson group 1"),
				StreamLearnerCounter: database.Int4(2),
				CourseIDs:            entities.CourseIDs{CourseIDs: database.TextArray([]string{"id-1", "id-3"})},
				LearnerIDs: entities.LearnerIDs{
					LearnerIDs: database.TextArray([]string{"id-4", "id-6"}),
				},
			},
			mockRepo: LessonRepoMock{
				create: func(ctx context.Context, db database.Ext, lesson *entities.Lesson) (*entities.Lesson, error) {
					return nil, fmt.Errorf("create lesson failed")
				},
				upsertLessonCourses: func(ctx context.Context, db database.Ext, lessonID pgtype.Text, courseIDs pgtype.TextArray) error {
					assert.Fail(t, "expected upsert lesson courses method repo not be called")
					return nil
				},
				upsertLessonTeachers: func(ctx context.Context, db database.Ext, lessonID pgtype.Text, teacherIDs pgtype.TextArray) error {
					assert.Fail(t, "expected upsert lesson teachers method repo not be called")
					return nil
				},
				upsertLessonMembers: func(ctx context.Context, db database.Ext, lessonID pgtype.Text, userIDs pgtype.TextArray) error {
					assert.Fail(t, "expected upsert lesson members method repo not be called")
					return nil
				},
			},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			builder := NewLessonBuilder(
				tc.mockRepo,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
			)
			builder.Lesson = tc.lesson
			err := builder.insertLesson(ctx, nil)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

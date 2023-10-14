package courses

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/services/courses/repo"
	topicsRepo "github.com/manabie-com/backend/internal/bob/services/topics/repo"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/multierr"
)

func TestCreatePresetStudyPlanWeekliesForLesson(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name                      string
		lesson                    *entities.Lesson
		courseRepo                repo.CourseRepo
		presetStudyPlanRepo       repo.PresetStudyPlanRepo
		presetStudyPlanWeeklyRepo repo.PresetStudyPlanWeeklyRepo
		topicRepo                 topicsRepo.TopicRepo
		hasError                  bool
	}{
		{
			name: "create preset study plan weekly for lesson successfully",
			lesson: &entities.Lesson{
				LessonID: database.Text("lesson-id-1"),
				Name:     database.Text("Introduction 1"),
				CourseIDs: entities.CourseIDs{
					CourseIDs: database.TextArray([]string{"course-id-1", "course-id-2"}),
				},
				StartTime: database.Timestamptz(time.Date(2021, 2, 3, 4, 5, 6, 7, time.Local)),
				EndTime:   database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)),
			},
			courseRepo: &repo.CourseRepoMock{
				FindByIDsMock: func(ctx context.Context, db database.Ext, courseIDs pgtype.TextArray) (map[pgtype.Text]*entities.Course, error) {
					assert.ElementsMatch(t, []string{"course-id-1", "course-id-2"}, database.FromTextArray(courseIDs))
					return map[pgtype.Text]*entities.Course{
						database.Text("course-id-1"): {
							ID:                database.Text("course-id-1"),
							PresetStudyPlanID: database.Text("preset-study-plan-id-1"),
							Country:           database.Text("vietnam"),
							Grade:             database.Int2(2),
							Subject:           database.Text("math"),
							SchoolID:          database.Int4(1),
						},
						database.Text("course-id-2"): {
							ID:                database.Text("course-id-2"),
							PresetStudyPlanID: database.Text("preset-study-plan-id-2"),
							Country:           database.Text("vietnam"),
							Grade:             database.Int2(5),
							Subject:           database.Text("physics"),
							SchoolID:          database.Int4(1),
						},
					}, nil
				},
			},
			topicRepo: &topicsRepo.TopicRepoMock{
				CreateMock: func(ctx context.Context, db database.Ext, plans []*entities.Topic) error {
					assert.Len(t, plans, 2)
					assert.NotEmpty(t, plans[0].ID.String)
					assert.NotZero(t, plans[0].PublishedAt.Time)
					assert.NotEmpty(t, plans[1].ID.String)
					assert.NotZero(t, plans[1].PublishedAt.Time)

					plansBySubject := make(map[string]*entities.Topic)
					for i := range plans {
						plansBySubject[plans[i].Subject.String] = plans[i]
					}

					e := &entities.Topic{}
					database.AllNullEntity(e)
					err := multierr.Combine(
						e.ID.Set(plansBySubject["math"].ID),
						e.Name.Set("Introduction 1"),
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
						e.Name.Set("Introduction 1"),
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

					return nil
				},
			},
			presetStudyPlanWeeklyRepo: repo.PresetStudyPlanWeeklyRepoMock{
				CreateMock: func(ctx context.Context, db database.Ext, plans []*entities.PresetStudyPlanWeekly) error {
					assert.Len(t, plans, 2)

					assert.NotEmpty(t, plans[0].ID.String)
					assert.NotEmpty(t, plans[0].TopicID.String)
					assert.Equal(t, int16(0), plans[0].Week.Int)
					assert.Equal(t, "lesson-id-1", plans[0].LessonID.String)
					assert.Equal(t, time.Date(2021, 2, 3, 4, 5, 6, 7, time.Local), plans[0].StartDate.Time)
					assert.Equal(t, time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local), plans[0].EndDate.Time)

					assert.NotEmpty(t, plans[1].ID.String)
					assert.NotEmpty(t, plans[1].TopicID.String)
					assert.Equal(t, int16(0), plans[1].Week.Int)
					assert.Equal(t, "lesson-id-1", plans[1].LessonID.String)
					assert.Equal(t, time.Date(2021, 2, 3, 4, 5, 6, 7, time.Local), plans[1].StartDate.Time)
					assert.Equal(t, time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local), plans[1].EndDate.Time)

					actualPresetStudyPlanID := []string{plans[0].PresetStudyPlanID.String, plans[1].PresetStudyPlanID.String}
					expectedPresetStudyPlanID := []string{"preset-study-plan-id-1", "preset-study-plan-id-2"}
					assert.ElementsMatch(t, expectedPresetStudyPlanID, actualPresetStudyPlanID)
					assert.NotEqual(t, plans[0].TopicID.String, plans[1].TopicID.String)

					return nil
				},
			},
		},
		{
			name: "create preset study plan weekly with course have not preset study plan id",
			lesson: &entities.Lesson{
				LessonID: database.Text("lesson-id-1"),
				Name:     database.Text("Introduction 1"),
				CourseIDs: entities.CourseIDs{
					CourseIDs: database.TextArray([]string{"course-id-1", "course-id-2"}),
				},
				StartTime: database.Timestamptz(time.Date(2021, 2, 3, 4, 5, 6, 7, time.Local)),
				EndTime:   database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)),
			},
			courseRepo: &repo.CourseRepoMock{
				FindByIDsMock: func(ctx context.Context, db database.Ext, courseIDs pgtype.TextArray) (map[pgtype.Text]*entities.Course, error) {
					assert.ElementsMatch(t, []string{"course-id-1", "course-id-2"}, database.FromTextArray(courseIDs))
					return map[pgtype.Text]*entities.Course{
						database.Text("course-id-1"): {
							ID:                database.Text("course-id-1"),
							PresetStudyPlanID: database.Text("preset-study-plan-id-1"),
							Country:           database.Text("vietnam"),
							Grade:             database.Int2(2),
							Subject:           database.Text("math"),
							SchoolID:          database.Int4(1),
						},
						database.Text("course-id-2"): {
							ID:       database.Text("course-id-2"),
							Country:  database.Text("vietnam"),
							Grade:    database.Int2(5),
							Subject:  database.Text("physics"),
							SchoolID: database.Int4(1),
						},
					}, nil
				},
			},
			hasError: true,
		},
	}

	for i := range tcs {
		t.Run(tcs[i].name, func(t *testing.T) {
			builder := NewPresetStudyPlanWeeklyBuilder(
				nil,
				tcs[i].courseRepo,
				tcs[i].topicRepo,
				tcs[i].presetStudyPlanRepo,
				tcs[i].presetStudyPlanWeeklyRepo,
			)
			err := builder.CreatePresetStudyPlanWeekliesForLesson(context.Background(), tcs[i].lesson)
			if tcs[i].hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

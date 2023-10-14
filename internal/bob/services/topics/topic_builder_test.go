package topics

import (
	"context"
	"errors"
	"testing"

	"github.com/manabie-com/backend/internal/bob/entities"
	coursesRepo "github.com/manabie-com/backend/internal/bob/services/courses/repo"
	"github.com/manabie-com/backend/internal/bob/services/topics/repo"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/multierr"
)

func TestCreateTopicByLiveLesson(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name       string
		lesson     *entities.Lesson
		courseRepo coursesRepo.CourseRepo
		topicRepo  repo.TopicRepo
		hasError   bool
	}{
		{
			name: "Create topic by live lesson successfully",
			lesson: &entities.Lesson{
				Name: database.Text("Introduction 1"),
				CourseIDs: entities.CourseIDs{
					CourseIDs: database.TextArray([]string{"course-id-1", "course-id-2"}),
				},
			},
			courseRepo: &coursesRepo.CourseRepoMock{
				FindByIDsMock: func(ctx context.Context, db database.Ext, courseIDs pgtype.TextArray) (map[pgtype.Text]*entities.Course, error) {
					assert.ElementsMatch(t, []string{"course-id-1", "course-id-2"}, database.FromTextArray(courseIDs))
					return map[pgtype.Text]*entities.Course{
						database.Text("course-id-1"): {
							ID:       database.Text("course-id-1"),
							Country:  database.Text("vietnam"),
							Grade:    database.Int2(2),
							Subject:  database.Text("math"),
							SchoolID: database.Int4(1),
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
			topicRepo: &repo.TopicRepoMock{
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
			hasError: false,
		},
		{
			name: "Create topic by live lesson with error when get course id",
			lesson: &entities.Lesson{
				Name:     database.Text("Math 1"),
				CourseID: database.Text("non-existing-id"),
			},
			courseRepo: &coursesRepo.CourseRepoMock{
				FindByIDsMock: func(ctx context.Context, db database.Ext, courseIDs pgtype.TextArray) (map[pgtype.Text]*entities.Course, error) {
					assert.Equal(t, "non-existing-id", courseIDs.Elements[0].String)
					return nil, errors.New("could not found course id")
				},
			},
			hasError: true,
		},
		{
			name: "Create topic by live lesson with non-existing course id",
			lesson: &entities.Lesson{
				Name:     database.Text("Math 1"),
				CourseID: database.Text("non-existing-id"),
			},
			courseRepo: &coursesRepo.CourseRepoMock{
				FindByIDsMock: func(ctx context.Context, db database.Ext, courseIDs pgtype.TextArray) (map[pgtype.Text]*entities.Course, error) {
					assert.Equal(t, "non-existing-id", courseIDs.Elements[0].String)
					return nil, nil
				},
			},
			topicRepo: &repo.TopicRepoMock{
				CreateMock: func(ctx context.Context, db database.Ext, plans []*entities.Topic) error {
					require.Fail(t, "expected not create any topic")
					return nil
				},
			},
			hasError: true,
		},
		{
			name: "create topic by live lesson with no course IDs",
			lesson: &entities.Lesson{
				Name: database.Text("Math 1"),
			},
			hasError: false,
		},
	}

	for i := range tcs {
		t.Run(tcs[i].name, func(t *testing.T) {
			builder := NewTopicBuilder(nil, tcs[i].courseRepo, tcs[i].topicRepo)
			_, err := builder.CreateTopicsByLiveLesson(context.Background(), tcs[i].lesson)
			if tcs[i].hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

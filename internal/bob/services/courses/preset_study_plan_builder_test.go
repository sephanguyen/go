package courses

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/services/courses/repo"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/multierr"
)

func TestCreatePresetStudyPlanByCourseID(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name            string
		courseIDs       []string
		courseRepo      repo.CourseRepo
		presetStudyRepo repo.PresetStudyPlanRepo
		hasError        bool
	}{
		{
			name:      "Create preset study plan by course id successfully",
			courseIDs: []string{"course-id-1", "course-id-2"},
			courseRepo: repo.CourseRepoMock{
				FindByIDsMock: func(ctx context.Context, db database.Ext, courseIDs pgtype.TextArray) (map[pgtype.Text]*entities.Course, error) {
					assert.ElementsMatch(t, []string{"course-id-1", "course-id-2"}, database.FromTextArray(courseIDs))
					return map[pgtype.Text]*entities.Course{
						database.Text("course-id-1"): {
							ID:        database.Text("course-id-1"),
							Name:      database.Text("math 1"),
							Country:   database.Text("vietnam"),
							Grade:     database.Int2(2),
							Subject:   database.Text("math"),
							SchoolID:  database.Int4(1),
							StartDate: database.Timestamptz(time.Date(2021, 2, 3, 4, 5, 6, 7, time.Local)),
						},
						database.Text("course-id-2"): {
							ID:        database.Text("course-id-2"),
							Name:      database.Text("physics 2"),
							Country:   database.Text("vietnam"),
							Grade:     database.Int2(5),
							Subject:   database.Text("physics"),
							SchoolID:  database.Int4(1),
							StartDate: database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)),
						},
					}, nil
				},
				UpsertMock: func(ctx context.Context, db database.Ext, cc []*entities.Course) error {
					assert.Len(t, cc, 2)
					courseIDs := make([]string, 0, len(cc))
					for _, c := range cc {
						assert.NotEmpty(t, c.ID.String)
						assert.NotEmpty(t, c.PresetStudyPlanID.String)
						courseIDs = append(courseIDs, c.ID.String)
					}
					assert.ElementsMatch(t, []string{"course-id-1", "course-id-2"}, courseIDs)

					return nil
				},
			},
			presetStudyRepo: repo.PresetStudyPlanRepoMock{
				CreatePresetStudyPlanMock: func(ctx context.Context, db database.Ext, presetStudyPlans []*entities.PresetStudyPlan) error {
					assert.Len(t, presetStudyPlans, 2)
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
						expected.Name.Set("math 1"),
						expected.Country.Set("vietnam"),
						expected.Grade.Set(2),
						expected.Subject.Set("math"),
						expected.StartDate.Set(time.Date(2021, 2, 3, 4, 5, 6, 7, time.Local)),
					)
					require.NoError(t, err)
					assert.Equal(t, expected, pspBySubject["math"])

					database.AllNullEntity(expected)
					err = multierr.Combine(
						expected.ID.Set(pspBySubject["physics"].ID),
						expected.Name.Set("physics 2"),
						expected.Country.Set("vietnam"),
						expected.Grade.Set(5),
						expected.Subject.Set("physics"),
						expected.StartDate.Set(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)),
					)
					require.NoError(t, err)
					assert.Equal(t, expected, pspBySubject["physics"])

					return nil
				},
			},
		},
		{
			name:      "Create preset study plan with non-existing course id",
			courseIDs: []string{"non-existing-id"},
			courseRepo: repo.CourseRepoMock{
				FindByIDsMock: func(ctx context.Context, db database.Ext, courseIDs pgtype.TextArray) (map[pgtype.Text]*entities.Course, error) {
					assert.Equal(t, "non-existing-id", courseIDs.Elements[0].String)
					return nil, nil
				},
			},
			presetStudyRepo: repo.PresetStudyPlanRepoMock{
				CreatePresetStudyPlanMock: func(ctx context.Context, db database.Ext, presetStudyPlans []*entities.PresetStudyPlan) error {
					require.Fail(t, "expected not create any preset study plan")
					return nil
				},
			},
		},
		{
			name:      "Create preset study plan by course id with some courses have preset study plan",
			courseIDs: []string{"course-id-1", "course-id-2", "course-id-3"},
			courseRepo: repo.CourseRepoMock{
				FindByIDsMock: func(ctx context.Context, db database.Ext, courseIDs pgtype.TextArray) (map[pgtype.Text]*entities.Course, error) {
					assert.ElementsMatch(t, []string{"course-id-1", "course-id-2", "course-id-3"}, database.FromTextArray(courseIDs))
					return map[pgtype.Text]*entities.Course{
						database.Text("course-id-1"): {
							ID:        database.Text("course-id-1"),
							Name:      database.Text("math 1"),
							Country:   database.Text("vietnam"),
							Grade:     database.Int2(2),
							Subject:   database.Text("math"),
							SchoolID:  database.Int4(1),
							StartDate: database.Timestamptz(time.Date(2021, 2, 3, 4, 5, 6, 7, time.Local)),
						},
						database.Text("course-id-2"): {
							ID:        database.Text("course-id-2"),
							Name:      database.Text("physics 2"),
							Country:   database.Text("vietnam"),
							Grade:     database.Int2(5),
							Subject:   database.Text("physics"),
							SchoolID:  database.Int4(1),
							StartDate: database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)),
						},
						database.Text("course-id-3"): {
							ID:                database.Text("course-id-3"),
							PresetStudyPlanID: database.Text("preset-study-plan-id-1"),
							Name:              database.Text("physics 3"),
							Country:           database.Text("vietnam"),
							Grade:             database.Int2(5),
							Subject:           database.Text("physics"),
							SchoolID:          database.Int4(1),
							StartDate:         database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)),
						},
					}, nil
				},
				UpsertMock: func(ctx context.Context, db database.Ext, cc []*entities.Course) error {
					assert.Len(t, cc, 2)
					courseIDs := make([]string, 0, len(cc))
					for _, c := range cc {
						assert.NotEmpty(t, c.ID.String)
						assert.NotEmpty(t, c.PresetStudyPlanID.String)
						courseIDs = append(courseIDs, c.ID.String)
					}
					assert.ElementsMatch(t, []string{"course-id-1", "course-id-2"}, courseIDs)

					return nil
				},
			},
			presetStudyRepo: repo.PresetStudyPlanRepoMock{
				CreatePresetStudyPlanMock: func(ctx context.Context, db database.Ext, presetStudyPlans []*entities.PresetStudyPlan) error {
					assert.Len(t, presetStudyPlans, 2)
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
						expected.Name.Set("math 1"),
						expected.Country.Set("vietnam"),
						expected.Grade.Set(2),
						expected.Subject.Set("math"),
						expected.StartDate.Set(time.Date(2021, 2, 3, 4, 5, 6, 7, time.Local)),
					)
					require.NoError(t, err)
					assert.Equal(t, expected, pspBySubject["math"])

					database.AllNullEntity(expected)
					err = multierr.Combine(
						expected.ID.Set(pspBySubject["physics"].ID),
						expected.Name.Set("physics 2"),
						expected.Country.Set("vietnam"),
						expected.Grade.Set(5),
						expected.Subject.Set("physics"),
						expected.StartDate.Set(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)),
					)
					require.NoError(t, err)
					assert.Equal(t, expected, pspBySubject["physics"])

					return nil
				},
			},
		},
	}

	for i := range tcs {
		t.Run(tcs[i].name, func(t *testing.T) {
			builder := NewPresetStudyPlanBuilder(nil, tcs[i].presetStudyRepo, tcs[i].courseRepo)
			err := builder.CreatePresetStudyPlansByCourseIDs(context.Background(), database.TextArray(tcs[i].courseIDs))
			if tcs[i].hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

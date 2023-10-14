package repositories

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/notification/modules/tagmgmt/entities"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_Upsert(t *testing.T) {
	t.Parallel()
	genMockArgs := func() []interface{} {
		e := &entities.Tag{}
		fields := database.GetFieldNames(e)
		values := database.GetScanFields(e, fields)
		mockValues := make([]interface{}, 0, len(values)+2)
		mockValues = append(mockValues, mock.Anything)
		mockValues = append(mockValues, mock.AnythingOfType("string"))
		for range values {
			mockValues = append(mockValues, mock.Anything)
		}
		return mockValues
	}
	db := &mock_database.Ext{}
	testCases := []struct {
		Name  string
		Ent   *entities.Tag
		Err   error
		Setup func(ctx context.Context)
	}{
		{
			Name: "happy case",
			Ent:  &entities.Tag{},
			Err:  nil,
			Setup: func(ctx context.Context) {
				db.On("Exec",
					genMockArgs()...,
				).Once().Return(pgconn.CommandTag([]byte(`1`)), nil)
			},
		},
		{
			Name: "Error pool closed",
			Ent:  &entities.Tag{},
			Err:  fmt.Errorf("TagRepo.Upsert: %w", puddle.ErrClosedPool),
			Setup: func(ctx context.Context) {
				db.On("Exec",
					genMockArgs()...,
				).Once().Return(pgconn.CommandTag([]byte(`0`)), puddle.ErrClosedPool)
			},
		},
		{
			Name: "No row affected",
			Ent:  &entities.Tag{},
			Err:  fmt.Errorf("TagRepo.Upsert: Tag is not inserted"),
			Setup: func(ctx context.Context) {
				db.On("Exec",
					genMockArgs()...,
				).Once().Return(pgconn.CommandTag([]byte(`0`)), nil)
			},
		},
	}
	tagRepo := &TagRepo{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	for _, testcase := range testCases {
		testcase.Setup(ctx)
		err := tagRepo.Upsert(ctx, db, testcase.Ent)
		if testcase.Err == nil {
			assert.Nil(t, err)
		} else {
			assert.Equal(t, testcase.Err.Error(), err.Error())
		}
	}
}

func Test_DoesTagNameExist(t *testing.T) {
	t.Parallel()
	mockDB := testutil.NewMockDB()
	db := mockDB.DB
	tagRepo := &TagRepo{}

	testCases := []struct {
		Name        string
		SearchTag   pgtype.Text
		ExistExpect bool
		ErrorExpect error
		Setup       func(ctx context.Context)
	}{
		{
			Name:        "happy case",
			SearchTag:   pgtype.Text{String: "manabie"},
			ExistExpect: true,
			ErrorExpect: nil,
			Setup: func(ctx context.Context) {
				countResult := int(1)
				mockDB.MockQueryRowArgs(t, mock.Anything, mock.AnythingOfType("string"), mock.Anything)
				mockDB.MockRowScanFields(nil, []string{""}, []interface{}{&countResult})
			},
		},
		{
			Name:        "count is 0",
			SearchTag:   pgtype.Text{String: "manabie"},
			ExistExpect: false,
			ErrorExpect: nil,
			Setup: func(ctx context.Context) {
				countResult := int(0)
				mockDB.MockQueryRowArgs(t, mock.Anything, mock.AnythingOfType("string"), mock.Anything)
				mockDB.MockRowScanFields(nil, []string{""}, []interface{}{&countResult})
			},
		},
		{
			Name:        "conn pool closed",
			SearchTag:   pgtype.Text{String: "manabie"},
			ExistExpect: false,
			ErrorExpect: fmt.Errorf("TagRepo.IsNameExist: %w", puddle.ErrClosedPool),
			Setup: func(ctx context.Context) {
				countResult := int(0)
				mockDB.MockQueryRowArgs(t, mock.Anything, mock.AnythingOfType("string"), mock.Anything)
				mockDB.MockRowScanFields(puddle.ErrClosedPool, []string{""}, []interface{}{&countResult})
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	for _, testCase := range testCases {
		testCase.Setup(ctx)
		t.Run(testCase.Name, func(t *testing.T) {
			isExist, err := tagRepo.DoesTagNameExist(ctx, db, testCase.SearchTag)
			if testCase.ErrorExpect == nil {
				assert.Equal(t, testCase.ExistExpect, isExist)
			} else {
				assert.Equal(t, testCase.ErrorExpect.Error(), err.Error())
			}
		})
	}
}

func Test_SoftDelete(t *testing.T) {
	t.Parallel()
	mockDB := testutil.NewMockDB()

	testCases := []struct {
		Name    string
		Ids     []string
		ExpcErr error
		Setup   func(ctx context.Context)
	}{
		{
			Name:    "happy case",
			Ids:     []string{"1"},
			ExpcErr: nil,
			Setup: func(ctx context.Context) {
				mockDB.MockExecArgs(t, pgconn.CommandTag([]byte(`1`)), nil, mock.Anything, mock.AnythingOfType("string"), mock.Anything)
			},
		},
		{
			Name:    "err conn closed",
			Ids:     []string{"1"},
			ExpcErr: puddle.ErrClosedPool,
			Setup: func(ctx context.Context) {
				mockDB.MockExecArgs(t, nil, puddle.ErrClosedPool, mock.Anything, mock.AnythingOfType("string"), mock.Anything)
			},
		},
	}

	tagRepo := &TagRepo{}
	ctx := context.Background()
	for _, testCase := range testCases {
		testCase.Setup(ctx)
		t.Run(testCase.Name, func(t *testing.T) {
			err := tagRepo.SoftDelete(ctx, mockDB.DB, database.TextArray(testCase.Ids))
			if testCase.ExpcErr == nil {
				assert.Nil(t, err)
			} else {
				assert.Equal(t, testCase.ExpcErr.Error(), err.Error())
			}
		})
	}
}

func Test_FindByID(t *testing.T) {
	t.Parallel()
	mockDB := testutil.NewMockDB()
	tagRepo := &TagRepo{}
	ctx := context.Background()
	tagID := "tag-id-1"
	tag := &entities.Tag{}
	database.AllRandomEntity(tag)
	fields, values := tag.FieldMap()
	t.Run("happy case", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.AnythingOfType("string"), database.Text(tagID))
		mockDB.MockScanArray(nil, fields, [][]interface{}{values})

		res, err := tagRepo.FindByID(ctx, mockDB.DB, database.Text(tagID))
		assert.Nil(t, err)
		assert.Equal(t, tag, res)
	})

	t.Run("error scan", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.AnythingOfType("string"), database.Text(tagID))
		mockDB.MockScanArray(pgx.ErrNoRows, fields, [][]interface{}{values})

		res, err := tagRepo.FindByID(ctx, mockDB.DB, database.Text(tagID))
		assert.Nil(t, res)
		assert.Equal(t, err, pgx.ErrNoRows)
	})
}

func Test_CheckTagIDsExist(t *testing.T) {
	t.Parallel()
	mockDB := testutil.NewMockDB()
	db := mockDB.DB
	tagRepo := &TagRepo{}
	type TestCase struct {
		Name         string
		Ids          []string
		ErrorExpect  error
		ResultExpect bool
		Setup        func(ctx context.Context, this *TestCase)
	}
	testCases := []TestCase{
		{
			Name:         "happy case",
			Ids:          []string{"id-1", "id-2", "id-3"},
			ErrorExpect:  nil,
			ResultExpect: true,
			Setup: func(ctx context.Context, this *TestCase) {
				result := true
				mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, database.TextArray(this.Ids))
				mockDB.MockRowScanFields(nil, []string{"result"}, []interface{}{&result})
			},
		},
		{
			Name:         "scan error",
			Ids:          []string{"id-1", "id-2", "id-3"},
			ErrorExpect:  fmt.Errorf("TagRepo.CheckTagIDsExist: %w", pgx.ErrNoRows),
			ResultExpect: false,
			Setup: func(ctx context.Context, this *TestCase) {
				mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, database.TextArray(this.Ids))
				mockDB.Row.On("Scan", mock.Anything).Once().Return(pgx.ErrNoRows)
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	for _, testCase := range testCases {
		testCase.Setup(ctx, &testCase)
		t.Run(testCase.Name, func(t *testing.T) {
			isExist, err := tagRepo.CheckTagIDsExist(ctx, db, database.TextArray(testCase.Ids))
			if testCase.ErrorExpect == nil {
				assert.Equal(t, testCase.ResultExpect, isExist)
			} else {
				assert.Equal(t, testCase.ErrorExpect.Error(), err.Error())
			}
		})
	}
}

func Test_FindByFilter(t *testing.T) {
	t.Parallel()
	mockDB := testutil.NewMockDB()
	db := mockDB.DB
	tagRepo := &TagRepo{}
	tag1 := &entities.Tag{}
	tag2 := &entities.Tag{}
	database.AllRandomEntity(tag1)
	database.AllRandomEntity(tag2)
	keyword := "test"
	_ = tag1.TagName.Set(idutil.ULIDNow() + keyword)
	_ = tag2.TagName.Set(idutil.ULIDNow() + keyword + idutil.ULIDNow())
	var total uint32 = 2
	type TestCase struct {
		Name      string
		Setup     func(ctx context.Context, this *TestCase)
		Filter    FindTagFilter
		ExpcErr   error
		ExpcTotal uint32
	}
	testCases := []TestCase{
		{
			Name:      "happy case",
			ExpcErr:   nil,
			ExpcTotal: total,
			Filter: func() FindTagFilter {
				f := NewFindTagFilter()
				f.Keyword.Set(keyword)
				f.WithCount.Set(true)
				return f
			}(),
			Setup: func(ctx context.Context, this *TestCase) {
				ctx, span := interceptors.StartSpan(ctx, "TagRepo.FindByFilter")
				defer span.End()
				fields1, values1 := tag1.FieldMap()
				_, values2 := tag2.FieldMap()

				mockDB.MockQueryArgs(t, nil, ctx, mock.AnythingOfType("string"), this.Filter.Keyword, this.Filter.IsArchived, this.Filter.Limit, this.Filter.Offset)
				mockDB.MockScanArray(nil, fields1, [][]interface{}{values1, values2})

				mockDB.MockQueryRowArgs(t, ctx, mock.AnythingOfType("string"), this.Filter.Keyword, this.Filter.IsArchived)
				mockDB.MockRowScanFields(nil, []string{"total"}, []interface{}{&total})
			},
		},
		{
			Name:      "no rows",
			ExpcErr:   pgx.ErrNoRows,
			ExpcTotal: 0,
			Filter: func() FindTagFilter {
				f := NewFindTagFilter()
				f.Keyword = database.Text(keyword)
				return f
			}(),
			Setup: func(ctx context.Context, this *TestCase) {
				ctx, span := interceptors.StartSpan(ctx, "TagRepo.FindByFilter")
				defer span.End()
				fields1, _ := tag1.FieldMap()

				mockDB.MockQueryArgs(t, pgx.ErrNoRows, ctx, mock.AnythingOfType("string"), this.Filter.Keyword, this.Filter.IsArchived, this.Filter.Limit, this.Filter.Offset)
				mockDB.MockScanArray(nil, fields1, [][]interface{}{nil})

				mockDB.MockQueryRowArgs(t, ctx, mock.AnythingOfType("string"), this.Filter.Keyword, this.Filter.IsArchived)
				mockDB.MockRowScanFields(nil, []string{"total"}, []interface{}{&total})
			},
		},
	}
	ctx := context.Background()
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			tc.Setup(ctx, &tc)
			res, total, err := tagRepo.FindByFilter(ctx, db, tc.Filter)
			if tc.ExpcErr != nil {
				assert.Equal(t, tc.ExpcErr.Error(), err.Error())
				assert.Equal(t, tc.ExpcTotal, total)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tc.ExpcTotal, total)
				assert.Equal(t, tag1, res[0])
				assert.Equal(t, tag2, res[1])
			}
		})
	}
}

func Test_BulkUpsert(t *testing.T) {
	t.Parallel()
	mockDB := testutil.NewMockDB()
	db := mockDB.DB
	repo := &TagRepo{}

	now := time.Now()
	testCases := []struct {
		Name  string
		Tags  []*entities.Tag
		Err   error
		Setup func(ctx context.Context)
	}{
		{
			Name: "happy case",
			Tags: []*entities.Tag{
				{
					TagID:      database.Text("tag-id-1"),
					TagName:    database.Text("tag-name-1"),
					CreatedAt:  database.Timestamptz(now),
					UpdatedAt:  database.Timestamptz(now),
					DeletedAt:  database.Timestamptz(now),
					IsArchived: database.Bool(false),
				},
				{
					TagID:      database.Text("tag-id-2"),
					TagName:    database.Text("tag-name-2"),
					CreatedAt:  database.Timestamptz(now),
					UpdatedAt:  database.Timestamptz(now),
					DeletedAt:  database.Timestamptz(now),
					IsArchived: database.Bool(false),
				},
			},
			Err: nil,
			Setup: func(ctx context.Context) {
				ctx, span := interceptors.StartSpan(ctx, "TagRepo.BulkUpsert")
				defer span.End()
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}

	ctx := context.Background()
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			tc.Setup(ctx)
			err := repo.BulkUpsert(ctx, db, tc.Tags)
			if tc.Err != nil {
				assert.Equal(t, tc.Err.Error(), err.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func Test_FindDuplicateTagNames(t *testing.T) {
	t.Parallel()
	mockDB := testutil.NewMockDB()
	db := mockDB.DB
	repo := &TagRepo{}

	now := time.Now()
	testCases := []struct {
		Name            string
		Tags            []*entities.Tag
		DuplicatedNames map[string]string
		Err             error
		Setup           func(ctx context.Context)
	}{
		{
			Name: "happy case",
			Tags: []*entities.Tag{
				{
					TagID:      database.Text("tag-id-1"),
					TagName:    database.Text("tag-name-1"),
					CreatedAt:  database.Timestamptz(now),
					UpdatedAt:  database.Timestamptz(now),
					DeletedAt:  database.Timestamptz(now),
					IsArchived: database.Bool(false),
				},
				{
					TagID:      database.Text("tag-id-2"),
					TagName:    database.Text("tag-name-2"),
					CreatedAt:  database.Timestamptz(now),
					UpdatedAt:  database.Timestamptz(now),
					DeletedAt:  database.Timestamptz(now),
					IsArchived: database.Bool(false),
				},
			},
			DuplicatedNames: map[string]string{"tag-id-2": "tag-name-2"},
			Err:             nil,
			Setup: func(ctx context.Context) {
				ctx, span := interceptors.StartSpan(ctx, "TagRepo.FindDuplicateTagNames")
				defer span.End()
				mockDB.MockQueryArgs(t, nil, ctx,
					mock.AnythingOfType("string"),
					database.TextArray([]string{"tag-name-1", "tag-name-2"}),
					database.TextArray([]string{"tag-id-1", "tag-id-2"}),
				)
				tagID := "tag-id-2"
				tagName := "tag-name-2"
				value := []interface{}{
					&tagID,
					&tagName,
				}
				mockDB.MockScanArray(nil, []string{"tag_id", "tag_name"}, [][]interface{}{
					value,
				})
			},
		},
	}

	ctx := context.Background()
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			tc.Setup(ctx)
			duplicatedNames, err := repo.FindDuplicateTagNames(ctx, db, tc.Tags)
			if tc.Err != nil {
				assert.Equal(t, tc.Err.Error(), err.Error())
				assert.Nil(t, duplicatedNames)
			} else {
				assert.Equal(t, tc.DuplicatedNames, duplicatedNames)
				assert.Nil(t, err)
			}
		})
	}
}

func Test_FindTagIDsNotExist(t *testing.T) {
	t.Parallel()
	mockDB := testutil.NewMockDB()
	db := mockDB.DB
	repo := &TagRepo{}

	testCases := []struct {
		Name           string
		TagIDs         []string
		TagIDsNotExist []string
		Err            error
		Setup          func(ctx context.Context)
	}{
		{
			Name:           "happy case",
			TagIDs:         []string{"tag-id-1", "tag-id-2"},
			TagIDsNotExist: []string{"tag-id-2"},
			Err:            nil,
			Setup: func(ctx context.Context) {
				ctx, span := interceptors.StartSpan(ctx, "TagRepo.FindTagIDsNotExist")
				defer span.End()
				mockDB.MockQueryArgs(t, nil, ctx,
					mock.AnythingOfType("string"),
					database.TextArray([]string{"tag-id-1", "tag-id-2"}),
				)
				tagID := "tag-id-2"
				value := []interface{}{
					&tagID,
				}
				mockDB.MockScanArray(nil, []string{"tag_id"}, [][]interface{}{
					value,
				})
			},
		},
	}

	ctx := context.Background()
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			tc.Setup(ctx)
			res, err := repo.FindTagIDsNotExist(ctx, db, database.TextArray(tc.TagIDs))
			if tc.Err != nil {
				assert.Equal(t, tc.Err.Error(), err.Error())
				assert.Nil(t, res)
			} else {
				assert.Equal(t, tc.TagIDsNotExist, res)
				assert.Nil(t, err)
			}
		})
	}
}

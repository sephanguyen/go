package postgres

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/book/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/modules/book/repository/postgres/dto"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestBookRepo_Upsert(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	mockDB := &mock_database.Ext{}
	bookRepo := &BookRepo{
		DB: mockDB,
	}
	validBookReq := []domain.Book{
		{
			ID:   "book-id-1",
			Name: "book-name-1",
		},
		{
			ID:   "book-id-2",
			Name: "book-name-2",
		},
	}

	t.Run("successfully", func(t *testing.T) {
		batchResults := &mock_database.BatchResults{}
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Once().Return(cmdTag, nil)
		batchResults.On("Exec").Once().Return(cmdTag, nil)
		batchResults.On("Close").Once().Return(nil)

		err := bookRepo.Upsert(ctx, validBookReq)
		require.Nil(t, err)

		mock.AssertExpectationsForObjects(
			t,
			mockDB,
		)
	})
	t.Run("error", func(t *testing.T) {
		batchResults := &mock_database.BatchResults{}
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Once().Return(cmdTag, nil)
		batchResults.On("Exec").Once().Return(cmdTag, puddle.ErrClosedPool)
		batchResults.On("Close").Once().Return(nil)

		err := bookRepo.Upsert(ctx, validBookReq)
		require.Error(t, err)

		mock.AssertExpectationsForObjects(
			t,
			mockDB,
		)
	})
}

func TestBookRepo_GetPublishedBookContent(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	t.Run("error on DB", func(t *testing.T) {
		// arrange
		mockDB := testutil.NewMockDB()
		bookRepo := &BookRepo{
			DB: mockDB.DB,
		}
		mockBookID := database.Text(idutil.ULIDNow())
		mockBookName := database.Text("name " + mockBookID.String)
		_, jsonStr := genFakeChapters()
		pgtypeJson := database.Text(jsonStr)
		values := []interface{}{
			&mockBookID,
			&mockBookName,
			&pgtypeJson,
		}
		mockDB.MockQueryArgs(t, nil, mock.Anything, queryPublishedBookContent, mockBookID.String)
		mockDB.MockScanFields(pgx.ErrNoRows, []string{"id", "name", "chapters"}, values)

		// act
		bookTree, err := bookRepo.GetPublishedBookContent(ctx, mockBookID.String)

		// assert
		assert.True(t, errors.CheckErrType(errors.ErrNoRowsExisted, err))
		assert.Equal(t, domain.Book{}, bookTree)
	})

	t.Run("successfully with book tree returned, include null material", func(t *testing.T) {
		// arrange
		mockDB := testutil.NewMockDB()
		bookRepo := &BookRepo{
			DB: mockDB.DB,
		}
		mockBookID := database.Text(idutil.ULIDNow())
		mockBookName := database.Text("name " + mockBookID.String)
		fakeChapters, jsonStr := genFakeChapters()
		pgtypeJson := database.Text(jsonStr)
		values := []interface{}{
			&mockBookID,
			&mockBookName,
			&pgtypeJson,
		}
		mockDB.MockQueryArgs(t, nil, mock.Anything, queryPublishedBookContent, mockBookID.String)
		mockDB.MockScanFields(nil, []string{"id", "name", "chapters"}, values)

		expectedBook := domain.Book{
			ID:       mockBookID.String,
			Name:     mockBookName.String,
			Chapters: fakeChapters,
		}

		// act
		actualBook, err := bookRepo.GetPublishedBookContent(ctx, mockBookID.String)

		// assert
		assert.Nil(t, err)
		assert.Equal(t, expectedBook, actualBook)
	})

	t.Run("return empty chapter dtos when no chapter found", func(t *testing.T) {
		// arrange
		mockDB := testutil.NewMockDB()
		bookRepo := &BookRepo{
			DB: mockDB.DB,
		}
		mockBookID := database.Text(idutil.ULIDNow())
		mockBookName := database.Text("name " + mockBookID.String)
		pgtypeJson := pgtype.Text{Status: pgtype.Null}
		values := []interface{}{
			&mockBookID,
			&mockBookName,
			&pgtypeJson,
		}
		mockDB.MockQueryArgs(t, nil, mock.Anything, queryPublishedBookContent, mockBookID.String)
		mockDB.MockScanFields(nil, []string{"id", "name", "chapters"}, values)

		expectedBook := domain.Book{
			ID:       mockBookID.String,
			Name:     mockBookName.String,
			Chapters: []domain.Chapter{},
		}

		// act
		actualBook, err := bookRepo.GetPublishedBookContent(ctx, mockBookID.String)

		// assert
		assert.Nil(t, err)
		assert.Equal(t, expectedBook, actualBook)
	})

	t.Run("success on null materials", func(t *testing.T) {
		// arrange
		mockDB := testutil.NewMockDB()
		bookRepo := &BookRepo{
			DB: mockDB.DB,
		}
		mockBookID := database.Text(idutil.ULIDNow())
		mockBookName := database.Text("name " + mockBookID.String)
		pgtypeJson := pgtype.Text{Status: pgtype.Null}
		values := []interface{}{
			&mockBookID,
			&mockBookName,
			&pgtypeJson,
		}
		mockDB.MockQueryArgs(t, nil, mock.Anything, queryPublishedBookContent, mockBookID.String)
		mockDB.MockScanFields(nil, []string{"id", "name", "chapters"}, values)

		expectedBook := domain.Book{
			ID:       mockBookID.String,
			Name:     mockBookName.String,
			Chapters: []domain.Chapter{},
		}

		// act
		actualBook, err := bookRepo.GetPublishedBookContent(ctx, mockBookID.String)

		// assert
		assert.Nil(t, err)
		assert.Equal(t, expectedBook, actualBook)
	})
}

func genFakeChapters() ([]domain.Chapter, string) {
	chapterDtos := []dto.Chapter{
		{
			ID:    "Chapter ID 1",
			Name:  "Chapter Name 1",
			Order: 1,
			Topics: []dto.Topic{
				{
					ID:                "TOPIC ID 1",
					Name:              "TOPIC NAME 1",
					IconURL:           "ICON 1",
					Order:             0,
					LearningMaterials: nil,
				},
			},
		},
		{
			ID:    "Chapter ID 2",
			Name:  "Chapter Name 2",
			Order: 2,
			Topics: []dto.Topic{
				{
					ID:      "TOPIC ID 2",
					Name:    "TOPIC NAME 2",
					IconURL: "ICON 2",
					Order:   0,
					LearningMaterials: []dto.BookContentLearningMaterial{
						{
							ID:    "Material ID 2",
							Name:  "Material Name 2",
							Order: 0,
						},
						{
							ID:    "Material ID 3",
							Name:  "Material Name 3",
							Order: 1,
						},
					},
				},
			},
		},
	}
	chapters := []domain.Chapter{
		{
			ID:           "Chapter ID 1",
			Name:         "Chapter Name 1",
			DisplayOrder: 1,
			Topics: []domain.Topic{
				{
					ID:                "TOPIC ID 1",
					Name:              "TOPIC NAME 1",
					IconURL:           "ICON 1",
					DisplayOrder:      0,
					LearningMaterials: []domain.LearningMaterial{},
				},
			},
		},
		{
			ID:           "Chapter ID 2",
			Name:         "Chapter Name 2",
			DisplayOrder: 2,
			Topics: []domain.Topic{
				{
					ID:           "TOPIC ID 2",
					Name:         "TOPIC NAME 2",
					IconURL:      "ICON 2",
					DisplayOrder: 0,
					LearningMaterials: []domain.LearningMaterial{
						{
							ID:           "Material ID 2",
							Name:         "Material Name 2",
							DisplayOrder: 0,
							Published:    true,
						},
						{
							ID:           "Material ID 3",
							Name:         "Material Name 3",
							DisplayOrder: 1,
							Published:    true,
						},
					},
				},
			},
		},
	}
	jsonStr, _ := json.Marshal(chapterDtos)

	return chapters, string(jsonStr)
}

func TestBookRepo_GetBookHierarchyFlattenByLearningMaterialID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDB := testutil.NewMockDB()
	bookRepo := &BookRepo{
		DB: mockDB.DB,
	}

	learningMaterialID := "learningMaterial_9"

	dto := dto.BookHierarchyFlatten{
		BookID:             database.Text("book_id_1"),
		ChapterID:          database.Text("chapter_id_1"),
		TopicID:            database.Text("topic_id_1"),
		LearningMaterialID: database.Text("lm_id_1"),
	}

	t.Run("Happy case", func(t *testing.T) {
		// arrange
		values := []interface{}{
			&dto.BookID,
			&dto.ChapterID,
			&dto.TopicID,
			&dto.LearningMaterialID,
		}
		mockDB.MockQueryArgs(t, nil, mock.Anything, queryGetBookHierarchyFlattenByLearningMaterialID, learningMaterialID)
		mockDB.MockScanFields(nil, []string{"book_id", "chapter_id", "topic_id", "learning_material_id"}, values)

		// act
		bHierarchyFlatten, err := bookRepo.GetBookHierarchyFlattenByLearningMaterialID(ctx, learningMaterialID)

		// assert
		expectBHierarchyFlatten := dto.ToEntity()
		assert.Nil(t, err)
		assert.Equal(t, expectBHierarchyFlatten, bHierarchyFlatten)
	})

	t.Run("NoRowsExisted", func(t *testing.T) {
		// arrange
		values := []interface{}{}
		mockDB.MockQueryArgs(t, nil, mock.Anything, queryGetBookHierarchyFlattenByLearningMaterialID, learningMaterialID)
		mockDB.MockScanFields(pgx.ErrNoRows, []string{}, values)

		// act
		bHierarchyFlatten, err := bookRepo.GetBookHierarchyFlattenByLearningMaterialID(ctx, learningMaterialID)

		// assert
		assert.True(t, errors.CheckErrType(errors.ErrNoRowsExisted, err))
		assert.Equal(t, domain.BookHierarchyFlatten{}, bHierarchyFlatten)
	})

	t.Run("error on DB", func(t *testing.T) {
		// arrange
		values := []interface{}{}
		mockDB.MockQueryArgs(t, puddle.ErrNotAvailable, mock.Anything, queryGetBookHierarchyFlattenByLearningMaterialID, learningMaterialID)
		mockDB.MockScanFields(nil, []string{}, values)

		// act
		bHierarchyFlatten, err := bookRepo.GetBookHierarchyFlattenByLearningMaterialID(ctx, learningMaterialID)

		// assert
		assert.True(t, errors.CheckErrType(errors.ErrDB, err))
		assert.Equal(t, domain.BookHierarchyFlatten{}, bHierarchyFlatten)
	})
}

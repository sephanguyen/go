package services

import (
	"context"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/bob/constants"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_eureka_repositories "github.com/manabie-com/backend/mock/eureka/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var ErrCanNotFindCountryGradeMap = status.Error(codes.InvalidArgument, "cannot find country grade map")

func TestChapterModifierService_UpsertChapter(t *testing.T) {
	t.Parallel()
	chapterRepo := &mock_eureka_repositories.MockChapterRepo{}
	bookChapterRepo := &mock_eureka_repositories.MockBookChapterRepo{}
	bookRepo := &mock_eureka_repositories.MockBookRepo{}

	mockDB := &mock_database.Ext{}
	mockTxer := &mock_database.Tx{}

	chapterModifierService := &ChapterModifierService{
		ChapterRepo:     chapterRepo,
		BookRepo:        bookRepo,
		BookChapterRepo: bookChapterRepo,
		DBTrace:         &database.DBTrace{DB: mockDB},
	}
	bookMock := &entities.Book{
		ID:                         database.Text("mock-book-id"),
		Name:                       database.Text("mock-book-name"),
		CurrentChapterDisplayOrder: database.Int4(0),
	}
	chapterMaps := make(map[string]*entities.Chapter)
	chapterMaps["mock-chapter-id"] = &entities.Chapter{
		ID: database.Text("mock-chapter-id"),
	}

	testCases := map[string]TestCase{
		"missing name": {
			req: &epb.UpsertChaptersRequest{
				Chapters: []*cpb.Chapter{{
					Info: &cpb.ContentBasicInfo{
						Id:           "id",
						Name:         "",
						Country:      cpb.Country_COUNTRY_VN,
						Subject:      cpb.Subject_SUBJECT_BIOLOGY,
						DisplayOrder: 1,
						SchoolId:     constants.ManabieSchool,
						Grade:        1,
					},
				}},
				BookId: "",
			},
			expectedErr: status.Error(codes.InvalidArgument, "chapter name cannot be empty"),
			setup: func(ctx context.Context) {
			},
		},
		"missing subject": {
			req: &epb.UpsertChaptersRequest{
				Chapters: []*cpb.Chapter{{
					Info: &cpb.ContentBasicInfo{
						Id:           "id",
						Name:         "name",
						Country:      cpb.Country_COUNTRY_VN,
						Subject:      cpb.Subject_SUBJECT_NONE,
						DisplayOrder: 1,
						SchoolId:     constants.ManabieSchool,
						Grade:        1,
					},
				}},
				BookId: "",
			},
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTxer, nil)
				mockTxer.On("Commit", mock.Anything).Return(nil)
				bookRepo.On("FindByID", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(bookMock, nil)
				chapterRepo.On("FindByIDs", ctx, mock.Anything, mock.Anything).Once().Return(chapterMaps, nil)
				chapterRepo.On("Upsert", ctx, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
		"missing school id": {
			req: &epb.UpsertChaptersRequest{
				Chapters: []*cpb.Chapter{{
					Info: &cpb.ContentBasicInfo{
						Id:           "id",
						Name:         "name",
						Country:      cpb.Country_COUNTRY_VN,
						Subject:      cpb.Subject_SUBJECT_BIOLOGY,
						DisplayOrder: 1,
						SchoolId:     0,
						Grade:        1,
					},
				}},
				BookId: "",
			},
			expectedErr: status.Error(codes.InvalidArgument, "missing school id of chapter name"),
			setup: func(ctx context.Context) {
			},
		},
		"display_order cannot be less than 0": {
			req: &epb.UpsertChaptersRequest{
				Chapters: []*cpb.Chapter{{
					Info: &cpb.ContentBasicInfo{
						Id:           "id",
						Name:         "name",
						Country:      cpb.Country_COUNTRY_VN,
						Subject:      cpb.Subject_SUBJECT_BIOLOGY,
						DisplayOrder: -1,
						SchoolId:     constants.ManabieSchool,
						Grade:        1,
					},
				}},
				BookId: "",
			},
			expectedErr: status.Error(codes.InvalidArgument, "display_order cannot be less than 0"),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTxer, nil)
				mockTxer.On("Commit", mock.Anything).Return(nil)
				bookRepo.On("FindByID", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(bookMock, nil)
				chapterRepo.On("FindByIDs", ctx, mock.Anything, mock.Anything).Once().Return(chapterMaps, nil)
				chapterRepo.On("Upsert", ctx, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
		"happy case": {
			req: &epb.UpsertChaptersRequest{
				Chapters: []*cpb.Chapter{{
					Info: &cpb.ContentBasicInfo{
						Id:           "id",
						Name:         "name",
						Country:      cpb.Country_COUNTRY_VN,
						Subject:      cpb.Subject_SUBJECT_BIOLOGY,
						DisplayOrder: 1,
						SchoolId:     constants.ManabieSchool,
						Grade:        1,
					},
				}},
				BookId: "",
			},
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTxer, nil)
				mockTxer.On("Commit", mock.Anything).Return(nil)
				bookRepo.On("FindByID", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(bookMock, nil)
				chapterRepo.On("FindByIDs", ctx, mock.Anything, mock.Anything).Once().Return(chapterMaps, nil)
				chapterRepo.On("Upsert", ctx, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
		"book not found": {
			req: &epb.UpsertChaptersRequest{
				Chapters: []*cpb.Chapter{{
					Info: &cpb.ContentBasicInfo{
						Id:           "id",
						Name:         "name",
						Country:      cpb.Country_COUNTRY_VN,
						Subject:      cpb.Subject_SUBJECT_BIOLOGY,
						DisplayOrder: 1,
						SchoolId:     constants.ManabieSchool,
						Grade:        1,
					},
				}},
				BookId: "book-id",
			},
			expectedErr: status.Error(codes.NotFound, "book not found"),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTxer, nil)
				mockTxer.On("Rollback", mock.Anything).Return(nil)
				chapterRepo.On("Upsert", ctx, mock.Anything, mock.Anything).Once().Return(nil)
				bookRepo.On("FindByID", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(bookMock, nil)
				chapterRepo.On("FindByIDs", ctx, mock.Anything, mock.Anything).Once().Return(chapterMaps, nil)
				bookRepo.On("FindByIDs", ctx, mock.Anything, mock.Anything).Once().Return(map[string]*entities.Book{
					"book-id-1": {ID: database.Text("book-id-1"), Name: database.Text("book-name")},
				}, nil)
			},
		},
		"with book id": {
			req: &epb.UpsertChaptersRequest{
				Chapters: []*cpb.Chapter{{
					Info: &cpb.ContentBasicInfo{
						Id:           "id",
						Name:         "name",
						Country:      cpb.Country_COUNTRY_VN,
						Subject:      cpb.Subject_SUBJECT_BIOLOGY,
						DisplayOrder: 1,
						SchoolId:     constants.ManabieSchool,
						Grade:        1,
					},
				}},
				BookId: "book-id",
			},
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTxer, nil)
				mockTxer.On("Commit", mock.Anything).Return(nil)

				chapterRepo.On("Upsert", ctx, mock.Anything, mock.Anything).Once().Return(nil)
				bookRepo.On("FindByID", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(bookMock, nil)
				chapterRepo.On("FindByIDs", ctx, mock.Anything, mock.Anything).Once().Return(chapterMaps, nil)
				bookRepo.On("FindByIDs", ctx, mock.Anything, mock.Anything).Once().Return(map[string]*entities.Book{
					"book-id": {ID: database.Text("book-id"), Name: database.Text("book-name")},
				}, nil)
				bookRepo.On("UpdateCurrentChapterDisplayOrder", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				bookChapterRepo.On("Upsert", ctx, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
	}

	for caseName, testCase := range testCases {
		t.Run(caseName, func(t *testing.T) {
			ctx := context.Background()
			testCase.setup(ctx)

			_, err := chapterModifierService.UpsertChapters(ctx, testCase.req.(*epb.UpsertChaptersRequest))
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}
}

func TestChapterModifierService_DeleteChapters(t *testing.T) {
	t.Parallel()

	chapterRepo := &mock_eureka_repositories.MockChapterRepo{}
	bookChapterRepo := &mock_eureka_repositories.MockBookChapterRepo{}

	mockDB := &mock_database.Ext{}
	tx := &mock_database.Tx{}

	s := &ChapterModifierService{
		DBTrace:         mockDB,
		ChapterRepo:     chapterRepo,
		BookChapterRepo: bookChapterRepo,
	}

	m := map[string]*entities.Chapter{
		"chapter-1": {
			ID: database.Text("chapter-1"),
		},
	}

	testCases := map[string]TestCase{
		"chapters not exist": {
			req: &epb.DeleteChaptersRequest{
				ChapterIds: []string{"wrong-chapter-id"},
			},
			expectedErr: status.Errorf(codes.InvalidArgument, fmt.Errorf("unable to find chapter by ids: %w", pgx.ErrNoRows).Error()),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				chapterRepo.On("FindByIDs", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		"err ChapterRepo.SoftDelete": {
			req: &epb.DeleteChaptersRequest{
				ChapterIds: []string{"chapter-1"},
			},
			expectedErr: status.Errorf(codes.Internal, fmt.Errorf("unable to delete chapters: %w", ErrSomethingWentWrong).Error()),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)

				chapterRepo.On("FindByIDs", ctx, mockDB, mock.Anything).Once().Return(m, nil)
				chapterRepo.On("SoftDelete", ctx, tx, mock.Anything).Once().Return(0, ErrSomethingWentWrong)
			},
		},
		"err BookChapterRepo.SoftDeleteByChapterIDs": {
			req: &epb.DeleteChaptersRequest{
				ChapterIds: []string{"chapter-1"},
			},
			expectedErr: status.Errorf(codes.Internal, fmt.Errorf("unable to delete books chapters: %w", ErrSomethingWentWrong).Error()),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)

				chapterRepo.On("FindByIDs", ctx, mockDB, mock.Anything).Once().Return(m, nil)
				chapterRepo.On("SoftDelete", ctx, tx, mock.Anything).Once().Return(0, nil)
				bookChapterRepo.On("SoftDeleteByChapterIDs", ctx, tx, mock.Anything).Once().Return(ErrSomethingWentWrong)
			},
		},
		"happy case": {
			req: &epb.DeleteChaptersRequest{
				ChapterIds: []string{"chapter-1"},
			},
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)

				chapterRepo.On("FindByIDs", ctx, mockDB, mock.Anything).Once().Return(m, nil)
				chapterRepo.On("SoftDelete", ctx, tx, mock.Anything).Once().Return(0, nil)
				bookChapterRepo.On("SoftDeleteByChapterIDs", ctx, tx, mock.Anything).Once().Return(nil)
			},
		},
	}
	for caseName, testCase := range testCases {
		t.Run(caseName, func(t *testing.T) {
			ctx := context.Background()
			testCase.setup(ctx)
			req := testCase.req.(*epb.DeleteChaptersRequest)
			if _, err := s.DeleteChapters(ctx, req); testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}
}

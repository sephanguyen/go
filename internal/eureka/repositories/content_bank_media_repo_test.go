package repositories

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
)

func TestContentBankMediaRepoUpsert(t *testing.T) {
	t.Parallel()
	now := time.Now()
	db := &mock_database.QueryExecer{}
	contentBankMediaRepo := &ContentBankMediaRepo{}
	validContentBankMediaReq := &entities.ContentBankMedia{
		ID:            pgtype.Text{String: "id", Status: pgtype.Present},
		Name:          pgtype.Text{String: "test.png", Status: pgtype.Present},
		Resource:      pgtype.Text{String: "gcs", Status: pgtype.Present},
		Type:          pgtype.Text{String: "image/png", Status: pgtype.Present},
		FileSizeBytes: pgtype.Int8{Int: 100, Status: pgtype.Present},
		CreatedBy:     pgtype.Text{String: "test", Status: pgtype.Present},
		CreatedAt:     pgtype.Timestamptz{Time: now, Status: pgtype.Present},
		UpdatedAt:     pgtype.Timestamptz{Time: now, Status: pgtype.Present},
	}
	expectedErr := fmt.Errorf("query row error")

	testCases := []TestCase{
		{
			name:        "happy case",
			req:         validContentBankMediaReq,
			expectedErr: nil,
			setup: func(context.Context) {
				_, fields := validContentBankMediaReq.FieldMap()
				mockRow := mock_database.NewRow(t)
				mockRow.On("Scan", mock.Anything).Once().Return(nil)

				args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, fields...)
				db.On("QueryRow", args...).Once().Return(mockRow, nil)
			},
		},
		{
			name:        "query row error",
			req:         validContentBankMediaReq,
			expectedErr: fmt.Errorf("db.QueryRow: %w", expectedErr),
			setup: func(context.Context) {
				_, fields := validContentBankMediaReq.FieldMap()
				mockRow := mock_database.NewRow(t)
				mockRow.On("Scan", mock.Anything).Once().Return(expectedErr)

				args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, fields...)
				db.On("QueryRow", args...).Once().Return(mockRow, nil)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			tc.setup(ctx)
			mediaID, err := contentBankMediaRepo.Upsert(ctx, db, tc.req.(*entities.ContentBankMedia))
			assert.NotNil(t, mediaID)
			assert.Equal(t, tc.expectedErr, err)
		})
	}
}

func TestContentBankMediaRepoFindByMediaNames(t *testing.T) {
	t.Parallel()

	db := &mock_database.QueryExecer{}
	contentBankMediaRepo := &ContentBankMediaRepo{}
	mediaNames := []string{"test.png", "test2.png", "test3.png"}

	contentBankMedia := &entities.ContentBankMedia{}
	fields, _ := contentBankMedia.FieldMap()
	scanFields := database.GetScanFields(contentBankMedia, fields)

	testCases := []TestCase{
		{
			name:        "happy case",
			req:         mediaNames,
			expectedErr: nil,
			setup: func(context.Context) {
				mockRows := mock_database.NewRows(t)

				args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, mediaNames)
				db.On("Query", args...).Once().Return(mockRows, nil)

				mockRows.On("Close").Once().Return(nil)

				mockRows.On("Next").Once().Return(true)
				mockRows.On("Scan", scanFields...).Once().Return(nil)
				mockRows.On("Next").Once().Return(true)
				mockRows.On("Scan", scanFields...).Once().Return(nil)
				mockRows.On("Next").Once().Return(false)
			},
		},
		{
			name:        "query error",
			req:         mediaNames,
			expectedErr: fmt.Errorf("db.Query: %w", fmt.Errorf("query error")),
			setup: func(context.Context) {
				args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, mediaNames)
				db.On("Query", args...).Once().Return(nil, fmt.Errorf("query error"))
			},
		},
		{
			name:        "scan error",
			req:         mediaNames,
			expectedErr: fmt.Errorf("rows.Scan: %w", fmt.Errorf("scan error")),
			setup: func(context.Context) {
				mockRows := mock_database.NewRows(t)

				args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, mediaNames)
				db.On("Query", args...).Once().Return(mockRows, nil)

				mockRows.On("Close").Once().Return(nil)

				mockRows.On("Next").Once().Return(true)
				mockRows.On("Scan", scanFields...).Once().Return(fmt.Errorf("scan error"))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			tc.setup(ctx)
			medias, err := contentBankMediaRepo.FindByMediaNames(ctx, db, tc.req.([]string))
			if err == nil {
				assert.NotNil(t, medias)
			}
			assert.Equal(t, tc.expectedErr, err)
		})
	}

}

func TestContentBankMediaRepoFindByID(t *testing.T) {
	t.Parallel()

	db := &mock_database.QueryExecer{}
	contentBankMediaRepo := &ContentBankMediaRepo{}
	mediaID := "01H838KN55EJBR564HW3DKHJAR"
	contentBankMedia := &entities.ContentBankMedia{}
	fields, _ := contentBankMedia.FieldMap()
	scanFields := database.GetScanFields(contentBankMedia, fields)

	testCases := []TestCase{
		{
			name:        "happy case",
			req:         mediaID,
			expectedErr: nil,
			setup: func(context.Context) {
				mockRow := mock_database.NewRow(t)

				args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, mediaID)
				db.On("QueryRow", args...).Once().Return(mockRow)
				mockRow.On("Scan", scanFields...).Once().Return(nil)
			},
		},
		{
			name:        "query row error",
			req:         mediaID,
			expectedErr: fmt.Errorf("db.QueryRow: %w", fmt.Errorf("query row error")),
			setup: func(context.Context) {
				mockRow := mock_database.NewRow(t)

				args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, mediaID)
				db.On("QueryRow", args...).Once().Return(mockRow)
				mockRow.On("Scan", scanFields...).Once().Return(fmt.Errorf("query row error"))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			tc.setup(ctx)
			media, err := contentBankMediaRepo.FindByID(ctx, db, tc.req.(string))
			if err == nil {
				assert.NotNil(t, media)
				assert.Equal(t, media, contentBankMedia)
			}
			assert.Equal(t, tc.expectedErr, err)
		})
	}
}

func TestContentBankMediaRepoDeleteByID(t *testing.T) {
	t.Parallel()

	db := &mock_database.QueryExecer{}
	contentBankMediaRepo := &ContentBankMediaRepo{}
	mediaID := "01H838KN55EJBR564HW3DKHJAR"

	testCases := []TestCase{
		{
			name:        "happy case",
			req:         mediaID,
			expectedErr: nil,
			setup: func(context.Context) {
				args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, mediaID)

				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("Exec", args...).Once().Return(cmdTag, nil)
			},
		},
		{
			name:        "exec error",
			req:         mediaID,
			expectedErr: fmt.Errorf("db.Exec: %w", fmt.Errorf("exec error")),
			setup: func(context.Context) {
				args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, mediaID)

				db.On("Exec", args...).Once().Return(nil, fmt.Errorf("exec error"))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			tc.setup(ctx)
			err := contentBankMediaRepo.DeleteByID(ctx, db, tc.req.(string))
			assert.Equal(t, tc.expectedErr, err)
		})
	}
}

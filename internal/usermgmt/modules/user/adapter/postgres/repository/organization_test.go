package repository

import (
	"context"
	"encoding/base64"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func OrganizationRepoSqlMock() (*OrganizationRepo, *testutil.MockDB) {
	r := &OrganizationRepo{}
	return r, testutil.NewMockDB()
}
func TestOrganization_DefaultOrganizationAuthValues(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, mockDB := OrganizationRepoSqlMock()
	t.Run("happy case default organization auth values local", func(t *testing.T) {
		env := "local"
		orgValue := generateDefaultOrganizationAuthValues(env)
		data := r.DefaultOrganizationAuthValues(env)
		assert.EqualValues(t, data, orgValue)
		assert.Nil(t, nil)
	})
	t.Run("happy case default organization auth values stag", func(t *testing.T) {
		env := "stag"
		orgValue := generateDefaultOrganizationAuthValues(env)
		data := r.DefaultOrganizationAuthValues(env)
		assert.EqualValues(t, data, orgValue)
		assert.Nil(t, nil)
	})
	t.Run("happy case default organization auth values uat", func(t *testing.T) {
		env := "uat"
		orgValue := generateDefaultOrganizationAuthValues(env)
		data := r.DefaultOrganizationAuthValues(env)
		assert.EqualValues(t, data, orgValue)
		assert.Nil(t, nil)
	})
	t.Run("happy case default organization auth values prod", func(t *testing.T) {
		env := "prod"
		orgValue := generateDefaultOrganizationAuthValues(env)
		data := r.DefaultOrganizationAuthValues(env)
		assert.EqualValues(t, data, orgValue)
		assert.Nil(t, nil)
	})
	t.Run("happy case with default value", func(t *testing.T) {
		env := "local"
		orgValue := generateDefaultOrganizationAuthValues(env)
		data := r.WithDefaultValue(env)
		assert.EqualValues(t, data.defaultOrganizationAuthValues, orgValue)
		assert.Nil(t, nil)
	})
	t.Run("happy case get tenantID by organizationID", func(t *testing.T) {
		orgID := "1"
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &orgID)
		mockDB.DB.On("QueryRow").Once().Return(mockDB.Row, nil)
		mockDB.Row.On("Scan", mock.Anything).Once().Return(nil)

		data := generateDefaultOrganizationAuthValues("local")
		r.defaultOrganizationAuthValues = data
		_, err := r.GetTenantIDByOrgID(ctx, mockDB.DB, orgID)
		assert.NoError(t, err)
	})
}

func TestOrganizationRepo_Find(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	orgID := database.Text(idutil.ULIDNow())
	_, orgValues := (&entity.Organization{}).FieldMap()
	argsOrgs := append([]interface{}{}, genSliceMock(len(orgValues))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := OrganizationRepoSqlMock()
		mockDB.DB.On("QueryRow", mock.Anything, mock.AnythingOfType("string"), &orgID).Once().Return(mockDB.Row, nil)
		mockDB.Row.On("Scan", argsOrgs...).Once().Return(nil)
		org, err := repo.Find(ctx, mockDB.DB, orgID)
		assert.Nil(t, err)
		assert.NotNil(t, org)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("query row error", func(t *testing.T) {
		repo, mockDB := OrganizationRepoSqlMock()
		mockDB.DB.On("QueryRow", mock.Anything, mock.AnythingOfType("string"), &orgID).Once().Return(mockDB.Row)
		mockDB.Row.On("Scan", argsOrgs...).Once().Return(puddle.ErrClosedPool)
		org, err := repo.Find(ctx, mockDB.DB, orgID)
		assert.Nil(t, org)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func generateDefaultOrganizationAuthValues(env string) string {
	r, _ := OrganizationRepoSqlMock()
	orgValue := r.DefaultOrganizationAuthValues(env)
	return orgValue
}

func TestOrganizationRepo_GetByTenantID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	organizationRepo := &OrganizationRepo{}
	mockDB := testutil.NewMockDB()
	rows := &mock_database.Rows{}
	organization := &entity.Organization{}
	_, scanFields := organization.FieldMap()

	testCases := []struct {
		name      string
		setup     func()
		expectErr error
	}{
		{
			name:      "happy case",
			expectErr: nil,
			setup: func() {
				mockDB.DB.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Scan", scanFields...).Once().Return(nil)
			},
		},
		{
			name:      "query return no rows err",
			expectErr: errors.Wrap(pgx.ErrNoRows, "Scan"),
			setup: func() {
				mockDB.DB.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Scan", scanFields...).Once().Return(pgx.ErrNoRows)
			},
		},
	}

	for index, testcase := range testCases {
		t.Run(fmt.Sprintf("%s-%d", testcase.name, index), func(t *testing.T) {
			testcase.setup()
			role, err := organizationRepo.GetByTenantID(ctx, mockDB.DB, "example-tenant-id")
			if err != nil {
				assert.EqualError(t, err, testcase.expectErr.Error())
			} else {
				assert.NotNil(t, role)
			}

			mock.AssertExpectationsForObjects(t, mockDB.DB, rows)
		})
	}
}

type mockOrganizationRepo struct {
	getOrganizationByTenantID func(ctx context.Context, db database.QueryExecer, tenantID string) (*entity.Organization, error)
}

func (m *mockOrganizationRepo) GetByTenantID(ctx context.Context, db database.QueryExecer, tenantID string) (*entity.Organization, error) {
	return m.getOrganizationByTenantID(ctx, db, tenantID)
}

func happyCaseMockOrganizationRepo() *mockOrganizationRepo {
	return &mockOrganizationRepo{
		func(ctx context.Context, db database.QueryExecer, tenantID string) (*entity.Organization, error) {
			organization := &entity.Organization{
				ScryptSignerKey:     database.Text("rG79ZtBaaMU5bdeu6W41Svw2s7db4cBYFpuPx8nzciwPW26pNNWB0MhSJ+11C+NXf3iROB4xYQqa\nqv2uLjjUGd/jU9FFmtEkXfvHUeSlz1pFEQqT8i5aIkZnu010UjVY"),
				ScryptSaltSeparator: database.Text("ltZsOrOl9PrsyPy68iu1zA=="),
				ScryptRounds:        database.Text("FbOQaVWt8wJGGnTI9Lz0tA=="),
				ScryptMemoryCost:    database.Text("OAIWCBNantBzaZCVvWNTsw=="),
			}
			return organization, nil
		},
	}
}

func TestTenantConfigRepo_GetTenantConfig(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	mockDB := testutil.NewMockDB()

	testCases := []struct {
		name             string
		tenantConfigRepo *TenantConfigRepo
		expectErr        error
	}{
		{
			name:             "tenant config repo is nil",
			tenantConfigRepo: nil,
			expectErr:        errors.New("TenantConfigRepo is nil"),
		},
		{
			name: "tenant config repo is nil",
			tenantConfigRepo: &TenantConfigRepo{
				QueryExecer: nil,
			},
			expectErr: errors.New("QueryExecer is nil"),
		},
		{
			name: "invalid base64 aes key",
			tenantConfigRepo: &TenantConfigRepo{
				QueryExecer:      mockDB.DB,
				OrganizationRepo: happyCaseMockOrganizationRepo(),
				ConfigAESKey:     "invalidBase64",
			},
			expectErr: errors.Wrap(base64.CorruptInputError(12), "failed to decode key with DecodeBase64()"),
		},
		{
			name: "invalid base64 aes iv",
			tenantConfigRepo: &TenantConfigRepo{
				QueryExecer:      mockDB.DB,
				OrganizationRepo: happyCaseMockOrganizationRepo(),
				ConfigAESKey:     "W4qOy896DmWHg22orCQc2NEM9vQuIVvNuj+TwJDV8J0=",
				ConfigAESIv:      "invalidBase64",
			},
			expectErr: errors.Wrap(base64.CorruptInputError(12), "failed to decode iv with DecodeBase64()"),
		},
		{
			name: "happy case",
			tenantConfigRepo: &TenantConfigRepo{
				QueryExecer:      mockDB.DB,
				ConfigAESKey:     "W4qOy896DmWHg22orCQc2NEM9vQuIVvNuj+TwJDV8J0=",
				ConfigAESIv:      "2/Ukd9Ue2ci6uRB5g3qPSA==",
				OrganizationRepo: happyCaseMockOrganizationRepo(),
			},
			expectErr: nil,
		},
	}

	for index, testcase := range testCases {
		t.Run(fmt.Sprintf("%s-%d", testcase.name, index), func(t *testing.T) {
			tenantConfig, err := testcase.tenantConfigRepo.GetTenantConfig(ctx, "example-tenant-id")

			if testcase.expectErr == nil {
				assert.Nil(t, err)
				assert.NotNil(t, tenantConfig)
			} else {
				assert.EqualError(t, err, testcase.expectErr.Error())
			}
		})
	}
}

func TestOrganizationRepo_GetByDomainName(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	organizationRepo := &OrganizationRepo{}
	mockDB := testutil.NewMockDB()
	rows := &mock_database.Rows{}
	organization := &entity.Organization{}
	_, scanFields := organization.FieldMap()

	testCases := []struct {
		name      string
		setup     func()
		expectErr error
	}{
		{
			name:      "happy case",
			expectErr: nil,
			setup: func() {
				mockDB.DB.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Scan", scanFields...).Once().Return(nil)
			},
		},
		{
			name:      "query return no rows err",
			expectErr: pgx.ErrNoRows,
			setup: func() {
				mockDB.DB.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Scan", scanFields...).Once().Return(pgx.ErrNoRows)
			},
		},
	}

	for index, testcase := range testCases {
		t.Run(fmt.Sprintf("%s-%d", testcase.name, index), func(t *testing.T) {
			testcase.setup()
			org, err := organizationRepo.GetByDomainName(ctx, mockDB.DB, "manabie")
			if err != nil {
				assert.EqualError(t, err, testcase.expectErr.Error())
			} else {
				assert.NotNil(t, org)
			}

			mock.AssertExpectationsForObjects(t, mockDB.DB, rows)
		})
	}
}

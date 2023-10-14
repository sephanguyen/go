package repositories

import (
	"context"
	"strings"
	"testing"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/mock"
)

func TestCreateSchool(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	row := &mock_database.Row{}

	school := &entities.School{}
	school.Name = database.Text("name")
	school.Country = database.Text("COUNTRY_VN")
	school.CityID = database.Int4(1)
	school.DistrictID = database.Int4(1)
	school.Point = pgtype.Point{
		P: pgtype.Vec2{
			X: 1.0,
			Y: 1.0,
		},
	}
	school.IsSystemSchool = pgtype.Bool{Bool: true, Status: pgtype.Present}

	row.On("Scan", mock.Anything).Return(nil)
	db.On(
		"QueryRow",
		mock.Anything,
		mock.Anything,
		&school.Name,
		&school.Country,
		&school.CityID,
		&school.DistrictID,
		&school.Point,
		&school.IsSystemSchool,
		mock.AnythingOfType("*pgtype.Timestamptz"),
		mock.AnythingOfType("*pgtype.Timestamptz"),
	).Return(row)

	schoolRepo := &SchoolRepo{}
	schoolRepo.Create(context.Background(), db, school)

	query := db.Calls[0].Arguments[1].(string)
	if strings.Contains(query, "point =") || strings.Contains(query, "is_system_school =") {
		t.Errorf("SchoolRepo.Create query must not update point and is_system_school fields.")
	}
}

func TestUpsertSchool(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	row := &mock_database.Row{}

	school := &entities.School{}
	school.Name = database.Text("name")
	school.Country = database.Text("COUNTRY_VN")
	school.CityID = database.Int4(1)
	school.DistrictID = database.Int4(1)
	school.Point = pgtype.Point{
		P: pgtype.Vec2{
			X: 1.0,
			Y: 1.0,
		},
	}
	school.IsSystemSchool = pgtype.Bool{Bool: true, Status: pgtype.Present}

	row.On("Scan", mock.Anything).Return(nil)
	db.On(
		"QueryRow",
		mock.Anything,
		mock.Anything,
		&school.Name,
		&school.Country,
		&school.CityID,
		&school.DistrictID,
		&school.Point,
		&school.IsSystemSchool,
		mock.Anything,
		mock.Anything,
	).Return(row)

	schoolRepo := &SchoolRepo{}
	schoolRepo.Upsert(context.Background(), db, school)

	query := db.Calls[0].Arguments[1].(string)
	if !strings.Contains(query, "point =") || !strings.Contains(query, "is_system_school =") {
		t.Errorf("SchoolRepo.Upsert query must update point and is_system_school fields.")
	}
}

func SchoolRepoWithSqlMock() (*SchoolRepo, *testutil.MockDB) {
	r := &SchoolRepo{}
	return r, testutil.NewMockDB()
}

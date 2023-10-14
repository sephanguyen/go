package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgtype"
	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"

	"go.uber.org/multierr"
)

type SchoolRepo struct{}

func (s *SchoolRepo) Get(ctx context.Context, db database.QueryExecer, schoolIDs []int32) (map[int32]*entities.School, error) {
	// var schools []*entities.SchoolExpand

	e := entities.School{}

	schools := entities.Schools{}

	fieldNames := database.GetFieldNames(&e)
	stmt := "SELECT %s FROM %s WHERE school_id = ANY($1)"
	query := fmt.Sprintf(stmt, strings.Join(fieldNames, ", "), e.TableName())
	err := database.Select(ctx, db, query, database.Int4Array(schoolIDs)).ScanAll(&schools)
	if err != nil {
		return nil, err
	}
	mSchools := make(map[int32]*entities.School)
	if len(schools) > 0 {
		for _, school := range schools {
			mSchools[school.ID.Int] = school
		}
	}
	return mSchools, nil
}

//Create create school
func (s *SchoolRepo) Create(ctx context.Context, db database.QueryExecer, school *entities.School) error {
	now := time.Now()
	school.CreatedAt.Set(now)
	school.UpdatedAt.Set(now)

	fieldNames := []string{"name", "country", "city_id", "district_id", "point", "is_system_school", "created_at", "updated_at"}
	placeHolders := "$1, $2, $3, $4, $5, $6, $7, $8"
	args := database.GetScanFields(school, fieldNames)

	query := fmt.Sprintf("INSERT INTO schools (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT school_un DO UPDATE SET updated_at = $8 RETURNING school_id;", strings.Join(fieldNames, ","), placeHolders)

	var id pgtype.Int4
	if err := db.QueryRow(ctx, query, args...).Scan(&id); err != nil {
		return fmt.Errorf("db.QueryRow: %w", err)
	}

	school.ID = id
	return nil
}

//Update update school
func (s *SchoolRepo) Update(ctx context.Context, db database.QueryExecer, school *entities.School) (*entities.School, error) {
	now := time.Now()
	err := multierr.Combine(
		school.UpdatedAt.Set(now),
	)
	if err != nil {
		return nil, fmt.Errorf("multierr.Combine: %w", err)
	}
	query := `UPDATE schools SET updated_at = $1, point = $2, name = $3,
		 phone_number = $4, country = $5, city_id = $6, district_id = $7 
		 WHERE school_id = $8`
	_, err = db.Exec(ctx, query, school.UpdatedAt, school.Point, school.Name,
		school.PhoneNumber, school.Country, school.CityID, school.DistrictID,
		school.ID)
	if err != nil {
		return nil, err
	}
	return school, err
}

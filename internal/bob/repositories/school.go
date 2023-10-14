package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
)

// SchoolRepo repository
type SchoolRepo struct{}

func (r *SchoolRepo) Import(ctx context.Context, db database.QueryExecer, ss []*entities.School) error {
	insertCity := func(c *entities.City) (pgtype.Int4, error) {
		fieldNames := []string{"name", "country", "created_at", "updated_at"}
		placeHolders := "$1, $2, $3, $4"

		query := fmt.Sprintf("INSERT INTO cities (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT city_un DO UPDATE SET updated_at = $4 RETURNING city_id;", strings.Join(fieldNames, ","), placeHolders)
		args := database.GetScanFields(c, fieldNames)

		var id pgtype.Int4
		if err := db.QueryRow(ctx, query, args...).Scan(&id); err != nil {
			return id, errors.Wrap(err, "r.Wrapper.QueryRowEx")
		}
		return id, nil
	}

	insertDistrict := func(d *entities.District) (pgtype.Int4, error) {
		fieldNames := []string{"name", "country", "city_id", "created_at", "updated_at"}
		placeHolders := "$1, $2, $3, $4, $5"

		query := fmt.Sprintf("INSERT INTO districts (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT district_un DO UPDATE SET updated_at = $5 RETURNING district_id;", strings.Join(fieldNames, ","), placeHolders)
		args := database.GetScanFields(d, fieldNames)

		var id pgtype.Int4
		if err := db.QueryRow(ctx, query, args...).Scan(&id); err != nil {
			return id, errors.Wrap(err, "r.Wrapper.QueryRowEx")
		}
		return id, nil
	}

	now := time.Now()
	cities := make(map[string]pgtype.Int4)
	districts := make(map[string]pgtype.Int4)

	var err error
	for _, s := range ss {
		// city key = country + city name
		cityKey := s.City.Country.String + s.City.Name.String
		cityID, ok := cities[cityKey]
		if !ok {
			s.City.CreatedAt.Set(now)
			s.City.UpdatedAt.Set(now)

			cityID, err = insertCity(s.City)
			if err != nil {
				return errors.Wrapf(err, "insertCity: %v", s.City)
			}
			cities[cityKey] = cityID
		}

		// district key = country + city + district name
		districtKey := s.District.Country.String + s.District.City.Name.String + s.District.Name.String
		districtID, ok := districts[districtKey]
		if !ok {
			s.District.CityID = cityID
			s.District.CreatedAt.Set(now)
			s.District.UpdatedAt.Set(now)

			districtID, err = insertDistrict(s.District)
			if err != nil {
				return errors.Wrapf(err, "insertDistrict: %v", s.District)
			}
			districts[districtKey] = districtID
		}

		s.CityID = cityID
		s.DistrictID = districtID
		s.CreatedAt.Set(now)
		s.UpdatedAt.Set(now)
		if err = r.Upsert(ctx, db, s); err != nil {
			return errors.Wrap(err, "r.Create")
		}
	}
	return nil
}

func (r *SchoolRepo) createOrUpsert(ctx context.Context, db database.QueryExecer, s *entities.School, isCreate bool) error {
	now := time.Now()
	s.CreatedAt.Set(now)
	s.UpdatedAt.Set(now)

	fieldNames := []string{"name", "country", "city_id", "district_id", "point", "is_system_school", "created_at", "updated_at"}
	placeHolders := "$1, $2, $3, $4, $5, $6, $7, $8"
	args := database.GetScanFields(s, fieldNames)

	var query string
	if isCreate {
		// postgres needs to update at least 1 field to let returning feature works, so we update updated_at field here.
		query = fmt.Sprintf("INSERT INTO schools (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT school_un DO UPDATE SET updated_at = $8 RETURNING school_id;", strings.Join(fieldNames, ","), placeHolders)
	} else {
		query = fmt.Sprintf("INSERT INTO schools (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT school_un DO UPDATE SET point = $5::point, is_system_school = $6, updated_at = $8 RETURNING school_id;", strings.Join(fieldNames, ","), placeHolders)
	}

	var id pgtype.Int4
	if err := db.QueryRow(ctx, query, args...).Scan(&id); err != nil {
		return errors.Wrap(err, "db.QueryRow")
	}

	s.ID = id
	return nil
}

// Upsert inserts a new school, and updates point, is_system_school and updated_at fields
// if the school already exists.
func (r *SchoolRepo) Upsert(ctx context.Context, db database.QueryExecer, s *entities.School) error {
	return r.createOrUpsert(ctx, db, s, false)
}

// Create inserts a new school, and only updates updated_at field if the school already exists.
func (r *SchoolRepo) Create(ctx context.Context, db database.QueryExecer, s *entities.School) error {
	return r.createOrUpsert(ctx, db, s, true)
}

func (r *SchoolRepo) RetrieveDistricts(ctx context.Context, db database.QueryExecer, country string, cityID int32) ([]*entities.District, error) {
	fields := database.GetFieldNames(&entities.District{})
	query := fmt.Sprintf("SELECT %s FROM districts WHERE country = $1 AND city_id = $2", strings.Join(fields, ", "))
	rows, err := db.Query(ctx, query, &country, &cityID)
	if err != nil {
		return nil, errors.Wrap(err, "r.Wrapper.QueryEx")
	}
	defer rows.Close()

	var districts []*entities.District
	for rows.Next() {
		d := new(entities.District)
		if err := rows.Scan(database.GetScanFields(d, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		districts = append(districts, d)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return districts, nil
}

func (r *SchoolRepo) RetrieveCountries(ctx context.Context, db database.QueryExecer, schoolIDs pgtype.Int4Array) ([]string, error) {
	query := "SELECT DISTINCT country FROM schools WHERE school_id = ANY($1)"
	rows, err := db.Query(ctx, query, &schoolIDs)
	if err != nil {
		return nil, errors.Wrap(err, "r.Wrapper.QueryEx")
	}
	defer rows.Close()

	var countries []string
	for rows.Next() {
		var country string
		if err := rows.Scan(&country); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		countries = append(countries, country)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return countries, nil
}

func (r *SchoolRepo) FindByID(ctx context.Context, db database.QueryExecer, id pgtype.Int4) (*entities.School, error) {
	ctx, span := interceptors.StartSpan(ctx, "SchoolRepo.Find")
	defer span.End()

	e := new(entities.School)
	fields := database.GetFieldNames(e)
	selectStmt := fmt.Sprintf("SELECT %s FROM %s WHERE school_id = $1", strings.Join(fields, ","), e.TableName())

	row := db.QueryRow(ctx, selectStmt, &id)
	if err := row.Scan(database.GetScanFields(e, fields)...); err != nil {
		return nil, err
	}

	return e, nil
}

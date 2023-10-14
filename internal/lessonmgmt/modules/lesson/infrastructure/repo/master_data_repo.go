package repo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

type Location struct {
	LocationID        pgtype.Text
	Name              pgtype.Text
	PartnerInternalID pgtype.Text
	UpdatedAt         pgtype.Timestamptz
	CreatedAt         pgtype.Timestamptz
	DeletedAt         pgtype.Timestamptz
}

func NewLocationFromEntity(center *domain.Location) (*Location, error) {
	e := &Location{}
	database.AllNullEntity(e)
	if err := multierr.Combine(
		e.LocationID.Set(center.LocationID),
		e.Name.Set(center.Name),
	); err != nil {
		return nil, err
	}
	return e, nil
}

func (l *Location) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"location_id", "name", "partner_internal_id", "updated_at", "created_at", "deleted_at"}
	values = []interface{}{&l.LocationID, &l.Name, &l.PartnerInternalID, &l.UpdatedAt, &l.CreatedAt, &l.DeletedAt}
	return
}

func (*Location) TableName() string {
	return "locations"
}

func (l *Location) ToCenterEntity() *domain.Location {
	return &domain.Location{
		LocationID: l.LocationID.String,
		Name:       l.Name.String,
		CreatedAt:  l.CreatedAt.Time,
		UpdatedAt:  l.UpdatedAt.Time,
	}
}
func (l *Location) PreInsert() error {
	now := time.Now()
	if err := multierr.Combine(
		l.CreatedAt.Set(now),
		l.UpdatedAt.Set(now),
	); err != nil {
		return err
	}
	return nil
}

type Course struct {
	CourseID  pgtype.Text
	Name      pgtype.Text
	UpdatedAt pgtype.Timestamptz
	CreatedAt pgtype.Timestamptz
	DeletedAt pgtype.Timestamptz
}

func (c *Course) ToCourseEntity() *domain.Course {
	return &domain.Course{
		CourseID:  c.CourseID.String,
		Name:      c.Name.String,
		CreatedAt: c.CreatedAt.Time,
		UpdatedAt: c.UpdatedAt.Time,
	}
}

func (c *Course) PreInsert() error {
	now := time.Now()
	if err := multierr.Combine(
		c.CreatedAt.Set(now),
		c.UpdatedAt.Set(now),
	); err != nil {
		return err
	}
	return nil
}
func NewCourseFromEntity(course *domain.Course) (*Course, error) {
	e := &Course{}
	database.AllNullEntity(e)
	if err := multierr.Combine(
		e.CourseID.Set(course.CourseID),
		e.Name.Set(course.Name),
	); err != nil {
		return nil, err
	}
	return e, nil
}

func (c *Course) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"course_id", "name", "updated_at", "created_at", "deleted_at"}
	values = []interface{}{&c.CourseID, &c.Name, &c.UpdatedAt, &c.CreatedAt, &c.DeletedAt}
	return
}

func (*Course) TableName() string {
	return "courses"
}

type CourseTeachingTime struct {
	CourseID        pgtype.Text
	PreparationTime pgtype.Int4
	BreakTime       pgtype.Int4
	UpdatedAt       pgtype.Timestamptz
	CreatedAt       pgtype.Timestamptz
	DeletedAt       pgtype.Timestamptz
}

func (*CourseTeachingTime) TableName() string {
	return "course_teaching_time"
}

func (c *CourseTeachingTime) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"course_id", "preparation_time", "break_time", "updated_at", "created_at", "deleted_at"}
	values = []interface{}{&c.CourseID, &c.PreparationTime, &c.BreakTime, &c.UpdatedAt, &c.CreatedAt, &c.DeletedAt}
	return
}

func (c *CourseTeachingTime) ToCourseEntity() *domain.Course {
	return &domain.Course{
		CourseID:        c.CourseID.String,
		PreparationTime: c.PreparationTime.Int,
		BreakTime:       c.BreakTime.Int,
		CreatedAt:       c.CreatedAt.Time,
		UpdatedAt:       c.UpdatedAt.Time,
	}
}

func NewCourseTeachingTimeFromEntity(course *domain.Course) (*CourseTeachingTime, error) {
	ctt := &CourseTeachingTime{}
	now := time.Now()
	database.AllNullEntity(ctt)
	if err := multierr.Combine(
		ctt.CourseID.Set(course.CourseID),
		ctt.PreparationTime.Set(course.PreparationTime),
		ctt.BreakTime.Set(course.BreakTime),
		ctt.CreatedAt.Set(now),
		ctt.UpdatedAt.Set(now),
	); err != nil {
		return nil, err
	}

	if !course.DeletedAt.IsZero() {
		if err := ctt.DeletedAt.Set(course.DeletedAt); err != nil {
			return nil, err
		}
	}
	return ctt, nil
}

type Class struct {
	ClassID   pgtype.Text
	Name      pgtype.Text
	UpdatedAt pgtype.Timestamptz
	CreatedAt pgtype.Timestamptz
	DeletedAt pgtype.Timestamptz
}

func (c *Class) ToClassEntity() *domain.Class {
	return &domain.Class{
		ClassID:   c.ClassID.String,
		Name:      c.Name.String,
		CreatedAt: c.CreatedAt.Time,
		UpdatedAt: c.UpdatedAt.Time,
	}
}

func (c *Class) PreInsert() error {
	now := time.Now()
	if err := multierr.Combine(
		c.CreatedAt.Set(now),
		c.UpdatedAt.Set(now),
	); err != nil {
		return err
	}
	return nil
}
func NewClassFromEntity(course *domain.Class) (*Class, error) {
	e := &Class{}
	database.AllNullEntity(e)
	if err := multierr.Combine(
		e.ClassID.Set(course.ClassID),
		e.Name.Set(course.Name),
	); err != nil {
		return nil, err
	}
	return e, nil
}

func (c *Class) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"class_id", "name", "updated_at", "created_at", "deleted_at"}
	values = []interface{}{&c.ClassID, &c.Name, &c.UpdatedAt, &c.CreatedAt, &c.DeletedAt}
	return
}

func (*Class) TableName() string {
	return "class"
}

type MasterDataRepo struct{}

func (m *MasterDataRepo) GetLocationByID(ctx context.Context, db database.Ext, id string) (*domain.Location, error) {
	ctx, span := interceptors.StartSpan(ctx, "MasterDataRepo.GetLocationByID")
	defer span.End()

	location := &Location{}
	fields, values := location.FieldMap()
	query := fmt.Sprintf(`
		SELECT %s FROM locations
		WHERE location_id = $1
			AND deleted_at IS NULL`,
		strings.Join(fields, ","),
	)

	err := db.QueryRow(ctx, query, &id).Scan(values...)
	if err != nil {
		return nil, err
	}

	return location.ToCenterEntity(), nil
}

func (m *MasterDataRepo) InsertCenter(ctx context.Context, db database.Ext, center *domain.Location) (*domain.Location, error) {
	ctx, span := interceptors.StartSpan(ctx, "MasterDataRepo.InsertCenter")
	defer span.End()

	dto, err := NewLocationFromEntity(center)
	if err != nil {
		return nil, err
	}
	if err = dto.PreInsert(); err != nil {
		return nil, fmt.Errorf("got error when preinsert centet: %w", err)
	}

	if _, err = database.Insert(ctx, dto, db.Exec); err != nil {
		return nil, err
	}

	return dto.ToCenterEntity(), nil
}

func (m *MasterDataRepo) GetCourseByID(ctx context.Context, db database.Ext, id string) (*domain.Course, error) {
	ctx, span := interceptors.StartSpan(ctx, "MasterDataRepo.GetCourseByID")
	defer span.End()

	course := &Course{}
	fields, values := course.FieldMap()
	query := fmt.Sprintf(`
		SELECT %s FROM courses
		WHERE course_id = $1
			AND deleted_at IS NULL`,
		strings.Join(fields, ","),
	)

	err := db.QueryRow(ctx, query, &id).Scan(values...)
	if err != nil {
		return nil, err
	}

	return course.ToCourseEntity(), nil
}

func (m *MasterDataRepo) GetCourseTeachingTimeByIDs(ctx context.Context, db database.Ext, ids []string) (map[string]*domain.Course, error) {
	ctx, span := interceptors.StartSpan(ctx, "MasterDataRepo.GetCourseTeachingTimeByIDs")
	defer span.End()
	course := &CourseTeachingTime{}
	fields, _ := course.FieldMap()

	query := `SELECT c.course_id, ctt.preparation_time, ctt.break_time, ctt.created_at, ctt.updated_at, ctt.deleted_at
		FROM courses c
		LEFT JOIN course_teaching_time ctt
			ON c.course_id  = ctt.course_id
		WHERE c.course_id = any($1)
			AND c.deleted_at IS NULL
			AND ctt.deleted_at IS NULL`

	rows, err := db.Query(ctx, query, &ids)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	defer rows.Close()

	results := make(map[string]*domain.Course)
	for rows.Next() {
		course := &CourseTeachingTime{}
		if err := rows.Scan(database.GetScanFields(course, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		results[course.CourseID.String] = course.ToCourseEntity()
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return results, nil
}

func (m *MasterDataRepo) GetClassByID(ctx context.Context, db database.Ext, id string) (*domain.Class, error) {
	ctx, span := interceptors.StartSpan(ctx, "MasterDataRepo.GetClassByID")
	defer span.End()

	class := &Class{}
	fields, values := class.FieldMap()
	query := fmt.Sprintf(`
		SELECT %s FROM class
		WHERE class_id = $1
			AND deleted_at IS NULL`,
		strings.Join(fields, ","),
	)

	err := db.QueryRow(ctx, query, &id).Scan(values...)
	if err != nil {
		return nil, err
	}

	return class.ToClassEntity(), nil
}

func (m *MasterDataRepo) GetLowestLocationsByPartnerInternalIDs(ctx context.Context, db database.Ext, ids []string) (map[string]*domain.Location, error) {
	ctx, span := interceptors.StartSpan(ctx, "MasterDataRepo.GetLocationByPartnerInternalID")
	defer span.End()

	location := &Location{}
	fields, _ := location.FieldMap()
	query := fmt.Sprintf(`
		SELECT %s FROM locations
		WHERE location_type NOT IN (
			SELECT DISTINCT parent_location_type_id FROM location_types 
				WHERE parent_location_type_id IS NOT NULL AND deleted_at IS NULL AND is_archived IS FALSE)
		AND parent_location_id IS NOT NULL
		AND partner_internal_id = any($1)
		AND deleted_at IS NULL
		AND is_archived IS FALSE`,
		strings.Join(fields, ","),
	)

	rows, err := db.Query(ctx, query, &ids)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	defer rows.Close()

	centerByPartnerID := make(map[string]*domain.Location)
	for rows.Next() {
		location := &Location{}
		if err := rows.Scan(database.GetScanFields(location, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		centerByPartnerID[location.PartnerInternalID.String] = location.ToCenterEntity()
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return centerByPartnerID, nil
}

func (m *MasterDataRepo) FindPermissionByNamesAndUserID(ctx context.Context, db database.QueryExecer, permissionNames []string, userID string) (*domain.UserPermissions, error) {
	ctx, span := interceptors.StartSpan(ctx, "MasterDataRepo.FindPermissionByNameAndUserID")
	defer span.End()

	gp := &GrantedPermission{}

	fields := database.GetFieldNames(gp)

	query := fmt.Sprintf(`SELECT %s FROM %s
		WHERE user_id = $1
		AND permission_name = any($2)`, strings.Join(fields, ","), gp.TableName())

	rows, err := db.Query(ctx, query, userID, permissionNames)
	if err != nil {
		return nil, fmt.Errorf("db.Query %w", err)
	}
	defer rows.Close()
	permissions := []*GrantedPermission{}
	for rows.Next() {
		p := &GrantedPermission{}
		if err := rows.Scan(database.GetScanFields(p, fields)...); err != nil {
			return nil, errors.Wrap(err, "row.Scan")
		}
		permissions = append(permissions, p)
	}

	return ToUserPermissionDomain(permissions), nil
}

func (m *MasterDataRepo) CheckLocationByIDs(ctx context.Context, db database.Ext, ids []string, locationName map[string]string) error {
	ctx, span := interceptors.StartSpan(ctx, "MasterDataRepo.CheckLocationByIDs")
	defer span.End()

	l := Location{}
	fields, _ := l.FieldMap()

	query := fmt.Sprintf(`SELECT %s FROM %s WHERE location_id = ANY($1) AND deleted_at IS NULL`,
		strings.Join(fields, ","), l.TableName())

	rows, err := db.Query(ctx, query, &ids)
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.ErrNotFound
		}
		return err
	}
	defer rows.Close()

	locations := []*Location{}
	for rows.Next() {
		location := &Location{}
		if err := rows.Scan(database.GetScanFields(location, fields)...); err != nil {
			return errors.Wrap(err, "rows.Scan")
		}

		if len(locationName) > 0 {
			locationID := location.LocationID.String
			lname, ok := locationName[locationID]
			if !ok {
				return fmt.Errorf("cannot find name of location %s", locationID)
			}
			if lname != location.Name.String {
				return fmt.Errorf("expected name %s but only got %s", location.Name.String, lname)
			}
		}
		locations = append(locations, location)
	}
	if len(locations) != len(ids) {
		return fmt.Errorf("received location IDs %v but only found %v", ids, locations)
	}

	return nil
}

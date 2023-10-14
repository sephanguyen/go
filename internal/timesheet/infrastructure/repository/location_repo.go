package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/timesheet/domain/entity"

	"github.com/jackc/pgtype"
)

type LocationRepoImpl struct{}

func (r *LocationRepoImpl) GetGrantedLocationOfStaff(ctx context.Context, db database.QueryExecer, staffID pgtype.Text) ([]*entity.Location, error) {
	ctx, span := interceptors.StartSpan(ctx, "LocationRepoImpl.GetGrantedLocationOfStaff")
	defer span.End()

	listLocation := &entity.Locations{}
	location := &entity.Location{}
	fields, _ := location.FieldMap()
	convertFields := []string{}
	for _, field := range fields {
		convertFields = append(convertFields, "lo."+field)
	}

	stmt := fmt.Sprintf(`
		SELECT %s FROM user_group_member ugm
		INNER JOIN granted_role gr ON ugm.user_group_id = gr.user_group_id AND ugm.deleted_at IS NULL
		INNER JOIN granted_role_access_path grap ON gr.granted_role_id  = grap.granted_role_id AND grap.deleted_at IS NULL
		INNER JOIN %s lo on grap.location_id = lo.location_id AND lo.deleted_at IS NULL WHERE ugm.user_id = $1
		AND gr.deleted_at IS NULL;`,
		strings.Join(convertFields, ", "), location.TableName())

	if err := database.Select(ctx, db, stmt, &staffID).ScanAll(listLocation); err != nil {
		return nil, err
	}

	return *listLocation, nil
}

func (r *LocationRepoImpl) GetListChildLocations(ctx context.Context, db database.QueryExecer, parentLocationID pgtype.Text) ([]*entity.Location, error) {
	ctx, span := interceptors.StartSpan(ctx, "LocationRepoImpl.GetListChildLocations")
	defer span.End()

	listLocation := &entity.Locations{}
	location := &entity.Location{}
	fields, _ := location.FieldMap()

	stmt := fmt.Sprintf(`
		SELECT %s FROM %s lo where
		lo.parent_location_id = $1 
		AND lo.deleted_at IS NULL`, strings.Join(fields, ","), location.TableName())

	if err := database.Select(ctx, db, stmt, &parentLocationID).ScanAll(listLocation); err != nil {
		return nil, err
	}

	return *listLocation, nil
}

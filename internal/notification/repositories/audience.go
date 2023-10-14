package repositories

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/notification/consts"
	"github.com/manabie-com/backend/internal/notification/entities"

	"github.com/jackc/pgtype"
	"golang.org/x/exp/slices"
)

type AudienceRepo struct {
	AudienceSQLBuilder AudienceSQLBuilder
}

type FindGroupAudienceFilter struct {
	GradeIDs         pgtype.TextArray
	GradeSelectType  pgtype.Text
	LocationIDs      pgtype.TextArray
	CourseIDs        pgtype.TextArray
	CourseSelectType pgtype.Text
	ClassIDs         pgtype.TextArray
	ClassSelectType  pgtype.Text
	SchoolIDs        pgtype.TextArray
	SchoolSelectType pgtype.Text

	UserGroups     pgtype.TextArray
	Keyword        pgtype.Text
	Limit          pgtype.Int8
	Offset         pgtype.Int8
	IncludeUserIds pgtype.TextArray
	ExcludeUserIds pgtype.TextArray

	StudentEnrollmentStatus pgtype.Text
}

func NewFindGroupAudienceFilter() *FindGroupAudienceFilter {
	f := &FindGroupAudienceFilter{}
	_ = f.GradeIDs.Set(nil)
	_ = f.GradeSelectType.Set(nil)
	_ = f.LocationIDs.Set(nil)
	_ = f.CourseIDs.Set(nil)
	_ = f.CourseSelectType.Set(nil)
	_ = f.ClassIDs.Set(nil)
	_ = f.ClassSelectType.Set(nil)
	_ = f.SchoolIDs.Set(nil)
	_ = f.SchoolSelectType.Set(nil)

	_ = f.UserGroups.Set(nil)
	_ = f.Keyword.Set(nil)
	_ = f.Limit.Set(nil)
	_ = f.Offset.Set(nil)
	_ = f.IncludeUserIds.Set(nil)
	_ = f.ExcludeUserIds.Set(nil)
	_ = f.StudentEnrollmentStatus.Set(nil)
	return f
}

func (f *FindGroupAudienceFilter) IsNull() bool {
	if f.GradeIDs.Status == pgtype.Null &&
		f.GradeSelectType.Status == pgtype.Null &&
		f.LocationIDs.Status == pgtype.Null &&
		f.CourseIDs.Status == pgtype.Null &&
		f.CourseSelectType.Status == pgtype.Null &&
		f.ClassIDs.Status == pgtype.Null &&
		f.ClassSelectType.Status == pgtype.Null &&

		f.UserGroups.Status == pgtype.Null &&
		f.Keyword.Status == pgtype.Null &&
		f.Limit.Status == pgtype.Null &&
		f.Offset.Status == pgtype.Null &&
		f.IncludeUserIds.Status == pgtype.Null &&
		f.ExcludeUserIds.Status == pgtype.Null {
		return true
	}
	return false
}

type FindAudienceOption struct {
	OrderByName string
	IsGetName   bool
}

func NewFindAudienceOption() *FindAudienceOption {
	f := &FindAudienceOption{
		OrderByName: consts.DefaultOrder,
		IsGetName:   false,
	}
	return f
}

func (f *FindAudienceOption) Validate() error {
	orderValues := []string{consts.DefaultOrder, consts.AscendingOrder, consts.DescendingOrder}
	if !slices.Contains(orderValues, f.OrderByName) {
		return fmt.Errorf("value of OrderByName need to be one of %v", orderValues)
	}

	return nil
}

// For Audience Selector - RetrieveAudienceGroup
// This function support query for both Student and Parent
// We will replace FindGroupAudiencesByFilter with this function when FE team has finished coding
// their new UI (move User Group to Target Group, load student + parent data into Individual Target)
// the worst case when run this function is it have no filters (locations, courses, grades, classes)
// => in reality we already avoid this case by using check CheckNoneSelectTargetGroup
// nolint
func (repo *AudienceRepo) FindGroupAudiencesByFilter(ctx context.Context, db database.QueryExecer, filter *FindGroupAudienceFilter, opts *FindAudienceOption) ([]*entities.Audience, error) {
	ctx, span := interceptors.StartSpan(ctx, "AudienceRepo.FindGroupAudiencesByFilter")
	defer span.End()

	err := opts.Validate()
	if err != nil {
		return nil, fmt.Errorf("error validate options: %v", err)
	}

	audiences := []*entities.Audience{}
	args := []interface{}{}
	argIdx := 1

	query, _ := repo.AudienceSQLBuilder.BuildFindGroupAudiencesByFilterSQL(filter, opts, &argIdx, &args)

	// paging
	if filter.Limit.Status == pgtype.Present && filter.Offset.Status == pgtype.Present {
		query += fmt.Sprintf(` LIMIT $%d OFFSET $%d`, argIdx, argIdx+1)
		args = append(args, filter.Limit, filter.Offset)
	}

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		item := &entities.Audience{}
		if opts.OrderByName != consts.DefaultOrder || opts.IsGetName || filter.Keyword.Status == pgtype.Present {
			err = rows.Scan(&item.UserID, &item.StudentID, &item.ParentID, &item.GradeID, &item.ChildIDs, &item.UserGroup, &item.IsIndividual, &item.Name, &item.Email)
		} else {
			err = rows.Scan(&item.UserID, &item.StudentID, &item.ParentID, &item.GradeID, &item.ChildIDs, &item.UserGroup, &item.IsIndividual)
		}
		if err != nil {
			return nil, err
		}
		audiences = append(audiences, item)
	}
	return audiences, nil
}

func (repo *AudienceRepo) CountGroupAudiencesByFilter(ctx context.Context, db database.QueryExecer, filter *FindGroupAudienceFilter, opts *FindAudienceOption) (uint32, error) {
	if filter.Limit.Status != pgtype.Present || filter.Offset.Status != pgtype.Present {
		return 0, nil
	}
	ctx, span := interceptors.StartSpan(ctx, "AudienceRepo.CountGroupAudiencesByFilter")
	defer span.End()

	err := opts.Validate()
	if err != nil {
		return 0, fmt.Errorf("error validate options: %v", err)
	}

	args := []interface{}{}
	argIdx := 1

	_, queryWithoutOrder := repo.AudienceSQLBuilder.BuildFindGroupAudiencesByFilterSQL(filter, opts, &argIdx, &args)

	totalAudiences := 0
	// query count
	countQuery := `
		SELECT COUNT(*) 
		FROM (
			` + queryWithoutOrder + `
		) AS temp_audiences
	`
	err = db.QueryRow(ctx, countQuery, args...).Scan(&totalAudiences)
	if err != nil {
		return uint32(0), err
	}
	return uint32(totalAudiences), nil
}

type FindIndividualAudienceFilter struct {
	LocationIDs        pgtype.TextArray
	UserIDs            pgtype.TextArray
	EnrollmentStatuses pgtype.TextArray
}

func NewFindIndividualAudienceFilter() *FindIndividualAudienceFilter {
	f := &FindIndividualAudienceFilter{}
	_ = f.UserIDs.Set(nil)
	_ = f.LocationIDs.Set(nil)
	_ = f.EnrollmentStatuses.Set(nil)
	return f
}

func (f *FindIndividualAudienceFilter) IsNull() bool {
	if f.UserIDs.Status == pgtype.Null &&
		f.LocationIDs.Status == pgtype.Null &&
		f.EnrollmentStatuses.Status == pgtype.Null {
		return true
	}
	return false
}

func (repo *AudienceRepo) FindIndividualAudiencesByFilter(ctx context.Context, db database.QueryExecer, filter *FindIndividualAudienceFilter) ([]*entities.Audience, error) {
	ctx, span := interceptors.StartSpan(ctx, "AudienceRepo.FindIndividualAudiencesByFilter")
	defer span.End()

	argIdx := 1
	args := []interface{}{}

	query := repo.AudienceSQLBuilder.BuildFindIndividualAudiencesByFilterSQL(filter, NewFindAudienceOption(), &argIdx, &args)

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	audiences := []*entities.Audience{}
	for rows.Next() {
		item := &entities.Audience{}
		err = rows.Scan(&item.UserID, &item.StudentID, &item.ParentID, &item.GradeID, &item.ChildIDs, &item.UserGroup, &item.IsIndividual)
		if err != nil {
			return nil, err
		}
		audiences = append(audiences, item)
	}

	return audiences, nil
}

type FindDraftAudienceFilter struct {
	GroupFilter      *FindGroupAudienceFilter
	IndividualFilter *FindIndividualAudienceFilter
	Limit            pgtype.Int8
	Offset           pgtype.Int8
}

func NewFindDraftAudienceFilter() *FindDraftAudienceFilter {
	f := &FindDraftAudienceFilter{}
	f.GroupFilter = NewFindGroupAudienceFilter()
	f.IndividualFilter = NewFindIndividualAudienceFilter()
	_ = f.Limit.Set(nil)
	_ = f.Offset.Set(nil)
	return f
}

func (filter *FindDraftAudienceFilter) Validate() error {
	if filter.GroupFilter == nil &&
		filter.IndividualFilter == nil &&
		filter.Limit.Status != pgtype.Present &&
		filter.Offset.Status != pgtype.Present {
		return fmt.Errorf("FindDraftAudienceFilter filter must be not null")
	}
	return nil
}

func (repo *AudienceRepo) buildFindDraftAudienceQuery(filter *FindDraftAudienceFilter, opts *FindAudienceOption) (string, []interface{}, int) {
	argIdx := 1
	args := []interface{}{}

	sqlBuilderOpts := NewFindAudienceOption()
	if opts.IsGetName || opts.OrderByName != consts.DefaultOrder {
		sqlBuilderOpts.IsGetName = opts.IsGetName
	}

	isNullGroupFilter := filter.GroupFilter.IsNull()
	isNullIndividualFilter := filter.IndividualFilter.IsNull()

	individualQuery, groupQueryWithoutOrder := "", ""
	if !isNullGroupFilter {
		_, groupQueryWithoutOrder = repo.AudienceSQLBuilder.BuildFindGroupAudiencesByFilterSQL(filter.GroupFilter, sqlBuilderOpts, &argIdx, &args)
	}

	if !isNullIndividualFilter {
		individualQuery = repo.AudienceSQLBuilder.BuildFindIndividualAudiencesByFilterSQL(filter.IndividualFilter, sqlBuilderOpts, &argIdx, &args)
	}

	query := ""
	// nolint: gocritic
	if !isNullGroupFilter && !isNullIndividualFilter {
		query = `
			WITH group_audiences AS ( 
				` + groupQueryWithoutOrder + ` 
			),  individual_audiences AS ( 
				` + individualQuery + ` 
			) (
				SELECT * 
				FROM group_audiences
			) UNION (
				SELECT * 
				FROM individual_audiences ia
				WHERE ia.user_id != ALL (
					SELECT user_id 
					FROM group_audiences
				)
			)
		`
	} else if !isNullGroupFilter {
		query = `
			WITH group_audiences AS ( 
				` + groupQueryWithoutOrder + ` 
			)
			SELECT * 
			FROM group_audiences
		`
	} else if !isNullIndividualFilter {
		query = `
			WITH individual_audiences AS ( 
				` + individualQuery + ` 
			)
			SELECT * 
			FROM individual_audiences
		`
	}

	return query, args, argIdx
}

func (repo *AudienceRepo) CountDraftAudiencesByFilter(ctx context.Context, db database.QueryExecer, filter *FindDraftAudienceFilter, opts *FindAudienceOption) (uint32, error) {
	if filter.Limit.Status != pgtype.Present || filter.Offset.Status != pgtype.Present {
		return uint32(0), nil
	}

	ctx, span := interceptors.StartSpan(ctx, "AudienceRepo.CountDraftAudiencesByFilter")
	defer span.End()

	var totalAudiences int

	if err := filter.Validate(); err != nil {
		return uint32(0), fmt.Errorf("error on ValidateFindDraftAudienceFilter: %v", err)
	}

	query, args, _ := repo.buildFindDraftAudienceQuery(filter, opts)

	// query count
	countQuery := `
		SELECT COUNT(*) 
		FROM (
			` + query + `
		) AS temp_audiences
	`
	err := db.QueryRow(ctx, countQuery, args...).Scan(&totalAudiences)
	if err != nil {
		return uint32(0), err
	}

	return uint32(totalAudiences), nil
}

func (repo *AudienceRepo) FindDraftAudiencesByFilter(ctx context.Context, db database.QueryExecer, filter *FindDraftAudienceFilter, opts *FindAudienceOption) ([]*entities.Audience, error) {
	ctx, span := interceptors.StartSpan(ctx, "AudienceRepo.FindDraftAudiencesByFilter")
	defer span.End()

	audiences := []*entities.Audience{}

	if err := filter.Validate(); err != nil {
		return nil, fmt.Errorf("error on ValidateFindDraftAudienceFilter: %v", err)
	}

	query, args, argIdx := repo.buildFindDraftAudienceQuery(filter, opts)

	// order by name
	if opts.OrderByName != consts.DefaultOrder {
		query += fmt.Sprintf(` ORDER BY name %s `, opts.OrderByName)
	}

	// paging
	if filter.Limit.Status == pgtype.Present && filter.Offset.Status == pgtype.Present {
		query += fmt.Sprintf(` LIMIT $%d OFFSET $%d`, argIdx, argIdx+1)
		args = append(args, filter.Limit, filter.Offset)
	}

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		item := &entities.Audience{}
		if opts.IsGetName || opts.OrderByName != consts.DefaultOrder {
			err = rows.Scan(&item.UserID, &item.StudentID, &item.ParentID, &item.GradeID, &item.ChildIDs, &item.UserGroup, &item.IsIndividual, &item.Name, &item.Email)
		} else {
			err = rows.Scan(&item.UserID, &item.StudentID, &item.ParentID, &item.GradeID, &item.ChildIDs, &item.UserGroup, &item.IsIndividual)
		}
		if err != nil {
			return nil, err
		}
		audiences = append(audiences, item)
	}

	return audiences, nil
}

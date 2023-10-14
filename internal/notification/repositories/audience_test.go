package repositories

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/notification/consts"
	"github.com/manabie-com/backend/internal/notification/entities"
	"github.com/manabie-com/backend/mock/testutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pbu "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// common variables for testing
var (
	courseIDs        = []string{"course-id-1", "course-id-2"}
	classIDs         = []string{"class-id-1", "class-id-2"}
	locationIDs      = []string{"location-id-1", "location-id-2"}
	studentIDs       = []string{"student-id-1", "student-id-2"}
	studentNames     = []string{"student-name-1", "student-name-2"}
	studentEmails    = []string{"student-email-1", "student-email-2"}
	parentIDs        = []string{"parent-id-1", "parent-id-2"}
	parentNames      = []string{"parent-name-1", "parent-name-2"}
	parentEmails     = []string{"parent-email-1", "parent-email-2"}
	userGroups       = []string{"USER_GROUP_STUDENT", "USER_GROUP_PARENT"}
	gradeIDs         = []string{"grade-id-1", "grade-id-2"}
	schoolIDs        = []string{"school-id-1", "school-id-2"}
	enrollmentStatus = pbu.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()
)

// common functions for testing
var makeFindGroupAudienceFilter = func(locationSelect, courseSelect, classSelect, gradeSelect, schoolSelect, userGroup, keyword string, offset, limit int8) *FindGroupAudienceFilter {
	filter := NewFindGroupAudienceFilter()
	switch locationSelect {
	case "all":
		_ = filter.LocationIDs.Set(locationIDs)
	case "list":
		_ = filter.LocationIDs.Set(locationIDs)
	}

	switch courseSelect {
	case "none":
		_ = filter.CourseSelectType.Set(consts.TargetGroupSelectTypeNone.String())
	case "all":
		_ = filter.CourseSelectType.Set(consts.TargetGroupSelectTypeAll.String())
	case "list":
		_ = filter.CourseIDs.Set(courseIDs)
		_ = filter.CourseSelectType.Set(consts.TargetGroupSelectTypeList.String())
	}

	switch classSelect {
	case "none":
		_ = filter.ClassSelectType.Set(consts.TargetGroupSelectTypeNone.String())
	case "all":
		_ = filter.ClassSelectType.Set(consts.TargetGroupSelectTypeAll.String())
	case "list":
		_ = filter.ClassIDs.Set(classIDs)
		_ = filter.ClassSelectType.Set(consts.TargetGroupSelectTypeList.String())
	}

	switch gradeSelect {
	case "none":
		_ = filter.GradeSelectType.Set(consts.TargetGroupSelectTypeNone.String())
	case "all":
		_ = filter.GradeSelectType.Set(consts.TargetGroupSelectTypeAll.String())
	case "list":
		_ = filter.GradeIDs.Set(gradeIDs)
		_ = filter.GradeSelectType.Set(consts.TargetGroupSelectTypeList.String())
	}

	switch schoolSelect {
	case "none":
		_ = filter.SchoolSelectType.Set(consts.TargetGroupSelectTypeNone.String())
	case "all":
		_ = filter.SchoolSelectType.Set(consts.TargetGroupSelectTypeAll.String())
	case "list":
		_ = filter.SchoolIDs.Set(schoolIDs)
		_ = filter.SchoolSelectType.Set(consts.TargetGroupSelectTypeList.String())
	}

	switch userGroup {
	case "student":
		_ = filter.UserGroups.Set([]string{cpb.UserGroup_USER_GROUP_STUDENT.String()})
	case "parent":
		_ = filter.UserGroups.Set([]string{cpb.UserGroup_USER_GROUP_PARENT.String()})
	case "student,parent":
		_ = filter.UserGroups.Set([]string{cpb.UserGroup_USER_GROUP_STUDENT.String(), cpb.UserGroup_USER_GROUP_PARENT.String()})
	default:
		_ = filter.UserGroups.Set(nil)
	}

	_ = filter.Offset.Set(offset)
	_ = filter.Limit.Set(limit)

	if keyword != "" {
		_ = filter.Keyword.Set(keyword)
	}

	_ = filter.StudentEnrollmentStatus.Set(enrollmentStatus)

	return filter
}

var makeFindDraftAudienceFilter = func(locationSelect, courseSelect, classSelect, gradeSelect, userGroup, keyword, userID, enrollment string, offset, limit int8) *FindDraftAudienceFilter {
	filter := NewFindDraftAudienceFilter()
	switch locationSelect {
	case "all":
		_ = filter.GroupFilter.LocationIDs.Set(locationIDs)
		_ = filter.IndividualFilter.LocationIDs.Set(locationIDs)
	case "list":
		_ = filter.GroupFilter.LocationIDs.Set(locationIDs)
	}

	switch courseSelect {
	case "none":
		_ = filter.GroupFilter.CourseSelectType.Set(consts.TargetGroupSelectTypeNone.String())
	case "all":
		_ = filter.GroupFilter.CourseSelectType.Set(consts.TargetGroupSelectTypeAll.String())
	case "list":
		_ = filter.GroupFilter.CourseIDs.Set(courseIDs)
		_ = filter.GroupFilter.CourseSelectType.Set(consts.TargetGroupSelectTypeList.String())
	}

	switch classSelect {
	case "none":
		_ = filter.GroupFilter.ClassSelectType.Set(consts.TargetGroupSelectTypeNone.String())
	case "all":
		_ = filter.GroupFilter.ClassSelectType.Set(consts.TargetGroupSelectTypeAll.String())
	case "list":
		_ = filter.GroupFilter.ClassIDs.Set(classIDs)
		_ = filter.GroupFilter.ClassSelectType.Set(consts.TargetGroupSelectTypeList.String())
	}

	switch gradeSelect {
	case "none":
		_ = filter.GroupFilter.GradeSelectType.Set(consts.TargetGroupSelectTypeNone.String())
	case "all":
		_ = filter.GroupFilter.GradeSelectType.Set(consts.TargetGroupSelectTypeAll.String())
	case "list":
		_ = filter.GroupFilter.GradeIDs.Set(gradeIDs)
		_ = filter.GroupFilter.GradeSelectType.Set(consts.TargetGroupSelectTypeList.String())
	}

	switch userGroup {
	case "student":
		_ = filter.GroupFilter.UserGroups.Set([]string{cpb.UserGroup_USER_GROUP_STUDENT.String()})
	case "parent":
		_ = filter.GroupFilter.UserGroups.Set([]string{cpb.UserGroup_USER_GROUP_PARENT.String()})
	case "student,parent":
		_ = filter.GroupFilter.UserGroups.Set([]string{cpb.UserGroup_USER_GROUP_STUDENT.String(), cpb.UserGroup_USER_GROUP_PARENT.String()})
	default:
		_ = filter.GroupFilter.UserGroups.Set(nil)
	}

	_ = filter.Offset.Set(offset)
	_ = filter.Limit.Set(limit)

	if keyword != "" {
		_ = filter.GroupFilter.Keyword.Set(keyword)
	}

	switch userID {
	case "student":
		_ = filter.IndividualFilter.UserIDs.Set(studentIDs)
	case "parent":
		_ = filter.IndividualFilter.UserIDs.Set(parentIDs)
	}

	switch enrollment {
	case "enrolled":
		_ = filter.GroupFilter.StudentEnrollmentStatus.Set(enrollmentStatus)
		_ = filter.IndividualFilter.EnrollmentStatuses.Set([]string{enrollmentStatus})
	}

	return filter
}

// common functions for testing
var makeAudienceListByFields = func(fields []string, user_group string, num int) []*entities.Audience {
	audiences := []*entities.Audience{}
	for i := 0; i < num; i++ {
		audience := &entities.Audience{}
		for _, fieldName := range fields {
			switch fieldName {
			case "user_id":
				_ = audience.UserID.Set(studentIDs[i])
			case "student_id":
				_ = audience.StudentID.Set(studentIDs[i])
			case "parent_id":
				_ = audience.ParentID.Set(parentIDs[i])
			case "grade_id":
				_ = audience.GradeID.Set(gradeIDs[i])
			case "is_individual":
				_ = audience.IsIndividual.Set(false)
			default: // other cases
				switch user_group {
				case "student":
					switch fieldName {
					case "child_ids":
						_ = audience.ChildIDs.Set(nil)
					case "user_group":
						_ = audience.UserGroup.Set(userGroups[0])
					case "name":
						_ = audience.Name.Set(studentNames[i])
					case "email":
						_ = audience.Email.Set(studentEmails[i])
					}
				case "parent":
					switch fieldName {
					case "child_ids":
						_ = audience.ChildIDs.Set(studentIDs)
					case "user_group":
						_ = audience.UserGroup.Set(userGroups[1])
					case "name":
						_ = audience.Name.Set(parentNames[i])
					case "email":
						_ = audience.Email.Set(parentEmails[i])
					}
				}
			}
		}
		audiences = append(audiences, audience)
	}
	return audiences
}

// common functions for testing
var makeFindAudienceOption = func(fields ...string) *FindAudienceOption {
	opts := NewFindAudienceOption()
	for _, f := range fields {
		switch f {
		case "name":
			opts.IsGetName = true
		case "order_asc":
			opts.OrderByName = consts.AscendingOrder
		case "order_desc":
			opts.OrderByName = consts.DescendingOrder
		}
	}
	return opts
}

func Test_FindAudienceOption_Validate(t *testing.T) {
	t.Parallel()
	t.Run("happy case default", func(t *testing.T) {
		opts := NewFindAudienceOption()
		err := opts.Validate()
		assert.Nil(t, err)
	})
	t.Run("happy case asc", func(t *testing.T) {
		opts := NewFindAudienceOption()
		opts.OrderByName = consts.AscendingOrder
		err := opts.Validate()
		assert.Nil(t, err)
	})
	t.Run("happy case desc", func(t *testing.T) {
		opts := NewFindAudienceOption()
		opts.OrderByName = consts.DescendingOrder
		err := opts.Validate()
		assert.Nil(t, err)
	})
	t.Run("happy case error", func(t *testing.T) {
		opts := FindAudienceOption{}
		err := opts.Validate()
		assert.NotNil(t, err)
	})
}

func TestAudienceRepo_FindGroupAudiencesByFilter(t *testing.T) {
	t.Parallel()
	mockDB := testutil.NewMockDB()
	repo := &AudienceRepo{
		AudienceSQLBuilder: AudienceSQLBuilder{},
	}
	scanFields := []string{"user_id", "student_id", "parent_id", "grade_id", "child_ids", "user_group", "is_individual"}
	testCases := []struct {
		Name            string
		Setup           func(ctx context.Context)
		Err             error
		Filter          *FindGroupAudienceFilter
		Opts            *FindAudienceOption
		ExpectAudiences []*entities.Audience
	}{
		{
			Name:   "case error query",
			Err:    fmt.Errorf("some error"),
			Filter: makeFindGroupAudienceFilter("all", "all", "all", "all", "all", "student,parent", "", 0, 0),
			Opts:   NewFindAudienceOption(),
			Setup: func(ctx context.Context) {
				ctx2, span := interceptors.StartSpan(ctx, "TestAudienceRepo_FindGroupAudienceByFilter")
				defer span.End()

				mockDB.MockQueryArgs(t, fmt.Errorf("some error"), ctx2, mock.Anything, database.Text(enrollmentStatus), database.TextArray(locationIDs), database.Int8(0), database.Int8(0))
			},
			ExpectAudiences: nil,
		},
		{
			Name:   "case error scan",
			Err:    fmt.Errorf("some error"),
			Filter: makeFindGroupAudienceFilter("all", "all", "all", "all", "none", "student,parent", "", 0, 0),
			Opts:   NewFindAudienceOption(),
			Setup: func(ctx context.Context) {
				ctx2, span := interceptors.StartSpan(ctx, "TestAudienceRepo_FindGroupAudienceByFilter")
				defer span.End()

				mockDB.MockQueryArgs(t, nil, ctx2, mock.Anything, database.Text(enrollmentStatus), database.TextArray(locationIDs), database.Int8(0), database.Int8(0))
				rows := mockDB.Rows
				rows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(fmt.Errorf("some error"))
				rows.On("Next").Once().Return(true)
				rows.On("Close").Once().Return(nil)
			},
			ExpectAudiences: nil,
		},
		{
			Name:   "case locations",
			Err:    nil,
			Filter: makeFindGroupAudienceFilter("all", "all", "all", "all", "none", "student,parent", "", 0, 10),
			Opts:   NewFindAudienceOption(),
			Setup: func(ctx context.Context) {
				ctx2, span := interceptors.StartSpan(ctx, "TestAudienceRepo_FindGroupAudienceByFilter")
				defer span.End()

				audiences := makeAudienceListByFields(scanFields, "student", 2)
				mockDB.MockQueryArgs(t, nil, ctx2, mock.Anything, database.Text(enrollmentStatus), database.TextArray(locationIDs), database.Int8(10), database.Int8(0))
				mockDB.MockScanArray(nil, scanFields, [][]interface{}{
					{
						&audiences[0].UserID,
						&audiences[0].StudentID,
						&audiences[0].ParentID,
						&audiences[0].GradeID,
						&audiences[0].ChildIDs,
						&audiences[0].UserGroup,
						&audiences[0].IsIndividual,
					},
					{
						&audiences[1].UserID,
						&audiences[1].StudentID,
						&audiences[1].ParentID,
						&audiences[1].GradeID,
						&audiences[1].ChildIDs,
						&audiences[1].UserGroup,
						&audiences[1].IsIndividual,
					},
				})
			},
			ExpectAudiences: makeAudienceListByFields(scanFields, "student", 2),
		},
		{
			Name:   "case locations + keyword",
			Err:    nil,
			Filter: makeFindGroupAudienceFilter("all", "all", "all", "all", "none", "student,parent", "keyword", 0, 10),
			Opts:   NewFindAudienceOption(),
			Setup: func(ctx context.Context) {
				ctx2, span := interceptors.StartSpan(ctx, "TestAudienceRepo_FindGroupAudienceByFilter")
				defer span.End()

				audiences := makeAudienceListByFields(append(scanFields, []string{"name", "email"}...), "student", 2)
				mockDB.MockQueryArgs(t, nil, ctx2, mock.Anything, database.Text(enrollmentStatus), database.Text("keyword"), database.TextArray(locationIDs), database.Int8(10), database.Int8(0))
				mockDB.MockScanArray(nil, append(scanFields, []string{"name", "email"}...), [][]interface{}{
					{
						&audiences[0].UserID,
						&audiences[0].StudentID,
						&audiences[0].ParentID,
						&audiences[0].GradeID,
						&audiences[0].ChildIDs,
						&audiences[0].UserGroup,
						&audiences[0].IsIndividual,
						&audiences[0].Name,
						&audiences[0].Email,
					},
					{
						&audiences[1].UserID,
						&audiences[1].StudentID,
						&audiences[1].ParentID,
						&audiences[1].GradeID,
						&audiences[1].ChildIDs,
						&audiences[1].UserGroup,
						&audiences[1].IsIndividual,
						&audiences[1].Name,
						&audiences[1].Email,
					},
				})
			},
			ExpectAudiences: makeAudienceListByFields(append(scanFields, []string{"name", "email"}...), "student", 2),
		},
		{
			Name:   "case locations + IsGetName",
			Err:    nil,
			Filter: makeFindGroupAudienceFilter("all", "all", "all", "all", "none", "student,parent", "keyword", 0, 10),
			Opts:   makeFindAudienceOption("name"),
			Setup: func(ctx context.Context) {
				ctx2, span := interceptors.StartSpan(ctx, "TestAudienceRepo_FindGroupAudienceByFilter")
				defer span.End()

				audiences := makeAudienceListByFields(append(scanFields, []string{"name", "email"}...), "student", 2)
				mockDB.MockQueryArgs(t, nil, ctx2, mock.Anything, database.Text(enrollmentStatus), database.Text("keyword"), database.TextArray(locationIDs), database.Int8(10), database.Int8(0))
				mockDB.MockScanArray(nil, append(scanFields, []string{"name", "email"}...), [][]interface{}{
					{
						&audiences[0].UserID,
						&audiences[0].StudentID,
						&audiences[0].ParentID,
						&audiences[0].GradeID,
						&audiences[0].ChildIDs,
						&audiences[0].UserGroup,
						&audiences[0].IsIndividual,
						&audiences[0].Name,
						&audiences[0].Email,
					},
					{
						&audiences[1].UserID,
						&audiences[1].StudentID,
						&audiences[1].ParentID,
						&audiences[1].GradeID,
						&audiences[1].ChildIDs,
						&audiences[1].UserGroup,
						&audiences[1].IsIndividual,
						&audiences[1].Name,
						&audiences[1].Email,
					},
				})
			},
			ExpectAudiences: makeAudienceListByFields(append(scanFields, []string{"name", "email"}...), "student", 2),
		},
		{
			Name:   "case locations + school",
			Err:    nil,
			Filter: makeFindGroupAudienceFilter("all", "all", "all", "all", "list", "student,parent", "", 0, 10),
			Opts:   NewFindAudienceOption(),
			Setup: func(ctx context.Context) {
				ctx2, span := interceptors.StartSpan(ctx, "TestAudienceRepo_FindGroupAudienceByFilter")
				defer span.End()

				audiences := makeAudienceListByFields(scanFields, "student", 2)
				mockDB.MockQueryArgs(t, nil, ctx2, mock.Anything, database.Text(enrollmentStatus), database.TextArray(locationIDs), database.TextArray(schoolIDs), database.Int8(10), database.Int8(0))
				mockDB.MockScanArray(nil, scanFields, [][]interface{}{
					{
						&audiences[0].UserID,
						&audiences[0].StudentID,
						&audiences[0].ParentID,
						&audiences[0].GradeID,
						&audiences[0].ChildIDs,
						&audiences[0].UserGroup,
						&audiences[0].IsIndividual,
					},
					{
						&audiences[1].UserID,
						&audiences[1].StudentID,
						&audiences[1].ParentID,
						&audiences[1].GradeID,
						&audiences[1].ChildIDs,
						&audiences[1].UserGroup,
						&audiences[1].IsIndividual,
					},
				})
			},
			ExpectAudiences: makeAudienceListByFields(scanFields, "student", 2),
		},
	}

	for _, tc := range testCases {
		ctx := context.Background()
		tc.Setup(ctx)
		actualAudiences, err := repo.FindGroupAudiencesByFilter(ctx, mockDB.DB, tc.Filter, tc.Opts)
		if tc.Err != nil {
			assert.Equal(t, tc.Err, err)
			assert.Nil(t, actualAudiences)
		} else {
			assert.Equal(t, tc.ExpectAudiences, actualAudiences)
			assert.Nil(t, err)
		}
	}
}

func TestAudienceRepo_CountGroupAudiencesByFilter(t *testing.T) {
	t.Parallel()
	mockDB := testutil.NewMockDB()
	repo := &AudienceRepo{
		AudienceSQLBuilder: AudienceSQLBuilder{},
	}

	testCases := []struct {
		Name        string
		Setup       func(ctx context.Context)
		Err         error
		Filter      *FindGroupAudienceFilter
		Opts        *FindAudienceOption
		ExpectCount uint32
	}{
		{
			Name:   "case error",
			Err:    fmt.Errorf("some error"),
			Filter: makeFindGroupAudienceFilter("all", "all", "all", "all", "none", "student,parent", "", 0, 0),
			Opts:   NewFindAudienceOption(),
			Setup: func(ctx context.Context) {
				ctx2, span := interceptors.StartSpan(ctx, "TestAudienceRepo_CountGroupAudiencesByFilter")
				defer span.End()

				countNum := 1
				mockDB.MockQueryRowArgs(t, ctx2, mock.Anything, database.Text(enrollmentStatus), database.TextArray(locationIDs))
				mockDB.MockRowScanFields(fmt.Errorf("some error"), []string{"count"}, []interface{}{&countNum})
			},
			ExpectCount: uint32(0),
		},
		{
			Name:   "case locations",
			Err:    nil,
			Filter: makeFindGroupAudienceFilter("all", "all", "all", "all", "none", "student,parent", "", 0, 0),
			Opts:   NewFindAudienceOption(),
			Setup: func(ctx context.Context) {
				ctx2, span := interceptors.StartSpan(ctx, "TestAudienceRepo_CountGroupAudiencesByFilter")
				defer span.End()

				countNum := 1
				mockDB.MockQueryRowArgs(t, ctx2, mock.Anything, database.Text(enrollmentStatus), database.TextArray(locationIDs))
				mockDB.MockRowScanFields(nil, []string{"count"}, []interface{}{&countNum})
			},
			ExpectCount: 1,
		},
		{
			Name:   "case locations + course",
			Err:    nil,
			Filter: makeFindGroupAudienceFilter("all", "list", "all", "all", "none", "student,parent", "", 0, 0),
			Opts:   NewFindAudienceOption(),
			Setup: func(ctx context.Context) {
				ctx2, span := interceptors.StartSpan(ctx, "TestAudienceRepo_CountGroupAudiencesByFilter")
				defer span.End()

				countNum := 1
				mockDB.MockQueryRowArgs(t, ctx2, mock.Anything, database.Text(enrollmentStatus), database.TextArray(courseIDs), database.TextArray(locationIDs))
				mockDB.MockRowScanFields(nil, []string{"count"}, []interface{}{&countNum})
			},
			ExpectCount: 1,
		},
		{
			Name:   "case locations + course + class",
			Err:    nil,
			Filter: makeFindGroupAudienceFilter("all", "list", "list", "all", "none", "student,parent", "", 0, 0),
			Opts:   NewFindAudienceOption(),
			Setup: func(ctx context.Context) {
				ctx2, span := interceptors.StartSpan(ctx, "TestAudienceRepo_CountGroupAudiencesByFilter")
				defer span.End()

				countNum := 2
				mockDB.MockQueryRowArgs(t, ctx2, mock.Anything, database.Text(enrollmentStatus), database.TextArray(courseIDs), database.TextArray(classIDs), database.TextArray(locationIDs))
				mockDB.MockRowScanFields(nil, []string{"count"}, []interface{}{&countNum})
			},
			ExpectCount: 2,
		},
		{
			Name:   "case locations + course + class + grade",
			Err:    nil,
			Filter: makeFindGroupAudienceFilter("all", "list", "list", "list", "none", "student,parent", "", 0, 0),
			Opts:   NewFindAudienceOption(),
			Setup: func(ctx context.Context) {
				ctx2, span := interceptors.StartSpan(ctx, "TestAudienceRepo_CountGroupAudiencesByFilter")
				defer span.End()

				countNum := 2
				mockDB.MockQueryRowArgs(t, ctx2, mock.Anything, database.Text(enrollmentStatus), database.TextArray(courseIDs), database.TextArray(classIDs), database.TextArray(locationIDs), database.TextArray(gradeIDs))
				mockDB.MockRowScanFields(nil, []string{"count"}, []interface{}{&countNum})
			},
			ExpectCount: 2,
		},
		{
			Name:   "case locations + course + class + grade + school",
			Err:    nil,
			Filter: makeFindGroupAudienceFilter("all", "list", "list", "list", "list", "student,parent", "", 0, 0),
			Opts:   NewFindAudienceOption(),
			Setup: func(ctx context.Context) {
				ctx2, span := interceptors.StartSpan(ctx, "TestAudienceRepo_CountGroupAudiencesByFilter")
				defer span.End()

				countNum := 2
				mockDB.MockQueryRowArgs(t, ctx2, mock.Anything, database.Text(enrollmentStatus), database.TextArray(courseIDs), database.TextArray(classIDs), database.TextArray(locationIDs), database.TextArray(gradeIDs), database.TextArray(schoolIDs))
				mockDB.MockRowScanFields(nil, []string{"count"}, []interface{}{&countNum})
			},
			ExpectCount: 2,
		},
	}

	for _, tc := range testCases {
		ctx := context.Background()
		tc.Setup(ctx)
		actualCount, err := repo.CountGroupAudiencesByFilter(ctx, mockDB.DB, tc.Filter, tc.Opts)
		if tc.Err != nil {
			assert.Equal(t, tc.Err, err)
			assert.Equal(t, tc.ExpectCount, actualCount)
		} else {
			assert.Nil(t, err)
			assert.Equal(t, tc.ExpectCount, actualCount)
		}
	}
}

func Test_FindIndividualAudiencesByFilter(t *testing.T) {
	t.Parallel()
	repo := &AudienceRepo{
		AudienceSQLBuilder: AudienceSQLBuilder{},
	}
	mockDB := testutil.NewMockDB()
	rows := mockDB.Rows

	t.Run("happy case, student", func(t *testing.T) {
		filter := NewFindIndividualAudienceFilter()
		filter.LocationIDs.Set(locationIDs)
		filter.UserIDs.Set(studentIDs)
		filter.EnrollmentStatuses.Set([]string{pbu.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()})

		ctx := context.Background()
		ctx2, span := interceptors.StartSpan(ctx, "TestAudienceRepo_FindIndividualAudiencesByFilter")
		defer span.End()
		mockDB.MockQueryArgs(t, nil, ctx2, mock.Anything, filter.EnrollmentStatuses, filter.UserIDs, filter.LocationIDs)
		mockDB.DB.On("Query").Once().Return(rows, nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
			reflect.ValueOf(args[0]).Elem().Set(reflect.ValueOf(pgtype.Text{Status: pgtype.Present, String: studentIDs[0]}))
			reflect.ValueOf(args[1]).Elem().Set(reflect.ValueOf(pgtype.Text{Status: pgtype.Present, String: studentIDs[0]}))
			reflect.ValueOf(args[2]).Elem().Set(reflect.ValueOf(pgtype.Text{Status: pgtype.Undefined, String: ""}))
			reflect.ValueOf(args[3]).Elem().Set(reflect.ValueOf(pgtype.Text{Status: pgtype.Present, String: gradeIDs[0]}))
			reflect.ValueOf(args[4]).Elem().Set(reflect.ValueOf(pgtype.TextArray{
				Status:     pgtype.Undefined,
				Dimensions: nil,
				Elements:   nil,
			}))
			reflect.ValueOf(args[5]).Elem().Set(reflect.ValueOf(pgtype.Text{Status: pgtype.Present, String: userGroups[0]}))
			reflect.ValueOf(args[6]).Elem().Set(reflect.ValueOf(pgtype.Bool{Status: pgtype.Present, Bool: true}))
		}).Return(nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
			reflect.ValueOf(args[0]).Elem().Set(reflect.ValueOf(pgtype.Text{Status: pgtype.Present, String: studentIDs[1]}))
			reflect.ValueOf(args[1]).Elem().Set(reflect.ValueOf(pgtype.Text{Status: pgtype.Present, String: studentIDs[1]}))
			reflect.ValueOf(args[2]).Elem().Set(reflect.ValueOf(pgtype.Text{Status: pgtype.Undefined, String: ""}))
			reflect.ValueOf(args[3]).Elem().Set(reflect.ValueOf(pgtype.Text{Status: pgtype.Present, String: gradeIDs[1]}))
			reflect.ValueOf(args[4]).Elem().Set(reflect.ValueOf(pgtype.TextArray{
				Status:     pgtype.Undefined,
				Dimensions: nil,
				Elements:   nil,
			}))
			reflect.ValueOf(args[5]).Elem().Set(reflect.ValueOf(pgtype.Text{Status: pgtype.Present, String: userGroups[0]}))
			reflect.ValueOf(args[6]).Elem().Set(reflect.ValueOf(pgtype.Bool{Status: pgtype.Present, Bool: true}))
		}).Return(nil)
		rows.On("Next").Once().Return(false)
		rows.On("Close").Once().Return(nil)
		rows.On("Err").Once().Return(nil)

		expectedAudiences := []*entities.Audience{}
		expectedAudiences = append(expectedAudiences, &entities.Audience{
			UserID:       database.Text(studentIDs[0]),
			StudentID:    database.Text(studentIDs[0]),
			GradeID:      database.Text(gradeIDs[0]),
			UserGroup:    database.Text(userGroups[0]),
			IsIndividual: database.Bool(true),
		})
		expectedAudiences = append(expectedAudiences, &entities.Audience{
			UserID:       database.Text(studentIDs[1]),
			StudentID:    database.Text(studentIDs[1]),
			GradeID:      database.Text(gradeIDs[1]),
			UserGroup:    database.Text(userGroups[0]),
			IsIndividual: database.Bool(true),
		})

		audiences, err := repo.FindIndividualAudiencesByFilter(ctx, mockDB.DB, filter)
		assert.Nil(t, err)
		assert.Equal(t, expectedAudiences, audiences)
	})

	t.Run("happy case, student and parent", func(t *testing.T) {
		filter := NewFindIndividualAudienceFilter()
		filter.LocationIDs.Set(locationIDs)
		filter.UserIDs.Set([]string{studentIDs[0], parentIDs[0]})
		filter.EnrollmentStatuses.Set([]string{pbu.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()})

		ctx := context.Background()
		ctx2, span := interceptors.StartSpan(ctx, "TestAudienceRepo_FindIndividualAudiencesByFilter")
		defer span.End()
		mockDB.MockQueryArgs(t, nil, ctx2, mock.Anything, filter.EnrollmentStatuses, filter.UserIDs, filter.LocationIDs)
		mockDB.DB.On("Query").Once().Return(rows, nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
			reflect.ValueOf(args[0]).Elem().Set(reflect.ValueOf(pgtype.Text{Status: pgtype.Present, String: studentIDs[0]}))
			reflect.ValueOf(args[1]).Elem().Set(reflect.ValueOf(pgtype.Text{Status: pgtype.Present, String: studentIDs[0]}))
			reflect.ValueOf(args[2]).Elem().Set(reflect.ValueOf(pgtype.Text{Status: pgtype.Undefined, String: ""}))
			reflect.ValueOf(args[3]).Elem().Set(reflect.ValueOf(pgtype.Text{Status: pgtype.Present, String: gradeIDs[0]}))
			reflect.ValueOf(args[4]).Elem().Set(reflect.ValueOf(pgtype.TextArray{
				Status:     pgtype.Undefined,
				Dimensions: nil,
				Elements:   nil,
			}))
			reflect.ValueOf(args[5]).Elem().Set(reflect.ValueOf(pgtype.Text{String: userGroups[0], Status: pgtype.Present}))
			reflect.ValueOf(args[6]).Elem().Set(reflect.ValueOf(pgtype.Bool{Bool: true, Status: pgtype.Present}))
		}).Return(nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
			reflect.ValueOf(args[0]).Elem().Set(reflect.ValueOf(pgtype.Text{Status: pgtype.Present, String: parentIDs[0]}))
			reflect.ValueOf(args[1]).Elem().Set(reflect.ValueOf(pgtype.Text{Status: pgtype.Undefined, String: ""}))
			reflect.ValueOf(args[2]).Elem().Set(reflect.ValueOf(pgtype.Text{Status: pgtype.Present, String: parentIDs[0]}))
			reflect.ValueOf(args[3]).Elem().Set(reflect.ValueOf(pgtype.Text{Status: pgtype.Undefined}))
			reflect.ValueOf(args[4]).Elem().Set(reflect.ValueOf(pgtype.TextArray{
				Status:     pgtype.Undefined,
				Dimensions: nil,
				Elements:   nil,
			}))
			reflect.ValueOf(args[5]).Elem().Set(reflect.ValueOf(pgtype.Text{Status: pgtype.Present, String: userGroups[1]}))
			reflect.ValueOf(args[6]).Elem().Set(reflect.ValueOf(pgtype.Bool{Bool: true, Status: pgtype.Present}))
		}).Return(nil)
		rows.On("Next").Once().Return(false)
		rows.On("Close").Once().Return(nil)
		rows.On("Err").Once().Return(nil)

		expectedAudiences := []*entities.Audience{}
		expectedAudiences = append(expectedAudiences, &entities.Audience{
			UserID:       database.Text(studentIDs[0]),
			StudentID:    database.Text(studentIDs[0]),
			GradeID:      database.Text(gradeIDs[0]),
			UserGroup:    database.Text(userGroups[0]),
			IsIndividual: database.Bool(true),
		})
		expectedAudiences = append(expectedAudiences, &entities.Audience{
			UserID:       database.Text(parentIDs[0]),
			ParentID:     database.Text(parentIDs[0]),
			UserGroup:    database.Text(userGroups[1]),
			IsIndividual: database.Bool(true),
		})

		audiences, err := repo.FindIndividualAudiencesByFilter(ctx, mockDB.DB, filter)
		assert.Nil(t, err)
		assert.Equal(t, expectedAudiences, audiences)
	})
}

func TestAudienceRepo_FindDraftAudiencesByFilter(t *testing.T) {
	t.Parallel()
	mockDB := testutil.NewMockDB()
	repo := &AudienceRepo{
		AudienceSQLBuilder: AudienceSQLBuilder{},
	}
	scanFields := []string{"user_id", "student_id", "parent_id", "grade_id", "child_ids", "user_group", "is_individual"}
	testCases := []struct {
		Name            string
		Setup           func(ctx context.Context)
		Err             error
		Filter          *FindDraftAudienceFilter
		Opts            *FindAudienceOption
		ExpectAudiences []*entities.Audience
	}{
		{
			Name:   "case error query",
			Err:    fmt.Errorf("some error"),
			Filter: makeFindDraftAudienceFilter("all", "all", "all", "all", "student,parent", "", "student", "enrolled", 0, 0),
			Opts:   NewFindAudienceOption(),
			Setup: func(ctx context.Context) {
				ctx2, span := interceptors.StartSpan(ctx, "TestAudienceRepo_FindGroupAudienceByFilter")
				defer span.End()

				mockDB.MockQueryArgs(t, fmt.Errorf("some error"),
					ctx2,
					mock.Anything,
					database.Text(enrollmentStatus),
					database.TextArray(locationIDs),
					database.TextArray([]string{pbu.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()}),
					database.TextArray(studentIDs),
					database.TextArray(locationIDs),
					database.Int8(0),
					database.Int8(0),
				)
			},
			ExpectAudiences: nil,
		},
		{
			Name:   "case error scan",
			Err:    fmt.Errorf("some error"),
			Filter: makeFindDraftAudienceFilter("all", "all", "all", "all", "student,parent", "", "student", "enrolled", 0, 0),
			Opts:   NewFindAudienceOption(),
			Setup: func(ctx context.Context) {
				ctx2, span := interceptors.StartSpan(ctx, "TestAudienceRepo_FindGroupAudienceByFilter")
				defer span.End()

				mockDB.MockQueryArgs(t, nil,
					ctx2,
					mock.Anything,
					database.Text(enrollmentStatus),
					database.TextArray(locationIDs),
					database.TextArray([]string{pbu.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()}),
					database.TextArray(studentIDs),
					database.TextArray(locationIDs),
					database.Int8(0),
					database.Int8(0),
				)
				rows := mockDB.Rows
				rows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(fmt.Errorf("some error"))
				rows.On("Next").Once().Return(true)
				rows.On("Close").Once().Return(nil)
			},
			ExpectAudiences: nil,
		},
		{
			Name:   "case locations",
			Err:    nil,
			Filter: makeFindDraftAudienceFilter("all", "all", "all", "all", "student,parent", "", "student", "enrolled", 0, 10),
			Opts:   NewFindAudienceOption(),
			Setup: func(ctx context.Context) {
				ctx2, span := interceptors.StartSpan(ctx, "TestAudienceRepo_FindGroupAudienceByFilter")
				defer span.End()

				audiences := makeAudienceListByFields(scanFields, "student", 2)
				mockDB.MockQueryArgs(t, nil,
					ctx2,
					mock.Anything,
					database.Text(enrollmentStatus),
					database.TextArray(locationIDs),
					database.TextArray([]string{pbu.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()}),
					database.TextArray(studentIDs),
					database.TextArray(locationIDs),
					database.Int8(10),
					database.Int8(0),
				)
				mockDB.MockScanArray(nil, scanFields, [][]interface{}{
					{
						&audiences[0].UserID,
						&audiences[0].StudentID,
						&audiences[0].ParentID,
						&audiences[0].GradeID,
						&audiences[0].ChildIDs,
						&audiences[0].UserGroup,
						&audiences[0].IsIndividual,
					},
					{
						&audiences[1].UserID,
						&audiences[1].StudentID,
						&audiences[1].ParentID,
						&audiences[1].GradeID,
						&audiences[1].ChildIDs,
						&audiences[1].UserGroup,
						&audiences[1].IsIndividual,
					},
				})
			},
			ExpectAudiences: makeAudienceListByFields(scanFields, "student", 2),
		},
		{
			Name:   "case locations + IsGetName",
			Err:    nil,
			Filter: makeFindDraftAudienceFilter("all", "all", "all", "all", "student,parent", "keyword", "student", "enrolled", 0, 10),
			Opts:   makeFindAudienceOption("name"),
			Setup: func(ctx context.Context) {
				ctx2, span := interceptors.StartSpan(ctx, "TestAudienceRepo_FindGroupAudienceByFilter")
				defer span.End()

				audiences := makeAudienceListByFields(append(scanFields, []string{"name", "email"}...), "student", 2)
				mockDB.MockQueryArgs(t, nil,
					ctx2,
					mock.Anything,
					database.Text(enrollmentStatus),
					database.Text("keyword"),
					database.TextArray(locationIDs),
					database.TextArray([]string{pbu.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()}),
					database.TextArray(studentIDs),
					database.TextArray(locationIDs),
					database.Int8(10),
					database.Int8(0),
				)
				mockDB.MockScanArray(nil, append(scanFields, []string{"name", "email"}...), [][]interface{}{
					{
						&audiences[0].UserID,
						&audiences[0].StudentID,
						&audiences[0].ParentID,
						&audiences[0].GradeID,
						&audiences[0].ChildIDs,
						&audiences[0].UserGroup,
						&audiences[0].IsIndividual,
						&audiences[0].Name,
						&audiences[0].Email,
					},
					{
						&audiences[1].UserID,
						&audiences[1].StudentID,
						&audiences[1].ParentID,
						&audiences[1].GradeID,
						&audiences[1].ChildIDs,
						&audiences[1].UserGroup,
						&audiences[1].IsIndividual,
						&audiences[1].Name,
						&audiences[1].Email,
					},
				})
			},
			ExpectAudiences: makeAudienceListByFields(append(scanFields, []string{"name", "email"}...), "student", 2),
		},
	}

	for _, tc := range testCases {
		ctx := context.Background()
		tc.Setup(ctx)
		actualAudiences, err := repo.FindDraftAudiencesByFilter(ctx, mockDB.DB, tc.Filter, tc.Opts)
		if tc.Err != nil {
			assert.Equal(t, tc.Err, err)
			assert.Nil(t, actualAudiences)
		} else {
			assert.Equal(t, tc.ExpectAudiences, actualAudiences)
			assert.Nil(t, err)
		}
	}
}

func TestAudienceRepo_CountDraftAudiencesByFilter(t *testing.T) {
	t.Parallel()
	mockDB := testutil.NewMockDB()
	repo := &AudienceRepo{
		AudienceSQLBuilder: AudienceSQLBuilder{},
	}

	testCases := []struct {
		Name        string
		Setup       func(ctx context.Context)
		Err         error
		Filter      *FindDraftAudienceFilter
		Opts        *FindAudienceOption
		ExpectCount uint32
	}{
		{
			Name:   "case error",
			Err:    fmt.Errorf("some error"),
			Filter: makeFindDraftAudienceFilter("all", "all", "all", "all", "student,parent", "", "student", "enrolled", 0, 0),
			Opts:   NewFindAudienceOption(),
			Setup: func(ctx context.Context) {
				ctx2, span := interceptors.StartSpan(ctx, "TestAudienceRepo_CountGroupAudiencesByFilter")
				defer span.End()

				countNum := 1
				mockDB.MockQueryRowArgs(t,
					ctx2,
					mock.Anything,
					database.Text(enrollmentStatus),
					database.TextArray(locationIDs),
					database.TextArray([]string{pbu.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()}),
					database.TextArray(studentIDs),
					database.TextArray(locationIDs),
				)
				mockDB.MockRowScanFields(fmt.Errorf("some error"), []string{"count"}, []interface{}{&countNum})
			},
			ExpectCount: uint32(0),
		},
		{
			Name:   "case locations + user group student parent + userID + enrolled",
			Err:    nil,
			Filter: makeFindDraftAudienceFilter("all", "all", "all", "all", "student,parent", "", "student", "enrolled", 0, 0),
			Opts:   NewFindAudienceOption(),
			Setup: func(ctx context.Context) {
				ctx2, span := interceptors.StartSpan(ctx, "TestAudienceRepo_CountGroupAudiencesByFilter")
				defer span.End()

				countNum := 1
				mockDB.MockQueryRowArgs(t,
					ctx2,
					mock.Anything,
					database.Text(enrollmentStatus),
					database.TextArray(locationIDs),
					database.TextArray([]string{pbu.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()}),
					database.TextArray(studentIDs),
					database.TextArray(locationIDs),
				)
				mockDB.MockRowScanFields(nil, []string{"count"}, []interface{}{&countNum})
			},
			ExpectCount: 1,
		},
		{
			Name:   "case locations + course + user group student parent + userID + enrolled",
			Err:    nil,
			Filter: makeFindDraftAudienceFilter("all", "list", "all", "all", "student,parent", "", "student", "enrolled", 0, 0),
			Opts:   NewFindAudienceOption(),
			Setup: func(ctx context.Context) {
				ctx2, span := interceptors.StartSpan(ctx, "TestAudienceRepo_CountGroupAudiencesByFilter")
				defer span.End()

				countNum := 1
				mockDB.MockQueryRowArgs(t,
					ctx2,
					mock.Anything,
					database.Text(enrollmentStatus),
					database.TextArray(courseIDs),
					database.TextArray(locationIDs),
					database.TextArray([]string{pbu.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()}),
					database.TextArray(studentIDs),
					database.TextArray(locationIDs),
				)
				mockDB.MockRowScanFields(nil, []string{"count"}, []interface{}{&countNum})
			},
			ExpectCount: 1,
		},
		{
			Name:   "case locations + course + class + user group student parent + userID + enrolled",
			Err:    nil,
			Filter: makeFindDraftAudienceFilter("all", "list", "list", "all", "student,parent", "", "student", "enrolled", 0, 0),
			Opts:   NewFindAudienceOption(),
			Setup: func(ctx context.Context) {
				ctx2, span := interceptors.StartSpan(ctx, "TestAudienceRepo_CountGroupAudiencesByFilter")
				defer span.End()

				countNum := 2
				mockDB.MockQueryRowArgs(t,
					ctx2,
					mock.Anything,
					database.Text(enrollmentStatus),
					database.TextArray(courseIDs),
					database.TextArray(classIDs),
					database.TextArray(locationIDs),
					database.TextArray([]string{pbu.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()}),
					database.TextArray(studentIDs),
					database.TextArray(locationIDs),
				)
				mockDB.MockRowScanFields(nil, []string{"count"}, []interface{}{&countNum})
			},
			ExpectCount: 2,
		},
		{
			Name:   "case locations + course + class + grade",
			Err:    nil,
			Filter: makeFindDraftAudienceFilter("all", "list", "list", "list", "student,parent", "", "student", "enrolled", 0, 0),
			Opts:   NewFindAudienceOption(),
			Setup: func(ctx context.Context) {
				ctx2, span := interceptors.StartSpan(ctx, "TestAudienceRepo_CountGroupAudiencesByFilter")
				defer span.End()

				countNum := 2
				mockDB.MockQueryRowArgs(t,
					ctx2,
					mock.Anything,
					database.Text(enrollmentStatus),
					database.TextArray(courseIDs),
					database.TextArray(classIDs),
					database.TextArray(locationIDs),
					database.TextArray(gradeIDs),
					database.TextArray([]string{pbu.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()}),
					database.TextArray(studentIDs),
					database.TextArray(locationIDs),
				)
				mockDB.MockRowScanFields(nil, []string{"count"}, []interface{}{&countNum})
			},
			ExpectCount: 2,
		},
	}

	for _, tc := range testCases {
		ctx := context.Background()
		tc.Setup(ctx)
		actualCount, err := repo.CountDraftAudiencesByFilter(ctx, mockDB.DB, tc.Filter, tc.Opts)
		if tc.Err != nil {
			assert.Equal(t, tc.Err, err)
			assert.Equal(t, tc.ExpectCount, actualCount)
		} else {
			assert.Nil(t, err)
			assert.Equal(t, tc.ExpectCount, actualCount)
		}
	}
}

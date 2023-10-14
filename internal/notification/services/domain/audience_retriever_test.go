package domain

import (
	"context"
	"fmt"
	"testing"

	bobEntities "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/notification/consts"
	"github.com/manabie-com/backend/internal/notification/entities"
	"github.com/manabie-com/backend/internal/notification/repositories"
	"github.com/manabie-com/backend/internal/notification/services/utils"
	mock_bob_repositories "github.com/manabie-com/backend/mock/bob/repositories"
	mock_repositories "github.com/manabie-com/backend/mock/notification/repositories"
	"github.com/manabie-com/backend/mock/testutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pbu "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAudienceRetriever_FindAudiences(t *testing.T) {
	t.Parallel()
	mockDB := testutil.NewMockDB()
	audienceRepo := &mock_repositories.MockAudienceRepo{}
	infoNotiAccessPathRepo := &mock_repositories.MockInfoNotificationAccessPathRepo{}
	locationRepo := &mock_repositories.MockLocationRepo{}
	notificationInternalUserRepo := &mock_repositories.MockNotificationInternalUserRepo{}
	studentParentRepo := &mock_bob_repositories.MockStudentParentRepo{}
	svc := &AudienceRetrieverService{
		AudienceRepo:                   audienceRepo,
		InfoNotificationAccessPathRepo: infoNotiAccessPathRepo,
		LocationRepo:                   locationRepo,
		NotificationInternalUserRepo:   notificationInternalUserRepo,
		StudentParentRepo:              studentParentRepo,
		Env:                            "prod",
	}
	userID := "user-id"
	notiID := "noti-id"
	studentIDs := []string{"student_id_1", "student_id_2", "student_id_3"}
	parentIDs := []string{"parent_student_id_1", "parent_student_id_2", "parent_student_id_3"}
	individualStudentIDs := []string{"ind_student_id_1", "ind_student_id_2", "ind_student_id_3"}
	individualParentIDs := []string{"ind_parent_id_1", "ind_parent_id_2", "ind_parent_id_3"}
	genericUserIDs := []string{}
	genericUserIDs = append(genericUserIDs, individualStudentIDs...)
	genericUserIDs = append(genericUserIDs, individualParentIDs...)
	locationIDs := []string{"loc-1", "loc-2"}
	courseIDs := []string{"course_id_1", "course_id_2", "course_id_3"}
	classIDs := []string{"class-id-1", "class-id-2", "class-id-3"}
	gradeIDs := []string{"grade-1", "grade-2", "grade-3"}
	schoolIDs := []string{"school-id-1", "school-id-2", "school-id-3"}
	notificationPermissions := []string{
		consts.NotificationWritePermission,
		consts.NotificationOwnerPermission,
	}
	manabieOrgID := fmt.Sprint(constants.ManabieSchool)
	notification := utils.GenNotificationEntity()
	// Reset all filter condition, each test will have its own data
	_ = notification.NotificationID.Set(notiID)
	_ = notification.CreatedUserID.Set(userID)
	_ = notification.Type.Set(cpb.NotificationType_NOTIFICATION_TYPE_NONE.String())
	_ = notification.TargetGroups.Set(nil)
	_ = notification.GenericReceiverIDs.Set(nil)
	_ = notification.ReceiverIDs.Set(nil)

	students := []*entities.Audience{
		{
			UserID:    database.Text(studentIDs[0]),
			StudentID: database.Text(studentIDs[0]),
			GradeID:   database.Text(gradeIDs[0]),
			ChildIDs:  database.TextArray(nil),
			UserGroup: database.Text(cpb.UserGroup_USER_GROUP_STUDENT.String()),
		},
		{
			UserID:    database.Text(studentIDs[1]),
			StudentID: database.Text(studentIDs[1]),
			GradeID:   database.Text(gradeIDs[1]),
			ChildIDs:  database.TextArray(nil),
			UserGroup: database.Text(cpb.UserGroup_USER_GROUP_STUDENT.String()),
		},
		{
			UserID:    database.Text(studentIDs[2]),
			StudentID: database.Text(studentIDs[2]),
			GradeID:   database.Text(gradeIDs[2]),
			ChildIDs:  database.TextArray(nil),
			UserGroup: database.Text(cpb.UserGroup_USER_GROUP_STUDENT.String()),
		},
	}
	parents := []*entities.Audience{
		{
			UserID:    database.Text(parentIDs[0]),
			GradeID:   database.Text(""),
			ChildIDs:  database.TextArray([]string{studentIDs[0]}),
			UserGroup: database.Text(cpb.UserGroup_USER_GROUP_PARENT.String()),
		},
		{
			UserID:    database.Text(parentIDs[1]),
			GradeID:   database.Text(""),
			ChildIDs:  database.TextArray([]string{studentIDs[1]}),
			UserGroup: database.Text(cpb.UserGroup_USER_GROUP_PARENT.String()),
		},
		{
			UserID:    database.Text(parentIDs[2]),
			GradeID:   database.Text(""),
			ChildIDs:  database.TextArray([]string{studentIDs[2]}),
			UserGroup: database.Text(cpb.UserGroup_USER_GROUP_PARENT.String()),
		},
	}
	parentIndividualWithUserGroups := []*entities.Audience{
		{
			UserID:       database.Text(parentIDs[0]),
			ParentID:     database.Text(parentIDs[0]),
			StudentID:    database.Text(studentIDs[0]),
			ChildIDs:     database.TextArray([]string{studentIDs[0]}),
			UserGroup:    database.Text(cpb.UserGroup_USER_GROUP_PARENT.String()),
			IsIndividual: database.Bool(true),
		},
		{
			UserID:       database.Text(parentIDs[1]),
			ParentID:     database.Text(parentIDs[1]),
			StudentID:    database.Text(studentIDs[1]),
			ChildIDs:     database.TextArray([]string{studentIDs[1]}),
			UserGroup:    database.Text(cpb.UserGroup_USER_GROUP_PARENT.String()),
			IsIndividual: database.Bool(true),
		},
		{
			UserID:       database.Text(parentIDs[2]),
			ParentID:     database.Text(parentIDs[2]),
			StudentID:    database.Text(studentIDs[2]),
			ChildIDs:     database.TextArray([]string{studentIDs[2]}),
			UserGroup:    database.Text(cpb.UserGroup_USER_GROUP_PARENT.String()),
			IsIndividual: database.Bool(true),
		},
	}

	individualStudents := []*entities.Audience{
		{
			UserID:       database.Text(individualStudentIDs[0]),
			GradeID:      database.Text("grade-4"),
			UserGroup:    database.Text(cpb.UserGroup_USER_GROUP_STUDENT.String()),
			IsIndividual: database.Bool(true),
		},
		{
			UserID:       database.Text(individualStudentIDs[1]),
			GradeID:      database.Text("grade-5"),
			UserGroup:    database.Text(cpb.UserGroup_USER_GROUP_STUDENT.String()),
			IsIndividual: database.Bool(true),
		},
		{
			UserID:       database.Text(individualStudentIDs[2]),
			GradeID:      database.Text("grade-6"),
			UserGroup:    database.Text(cpb.UserGroup_USER_GROUP_STUDENT.String()),
			IsIndividual: database.Bool(true),
		},
	}
	individualParents := []*entities.Audience{
		{
			UserID:       database.Text(individualParentIDs[0]),
			GradeID:      database.Text(""),
			ChildIDs:     database.TextArray([]string{individualStudentIDs[0]}),
			UserGroup:    database.Text(cpb.UserGroup_USER_GROUP_PARENT.String()),
			IsIndividual: database.Bool(true),
		},
		{
			UserID:       database.Text(individualParentIDs[1]),
			GradeID:      database.Text(""),
			ChildIDs:     database.TextArray([]string{individualStudentIDs[0]}),
			UserGroup:    database.Text(cpb.UserGroup_USER_GROUP_PARENT.String()),
			IsIndividual: database.Bool(true),
		},
		{
			UserID:       database.Text(individualParentIDs[2]),
			GradeID:      database.Text(""),
			ChildIDs:     database.TextArray([]string{individualStudentIDs[0]}),
			UserGroup:    database.Text(cpb.UserGroup_USER_GROUP_PARENT.String()),
			IsIndividual: database.Bool(true),
		},
	}

	studentParents := []*bobEntities.StudentParent{
		{
			StudentID: database.Text(studentIDs[0]),
			ParentID:  database.Text(parentIDs[0]),
		},
		{
			StudentID: database.Text(studentIDs[1]),
			ParentID:  database.Text(parentIDs[1]),
		},
		{
			StudentID: database.Text(studentIDs[2]),
			ParentID:  database.Text(parentIDs[2]),
		},
	}

	makeGroupAudienceFilter := func(locationSelect, courseSelect, classSelect, gradeSelect, schoolSelect, userGroup string, individualIDs []string, notificationType string) *repositories.FindGroupAudienceFilter {
		filter := repositories.NewFindGroupAudienceFilter()
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

		if notificationType == cpb.NotificationType_NOTIFICATION_TYPE_COMPOSED.String() {
			filter.StudentEnrollmentStatus = database.Text(pbu.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String())
		}

		return filter
	}

	makeIndividualFilter := func(locationIDs, genericUserIDs []string, notificationType string) *repositories.FindIndividualAudienceFilter {
		individualFilter := repositories.NewFindIndividualAudienceFilter()
		_ = individualFilter.LocationIDs.Set(locationIDs)
		_ = individualFilter.UserIDs.Set(genericUserIDs)

		if notificationType == cpb.NotificationType_NOTIFICATION_TYPE_COMPOSED.String() {
			_ = individualFilter.EnrollmentStatuses.Set([]string{pbu.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()})
		}

		return individualFilter
	}

	claim := &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: manabieOrgID,
			UserID:       userID,
		},
	}
	ctx := interceptors.ContextWithJWTClaims(context.Background(), claim)

	// TIL: test order affects test coverage
	t.Run("case granted locations changed", func(t *testing.T) {
		targetGroupAudienceSelector := &entities.InfoNotificationTarget{
			LocationFilter:  entities.InfoNotificationTarget_LocationFilter{LocationIDs: locationIDs, Type: consts.TargetGroupSelectTypeAll.String()},
			CourseFilter:    entities.InfoNotificationTarget_CourseFilter{CourseIDs: []string{}, Type: consts.TargetGroupSelectTypeNone.String()},
			ClassFilter:     entities.InfoNotificationTarget_ClassFilter{ClassIDs: []string{}, Type: consts.TargetGroupSelectTypeNone.String()},
			GradeFilter:     entities.InfoNotificationTarget_GradeFilter{GradeIDs: []string{}, Type: consts.TargetGroupSelectTypeNone.String()},
			SchoolFilter:    entities.InfoNotificationTarget_SchoolFilter{SchoolIDs: []string{}, Type: consts.TargetGroupSelectTypeNone.String()},
			UserGroupFilter: entities.InfoNotificationTarget_UserGroupFilter{UserGroups: []string{cpb.UserGroup_USER_GROUP_STUDENT.String(), cpb.UserGroup_USER_GROUP_PARENT.String()}},
		}

		_ = notification.Type.Set(cpb.NotificationType_NOTIFICATION_TYPE_COMPOSED.String())
		_ = notification.TargetGroups.Set(targetGroupAudienceSelector)
		_ = notification.GenericReceiverIDs.Set(nil)
		_ = notification.ReceiverIDs.Set(nil)

		audienceFilter := makeGroupAudienceFilter("all", "none", "none", "none", "none", "student,parent", nil, notification.Type.String)
		infoNotiAccessPathRepo.On("GetLocationIDsByNotificationID", ctx, mockDB.DB, notiID).Once().Return(locationIDs, nil)
		locationRepo.On("GetLowestLocationIDsByIDs", ctx, mockDB.DB, locationIDs).Once().Return([]string{}, nil)
		internalUser := &entities.NotificationInternalUser{
			UserID: database.Text("internal-user-id"),
		}
		notificationInternalUserRepo.On("GetByOrgID", ctx, mockDB.DB, manabieOrgID).Once().Return(internalUser, nil)
		internalUserCtx := interceptors.ContextWithJWTClaims(ctx, &interceptors.CustomClaims{
			Manabie: &interceptors.ManabieClaims{
				ResourcePath: manabieOrgID,
				UserID:       "internal-user-id",
				UserGroup:    cpb.UserGroup_USER_GROUP_SCHOOL_ADMIN.String(),
			},
		})
		infoNotiAccessPathRepo.On("GetLocationIDsByNotificationID", internalUserCtx, mockDB.DB, notiID).Once().Return(locationIDs, nil)

		rawGroupAudiences := []*entities.Audience{}
		rawGroupAudiences = append(rawGroupAudiences, students...)
		rawGroupAudiences = append(rawGroupAudiences, parents...)
		audienceRepo.On("FindGroupAudiencesByFilter", ctx, mockDB.DB, audienceFilter, repositories.NewFindAudienceOption()).Once().Return(rawGroupAudiences, nil)

		actualAudiences, err := svc.FindAudiences(ctx, mockDB.DB, &notification)
		assert.Nil(t, err)

		expectedAudiences := []*entities.Audience{}
		expectedAudiences = append(expectedAudiences, rawGroupAudiences...)
		assert.Equal(t, expectedAudiences, actualAudiences)
	})
	t.Run("Happy case audience selector with selected none for all, user group student and parent, generic individual student and parent", func(t *testing.T) {
		targetGroupAudienceSelector := &entities.InfoNotificationTarget{
			LocationFilter:  entities.InfoNotificationTarget_LocationFilter{LocationIDs: locationIDs, Type: consts.TargetGroupSelectTypeList.String()},
			CourseFilter:    entities.InfoNotificationTarget_CourseFilter{CourseIDs: []string{}, Type: consts.TargetGroupSelectTypeNone.String()},
			ClassFilter:     entities.InfoNotificationTarget_ClassFilter{ClassIDs: []string{}, Type: consts.TargetGroupSelectTypeNone.String()},
			GradeFilter:     entities.InfoNotificationTarget_GradeFilter{GradeIDs: []string{}, Type: consts.TargetGroupSelectTypeNone.String()},
			SchoolFilter:    entities.InfoNotificationTarget_SchoolFilter{SchoolIDs: []string{}, Type: consts.TargetGroupSelectTypeNone.String()},
			UserGroupFilter: entities.InfoNotificationTarget_UserGroupFilter{UserGroups: []string{cpb.UserGroup_USER_GROUP_STUDENT.String(), cpb.UserGroup_USER_GROUP_PARENT.String()}},
		}

		_ = notification.Type.Set(cpb.NotificationType_NOTIFICATION_TYPE_COMPOSED.String())
		_ = notification.TargetGroups.Set(targetGroupAudienceSelector)
		_ = notification.GenericReceiverIDs.Set(genericUserIDs)
		_ = notification.ReceiverIDs.Set(nil)

		audienceFilter := makeGroupAudienceFilter("list", "none", "none", "none", "none", "student,parent", nil, notification.Type.String)
		rawGroupAudiences := []*entities.Audience{}
		rawGroupAudiences = append(rawGroupAudiences, students...)
		rawGroupAudiences = append(rawGroupAudiences, parents...)
		audienceRepo.On("FindGroupAudiencesByFilter", ctx, mockDB.DB, audienceFilter, repositories.NewFindAudienceOption()).Once().Return(rawGroupAudiences, nil)

		infoNotiAccessPathRepo.On("GetLocationIDsByNotificationID", ctx, mockDB.DB, notiID).Once().Return(locationIDs, nil)
		locationRepo.On("GetLowestLocationIDsByIDs", ctx, mockDB.DB, locationIDs).Once().Return(locationIDs, nil)
		locationRepo.On("GetLowestGrantedLocationsByUserIDAndPermissions", ctx, mockDB.DB, userID, notificationPermissions).Once().Return(locationIDs, map[string]string{}, nil)

		individualFilter := makeIndividualFilter(locationIDs, genericUserIDs, notification.Type.String)
		individualAudience := []*entities.Audience{}
		individualAudience = append(individualAudience, individualStudents...)
		individualAudience = append(individualAudience, individualParents...)
		audienceRepo.On("FindIndividualAudiencesByFilter", ctx, mockDB.DB, individualFilter).Once().Return(individualAudience, nil)

		actualAudiences, err := svc.FindAudiences(ctx, mockDB.DB, &notification)
		assert.Nil(t, err)

		expectedAudiences := []*entities.Audience{}
		expectedAudiences = append(expectedAudiences, rawGroupAudiences...)
		expectedAudiences = append(expectedAudiences, individualAudience...)
		assert.Equal(t, expectedAudiences, actualAudiences)
	})

	t.Run("Happy case audience selector, user group student, individual student and parent", func(t *testing.T) {
		targetGroupAudienceSelector := &entities.InfoNotificationTarget{
			LocationFilter:  entities.InfoNotificationTarget_LocationFilter{LocationIDs: locationIDs, Type: consts.TargetGroupSelectTypeList.String()},
			CourseFilter:    entities.InfoNotificationTarget_CourseFilter{CourseIDs: courseIDs, Type: consts.TargetGroupSelectTypeList.String()},
			ClassFilter:     entities.InfoNotificationTarget_ClassFilter{ClassIDs: []string{}, Type: consts.TargetGroupSelectTypeNone.String()},
			GradeFilter:     entities.InfoNotificationTarget_GradeFilter{GradeIDs: gradeIDs, Type: consts.TargetGroupSelectTypeList.String()},
			SchoolFilter:    entities.InfoNotificationTarget_SchoolFilter{SchoolIDs: []string{}, Type: consts.TargetGroupSelectTypeNone.String()},
			UserGroupFilter: entities.InfoNotificationTarget_UserGroupFilter{UserGroups: []string{cpb.UserGroup_USER_GROUP_STUDENT.String()}},
		}

		_ = notification.Type.Set(cpb.NotificationType_NOTIFICATION_TYPE_COMPOSED.String())
		_ = notification.TargetGroups.Set(targetGroupAudienceSelector)
		_ = notification.GenericReceiverIDs.Set(genericUserIDs)
		_ = notification.ReceiverIDs.Set(nil)

		audienceFilter := makeGroupAudienceFilter("list", "list", "none", "list", "none", "student", nil, notification.Type.String)
		audienceRepo.On("FindGroupAudiencesByFilter", ctx, mockDB.DB, audienceFilter, repositories.NewFindAudienceOption()).Once().Return(students, nil)

		infoNotiAccessPathRepo.On("GetLocationIDsByNotificationID", ctx, mockDB.DB, notiID).Once().Return(locationIDs, nil)
		locationRepo.On("GetLowestLocationIDsByIDs", ctx, mockDB.DB, locationIDs).Once().Return(locationIDs, nil)
		locationRepo.On("GetLowestGrantedLocationsByUserIDAndPermissions", ctx, mockDB.DB, userID, notificationPermissions).Once().Return(locationIDs, map[string]string{}, nil)

		individualFilter := makeIndividualFilter(locationIDs, genericUserIDs, notification.Type.String)
		individualAudience := []*entities.Audience{}
		individualAudience = append(individualAudience, individualStudents...)
		individualAudience = append(individualAudience, individualParents...)
		audienceRepo.On("FindIndividualAudiencesByFilter", ctx, mockDB.DB, individualFilter).Once().Return(individualAudience, nil)

		actualAudiences, err := svc.FindAudiences(ctx, mockDB.DB, &notification)
		assert.Nil(t, err)

		expectedAudiences := []*entities.Audience{}
		expectedAudiences = append(expectedAudiences, students...)
		expectedAudiences = append(expectedAudiences, individualAudience...)
		assert.Equal(t, expectedAudiences, actualAudiences)
	})

	t.Run("Happy case audience selector, user group student, individual student", func(t *testing.T) {
		targetGroupAudienceSelector := &entities.InfoNotificationTarget{
			LocationFilter:  entities.InfoNotificationTarget_LocationFilter{LocationIDs: locationIDs, Type: consts.TargetGroupSelectTypeList.String()},
			CourseFilter:    entities.InfoNotificationTarget_CourseFilter{CourseIDs: courseIDs, Type: consts.TargetGroupSelectTypeList.String()},
			ClassFilter:     entities.InfoNotificationTarget_ClassFilter{ClassIDs: []string{}, Type: consts.TargetGroupSelectTypeNone.String()},
			GradeFilter:     entities.InfoNotificationTarget_GradeFilter{GradeIDs: gradeIDs, Type: consts.TargetGroupSelectTypeList.String()},
			SchoolFilter:    entities.InfoNotificationTarget_SchoolFilter{SchoolIDs: schoolIDs, Type: consts.TargetGroupSelectTypeList.String()},
			UserGroupFilter: entities.InfoNotificationTarget_UserGroupFilter{UserGroups: []string{cpb.UserGroup_USER_GROUP_STUDENT.String(), cpb.UserGroup_USER_GROUP_PARENT.String()}},
		}

		_ = notification.Type.Set(cpb.NotificationType_NOTIFICATION_TYPE_COMPOSED.String())
		_ = notification.TargetGroups.Set(targetGroupAudienceSelector)
		_ = notification.GenericReceiverIDs.Set(individualStudentIDs)
		_ = notification.ReceiverIDs.Set(nil)

		audienceFilter := makeGroupAudienceFilter("list", "list", "none", "list", "list", "student,parent", nil, notification.Type.String)
		audienceRepo.On("FindGroupAudiencesByFilter", ctx, mockDB.DB, audienceFilter, repositories.NewFindAudienceOption()).Once().Return(students, nil)

		infoNotiAccessPathRepo.On("GetLocationIDsByNotificationID", ctx, mockDB.DB, notiID).Once().Return(locationIDs, nil)
		locationRepo.On("GetLowestLocationIDsByIDs", ctx, mockDB.DB, locationIDs).Once().Return(locationIDs, nil)
		locationRepo.On("GetLowestGrantedLocationsByUserIDAndPermissions", ctx, mockDB.DB, userID, notificationPermissions).Once().Return(locationIDs, map[string]string{}, nil)

		individualFilter := makeIndividualFilter(locationIDs, individualStudentIDs, notification.Type.String)
		individualAudience := []*entities.Audience{}
		individualAudience = append(individualAudience, individualStudents...)
		audienceRepo.On("FindIndividualAudiencesByFilter", ctx, mockDB.DB, individualFilter).Once().Return(individualAudience, nil)

		actualAudiences, err := svc.FindAudiences(ctx, mockDB.DB, &notification)
		assert.Nil(t, err)

		expectedAudiences := []*entities.Audience{}
		expectedAudiences = append(expectedAudiences, students...)
		expectedAudiences = append(expectedAudiences, individualAudience...)
		assert.Equal(t, expectedAudiences, actualAudiences)
	})

	t.Run("Happy case audience selector, user group parent, individual parent", func(t *testing.T) {
		targetGroupAudienceSelector := &entities.InfoNotificationTarget{
			LocationFilter:  entities.InfoNotificationTarget_LocationFilter{LocationIDs: locationIDs, Type: consts.TargetGroupSelectTypeList.String()},
			CourseFilter:    entities.InfoNotificationTarget_CourseFilter{CourseIDs: courseIDs, Type: consts.TargetGroupSelectTypeList.String()},
			ClassFilter:     entities.InfoNotificationTarget_ClassFilter{ClassIDs: []string{}, Type: consts.TargetGroupSelectTypeNone.String()},
			GradeFilter:     entities.InfoNotificationTarget_GradeFilter{GradeIDs: gradeIDs, Type: consts.TargetGroupSelectTypeList.String()},
			SchoolFilter:    entities.InfoNotificationTarget_SchoolFilter{SchoolIDs: schoolIDs, Type: consts.TargetGroupSelectTypeList.String()},
			UserGroupFilter: entities.InfoNotificationTarget_UserGroupFilter{UserGroups: []string{cpb.UserGroup_USER_GROUP_PARENT.String()}},
		}

		_ = notification.Type.Set(cpb.NotificationType_NOTIFICATION_TYPE_COMPOSED.String())
		_ = notification.TargetGroups.Set(targetGroupAudienceSelector)
		_ = notification.GenericReceiverIDs.Set(individualParentIDs)
		_ = notification.ReceiverIDs.Set(nil)

		audienceFilter := makeGroupAudienceFilter("list", "list", "none", "list", "list", "parent", nil, notification.Type.String)
		audienceRepo.On("FindGroupAudiencesByFilter", ctx, mockDB.DB, audienceFilter, repositories.NewFindAudienceOption()).Once().Return(students, nil)

		infoNotiAccessPathRepo.On("GetLocationIDsByNotificationID", ctx, mockDB.DB, notiID).Once().Return(locationIDs, nil)
		locationRepo.On("GetLowestLocationIDsByIDs", ctx, mockDB.DB, locationIDs).Once().Return(locationIDs, nil)
		locationRepo.On("GetLowestGrantedLocationsByUserIDAndPermissions", ctx, mockDB.DB, userID, notificationPermissions).Once().Return(locationIDs, map[string]string{}, nil)

		individualFilter := makeIndividualFilter(locationIDs, individualParentIDs, notification.Type.String)
		individualAudience := []*entities.Audience{}
		individualAudience = append(individualAudience, individualParents...)
		audienceRepo.On("FindIndividualAudiencesByFilter", ctx, mockDB.DB, individualFilter).Once().Return(individualAudience, nil)

		actualAudiences, err := svc.FindAudiences(ctx, mockDB.DB, &notification)
		assert.Nil(t, err)

		expectedAudiences := []*entities.Audience{}
		expectedAudiences = append(expectedAudiences, students...)
		expectedAudiences = append(expectedAudiences, individualAudience...)
		assert.Equal(t, expectedAudiences, actualAudiences)
	})

	t.Run("case no target recipients, only individual students", func(t *testing.T) {
		targetGroupAudienceSelector := &entities.InfoNotificationTarget{
			LocationFilter:  entities.InfoNotificationTarget_LocationFilter{LocationIDs: []string{}, Type: consts.TargetGroupSelectTypeNone.String()},
			CourseFilter:    entities.InfoNotificationTarget_CourseFilter{CourseIDs: []string{}, Type: consts.TargetGroupSelectTypeNone.String()},
			ClassFilter:     entities.InfoNotificationTarget_ClassFilter{ClassIDs: []string{}, Type: consts.TargetGroupSelectTypeNone.String()},
			GradeFilter:     entities.InfoNotificationTarget_GradeFilter{GradeIDs: []string{}, Type: consts.TargetGroupSelectTypeNone.String()},
			SchoolFilter:    entities.InfoNotificationTarget_SchoolFilter{SchoolIDs: []string{}, Type: consts.TargetGroupSelectTypeNone.String()},
			UserGroupFilter: entities.InfoNotificationTarget_UserGroupFilter{UserGroups: []string{cpb.UserGroup_USER_GROUP_STUDENT.String()}},
		}

		_ = notification.Type.Set(cpb.NotificationType_NOTIFICATION_TYPE_COMPOSED.String())
		_ = notification.TargetGroups.Set(targetGroupAudienceSelector)
		_ = notification.GenericReceiverIDs.Set(individualStudentIDs)
		_ = notification.ReceiverIDs.Set(nil)

		infoNotiAccessPathRepo.On("GetLocationIDsByNotificationID", ctx, mockDB.DB, notiID).Once().Return(locationIDs, nil)
		locationRepo.On("GetLowestLocationIDsByIDs", ctx, mockDB.DB, locationIDs).Once().Return(locationIDs, nil)
		locationRepo.On("GetLowestGrantedLocationsByUserIDAndPermissions", ctx, mockDB.DB, userID, notificationPermissions).Once().Return(locationIDs, map[string]string{}, nil)

		individualFilter := makeIndividualFilter(locationIDs, individualStudentIDs, notification.Type.String)
		individualAudience := []*entities.Audience{}
		individualAudience = append(individualAudience, individualStudents...)
		audienceRepo.On("FindIndividualAudiencesByFilter", ctx, mockDB.DB, individualFilter).Once().Return(individualAudience, nil)
		actualAudiences, err := svc.FindAudiences(ctx, mockDB.DB, &notification)
		assert.Nil(t, err)

		expectedAudiences := []*entities.Audience{}
		expectedAudiences = append(expectedAudiences, individualAudience...)
		assert.Equal(t, expectedAudiences, actualAudiences)
	})

	t.Run("case selected none for all target, individual parents", func(t *testing.T) {
		targetGroupAudienceSelector := &entities.InfoNotificationTarget{
			LocationFilter:  entities.InfoNotificationTarget_LocationFilter{LocationIDs: []string{}, Type: consts.TargetGroupSelectTypeNone.String()},
			CourseFilter:    entities.InfoNotificationTarget_CourseFilter{CourseIDs: []string{}, Type: consts.TargetGroupSelectTypeNone.String()},
			ClassFilter:     entities.InfoNotificationTarget_ClassFilter{ClassIDs: []string{}, Type: consts.TargetGroupSelectTypeNone.String()},
			GradeFilter:     entities.InfoNotificationTarget_GradeFilter{GradeIDs: []string{}, Type: consts.TargetGroupSelectTypeNone.String()},
			SchoolFilter:    entities.InfoNotificationTarget_SchoolFilter{SchoolIDs: []string{}, Type: consts.TargetGroupSelectTypeNone.String()},
			UserGroupFilter: entities.InfoNotificationTarget_UserGroupFilter{UserGroups: []string{cpb.UserGroup_USER_GROUP_PARENT.String()}},
		}

		_ = notification.Type.Set(cpb.NotificationType_NOTIFICATION_TYPE_COMPOSED.String())
		_ = notification.TargetGroups.Set(targetGroupAudienceSelector)
		_ = notification.GenericReceiverIDs.Set(individualParentIDs)
		_ = notification.ReceiverIDs.Set(nil)

		infoNotiAccessPathRepo.On("GetLocationIDsByNotificationID", ctx, mockDB.DB, notiID).Once().Return(locationIDs, nil)
		locationRepo.On("GetLowestLocationIDsByIDs", ctx, mockDB.DB, locationIDs).Once().Return(locationIDs, nil)
		locationRepo.On("GetLowestGrantedLocationsByUserIDAndPermissions", ctx, mockDB.DB, userID, notificationPermissions).Once().Return(locationIDs, map[string]string{}, nil)

		individualFilter := makeIndividualFilter(locationIDs, individualParentIDs, notification.Type.String)
		individualAudience := []*entities.Audience{}
		individualAudience = append(individualAudience, individualParents...)
		audienceRepo.On("FindIndividualAudiencesByFilter", ctx, mockDB.DB, individualFilter).Once().Return(individualAudience, nil)

		actualAudiences, err := svc.FindAudiences(ctx, mockDB.DB, &notification)
		assert.Nil(t, err)

		expectedAudiences := []*entities.Audience{}
		expectedAudiences = append(expectedAudiences, individualAudience...)
		assert.Equal(t, expectedAudiences, actualAudiences)
	})

	t.Run("case selected none for all target, only individual student with user filter, composed notification", func(t *testing.T) {
		targetGroupAudienceSelector := &entities.InfoNotificationTarget{
			LocationFilter:  entities.InfoNotificationTarget_LocationFilter{LocationIDs: []string{}, Type: consts.TargetGroupSelectTypeNone.String()},
			CourseFilter:    entities.InfoNotificationTarget_CourseFilter{CourseIDs: []string{}, Type: consts.TargetGroupSelectTypeNone.String()},
			ClassFilter:     entities.InfoNotificationTarget_ClassFilter{ClassIDs: []string{}, Type: consts.TargetGroupSelectTypeNone.String()},
			GradeFilter:     entities.InfoNotificationTarget_GradeFilter{GradeIDs: []string{}, Type: consts.TargetGroupSelectTypeNone.String()},
			SchoolFilter:    entities.InfoNotificationTarget_SchoolFilter{SchoolIDs: []string{}, Type: consts.TargetGroupSelectTypeNone.String()},
			UserGroupFilter: entities.InfoNotificationTarget_UserGroupFilter{UserGroups: []string{cpb.UserGroup_USER_GROUP_PARENT.String()}},
		}

		_ = notification.Type.Set(cpb.NotificationType_NOTIFICATION_TYPE_COMPOSED.String())
		_ = notification.TargetGroups.Set(targetGroupAudienceSelector)
		_ = notification.GenericReceiverIDs.Set(nil)
		_ = notification.ReceiverIDs.Set(studentIDs)

		infoNotiAccessPathRepo.On("GetLocationIDsByNotificationID", ctx, mockDB.DB, notiID).Once().Return(locationIDs, nil)
		locationRepo.On("GetLowestLocationIDsByIDs", ctx, mockDB.DB, locationIDs).Once().Return(locationIDs, nil)
		locationRepo.On("GetLowestGrantedLocationsByUserIDAndPermissions", ctx, mockDB.DB, userID, notificationPermissions).Once().Return(locationIDs, map[string]string{}, nil)

		individualFilter := makeIndividualFilter(locationIDs, studentIDs, notification.Type.String)
		audienceRepo.On("FindIndividualAudiencesByFilter", ctx, mockDB.DB, individualFilter).Once().Return(students, nil)

		studentParentRepo.On("GetStudentParents", ctx, mockDB.DB, database.TextArray(studentIDs)).Once().Return(studentParents, nil)

		actualAudiences, err := svc.FindAudiences(ctx, mockDB.DB, &notification)
		assert.Nil(t, err)

		expectedAudiences := []*entities.Audience{}
		expectedAudiences = append(expectedAudiences, parentIndividualWithUserGroups...)
		assert.Equal(t, expectedAudiences, actualAudiences)
		fmt.Printf("[%v]", actualAudiences)
	})

	t.Run("case selected none for all target, only individual student with user filter, async notification", func(t *testing.T) {
		targetGroupAudienceSelector := &entities.InfoNotificationTarget{
			LocationFilter:  entities.InfoNotificationTarget_LocationFilter{LocationIDs: []string{}, Type: consts.TargetGroupSelectTypeNone.String()},
			CourseFilter:    entities.InfoNotificationTarget_CourseFilter{CourseIDs: []string{}, Type: consts.TargetGroupSelectTypeNone.String()},
			ClassFilter:     entities.InfoNotificationTarget_ClassFilter{ClassIDs: []string{}, Type: consts.TargetGroupSelectTypeNone.String()},
			GradeFilter:     entities.InfoNotificationTarget_GradeFilter{GradeIDs: []string{}, Type: consts.TargetGroupSelectTypeNone.String()},
			SchoolFilter:    entities.InfoNotificationTarget_SchoolFilter{SchoolIDs: []string{}, Type: consts.TargetGroupSelectTypeNone.String()},
			UserGroupFilter: entities.InfoNotificationTarget_UserGroupFilter{UserGroups: []string{cpb.UserGroup_USER_GROUP_PARENT.String()}},
		}

		_ = notification.Type.Set(cpb.NotificationType_NOTIFICATION_TYPE_NATS_ASYNC.String())
		_ = notification.TargetGroups.Set(targetGroupAudienceSelector)
		_ = notification.GenericReceiverIDs.Set(nil)
		_ = notification.ReceiverIDs.Set(studentIDs)

		infoNotiAccessPathRepo.On("GetLocationIDsByNotificationID", ctx, mockDB.DB, notiID).Once().Return(locationIDs, nil)
		locationRepo.On("GetLowestLocationIDsByIDs", ctx, mockDB.DB, locationIDs).Once().Return(locationIDs, nil)
		locationRepo.On("GetLowestGrantedLocationsByUserIDAndPermissions", ctx, mockDB.DB, userID, notificationPermissions).Once().Return(locationIDs, map[string]string{}, nil)

		individualFilter := makeIndividualFilter(locationIDs, studentIDs, notification.Type.String)
		audienceRepo.On("FindIndividualAudiencesByFilter", ctx, mockDB.DB, individualFilter).Once().Return(students, nil)

		studentParentRepo.On("GetStudentParents", ctx, mockDB.DB, database.TextArray(studentIDs)).Once().Return(studentParents, nil)

		actualAudiences, err := svc.FindAudiences(ctx, mockDB.DB, &notification)
		assert.Nil(t, err)

		expectedAudiences := []*entities.Audience{}
		expectedAudiences = append(expectedAudiences, parentIndividualWithUserGroups...)
		assert.Equal(t, expectedAudiences, actualAudiences)
		fmt.Printf("[%v]", actualAudiences)
	})
}

func TestAudienceRetriever_FindGroupAudiencesWithPaging(t *testing.T) {
	t.Parallel()
	mockDB := testutil.NewMockDB()
	audienceRepo := &mock_repositories.MockAudienceRepo{}
	infoNotiAccessPathRepo := &mock_repositories.MockInfoNotificationAccessPathRepo{}
	locationRepo := &mock_repositories.MockLocationRepo{}
	userRepo := &mock_repositories.MockUserRepo{}
	gradeRepo := &mock_repositories.MockGradeRepo{}
	svc := &AudienceRetrieverService{
		AudienceRepo:                   audienceRepo,
		InfoNotificationAccessPathRepo: infoNotiAccessPathRepo,
		LocationRepo:                   locationRepo,
		UserRepo:                       userRepo,
		GradeRepo:                      gradeRepo,
		Env:                            "prod",
	}

	// notiID := "noti-id"
	studentNames := []string{"student_name_1", "student_name_2", "student_name_3"}
	studentIDs := []string{"student_id_1", "student_id_2", "student_id_3"}
	studentNameMap := make(map[string]*entities.User)
	studentNameMap[studentIDs[0]] = &entities.User{
		UserID: database.Text(studentIDs[0]),
		Name:   database.Text(studentNames[0]),
	}
	studentNameMap[studentIDs[1]] = &entities.User{
		UserID: database.Text(studentIDs[1]),
		Name:   database.Text(studentNames[1]),
	}
	studentNameMap[studentIDs[2]] = &entities.User{
		UserID: database.Text(studentIDs[2]),
		Name:   database.Text(studentNames[2]),
	}
	parentIDs := []string{"parent_id_1", "parent_id_2", "parent_id_3"}
	locationIDs := []string{"loc-1", "loc-2"}
	courseIDs := []string{"course_id_1", "course_id_2", "course_id_3"}
	classIDs := []string{"class-id-1", "class-id-2"}
	grades := []string{"grade 1", "grade 2", "grade 3"}
	gradeIDs := []string{"grade-id-1", "grade-id-2", "grade-id-3"}
	gradeMap := make(map[string]string)
	gradeMap[gradeIDs[0]] = grades[0]
	gradeMap[gradeIDs[1]] = grades[1]
	gradeMap[gradeIDs[2]] = grades[2]
	keyword := "keyword"
	limit := 10
	offset := 0
	students := []*entities.Audience{
		{
			UserID:       database.Text(studentIDs[0]),
			CurrentGrade: database.Int2(1),
			GradeID:      database.Text(gradeIDs[0]),
			GradeName:    database.Text(grades[0]),
		},
		{
			UserID:       database.Text(studentIDs[1]),
			CurrentGrade: database.Int2(2),
			GradeID:      database.Text(gradeIDs[1]),
			GradeName:    database.Text(grades[1]),
		},
		{
			UserID:       database.Text(studentIDs[2]),
			CurrentGrade: database.Int2(3),
			GradeID:      database.Text(gradeIDs[2]),
			GradeName:    database.Text(grades[2]),
		},
	}

	parents := []*entities.Audience{
		{
			UserID: database.Text(parentIDs[0]),
			ChildIDs: database.TextArray([]string{
				studentIDs[0],
			}),
		},
		{
			UserID: database.Text(parentIDs[1]),
			ChildIDs: database.TextArray([]string{
				studentIDs[1],
			}),
		},
		{
			UserID: database.Text(parentIDs[2]),
			ChildIDs: database.TextArray([]string{
				studentIDs[2],
			}),
		},
	}

	userID := "user-id"
	claim := &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: fmt.Sprint(constants.ManabieSchool),
			UserID:       userID,
		},
	}
	ctx := interceptors.ContextWithUserID(context.Background(), userID)
	ctx = interceptors.ContextWithJWTClaims(ctx, claim)
	t.Run("case fail AudienceRepo", func(t *testing.T) {
		targetGroupAudienceSelector := &entities.InfoNotificationTarget{
			LocationFilter:  entities.InfoNotificationTarget_LocationFilter{LocationIDs: locationIDs, Type: consts.TargetGroupSelectTypeList.String()},
			CourseFilter:    entities.InfoNotificationTarget_CourseFilter{CourseIDs: courseIDs, Type: consts.TargetGroupSelectTypeList.String()},
			ClassFilter:     entities.InfoNotificationTarget_ClassFilter{ClassIDs: classIDs, Type: consts.TargetGroupSelectTypeList.String()},
			GradeFilter:     entities.InfoNotificationTarget_GradeFilter{GradeIDs: gradeIDs, Type: consts.TargetGroupSelectTypeNone.String()},
			UserGroupFilter: entities.InfoNotificationTarget_UserGroupFilter{UserGroups: []string{cpb.UserGroup_USER_GROUP_STUDENT.String(), cpb.UserGroup_USER_GROUP_PARENT.String()}},
		}

		audiences := make([]*entities.Audience, 0)
		audiences = append(audiences, students...)
		audiences = append(audiences, parents...)

		opts := repositories.NewFindAudienceOption()
		opts.OrderByName = consts.AscendingOrder
		opts.IsGetName = true

		audienceRepo.On("FindGroupAudiencesByFilter", ctx, mockDB.DB, mock.Anything, opts).Once().Return(audiences, fmt.Errorf("error1"))
		audienceRepo.On("CountGroupAudiencesByFilter", ctx, mockDB.DB, mock.Anything, opts).Once().Return(uint32(len(audiences)), fmt.Errorf("error2"))

		actualAudiences, actualCount, err := svc.FindGroupAudiencesWithPaging(ctx, mockDB.DB, "", targetGroupAudienceSelector, keyword, []string{}, limit, offset)

		assert.ErrorContains(t, err, "error1")
		assert.ErrorContains(t, err, "error2")
		assert.Nil(t, actualAudiences)
		assert.Equal(t, uint32(0), actualCount)
	})
	t.Run("Happy case with selected courses locations classes grades", func(t *testing.T) {
		targetGroupAudienceSelector := &entities.InfoNotificationTarget{
			LocationFilter:  entities.InfoNotificationTarget_LocationFilter{LocationIDs: locationIDs, Type: consts.TargetGroupSelectTypeList.String()},
			CourseFilter:    entities.InfoNotificationTarget_CourseFilter{CourseIDs: courseIDs, Type: consts.TargetGroupSelectTypeList.String()},
			ClassFilter:     entities.InfoNotificationTarget_ClassFilter{ClassIDs: classIDs, Type: consts.TargetGroupSelectTypeList.String()},
			GradeFilter:     entities.InfoNotificationTarget_GradeFilter{GradeIDs: gradeIDs, Type: consts.TargetGroupSelectTypeNone.String()},
			UserGroupFilter: entities.InfoNotificationTarget_UserGroupFilter{UserGroups: []string{cpb.UserGroup_USER_GROUP_STUDENT.String(), cpb.UserGroup_USER_GROUP_PARENT.String()}},
		}

		audiences := make([]*entities.Audience, 0)
		audiences = append(audiences, students...)
		audiences = append(audiences, parents...)

		opts := repositories.NewFindAudienceOption()
		opts.OrderByName = consts.AscendingOrder
		opts.IsGetName = true

		audienceRepo.On("FindGroupAudiencesByFilter", ctx, mockDB.DB, mock.Anything, opts).Once().Return(audiences, nil)
		audienceRepo.On("CountGroupAudiencesByFilter", ctx, mockDB.DB, mock.Anything, opts).Once().Return(uint32(len(audiences)), nil)

		gradeRepo.On("GetGradesByOrg", ctx, mockDB.DB, fmt.Sprint(constants.ManabieSchool)).
			Once().Return(gradeMap, nil)

		userRepo.On("FindUser", ctx, mockDB.DB, mock.Anything).
			Once().Return(nil, studentNameMap, nil)

		actualAudiences, actualCount, err := svc.FindGroupAudiencesWithPaging(ctx, mockDB.DB, "", targetGroupAudienceSelector, keyword, []string{}, limit, offset)
		assert.Nil(t, err)

		assert.Equal(t, audiences, actualAudiences)
		assert.Equal(t, uint32(len(audiences)), actualCount)
	})
	t.Run("Happy case with selected courses classes grades, no locations", func(t *testing.T) {
		targetGroupAudienceSelector := &entities.InfoNotificationTarget{
			LocationFilter:  entities.InfoNotificationTarget_LocationFilter{LocationIDs: []string{}, Type: consts.TargetGroupSelectTypeAll.String()},
			CourseFilter:    entities.InfoNotificationTarget_CourseFilter{CourseIDs: courseIDs, Type: consts.TargetGroupSelectTypeList.String()},
			ClassFilter:     entities.InfoNotificationTarget_ClassFilter{ClassIDs: classIDs, Type: consts.TargetGroupSelectTypeList.String()},
			GradeFilter:     entities.InfoNotificationTarget_GradeFilter{GradeIDs: gradeIDs, Type: consts.TargetGroupSelectTypeNone.String()},
			UserGroupFilter: entities.InfoNotificationTarget_UserGroupFilter{UserGroups: []string{cpb.UserGroup_USER_GROUP_STUDENT.String(), cpb.UserGroup_USER_GROUP_PARENT.String()}},
		}

		notificationPermissions := []string{
			consts.NotificationWritePermission,
			consts.NotificationOwnerPermission,
		}
		locationRepo.On("GetLowestGrantedLocationsByUserIDAndPermissions", ctx, mockDB.DB, userID, notificationPermissions).
			Once().Return(locationIDs, map[string]string{}, nil)

		audiences := make([]*entities.Audience, 0)
		audiences = append(audiences, students...)
		audiences = append(audiences, parents...)

		opts := repositories.NewFindAudienceOption()
		opts.OrderByName = consts.AscendingOrder
		opts.IsGetName = true

		audienceRepo.On("FindGroupAudiencesByFilter", ctx, mockDB.DB, mock.Anything, opts).Once().Return(audiences, nil)
		audienceRepo.On("CountGroupAudiencesByFilter", ctx, mockDB.DB, mock.Anything, opts).Once().Return(uint32(len(audiences)), nil)

		gradeRepo.On("GetGradesByOrg", ctx, mockDB.DB, fmt.Sprint(constants.ManabieSchool)).
			Once().Return(gradeMap, nil)

		userRepo.On("FindUser", ctx, mockDB.DB, mock.Anything).
			Once().Return(nil, studentNameMap, nil)

		actualAudiences, actualCount, err := svc.FindGroupAudiencesWithPaging(ctx, mockDB.DB, "", targetGroupAudienceSelector, keyword, []string{}, limit, offset)
		assert.Nil(t, err)

		assert.Equal(t, audiences, actualAudiences)
		assert.Equal(t, uint32(len(audiences)), actualCount)
	})
	t.Run("Happy case include user ids", func(t *testing.T) {
		targetGroupAudienceSelector := &entities.InfoNotificationTarget{
			LocationFilter:  entities.InfoNotificationTarget_LocationFilter{LocationIDs: []string{}, Type: consts.TargetGroupSelectTypeAll.String()},
			CourseFilter:    entities.InfoNotificationTarget_CourseFilter{CourseIDs: courseIDs, Type: consts.TargetGroupSelectTypeList.String()},
			ClassFilter:     entities.InfoNotificationTarget_ClassFilter{ClassIDs: classIDs, Type: consts.TargetGroupSelectTypeList.String()},
			GradeFilter:     entities.InfoNotificationTarget_GradeFilter{GradeIDs: gradeIDs, Type: consts.TargetGroupSelectTypeNone.String()},
			UserGroupFilter: entities.InfoNotificationTarget_UserGroupFilter{UserGroups: []string{cpb.UserGroup_USER_GROUP_STUDENT.String(), cpb.UserGroup_USER_GROUP_PARENT.String()}},
		}

		notificationPermissions := []string{
			consts.NotificationWritePermission,
			consts.NotificationOwnerPermission,
		}
		locationRepo.On("GetLowestGrantedLocationsByUserIDAndPermissions", ctx, mockDB.DB, userID, notificationPermissions).
			Once().Return(locationIDs, map[string]string{}, nil)

		audiences := make([]*entities.Audience, 0)

		opts := repositories.NewFindAudienceOption()
		opts.OrderByName = consts.AscendingOrder
		opts.IsGetName = true

		audienceRepo.On("FindGroupAudiencesByFilter", ctx, mockDB.DB, mock.Anything, opts).Once().Return(audiences, nil)
		audienceRepo.On("CountGroupAudiencesByFilter", ctx, mockDB.DB, mock.Anything, opts).Once().Return(uint32(len(audiences)), nil)

		gradeRepo.On("GetGradesByOrg", ctx, mockDB.DB, fmt.Sprint(constants.ManabieSchool)).
			Once().Return(make(map[string]string), nil)

		userRepo.On("FindUser", ctx, mockDB.DB, mock.Anything).
			Once().Return(nil, make(map[string]*entities.User), nil)

		actualAudiences, actualCount, err := svc.FindGroupAudiencesWithPaging(ctx, mockDB.DB, "", targetGroupAudienceSelector, keyword, []string{}, limit, offset)
		assert.Nil(t, err)

		assert.Equal(t, audiences, actualAudiences)
		assert.Equal(t, uint32(len(audiences)), actualCount)
	})
}

func TestAudienceRetriever_FindDraftAudiencesWithPaging(t *testing.T) {
	t.Parallel()
	mockDB := testutil.NewMockDB()
	audienceRepo := &mock_repositories.MockAudienceRepo{}
	infoNotiAccessPathRepo := &mock_repositories.MockInfoNotificationAccessPathRepo{}
	locationRepo := &mock_repositories.MockLocationRepo{}
	userRepo := &mock_repositories.MockUserRepo{}
	gradeRepo := &mock_repositories.MockGradeRepo{}
	svc := &AudienceRetrieverService{
		AudienceRepo:                   audienceRepo,
		InfoNotificationAccessPathRepo: infoNotiAccessPathRepo,
		LocationRepo:                   locationRepo,
		UserRepo:                       userRepo,
		GradeRepo:                      gradeRepo,
		Env:                            "prod",
	}

	// notiID := "noti-id"
	studentNames := []string{"student_name_1", "student_name_2", "student_name_3", "student_individual_name_4"}
	studentIDs := []string{"student_id_1", "student_id_2", "student_id_3", "student_individual_id_4"}
	studentNameMap := make(map[string]*entities.User)
	studentNameMap[studentIDs[0]] = &entities.User{
		UserID: database.Text(studentIDs[0]),
		Name:   database.Text(studentNames[0]),
	}
	studentNameMap[studentIDs[1]] = &entities.User{
		UserID: database.Text(studentIDs[1]),
		Name:   database.Text(studentNames[1]),
	}
	studentNameMap[studentIDs[2]] = &entities.User{
		UserID: database.Text(studentIDs[2]),
		Name:   database.Text(studentNames[2]),
	}
	parentIDs := []string{"parent_id_1", "parent_id_2", "parent_id_3"}
	locationIDs := []string{"loc-1", "loc-2"}
	courseIDs := []string{"course_id_1", "course_id_2", "course_id_3"}
	classIDs := []string{"class-id-1", "class-id-2"}
	grades := []string{"grade 1", "grade 2", "grade 3"}
	gradeIDs := []string{"grade-id-1", "grade-id-2", "grade-id-3"}
	gradeMap := make(map[string]string)
	gradeMap[gradeIDs[0]] = grades[0]
	gradeMap[gradeIDs[1]] = grades[1]
	gradeMap[gradeIDs[2]] = grades[2]
	notificationID := "notification-id"
	limit := 10
	offset := 0
	students := []*entities.Audience{
		{
			UserID:       database.Text(studentIDs[0]),
			CurrentGrade: database.Int2(1),
			GradeID:      database.Text(gradeIDs[0]),
			GradeName:    database.Text(grades[0]),
		},
		{
			UserID:       database.Text(studentIDs[1]),
			CurrentGrade: database.Int2(2),
			GradeID:      database.Text(gradeIDs[1]),
			GradeName:    database.Text(grades[1]),
		},
		{
			UserID:       database.Text(studentIDs[2]),
			CurrentGrade: database.Int2(3),
			GradeID:      database.Text(gradeIDs[2]),
			GradeName:    database.Text(grades[2]),
		},
	}

	individualStudents := []*entities.Audience{
		{
			UserID:       database.Text(studentIDs[3]),
			CurrentGrade: database.Int2(1),
			GradeID:      database.Text(gradeIDs[0]),
			GradeName:    database.Text(grades[0]),
		},
	}

	parents := []*entities.Audience{
		{
			UserID: database.Text(parentIDs[0]),
			ChildIDs: database.TextArray([]string{
				studentIDs[0],
			}),
		},
		{
			UserID: database.Text(parentIDs[1]),
			ChildIDs: database.TextArray([]string{
				studentIDs[1],
			}),
		},
		{
			UserID: database.Text(parentIDs[2]),
			ChildIDs: database.TextArray([]string{
				studentIDs[2],
			}),
		},
	}

	userID := "user-id"
	claim := &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: fmt.Sprint(constants.ManabieSchool),
			UserID:       userID,
		},
	}

	notificationPermissions := []string{
		consts.NotificationWritePermission,
		consts.NotificationOwnerPermission,
	}
	ctx := interceptors.ContextWithUserID(context.Background(), userID)
	ctx = interceptors.ContextWithJWTClaims(ctx, claim)
	t.Run("Happy case with selected courses locations classes grades and no individual", func(t *testing.T) {
		targetGroupAudienceSelector := &entities.InfoNotificationTarget{
			LocationFilter:  entities.InfoNotificationTarget_LocationFilter{LocationIDs: locationIDs, Type: consts.TargetGroupSelectTypeList.String()},
			CourseFilter:    entities.InfoNotificationTarget_CourseFilter{CourseIDs: courseIDs, Type: consts.TargetGroupSelectTypeList.String()},
			ClassFilter:     entities.InfoNotificationTarget_ClassFilter{ClassIDs: classIDs, Type: consts.TargetGroupSelectTypeList.String()},
			GradeFilter:     entities.InfoNotificationTarget_GradeFilter{GradeIDs: gradeIDs, Type: consts.TargetGroupSelectTypeNone.String()},
			UserGroupFilter: entities.InfoNotificationTarget_UserGroupFilter{UserGroups: []string{cpb.UserGroup_USER_GROUP_STUDENT.String(), cpb.UserGroup_USER_GROUP_PARENT.String()}},
		}
		genericReceiverIDs := []string{}

		groupFilter, err := svc.makeGroupAudienceFilter(ctx, mockDB.DB, notificationID, cpb.NotificationType_NOTIFICATION_TYPE_COMPOSED.String(), targetGroupAudienceSelector, nil)
		individualFilter := repositories.NewFindIndividualAudienceFilter()

		draftFilter := repositories.NewFindDraftAudienceFilter()
		draftFilter.GroupFilter = groupFilter
		draftFilter.IndividualFilter = individualFilter
		draftFilter.Limit.Set(limit)
		draftFilter.Offset.Set(offset)

		audiences := make([]*entities.Audience, 0)
		audiences = append(audiences, students...)
		audiences = append(audiences, parents...)

		opts := repositories.NewFindAudienceOption()
		opts.OrderByName = consts.AscendingOrder
		opts.IsGetName = true

		audienceRepo.On("FindDraftAudiencesByFilter", ctx, mockDB.DB, draftFilter, opts).Once().Return(audiences, nil)
		audienceRepo.On("CountDraftAudiencesByFilter", ctx, mockDB.DB, draftFilter, opts).Once().Return(uint32(len(audiences)), nil)
		gradeRepo.On("GetGradesByOrg", ctx, mockDB.DB, fmt.Sprint(constants.ManabieSchool)).Once().Return(gradeMap, nil)
		userRepo.On("FindUser", ctx, mockDB.DB, mock.Anything).Once().Return(nil, studentNameMap, nil)

		actualAudiences, actualCount, err := svc.FindDraftAudiencesWithPaging(ctx, mockDB.DB, notificationID, targetGroupAudienceSelector, genericReceiverIDs, nil, limit, offset)
		assert.Nil(t, err)

		assert.Equal(t, audiences, actualAudiences)
		assert.Equal(t, uint32(len(audiences)), actualCount)
	})

	t.Run("Happy case with selected courses classes grades, no locations, and no individual", func(t *testing.T) {
		targetGroupAudienceSelector := &entities.InfoNotificationTarget{
			LocationFilter:  entities.InfoNotificationTarget_LocationFilter{LocationIDs: locationIDs, Type: consts.TargetGroupSelectTypeAll.String()},
			CourseFilter:    entities.InfoNotificationTarget_CourseFilter{CourseIDs: courseIDs, Type: consts.TargetGroupSelectTypeList.String()},
			ClassFilter:     entities.InfoNotificationTarget_ClassFilter{ClassIDs: classIDs, Type: consts.TargetGroupSelectTypeList.String()},
			GradeFilter:     entities.InfoNotificationTarget_GradeFilter{GradeIDs: gradeIDs, Type: consts.TargetGroupSelectTypeList.String()},
			UserGroupFilter: entities.InfoNotificationTarget_UserGroupFilter{UserGroups: []string{cpb.UserGroup_USER_GROUP_STUDENT.String(), cpb.UserGroup_USER_GROUP_PARENT.String()}},
		}

		genericReceiverIDs := []string{}

		locationRepo.On("GetLowestLocationIDsByIDs", ctx, mockDB.DB, locationIDs).Once().Return(locationIDs, nil)
		individualFilter := repositories.NewFindIndividualAudienceFilter()

		groupFilter, err := svc.makeGroupAudienceFilter(ctx, mockDB.DB, notificationID, cpb.NotificationType_NOTIFICATION_TYPE_COMPOSED.String(), targetGroupAudienceSelector, nil)
		locationRepo.On("GetLowestLocationIDsByIDs", ctx, mockDB.DB, locationIDs).Once().Return(locationIDs, nil)

		draftFilter := repositories.NewFindDraftAudienceFilter()
		draftFilter.GroupFilter = groupFilter
		draftFilter.IndividualFilter = individualFilter
		draftFilter.Limit.Set(limit)
		draftFilter.Offset.Set(offset)

		audiences := make([]*entities.Audience, 0)
		audiences = append(audiences, students...)
		audiences = append(audiences, parents...)

		opts := repositories.NewFindAudienceOption()
		opts.OrderByName = consts.AscendingOrder
		opts.IsGetName = true

		audienceRepo.On("FindDraftAudiencesByFilter", ctx, mockDB.DB, draftFilter, opts).Once().Return(audiences, nil)
		audienceRepo.On("CountDraftAudiencesByFilter", ctx, mockDB.DB, draftFilter, opts).Once().Return(uint32(len(audiences)), nil)
		gradeRepo.On("GetGradesByOrg", ctx, mockDB.DB, fmt.Sprint(constants.ManabieSchool)).Once().Return(gradeMap, nil)
		userRepo.On("FindUser", ctx, mockDB.DB, mock.Anything).Once().Return(nil, studentNameMap, nil)

		actualAudiences, actualCount, err := svc.FindDraftAudiencesWithPaging(ctx, mockDB.DB, notificationID, targetGroupAudienceSelector, genericReceiverIDs, nil, limit, offset)
		assert.Nil(t, err)

		assert.Equal(t, audiences, actualAudiences)
		assert.Equal(t, uint32(len(audiences)), actualCount)
	})

	t.Run("Happy case with selected courses locations classes grades and individual", func(t *testing.T) {
		targetGroupAudienceSelector := &entities.InfoNotificationTarget{
			LocationFilter:  entities.InfoNotificationTarget_LocationFilter{LocationIDs: locationIDs, Type: consts.TargetGroupSelectTypeList.String()},
			CourseFilter:    entities.InfoNotificationTarget_CourseFilter{CourseIDs: courseIDs, Type: consts.TargetGroupSelectTypeList.String()},
			ClassFilter:     entities.InfoNotificationTarget_ClassFilter{ClassIDs: classIDs, Type: consts.TargetGroupSelectTypeList.String()},
			GradeFilter:     entities.InfoNotificationTarget_GradeFilter{GradeIDs: gradeIDs, Type: consts.TargetGroupSelectTypeNone.String()},
			UserGroupFilter: entities.InfoNotificationTarget_UserGroupFilter{UserGroups: []string{cpb.UserGroup_USER_GROUP_STUDENT.String(), cpb.UserGroup_USER_GROUP_PARENT.String()}},
		}
		genericReceiverIDs := []string{"student_individual_id_4"}

		groupFilter, err := svc.makeGroupAudienceFilter(ctx, mockDB.DB, notificationID, cpb.NotificationType_NOTIFICATION_TYPE_COMPOSED.String(), targetGroupAudienceSelector, nil)

		locationRepo.On("GetLowestGrantedLocationsByUserIDAndPermissions", ctx, mockDB.DB, userID, notificationPermissions).Once().Return(locationIDs, map[string]string{}, nil)

		individualFilter := repositories.NewFindIndividualAudienceFilter()
		individualFilter.LocationIDs.Set(locationIDs)
		individualFilter.UserIDs.Set(genericReceiverIDs)
		individualFilter.EnrollmentStatuses.Set([]string{pbu.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()})

		draftFilter := repositories.NewFindDraftAudienceFilter()
		draftFilter.GroupFilter = groupFilter
		draftFilter.IndividualFilter = individualFilter
		draftFilter.Limit.Set(limit)
		draftFilter.Offset.Set(offset)

		audiences := make([]*entities.Audience, 0)
		audiences = append(audiences, students...)
		audiences = append(audiences, parents...)
		audiences = append(audiences, individualStudents...)

		opts := repositories.NewFindAudienceOption()
		opts.OrderByName = consts.AscendingOrder
		opts.IsGetName = true

		audienceRepo.On("FindDraftAudiencesByFilter", ctx, mockDB.DB, draftFilter, opts).Once().Return(audiences, nil)
		audienceRepo.On("CountDraftAudiencesByFilter", ctx, mockDB.DB, draftFilter, opts).Once().Return(uint32(len(audiences)), nil)
		gradeRepo.On("GetGradesByOrg", ctx, mockDB.DB, fmt.Sprint(constants.ManabieSchool)).Once().Return(gradeMap, nil)
		userRepo.On("FindUser", ctx, mockDB.DB, mock.Anything).Once().Return(nil, studentNameMap, nil)

		actualAudiences, actualCount, err := svc.FindDraftAudiencesWithPaging(ctx, mockDB.DB, notificationID, targetGroupAudienceSelector, genericReceiverIDs, nil, limit, offset)
		assert.Nil(t, err)

		assert.Equal(t, audiences, actualAudiences)
		assert.Equal(t, uint32(len(audiences)), actualCount)
	})
}

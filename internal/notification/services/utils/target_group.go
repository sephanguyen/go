package utils

import (
	"github.com/manabie-com/backend/internal/notification/consts"
	"github.com/manabie-com/backend/internal/notification/entities"
)

func CheckNoneSelectTargetGroup(targetGroup *entities.InfoNotificationTarget) bool {
	if targetGroup.LocationFilter.Type == consts.TargetGroupSelectTypeNone.String() &&
		targetGroup.CourseFilter.Type == consts.TargetGroupSelectTypeNone.String() &&
		targetGroup.GradeFilter.Type == consts.TargetGroupSelectTypeNone.String() &&
		targetGroup.ClassFilter.Type == consts.TargetGroupSelectTypeNone.String() &&
		targetGroup.SchoolFilter.Type == consts.TargetGroupSelectTypeNone.String() {
		return true
	}
	return false
}

func CourseTargetGroupToCourseFilter(courseTargetGroup entities.InfoNotificationTarget_CourseFilter) (courseIDs []string, courseSelectType string) {
	switch courseTargetGroup.Type {
	case consts.TargetGroupSelectTypeNone.String():
		return nil, consts.TargetGroupSelectTypeNone.String()
	case consts.TargetGroupSelectTypeAll.String():
		return nil, consts.TargetGroupSelectTypeAll.String()
	case consts.TargetGroupSelectTypeList.String():
		return courseTargetGroup.CourseIDs, consts.TargetGroupSelectTypeList.String()
	}
	return
}

func ClassTargetGroupToClassFilter(classTargetGroup entities.InfoNotificationTarget_ClassFilter) (classIDs []string, classSelectType string) {
	switch classTargetGroup.Type {
	case consts.TargetGroupSelectTypeNone.String():
		return nil, consts.TargetGroupSelectTypeNone.String()
	case consts.TargetGroupSelectTypeAll.String():
		return nil, consts.TargetGroupSelectTypeAll.String()
	case consts.TargetGroupSelectTypeList.String():
		return classTargetGroup.ClassIDs, consts.TargetGroupSelectTypeList.String()
	}
	return
}

func GradeTargetGroupToGradeFilter(gradeTargetGroup entities.InfoNotificationTarget_GradeFilter) (gradeIDs []string, gradeSelectType string) {
	switch gradeTargetGroup.Type {
	case consts.TargetGroupSelectTypeNone.String():
		return nil, consts.TargetGroupSelectTypeNone.String()
	case consts.TargetGroupSelectTypeAll.String():
		return nil, consts.TargetGroupSelectTypeAll.String()
	case consts.TargetGroupSelectTypeList.String():
		return gradeTargetGroup.GradeIDs, consts.TargetGroupSelectTypeList.String()
	}
	return
}

func SchoolTargetGroupToSchoolFilter(schoolTargetGroup entities.InfoNotificationTarget_SchoolFilter) (schoolIDs []string, schoolSelectType string) {
	switch schoolTargetGroup.Type {
	case consts.TargetGroupSelectTypeNone.String():
		return nil, consts.TargetGroupSelectTypeNone.String()
	case consts.TargetGroupSelectTypeAll.String():
		return nil, consts.TargetGroupSelectTypeAll.String()
	case consts.TargetGroupSelectTypeList.String():
		return schoolTargetGroup.SchoolIDs, consts.TargetGroupSelectTypeList.String()
	}
	return
}

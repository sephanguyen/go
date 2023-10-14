package aggregate

import (
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
)

type DomainStudent struct {
	// aggregate root
	entity.DomainStudent
	IndexAttr int // additional info

	LegacyUserGroups          entity.LegacyUserGroups
	UserGroupMembers          entity.DomainUserGroupMembers
	UserAccessPaths           entity.DomainUserAccessPaths
	Locations                 entity.DomainLocations
	UserPhoneNumbers          entity.DomainUserPhoneNumbers
	TaggedUsers               entity.DomainTaggedUsers
	Tags                      entity.DomainTags  // this will be removed soon
	Grade                     entity.DomainGrade // this will be removed soon
	SchoolHistories           entity.DomainSchoolHistories
	SchoolInfos               entity.DomainSchools
	SchoolCourses             entity.DomainSchoolCourses
	UserAddress               entity.DomainUserAddress
	Prefecture                entity.DomainPrefecture
	EnrollmentStatusHistories entity.DomainEnrollmentStatusHistories
	Courses                   entity.DomainStudentCourses
}

func (student DomainStudent) Index() int {
	return student.IndexAttr
}

type DomainStudentWithAssignedParent struct {
	DomainStudent
	Parents DomainParents
}

type NullStudent struct {
	entity.NullDomainStudent
	LegacyUserGroups entity.LegacyUserGroups
	UserGroupMembers entity.DomainUserGroupMembers
	UserAddresses    entity.DomainUserAddresses
	UserPhoneNumbers entity.DomainUserPhoneNumbers
	SchoolHistories  entity.DomainSchoolHistories
	Locations        entity.DomainLocations
	Grade            entity.Grade
}

func ValidStudent(student DomainStudent, isEnableUsername bool) error {
	if err := entity.ValidUser(isEnableUsername, student); err != nil {
		return err
	}
	if err := entity.ValidStudent(student); err != nil {
		return err
	}
	return nil
}

type DomainStudents []DomainStudent

func (students DomainStudents) Users() entity.Users {
	users := make(entity.Users, len(students))
	for i := range students {
		users[i] = students[i]
	}
	return users
}

func (students DomainStudents) StudentIDs() []string {
	studentIDs := make([]string, 0, len(students))
	for _, student := range students {
		studentIDs = append(studentIDs, student.UserID().String())
	}
	return studentIDs
}

func (students DomainStudents) EnrollmentStatusHistories() entity.DomainEnrollmentStatusHistories {
	enrollmentStatusHistories := make(entity.DomainEnrollmentStatusHistories, 0)
	for _, student := range students {
		enrollmentStatusHistories = append(enrollmentStatusHistories, student.EnrollmentStatusHistories...)
	}
	return enrollmentStatusHistories
}
func (students DomainStudents) TagPartnerIDs() []string {
	mapPartnerIDs := map[string]struct{}{}
	for _, student := range students {
		for _, tag := range student.Tags {
			if !tag.PartnerInternalID().IsEmpty() {
				mapPartnerIDs[tag.PartnerInternalID().String()] = struct{}{}
			}
		}
	}
	tagPartnerIDs := []string{}
	for partnerID := range mapPartnerIDs {
		tagPartnerIDs = append(tagPartnerIDs, partnerID)
	}

	return tagPartnerIDs
}

func (students DomainStudents) SchoolPartnerIDs() []string {
	mapPartnerIDs := map[string]struct{}{}
	for _, student := range students {
		for _, school := range student.SchoolInfos {
			if !school.PartnerInternalID().IsEmpty() {
				mapPartnerIDs[school.PartnerInternalID().String()] = struct{}{}
			}
		}
	}
	schoolPartnerIDs := []string{}
	for partnerID := range mapPartnerIDs {
		schoolPartnerIDs = append(schoolPartnerIDs, partnerID)
	}

	return schoolPartnerIDs
}

func (students DomainStudents) SchoolCoursePartnerIDs() []string {
	mapPartnerIDs := map[string]struct{}{}
	for _, student := range students {
		for _, schoolCourse := range student.SchoolCourses {
			if !schoolCourse.PartnerInternalID().IsEmpty() {
				mapPartnerIDs[schoolCourse.PartnerInternalID().String()] = struct{}{}
			}
		}
	}
	schoolCoursePartnerIDs := []string{}
	for partnerID := range mapPartnerIDs {
		schoolCoursePartnerIDs = append(schoolCoursePartnerIDs, partnerID)
	}

	return schoolCoursePartnerIDs
}

func (students DomainStudents) PrefectureCodes() []string {
	mapPrefectureCodes := map[string]struct{}{}
	for _, student := range students {
		if student.Prefecture != nil {
			if !student.Prefecture.PrefectureCode().IsEmpty() {
				mapPrefectureCodes[student.Prefecture.PrefectureCode().String()] = struct{}{}
			}
		}
	}
	prefectureCodes := []string{}
	for code := range mapPrefectureCodes {
		prefectureCodes = append(prefectureCodes, code)
	}

	return prefectureCodes
}

type DomainStudentWithAssignedParents []DomainStudentWithAssignedParent

func (students DomainStudentWithAssignedParents) Students() DomainStudents {
	listStudents := make([]DomainStudent, 0, len(students))
	for _, student := range students {
		listStudents = append(listStudents, student.DomainStudent)
	}

	return listStudents
}

type ExistingEntitiesRelatedStudent struct {
	MapPartnerIDAndGrade              map[string]entity.DomainGrade
	MapPartnerIDAndTag                map[string]entity.DomainTag
	MapPartnerIDAndSchool             map[string]entity.DomainSchool
	MapPartnerIDAndSchoolCourse       map[string]entity.DomainSchoolCourse
	MapEmailAndUser                   map[string]entity.User
	MapUserIDAndUser                  map[string]entity.User
	MapExternalUserIDAndUser          map[string]entity.User
	MapExternalUserIDAndStudentUser   map[string]entity.User
	MapPartnerIDAndLowestLocation     map[string]entity.DomainLocation
	ExistingEnrollmentStatusHistories entity.DomainEnrollmentStatusHistories
	MapPrefectureCodeAndPrefecture    map[string]entity.DomainPrefecture
	MapUserNameAndUser                map[string]entity.User
	DomainUserGroup                   entity.DomainUserGroup
}

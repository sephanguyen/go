package grpc

import (
	"strings"

	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/aggregate"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/utils"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func ToDomainStudents(students []*pb.StudentProfileV2, isEnableUsername bool) []aggregate.DomainStudent {
	domainStudents := make([]aggregate.DomainStudent, 0, len(students))
	for _, student := range students {
		contactPreference := *student.StudentPhoneNumbers.ContactPreference.Enum()
		fullNamePhonetic := field.NewString(utils.CombineFirstNamePhoneticAndLastNamePhoneticToFullName(student.FirstNamePhonetic, student.LastNamePhonetic))
		if fullNamePhonetic.TrimSpace().IsEmpty() {
			fullNamePhonetic = field.NewNullString()
		}
		firstNamePhonetic := field.NewString(student.FirstNamePhonetic)
		if firstNamePhonetic.TrimSpace().IsEmpty() {
			firstNamePhonetic = field.NewNullString()
		}
		lastNamePhonetic := field.NewString(student.LastNamePhonetic)
		if lastNamePhonetic.TrimSpace().IsEmpty() {
			lastNamePhonetic = field.NewNullString()
		}
		externalUserID := field.NewString(strings.TrimSpace(student.ExternalUserId))
		if externalUserID.IsEmpty() {
			externalUserID = field.NewNullString()
		}
		loginEmail := student.Email
		profile := &DomainStudentImpl{
			UserIDAttr:            field.NewString(student.Id),
			ExternalUserIDAttr:    externalUserID,
			UserNameAttr:          toUsername(student, isEnableUsername),
			GradeIDAttr:           field.NewString(student.GradeId),
			FirstNameAttr:         field.NewString(student.FirstName),
			LastNameAttr:          field.NewString(student.LastName),
			FullNameAttr:          field.NewString(utils.CombineFirstNameAndLastNameToFullName(student.FirstName, student.LastName)),
			FirstNamePhoneticAttr: firstNamePhonetic,
			LastNamePhoneticAttr:  lastNamePhonetic,
			FullNamePhoneticAttr:  fullNamePhonetic,
			EmailAttr:             field.NewString(student.Email),
			NoteAttr:              field.NewString(student.StudentNote),
			PasswordAttr:          field.NewString(student.Password),
			ContactPreferenceAttr: field.NewInt32(int32(contactPreference)),
			GenderAttr:            field.NewInt32(int32(student.Gender)),
			BirthdayAttr:          field.NewDate(student.Birthday.AsTime()),
			EnrollmentStatusAttr:  toDomainEnrollmentStatus(student),
			ExternalStudentIDAttr: field.NewString(student.StudentExternalId),
			LoginEmailAttr:        field.NewString(loginEmail),
		}

		domainStudentAgg := aggregate.DomainStudent{
			DomainStudent:             profile,
			UserPhoneNumbers:          toUserPhoneNumbers(student),
			UserAccessPaths:           toUserAccessPaths(student.Id, student.LocationIds),
			TaggedUsers:               toTaggedUsers(student),
			SchoolHistories:           toSchoolHistoriesAgg(student.SchoolHistories),
			EnrollmentStatusHistories: toDomainEnrollmentStatusHistories(student.EnrollmentStatusHistories),
		}

		if len(student.GetUserAddresses()) > 0 {
			address := toUserAddressAgg(student.GetUserAddresses()[0])
			domainStudentAgg.UserAddress = address
		}

		// Generally, We will use Locations and EnrollmentStatus from proto
		// when EnrollmentStatusHistories was passed, we will use location & EnrollmentStatus from EnrollmentStatusHistories instead
		if len(domainStudentAgg.EnrollmentStatusHistories) > 0 {
			profile.EnrollmentStatusAttr = domainStudentAgg.EnrollmentStatusHistories[0].EnrollmentStatus()

			locationIDs := make([]string, 0, len(domainStudentAgg.EnrollmentStatusHistories))
			for _, enrollmentStatusHistory := range domainStudentAgg.EnrollmentStatusHistories {
				locationIDs = append(locationIDs, enrollmentStatusHistory.LocationID().String())
			}
			domainStudentAgg.UserAccessPaths = toUserAccessPaths(student.Id, locationIDs)
		}

		domainStudents = append(domainStudents, domainStudentAgg)
	}
	return domainStudents
}

func toUsername(student *pb.StudentProfileV2, isUserNameStudentParentEnabled bool) field.String {
	var userName field.String

	if isUserNameStudentParentEnabled {
		userName = field.NewString(student.Username)
	} else {
		userName = field.NewString(student.Email)
	}

	return userName
}

func UpsertStudentProfiles(students []aggregate.DomainStudent) []*pb.StudentProfileV2 {
	studentProfiles := make([]*pb.StudentProfileV2, 0, len(students))

	for _, student := range students {
		schoolHistories := make([]*pb.SchoolHistory, 0, len(student.SchoolHistories))
		for _, schoolHistory := range student.SchoolHistories {
			schoolHistories = append(schoolHistories, &pb.SchoolHistory{
				SchoolId:       schoolHistory.SchoolID().String(),
				SchoolCourseId: schoolHistory.SchoolCourseID().String(),
				StartDate:      timestamppb.New(schoolHistory.StartDate().Time()),
				EndDate:        timestamppb.New(schoolHistory.EndDate().Time()),
			})
		}

		updateStudentPhoneNumber := new(pb.StudentPhoneNumbers)
		updateStudentPhoneNumber.ContactPreference = pb.StudentContactPreference(pb.StudentContactPreference_value[student.ContactPreference().String()])
		for _, userPhoneNumber := range student.UserPhoneNumbers {
			updateStudentPhoneNumber.StudentPhoneNumberWithIds = append(updateStudentPhoneNumber.StudentPhoneNumberWithIds, &pb.StudentPhoneNumberWithID{
				StudentPhoneNumberId: userPhoneNumber.UserPhoneNumberID().String(),
				PhoneNumberType:      pb.StudentPhoneNumberType(pb.StudentPhoneNumberType_value[userPhoneNumber.Type().String()]),
				PhoneNumber:          userPhoneNumber.PhoneNumber().String(),
			})
		}

		enrollmentStatusHistories := make([]*pb.EnrollmentStatusHistory, 0, len(student.EnrollmentStatusHistories))
		for _, enrollmentStatusHistory := range student.EnrollmentStatusHistories {
			enrollmentStatusHistories = append(enrollmentStatusHistories,
				&pb.EnrollmentStatusHistory{
					StudentId:        student.UserID().String(),
					LocationId:       enrollmentStatusHistory.LocationID().String(),
					EnrollmentStatus: pb.StudentEnrollmentStatus(pb.StudentEnrollmentStatus_value[enrollmentStatusHistory.EnrollmentStatus().String()]),
					StartDate:        timestamppb.New(enrollmentStatusHistory.StartDate().Time()),
					EndDate:          timestamppb.New(enrollmentStatusHistory.EndDate().Time()),
				})
		}

		var userAddress []*pb.UserAddress
		if student.UserAddress != nil {
			userAddress = []*pb.UserAddress{
				{
					AddressId:    student.UserAddress.UserAddressID().String(),
					AddressType:  pb.AddressType(pb.AddressType_value[student.UserAddress.AddressType().String()]),
					PostalCode:   student.UserAddress.PostalCode().String(),
					Prefecture:   student.UserAddress.PrefectureID().String(),
					City:         student.UserAddress.City().String(),
					FirstStreet:  student.UserAddress.FirstStreet().String(),
					SecondStreet: student.UserAddress.SecondStreet().String(),
				},
			}
		}

		studentProfiles = append(studentProfiles, &pb.StudentProfileV2{
			Id:                        student.UserID().String(),
			StudentExternalId:         student.ExternalStudentID().String(),
			ExternalUserId:            student.ExternalUserID().String(),
			Username:                  student.UserName().String(),
			FirstName:                 student.FirstName().String(),
			LastName:                  student.LastName().String(),
			FirstNamePhonetic:         student.FirstNamePhonetic().String(),
			LastNamePhonetic:          student.LastNamePhonetic().String(),
			Email:                     student.Email().String(),
			Password:                  student.Password().String(),
			GradeId:                   student.GradeID().String(),
			StudentNote:               student.StudentNote().String(),
			Birthday:                  timestamppb.New(student.Birthday().Date()),
			Gender:                    pb.Gender(pb.Gender_value[student.Gender().String()]),
			StudentPhoneNumbers:       updateStudentPhoneNumber,
			TagIds:                    student.Tags.TagIDs(),
			LocationIds:               student.Locations.LocationIDs(),
			SchoolHistories:           schoolHistories,
			EnrollmentStatusHistories: enrollmentStatusHistories,
			UserAddresses:             userAddress,
		})
	}

	return studentProfiles
}

func toUserPhoneNumbers(student *pb.StudentProfileV2) entity.DomainUserPhoneNumbers {
	userPhoneNumbers := new(DomainPhoneNumberImpl)
	domainUserPhoneNumbers := make(entity.DomainUserPhoneNumbers, 0, len(student.StudentPhoneNumbers.GetStudentPhoneNumberWithIds()))
	if student.StudentPhoneNumbers != nil {
		userPhoneNumbers.ContactPreference = field.NewInt32(int32(*student.StudentPhoneNumbers.ContactPreference.Enum()))
		for _, phoneNumber := range student.StudentPhoneNumbers.GetStudentPhoneNumberWithIds() {
			domainUserPhoneNumbers = append(domainUserPhoneNumbers, &PhoneNumberAttr{
				PhoneIDAttr:     field.NewString(phoneNumber.GetStudentPhoneNumberId()),
				PhoneTypeAttr:   field.NewString(MapStudentPhoneNumberType[phoneNumber.PhoneNumberType.String()]),
				PhoneNumberAttr: field.NewString(phoneNumber.GetPhoneNumber()),
			})
		}
	}
	return domainUserPhoneNumbers
}

func toTaggedUsers(student *pb.StudentProfileV2) entity.DomainTaggedUsers {
	taggedUsers := make(entity.DomainTaggedUsers, 0, len(student.TagIds))
	for _, tagID := range student.TagIds {
		taggedUsers = append(taggedUsers, &DomainTaggedUserImpl{
			UserIDAttr: field.NewString(student.Id),
			TagIDAttr:  field.NewString(tagID),
		})
	}
	return taggedUsers
}

func toSchoolHistoriesAgg(schoolHistories []*pb.SchoolHistory) entity.DomainSchoolHistories {
	domainSchoolHistories := make(entity.DomainSchoolHistories, 0, len(schoolHistories))
	for _, schoolHistory := range schoolHistories {
		domainSchoolHistory := &DomainSchoolHistoryImpl{
			SchoolIDAttr:       field.NewString(schoolHistory.SchoolId),
			SchoolCourseIDAttr: field.NewString(schoolHistory.SchoolCourseId),
			StartDateAttr:      field.NewTime(schoolHistory.StartDate.AsTime()),
			EndDateAttr:        field.NewTime(schoolHistory.EndDate.AsTime()),
		}
		domainSchoolHistories = append(domainSchoolHistories, domainSchoolHistory)
	}

	return domainSchoolHistories
}

func toDomainEnrollmentStatus(student *pb.StudentProfileV2) field.String {
	// if EnrollmentStatus was not passed, we will use EnrollmentStatusStr instead
	enrollmentStatus := student.EnrollmentStatus
	if enrollmentStatus == pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_NONE {
		enrollmentStatus = pb.StudentEnrollmentStatus(pb.StudentEnrollmentStatus_value[student.EnrollmentStatusStr])
		if enrollmentStatus == pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_NONE {
			return field.NewNullString()
		}
	}
	return field.NewString(enrollmentStatus.String())
}

func toDomainEnrollmentStatusHistories(enrollmentStatusHistories []*pb.EnrollmentStatusHistory) entity.DomainEnrollmentStatusHistories {
	domainEnrollmentStatusHistories := make(entity.DomainEnrollmentStatusHistories, 0, len(enrollmentStatusHistories))
	for _, enrollmentStatusHistory := range enrollmentStatusHistories {
		domainEnrollmentStatusHistories = append(domainEnrollmentStatusHistories, &DomainEnrollmentStatusHistoryImpl{
			EnrollmentStatusAttr: field.NewInt32(int32(*enrollmentStatusHistory.GetEnrollmentStatus().Enum())),
			LocationAttr:         field.NewString(enrollmentStatusHistory.GetLocationId()),
			StartDateAttr:        field.NewTime(enrollmentStatusHistory.GetStartDate().AsTime()),
			EndDateAttr:          field.NewTime(enrollmentStatusHistory.GetEndDate().AsTime()),
			StudentIDAttr:        field.NewString(enrollmentStatusHistory.GetStudentId()),
		})
	}
	return domainEnrollmentStatusHistories
}

func toUserAddressAgg(address *pb.UserAddress) *DomainUserAddressImpl {
	domainAddress := &DomainUserAddressImpl{
		AddressIDAttr:    field.NewString(address.AddressId),
		AddressTypeAttr:  field.NewString(address.AddressType.String()),
		PostalCodeAttr:   field.NewString(address.PostalCode),
		PrefectureAttr:   field.NewString(address.Prefecture),
		CityAttr:         field.NewString(address.City),
		FirstStreetAttr:  field.NewString(address.FirstStreet),
		SecondStreetAttr: field.NewString(address.SecondStreet),
	}
	return domainAddress
}

func toUserAccessPaths(studentID string, locationIDs []string) entity.DomainUserAccessPaths {
	userAccessPaths := make(entity.DomainUserAccessPaths, 0, len(locationIDs))
	for _, locationID := range locationIDs {
		userAccessPaths = append(userAccessPaths, &DomainUserAccessPathImpl{
			UserIDAttr:     field.NewString(studentID),
			LocationIDAttr: field.NewString(locationID),
		})
	}
	return userAccessPaths
}

package grpc

import (
	"strings"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/aggregate"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/utils"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestToDomainStudents(t *testing.T) {
	type args struct {
		students         []*pb.StudentProfileV2
		isUsernameEnable bool
	}
	locationID := idutil.ULIDNow()
	tests := []struct {
		name string
		args func(t *testing.T) args

		want1 []aggregate.DomainStudent
	}{
		{
			name: "happy case",
			args: func(t *testing.T) args {
				return args{
					students: []*pb.StudentProfileV2{
						{
							Id:                  idutil.ULIDNow(),
							FirstName:           "FirstNameAttr",
							LastName:            "LastNameAttr",
							FirstNamePhonetic:   "FirstNamePhonetic",
							LastNamePhonetic:    "LastNamePhonetic",
							Username:            "Username",
							Email:               "Email",
							GradeId:             "GradeId",
							StudentNote:         "StudentNote",
							Birthday:            timestamppb.Now(),
							Gender:              pb.Gender_MALE,
							TagIds:              []string{idutil.ULIDNow()},
							LocationIds:         []string{locationID},
							EnrollmentStatus:    pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_GRADUATED,
							EnrollmentStatusStr: "EnrollmentStatusStr",
							Password:            "Password",
							ExternalUserId:      "ExternalUserId",
							StudentPhoneNumbers: &pb.StudentPhoneNumbers{
								StudentPhoneNumberWithIds: []*pb.StudentPhoneNumberWithID{
									{
										StudentPhoneNumberId: "",
										PhoneNumberType:      pb.StudentPhoneNumberType_HOME_PHONE_NUMBER,
										PhoneNumber:          "098888889",
									},
									{
										StudentPhoneNumberId: idutil.ULIDNow(),
										PhoneNumberType:      pb.StudentPhoneNumberType_HOME_PHONE_NUMBER,
										PhoneNumber:          "098888888",
									},
								},
								ContactPreference: pb.StudentContactPreference_PARENT_PRIMARY_PHONE_NUMBER,
							},
							SchoolHistories: []*pb.SchoolHistory{{
								SchoolId:       idutil.ULIDNow(),
								SchoolCourseId: idutil.ULIDNow(),
								StartDate:      timestamppb.Now(),
							}},
							EnrollmentStatusHistories: []*pb.EnrollmentStatusHistory{{
								LocationId:       locationID,
								EnrollmentStatus: pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
								StartDate:        timestamppb.Now(),
							}},
							UserAddresses: []*pb.UserAddress{{
								AddressId:    idutil.ULIDNow(),
								AddressType:  pb.AddressType_BILLING_ADDRESS,
								PostalCode:   "PostalCode",
								Prefecture:   "Prefecture",
								City:         "City",
								FirstStreet:  "FirstStreet",
								SecondStreet: "SecondStreet",
							}},
						},
					},
					isUsernameEnable: true,
				}
			},
		},
		{
			name: "happy case: with enrollment status & location and without enrollment status history",
			args: func(t *testing.T) args {
				return args{
					students: []*pb.StudentProfileV2{
						{
							Id:                  idutil.ULIDNow(),
							FirstName:           "FirstNameAttr",
							LastName:            "LastNameAttr",
							FirstNamePhonetic:   "FirstNamePhonetic",
							LastNamePhonetic:    "LastNamePhonetic",
							Username:            "Username",
							Email:               "Email",
							GradeId:             "GradeId",
							StudentNote:         "StudentNote",
							Birthday:            timestamppb.Now(),
							Gender:              pb.Gender_MALE,
							TagIds:              []string{idutil.ULIDNow()},
							LocationIds:         []string{locationID},
							EnrollmentStatus:    pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_GRADUATED,
							EnrollmentStatusStr: "EnrollmentStatusStr",
							Password:            "Password",
							ExternalUserId:      "ExternalUserId",
							StudentPhoneNumbers: &pb.StudentPhoneNumbers{
								ContactPreference: pb.StudentContactPreference_PARENT_PRIMARY_PHONE_NUMBER,
							},
						},
					},
					isUsernameEnable: true,
				}
			},
		},
		{
			name: "happy case: without enrollment status & location and with enrollment status history",
			args: func(t *testing.T) args {
				return args{
					students: []*pb.StudentProfileV2{
						{
							Id:                idutil.ULIDNow(),
							FirstName:         "FirstNameAttr",
							LastName:          "LastNameAttr",
							FirstNamePhonetic: "FirstNamePhonetic",
							LastNamePhonetic:  "LastNamePhonetic",
							Username:          "Username",
							Email:             "Email",
							GradeId:           "GradeId",
							StudentNote:       "StudentNote",
							Birthday:          timestamppb.Now(),
							Gender:            pb.Gender_MALE,
							TagIds:            []string{idutil.ULIDNow()},
							Password:          "Password",
							ExternalUserId:    "ExternalUserId",
							StudentPhoneNumbers: &pb.StudentPhoneNumbers{
								ContactPreference: pb.StudentContactPreference_PARENT_PRIMARY_PHONE_NUMBER,
							},
							EnrollmentStatusHistories: []*pb.EnrollmentStatusHistory{{
								LocationId:       locationID,
								EnrollmentStatus: pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
								StartDate:        timestamppb.Now(),
							}},
						},
					},
					isUsernameEnable: true,
				}
			},
		},
		{
			name: "happy case: without birthday",
			args: func(t *testing.T) args {
				return args{
					students: []*pb.StudentProfileV2{
						{
							Id:                  idutil.ULIDNow(),
							FirstName:           "FirstNameAttr",
							LastName:            "LastNameAttr",
							FirstNamePhonetic:   "FirstNamePhonetic",
							LastNamePhonetic:    "LastNamePhonetic",
							Username:            "Username",
							Email:               "Email",
							GradeId:             "GradeId",
							StudentNote:         "StudentNote",
							Gender:              pb.Gender_MALE,
							TagIds:              []string{idutil.ULIDNow()},
							LocationIds:         []string{locationID},
							EnrollmentStatus:    pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_GRADUATED,
							EnrollmentStatusStr: "EnrollmentStatusStr",
							Password:            "Password",
							ExternalUserId:      "ExternalUserId",
							StudentPhoneNumbers: &pb.StudentPhoneNumbers{
								ContactPreference: pb.StudentContactPreference_PARENT_PRIMARY_PHONE_NUMBER,
							},
						},
					},
					isUsernameEnable: true,
				}
			},
		},
		{
			name: "happy case: without start_day of school history",
			args: func(t *testing.T) args {
				return args{
					students: []*pb.StudentProfileV2{
						{
							Id:                  idutil.ULIDNow(),
							FirstName:           "FirstNameAttr",
							LastName:            "LastNameAttr",
							FirstNamePhonetic:   "FirstNamePhonetic",
							LastNamePhonetic:    "LastNamePhonetic",
							Username:            "Username",
							Email:               "Email",
							GradeId:             "GradeId",
							StudentNote:         "StudentNote",
							Gender:              pb.Gender_MALE,
							TagIds:              []string{idutil.ULIDNow()},
							LocationIds:         []string{locationID},
							EnrollmentStatus:    pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_GRADUATED,
							EnrollmentStatusStr: "EnrollmentStatusStr",
							Password:            "Password",
							ExternalUserId:      "ExternalUserId",
							SchoolHistories: []*pb.SchoolHistory{{
								SchoolId:       idutil.ULIDNow(),
								SchoolCourseId: idutil.ULIDNow(),
								EndDate:        timestamppb.Now(),
							}},
							StudentPhoneNumbers: &pb.StudentPhoneNumbers{
								ContactPreference: pb.StudentContactPreference_PARENT_PRIMARY_PHONE_NUMBER,
							},
						},
					},
					isUsernameEnable: true,
				}
			},
		},
		{
			name: "happy case: without end_day of school history",
			args: func(t *testing.T) args {
				return args{
					students: []*pb.StudentProfileV2{
						{
							Id:                  idutil.ULIDNow(),
							FirstName:           "FirstNameAttr",
							LastName:            "LastNameAttr",
							FirstNamePhonetic:   "FirstNamePhonetic",
							LastNamePhonetic:    "LastNamePhonetic",
							Username:            "Username",
							Email:               "Email",
							GradeId:             "GradeId",
							StudentNote:         "StudentNote",
							Gender:              pb.Gender_MALE,
							TagIds:              []string{idutil.ULIDNow()},
							LocationIds:         []string{locationID},
							EnrollmentStatus:    pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_GRADUATED,
							EnrollmentStatusStr: "EnrollmentStatusStr",
							Password:            "Password",
							ExternalUserId:      "ExternalUserId",
							SchoolHistories: []*pb.SchoolHistory{{
								SchoolId:       idutil.ULIDNow(),
								SchoolCourseId: idutil.ULIDNow(),
								StartDate:      timestamppb.Now(),
							}},
							StudentPhoneNumbers: &pb.StudentPhoneNumbers{
								ContactPreference: pb.StudentContactPreference_PARENT_PRIMARY_PHONE_NUMBER,
							},
						},
					},
					isUsernameEnable: true,
				}
			},
		},
		{
			name: "should return profile correctly when username feature is unable",
			args: func(t *testing.T) args {
				return args{
					students: []*pb.StudentProfileV2{
						{
							Id:                  idutil.ULIDNow(),
							FirstName:           "FirstNameAttr",
							LastName:            "LastNameAttr",
							FirstNamePhonetic:   "FirstNamePhonetic",
							LastNamePhonetic:    "LastNamePhonetic",
							Username:            "Username",
							Email:               "Email",
							GradeId:             "GradeId",
							StudentNote:         "StudentNote",
							Gender:              pb.Gender_MALE,
							TagIds:              []string{idutil.ULIDNow()},
							LocationIds:         []string{locationID},
							EnrollmentStatus:    pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_GRADUATED,
							EnrollmentStatusStr: "EnrollmentStatusStr",
							Password:            "Password",
							ExternalUserId:      "ExternalUserId",
							SchoolHistories: []*pb.SchoolHistory{{
								SchoolId:       idutil.ULIDNow(),
								SchoolCourseId: idutil.ULIDNow(),
								StartDate:      timestamppb.Now(),
							}},
							StudentPhoneNumbers: &pb.StudentPhoneNumbers{
								ContactPreference: pb.StudentContactPreference_PARENT_PRIMARY_PHONE_NUMBER,
							},
						},
					},
					isUsernameEnable: false,
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tArgs := tt.args(t)

			profiles := ToDomainStudents(tArgs.students, tArgs.isUsernameEnable)
			for i, profile := range profiles {
				student := tArgs.students[i]
				assert.Equal(t, profile.UserID().String(), student.Id)
				assert.Equal(t, profile.FirstName().String(), student.FirstName)
				assert.Equal(t, profile.LastName().String(), student.LastName)
				assert.Equal(t, profile.FirstNamePhonetic().String(), student.FirstNamePhonetic)
				assert.Equal(t, profile.LastNamePhonetic().String(), student.LastNamePhonetic)
				assert.Equal(t, profile.Email().String(), student.Email)
				assert.Equal(t, profile.GradeID().String(), student.GradeId)
				assert.Equal(t, profile.StudentNote().String(), student.StudentNote)
				assert.Equal(t, profile.Gender().String(), student.Gender.String())
				assert.Equal(t, profile.TaggedUsers.TagIDs(), student.TagIds)
				assert.Equal(t, profile.Password().String(), student.Password)
				assert.Equal(t, profile.ExternalUserID().String(), student.ExternalUserId)
				assert.Equal(t, profile.ContactPreference().String(), student.StudentPhoneNumbers.ContactPreference.String())
				assert.Equal(t, profile.LoginEmail().String(), student.Email)
				if tArgs.isUsernameEnable {
					assert.Equal(t, profile.UserName().String(), strings.ToLower(student.Username))
				} else {
					assert.Equal(t, profile.UserName().String(), strings.ToLower(student.Email))
				}
				if student.Birthday == nil {
					assert.True(t, field.IsNull(profile.Birthday()))
				} else {
					assert.True(t, profile.Birthday().Date().Equal(utils.TruncateToDay(student.Birthday.AsTime())))
				}

				if len(profile.EnrollmentStatusHistories) > 0 {
					locationIDs := make([]string, 0, len(profile.EnrollmentStatusHistories))
					for _, history := range profile.EnrollmentStatusHistories {
						locationIDs = append(locationIDs, history.LocationID().String())
					}
					assert.Equal(t, profile.UserAccessPaths.LocationIDs(), locationIDs)
					assert.Equal(t, profile.EnrollmentStatus().String(), profile.EnrollmentStatusHistories[0].EnrollmentStatus().String())
				} else {
					assert.Equal(t, profile.EnrollmentStatus().String(), student.EnrollmentStatus.String())
					assert.Equal(t, profile.UserAccessPaths.LocationIDs(), student.LocationIds)
				}

				for j, userPhoneNumber := range profile.UserPhoneNumbers {
					studentPhone := student.StudentPhoneNumbers.GetStudentPhoneNumberWithIds()[j]
					if studentPhone.GetStudentPhoneNumberId() != "" {
						assert.Equal(t, userPhoneNumber.UserPhoneNumberID().String(), studentPhone.GetStudentPhoneNumberId())
					} else {
						assert.NotEqual(t, userPhoneNumber.UserPhoneNumberID().String(), "")
					}
					assert.Equal(t, userPhoneNumber.PhoneNumber().String(), studentPhone.PhoneNumber)
					assert.Equal(t, userPhoneNumber.Type().String(), MapStudentPhoneNumberType[studentPhone.PhoneNumberType.String()])
				}

				for j, schoolHistory := range profile.SchoolHistories {
					sh := student.SchoolHistories[j]
					assert.Equal(t, schoolHistory.SchoolCourseID().String(), sh.GetSchoolCourseId())
					assert.Equal(t, schoolHistory.SchoolID().String(), sh.GetSchoolId())
					if sh.GetStartDate() != nil {
						assert.Equal(t, utils.TruncateToDay(schoolHistory.StartDate().Time()), utils.TruncateToDay(sh.GetStartDate().AsTime()))
					} else {
						assert.True(t, field.IsNull(schoolHistory.StartDate()))
					}

					if sh.GetEndDate() != nil {
						assert.Equal(t, utils.TruncateToDay(schoolHistory.EndDate().Time()), utils.TruncateToDay(sh.GetEndDate().AsTime()))
					} else {
						assert.True(t, field.IsNull(schoolHistory.EndDate()))
					}
				}

				for j, enrollmentStatusHistory := range profile.EnrollmentStatusHistories {
					esh := student.EnrollmentStatusHistories[j]
					assert.Equal(t, enrollmentStatusHistory.EnrollmentStatus().String(), esh.EnrollmentStatus.String())
					assert.Equal(t, enrollmentStatusHistory.LocationID().String(), esh.GetLocationId())
					assert.True(t, enrollmentStatusHistory.StartDate().Time().Equal(esh.GetStartDate().AsTime()))
				}

				if student.UserAddresses != nil {
					studentAddress := student.UserAddresses[0]
					assert.Equal(t, profile.UserAddress.AddressType().String(), studentAddress.AddressType.String())
					assert.Equal(t, profile.UserAddress.UserAddressID().String(), studentAddress.AddressId)
					assert.Equal(t, profile.UserAddress.PostalCode().String(), studentAddress.PostalCode)
					assert.Equal(t, profile.UserAddress.PrefectureID().String(), studentAddress.Prefecture)
					assert.Equal(t, profile.UserAddress.City().String(), studentAddress.City)
					assert.Equal(t, profile.UserAddress.FirstStreet().String(), studentAddress.FirstStreet)
					assert.Equal(t, profile.UserAddress.SecondStreet().String(), studentAddress.SecondStreet)
				}
			}
		})
	}
}

func Test_toUsername(t *testing.T) {
	type args struct {
		student                        *pb.StudentProfileV2
		isUserNameStudentParentEnabled bool
	}
	tests := []struct {
		name string
		args args
		want field.String
	}{
		{
			name: "if username feature was disabled, fill username by email",
			args: args{
				student: &pb.StudentProfileV2{
					Email:    "email",
					Username: "username",
				},
				isUserNameStudentParentEnabled: false,
			},
			want: field.NewString("email"),
		},
		{
			name: "if username feature was able, fill trimmed space username",
			args: args{
				student: &pb.StudentProfileV2{
					Email:    "email",
					Username: "",
				},
				isUserNameStudentParentEnabled: true,
			},
			want: field.NewString(""),
		},
		{
			name: "if username feature was able, fill trimmed space username",
			args: args{
				student: &pb.StudentProfileV2{
					Email:    "email",
					Username: "username",
				},
				isUserNameStudentParentEnabled: true,
			},
			want: field.NewString("username"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, toUsername(tt.args.student, tt.args.isUserNameStudentParentEnabled), "toUsername(%v, %v)", tt.args.student, tt.args.isUserNameStudentParentEnabled)
		})
	}
}

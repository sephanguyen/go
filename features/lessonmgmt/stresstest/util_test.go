package stresstest

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/manabie-com/backend/features/common"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateAcc(t *testing.T) {
	// some config to create student acc
	cfg := &common.Config{
		FirebaseAPIKey:     "",
		BobHasuraAdminURL:  "https://admin.[env].manabie.io:[port]",
		IdentityToolkitAPI: "https://identitytoolkit.googleapis.com/v1",
	}
	const courseID = ""
	const schoolID = 0
	adminAcc := &AccountInfo{
		Email:    "",
		Password: "",
	}
	const yourCompanyEmail = "firstname.lastname@manabie.com"
	const studentNameSuffix = "stress-test-04-28"
	const numStudent = 10
	const numTeacher = 2

	// gen email address template
	components := strings.Split(yourCompanyEmail, "@")
	fmt.Println("email.Name", components[0])
	emailTemple := fmt.Sprintf(`%s+:T.%s:N@manabie.com`, components[0], studentNameSuffix)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	st, err := NewStressTest(
		cfg,
		schoolID,
		"manabie-p7muf",
		"./admin-accounts.json",
	)

	require.NoError(t, err)
	res, err := st.UserSignInWithPassword(ctx, adminAcc, true)
	require.NoError(t, err)

	jwt, err := st.ExchangeUserToken(ctx, res.IdToken)
	require.NoError(t, err)

	suite := st.NewSuite()
	suite.lessonSuite.CommonSuite.UserMgmtConn = suite.lessonSuite.CommonSuite.BobConn
	suite.lessonSuite.CommonSuite.StepState.AuthToken = jwt
	ctx = common.StepStateToContext(ctx, suite.lessonSuite.CommonSuite.StepState)

	err = suite.lessonSuite.RetrieveLowestLevelLocations(ctx)
	require.NoError(t, err)
	suite.lessonSuite.CommonSuite.CenterIDs = suite.lessonSuite.CommonSuite.LowestLevelLocationIDs
	fmt.Printf("Location ID %s which assign for StudentCoursePackage: \n", suite.lessonSuite.CommonSuite.CenterIDs[0])

	// create students
	students := make([]*AccountInfo, 0, numStudent)
	for i := 0; i < numStudent; i++ {
		email := strings.ReplaceAll(emailTemple, ":T", "student")
		email = strings.ReplaceAll(email, ":N", strconv.Itoa(i))
		studentInfo, err := suite.lessonSuite.CommonSuite.CreateStudentByStudentInfo(ctx, &upb.CreateStudentRequest{
			SchoolId: -2147483648,
			StudentProfile: &upb.CreateStudentRequest_StudentProfile{
				LocationIds:      suite.lessonSuite.CommonSuite.LowestLevelLocationIDs,
				Email:            email,
				Password:         "123456",
				Name:             "Test_Stress_Test_Student_" + strconv.Itoa(i),
				CountryCode:      cpb.Country_COUNTRY_VN,
				Grade:            5,
				EnrollmentStatus: upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
			},
		})
		if err != nil {
			fmt.Printf("CreateStudentByStudentInfo: %v \n", err)
			break
		}
		assert.Equal(t, "123456", studentInfo.StudentPassword)

		// add course to student
		_, err = suite.lessonSuite.CommonSuite.
			UpsertStudentCoursePackages(ctx, courseID, suite.lessonSuite.CommonSuite.CenterIDs[0], studentInfo.Student.UserProfile.UserId)
		if suite.lessonSuite.CommonSuite.StepState.ResponseErr != nil {
			err = suite.lessonSuite.CommonSuite.StepState.ResponseErr
		}
		if err != nil {
			fmt.Printf("UpsertStudentCoursePackages: %v \n", err)

			break
		}

		students = append(students, &AccountInfo{
			Email:    email,
			Password: "123456",
		})

		// delay avoiding to time out
		time.Sleep(2 * time.Second)
	}
	time.Sleep(2 * time.Second)

	teachers := make([]*AccountInfo, 0, numTeacher)
	for i := 0; i < numTeacher; i++ {
		email := strings.ReplaceAll(emailTemple, ":T", "teacher")
		email = strings.ReplaceAll(email, ":N", strconv.Itoa(i))
		req := &upb.CreateStaffRequest{
			Staff: &upb.CreateStaffRequest_StaffProfile{
				Name:           fmt.Sprintf("Test_Stress_Test_Teacher_%d", i),
				OrganizationId: strconv.Itoa(schoolID),
				UserGroup:      upb.UserGroup_USER_GROUP_TEACHER,
				Country:        cpb.Country_COUNTRY_VN,
				Email:          email,
				LocationIds:    suite.lessonSuite.CommonSuite.LowestLevelLocationIDs,
			},
		}

		_, err = suite.lessonSuite.CommonSuite.CreateTeacherByTeacherInfo(ctx, req)
		if suite.lessonSuite.CommonSuite.StepState.ResponseErr != nil {
			err = suite.lessonSuite.CommonSuite.StepState.ResponseErr
		}
		if err != nil {
			fmt.Printf("CreateTeacherByTeacherInfo: %v \n", err)
			break
		}

		teachers = append(teachers, &AccountInfo{
			Email: email,
		})
		// delay avoiding to time out
		time.Sleep(2 * time.Second)
	}

	err = WriteAccounts("./created-accounts.json", nil, teachers, students)
	require.NoError(t, err)
}

func TestCreateCourse(t *testing.T) {
	const locationID = ""
	st, err := NewStressTest(
		&common.Config{
			FirebaseAPIKey:     "",
			BobHasuraAdminURL:  "https://admin.[env].manabie.io:[port]",
			IdentityToolkitAPI: "https://identitytoolkit.googleapis.com/v1",
		},
		0,
		"manabie-p7muf",
		"./admin-accounts.json",
	)

	suite := st.NewSuite()
	suite.lessonSuite.CommonSuite.StepState.AuthToken = st.adminAccounts[0].Token
	suite.lessonSuite.CommonSuite.MasterMgmtConn = suite.lessonSuite.CommonSuite.BobConn
	ctx := common.StepStateToContext(context.Background(), suite.lessonSuite.CommonSuite.StepState)
	_, err = suite.lessonSuite.CommonSuite.UserUpsertCourses(ctx, "Test_Stress_Test_28/4", []string{locationID})
	require.NoError(t, err)
	require.NoError(t, suite.lessonSuite.CommonSuite.StepState.ResponseErr)
	fmt.Println(suite.lessonSuite.CommonSuite.StepState.Response)
}

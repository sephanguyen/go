package helpers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/conversationmgmt/common/entities"
	"github.com/manabie-com/backend/features/eibanam/communication/entity"
	"github.com/manabie-com/backend/features/eibanam/communication/util"
	featuresHelper "github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type CreateStudentsWithSameParentOpt struct {
	StudentsName    []string
	NumberOfStudent int
	ParentName      string
}

type CreateStudentWithParentOpt struct {
	StudentName string
	ParentName  string
}

type CreateStudentOpt struct {
	Name string
}

func (helper *ConversationMgmtHelper) CreateStudent(schoolAdmin *entities.Staff, grade *entities.GradeMaster, locIDs []string, hasParent bool, numOfParent int, schoolID string) (*entities.Student, error) {
	organizationID := schoolAdmin.OrganizationIDs[0]
	ctxWithToken, cancel := contextWithTokenAndTimeOut(context.Background(), schoolAdmin.Token)
	defer cancel()

	// create student
	student, createStudentRequest := helper.newstudentInfo(organizationID, grade, locIDs, nil, schoolID)
	// nolint
	createStudentResponse, err := upb.NewUserModifierServiceClient(helper.UserMgmtGRPCConn).CreateStudent(ctxWithToken, createStudentRequest)
	if err != nil {
		return nil, err
	}
	student.ID = createStudentResponse.StudentProfile.Student.UserProfile.UserId

	err = helper.createStudentEnrollmentStatus(ctxWithToken, fmt.Sprint(organizationID), student.ID, locIDs)
	if err != nil {
		return nil, fmt.Errorf("err create student enrollment status: %v", err)
	}

	// if request has parents, create parent info
	if hasParent {
		parentProfiles := make([]*upb.CreateParentsAndAssignToStudentRequest_ParentProfile, numOfParent)
		for index := range parentProfiles {
			parent := entities.User{
				Name:     helper.PickName(),
				Email:    fmt.Sprintf("%v-parent-%v@example.com", student.ID, index),
				Phone:    util.RandPhoneNumber(),
				Password: idutil.ULIDNow(),
				Group:    cpb.UserGroup_USER_GROUP_PARENT.String(),
			}
			student.Parents = append(student.Parents, &parent)

			// reassign entity eibanam to parent
			parentProfiles[index] = &upb.CreateParentsAndAssignToStudentRequest_ParentProfile{
				Name:        parent.Name,
				CountryCode: cpb.Country_COUNTRY_VN,
				PhoneNumber: parent.Phone,
				Email:       parent.Email,
				// username can accept email format
				Username:     parent.Email,
				Relationship: upb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER,
				Password:     parent.Password,
			}
		}

		createParentsAndAssignToStudentRes, err := upb.NewUserModifierServiceClient(helper.UserMgmtGRPCConn).CreateParentsAndAssignToStudent(
			ctxWithToken,
			&upb.CreateParentsAndAssignToStudentRequest{
				SchoolId: organizationID,
				StudentId: createStudentResponse.
					StudentProfile.
					Student.
					UserProfile.
					UserId,
				ParentProfiles: parentProfiles,
			},
		)
		if err != nil {
			return nil, err
		}

		for index := range parentProfiles {
			student.Parents[index].ID = createParentsAndAssignToStudentRes.
				GetParentProfiles()[index].
				GetParent().
				GetUserProfile().
				GetUserId()
		}
	}

	return student, nil
}

func (helper *ConversationMgmtHelper) CreateStudentWithName(schoolAdmin *entities.Staff, grade *entities.GradeMaster, locIDs []string, schoolID string, firstName, lastName string) (*entities.Student, error) {
	organizationID := schoolAdmin.OrganizationIDs[0]
	ctxWithToken, cancel := contextWithTokenAndTimeOut(context.Background(), schoolAdmin.Token)
	defer cancel()

	// create student
	student, createStudentRequest := helper.newstudentInfo(organizationID, grade, locIDs, &CreateStudentOpt{
		Name: firstName + " " + lastName,
	}, schoolID)
	// nolint
	createStudentResponse, err := upb.NewUserModifierServiceClient(helper.UserMgmtGRPCConn).CreateStudent(ctxWithToken, createStudentRequest)
	if err != nil {
		return nil, err
	}
	student.ID = createStudentResponse.StudentProfile.Student.UserProfile.UserId

	err = helper.createStudentEnrollmentStatus(ctxWithToken, fmt.Sprint(organizationID), student.ID, locIDs)
	if err != nil {
		return nil, fmt.Errorf("err create student enrollment status: %v", err)
	}

	return student, nil
}

func (helper *ConversationMgmtHelper) RemoveParentFromStudent(admin *entity.Admin, student *entity.Student, parent *entity.User) error {
	ctx, cancel := contextWithTokenAndTimeOut(context.Background(), admin.Token)
	defer cancel()
	_, err := upb.NewUserModifierServiceClient(helper.UserMgmtGRPCConn).RemoveParentFromStudent(ctx, &upb.RemoveParentFromStudentRequest{
		StudentId: student.ID,
		ParentId:  parent.ID,
	})
	return err
}

func (helper *ConversationMgmtHelper) UpdateStudentWithParent(admin *entity.Admin, student *entity.Student) error {
	ctx, cancel := contextWithTokenAndTimeOut(context.Background(), admin.Token)
	defer cancel()

	parentProfiles := make([]*upb.UpdateParentsAndFamilyRelationshipRequest_ParentProfile, 0)

	for _, p := range student.Parents {
		firstName, lastName := featuresHelper.SplitNameToFirstNameAndLastName(p.Name)
		parentProfiles = append(parentProfiles, &upb.UpdateParentsAndFamilyRelationshipRequest_ParentProfile{
			Id:    p.ID,
			Email: p.Email,
			UserNameFields: &upb.UserNameFields{
				FirstName: firstName,
				LastName:  lastName,
			},
			Relationship: upb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER,
		})
	}

	_, err := upb.NewUserModifierServiceClient(helper.UserMgmtGRPCConn).UpdateParentsAndFamilyRelationship(ctx, &upb.UpdateParentsAndFamilyRelationshipRequest{
		SchoolId:       student.SchoolID,
		StudentId:      student.ID,
		ParentProfiles: parentProfiles,
	})

	if err != nil {
		return fmt.Errorf("error when call upb.UpdateParentsAndFamilyRelationship: %w", err)
	}

	return nil
}

func (helper *ConversationMgmtHelper) AddCourseToStudentWithLocation(admin *entities.Staff, student *entities.Student, courses []*entities.Course, mapCourseIDAndStudentIDs map[string][]string, locationIDs []string) error {
	ctx, cancel := contextWithTokenAndTimeOut(context.Background(), admin.Token)
	defer cancel()
	packages := []*entities.StudentPackage{}

	for _, course := range courses {
		upsertStdPkgRequest := &upb.UpsertStudentCoursePackageRequest{
			StudentId: student.ID,
			StudentPackageProfiles: []*upb.UpsertStudentCoursePackageRequest_StudentPackageProfile{
				{
					Id: &upb.UpsertStudentCoursePackageRequest_StudentPackageProfile_CourseId{
						CourseId: course.ID,
					},
					StartTime:   timestamppb.New(time.Now().Add(time.Hour * -20)),
					EndTime:     timestamppb.New(time.Now().Add(time.Hour * 24 * 10)),
					LocationIds: locationIDs,
					StudentPackageExtra: []*upb.StudentPackageExtra{
						{
							LocationId: locationIDs[0],
							ClassId:    "",
						},
					},
				},
			},
		}
		if len(course.Classes) > 0 {
			upsertStdPkgRequest.StudentPackageProfiles[0].StudentPackageExtra = []*upb.StudentPackageExtra{
				{
					ClassId:    course.Classes[0].ID,
					LocationId: course.LocationIDs[0],
				},
			}
		}
		res, err := upb.NewUserModifierServiceClient(helper.UserMgmtGRPCConn).UpsertStudentCoursePackage(ctx, upsertStdPkgRequest)

		if err != nil {
			return fmt.Errorf("school admin unable to add course to student: %w", err)
		}

		if len(res.StudentPackageProfiles) > 0 {
			for _, profile := range res.StudentPackageProfiles {
				stuPkg := &entities.StudentPackage{
					ID:       profile.StudentCoursePackageId,
					CourseID: profile.CourseId,
				}
				if len(profile.StudentPackageExtra) > 0 {
					stuPkg.ClassID = profile.StudentPackageExtra[0].ClassId
					stuPkg.LocationID = profile.StudentPackageExtra[0].LocationId
				}
				packages = append(packages, stuPkg)
			}
		}

		mapCourseIDAndStudentIDs[course.ID] = append(mapCourseIDAndStudentIDs[course.ID], student.ID)
	}
	student.Packages = packages
	student.Courses = courses

	return nil
}

func (helper *ConversationMgmtHelper) CreateStudentWithParent(admin *entities.Staff, organizationID int32, grade *entities.GradeMaster, locIDs []string, optP *CreateStudentWithParentOpt, schoolID string) (*entities.Student, *entities.User, error) {
	opt := CreateStudentWithParentOpt{}
	if optP != nil {
		opt = *optP
	}

	student, req := helper.newstudentInfo(organizationID, grade, locIDs, &CreateStudentOpt{
		Name: opt.StudentName,
	}, schoolID)
	// nolint
	res, err := upb.NewUserModifierServiceClient(helper.UserMgmtGRPCConn).CreateStudent(contextWithToken(context.Background(), admin.Token), req)
	if err != nil {
		return nil, nil, fmt.Errorf("upb.NewUserModifierServiceClient.CreateStudent: %v", err)
	}
	student.ID = res.StudentProfile.Student.UserProfile.UserId

	err = helper.createStudentEnrollmentStatus(context.Background(), fmt.Sprint(organizationID), student.ID, locIDs)
	if err != nil {
		return nil, nil, fmt.Errorf("err create student enrollment status: %v", err)
	}

	parent := &entities.User{
		Name:     helper.PickName(),
		Email:    fmt.Sprintf("%v-parent-1@example.com", student.ID),
		Phone:    util.RandPhoneNumber(),
		Password: idutil.ULIDNow(),
		Group:    cpb.UserGroup_USER_GROUP_PARENT.String(),
	}
	student.Parents = append(student.Parents, parent)

	req2 := &upb.CreateParentsAndAssignToStudentRequest{
		SchoolId:  organizationID,
		StudentId: res.StudentProfile.Student.UserProfile.UserId,
		ParentProfiles: []*upb.CreateParentsAndAssignToStudentRequest_ParentProfile{
			{
				Name:        parent.Name,
				CountryCode: cpb.Country_COUNTRY_VN,
				PhoneNumber: parent.Phone,
				Email:       parent.Email,
				// username can accept email format
				Username:     parent.Email,
				Relationship: upb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER,
				Password:     parent.Password,
			},
		},
	}
	res2, err := upb.NewUserModifierServiceClient(helper.UserMgmtGRPCConn).CreateParentsAndAssignToStudent(contextWithToken(context.Background(), admin.Token), req2)
	if err != nil {
		return nil, nil, fmt.Errorf("upb.NewUserModifierServiceClient.CreateParentsAndAssignToStudent: %v", err)
	}

	student.Parents[0].ID = res2.
		GetParentProfiles()[0].
		GetParent().
		GetUserProfile().
		GetUserId()

	return student, parent, nil
}

func (helper *ConversationMgmtHelper) CreateStudentsWithSameParent(admin *entities.Staff, organizationID, _ int32, grade *entities.GradeMaster, locIDs []string, optP *CreateStudentsWithSameParentOpt, schoolID string) ([]*entities.Student, *entities.User, error) {
	opt := CreateStudentsWithSameParentOpt{}
	if optP != nil {
		opt = *optP
	}

	if len(opt.StudentsName) != opt.NumberOfStudent {
		opt.StudentsName = make([]string, opt.NumberOfStudent)

		for i := 0; i < len(opt.StudentsName); i++ {
			name := idutil.ULIDNow()
			opt.StudentsName[i] = name
		}
	}

	firstStudent, parent, err := helper.CreateStudentWithParent(admin, organizationID, grade, locIDs, &CreateStudentWithParentOpt{
		StudentName: opt.StudentsName[0],
		ParentName:  opt.ParentName,
	}, schoolID)

	if err != nil {
		return nil, nil, fmt.Errorf("s.CreateStudentWithParent: %v", err)
	}

	students := make([]*entities.Student, 0)
	students = append(students, firstStudent)

	for i := 0; i < opt.NumberOfStudent-1; i++ {
		student, req := helper.newstudentInfo(organizationID, grade, locIDs, &CreateStudentOpt{
			Name: opt.StudentsName[i+1],
		}, schoolID)
		// nolint
		res, err := upb.NewUserModifierServiceClient(helper.UserMgmtGRPCConn).CreateStudent(contextWithToken(context.Background(), admin.Token), req)
		if err != nil {
			return nil, nil, err
		}
		student.ID = res.StudentProfile.Student.UserProfile.UserId

		err = helper.createStudentEnrollmentStatus(context.Background(), fmt.Sprint(organizationID), student.ID, locIDs)
		if err != nil {
			return nil, nil, fmt.Errorf("err create student enrollment status: %v", err)
		}

		firstName, lastName := featuresHelper.SplitNameToFirstNameAndLastName(parent.Name)
		reqUpdateParentsAndFamilyRelationshipRequest := &upb.UpdateParentsAndFamilyRelationshipRequest{
			SchoolId:  organizationID,
			StudentId: student.ID,
			ParentProfiles: []*upb.UpdateParentsAndFamilyRelationshipRequest_ParentProfile{
				{
					Id:    parent.ID,
					Email: parent.Email,
					// username can accept email format
					Username: parent.Email,
					UserNameFields: &upb.UserNameFields{
						FirstName: firstName,
						LastName:  lastName,
					},
					Relationship: upb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER,
				},
			},
		}

		_, err = upb.NewUserModifierServiceClient(helper.UserMgmtGRPCConn).UpdateParentsAndFamilyRelationship(contextWithToken(context.Background(), admin.Token), reqUpdateParentsAndFamilyRelationshipRequest)
		if err != nil {
			return nil, nil, err
		}

		students = append(students, student)
	}

	return students, parent, nil
}

func (helper *ConversationMgmtHelper) newstudentInfo(organizationID int32, grade *entities.GradeMaster, locIDs []string, optP *CreateStudentOpt, schoolID string) (*entities.Student, *upb.CreateStudentRequest) {
	opt := CreateStudentOpt{}
	if optP != nil {
		opt = *optP
	}

	// create student
	student := &entities.Student{}
	randomID := idutil.ULIDNow()
	student.Password = fmt.Sprintf("password-%v", randomID)
	student.Email = fmt.Sprintf("student-%v@example.com", randomID)
	student.Name = helper.PickName()
	if opt.Name != "" {
		splitNameArr := strings.Split(opt.Name, " ")
		if len(splitNameArr) > 0 {
			student.FirstName = splitNameArr[0]
		}
		if len(splitNameArr) > 1 {
			student.LastName = splitNameArr[1]
		}
		student.Name = opt.Name
	}
	student.Phone = util.RandPhoneNumber()
	student.OrganizationID = organizationID
	student.Group = cpb.UserGroup_USER_GROUP_STUDENT.String()

	student.GradeMaster = grade

	createStudentRequest := &upb.CreateStudentRequest{
		SchoolId: student.OrganizationID,
		StudentProfile: &upb.CreateStudentRequest_StudentProfile{
			Email:            student.Email,
			Password:         student.Password,
			Name:             student.Name,
			FirstName:        student.FirstName,
			LastName:         student.LastName,
			CountryCode:      cpb.Country_COUNTRY_VN,
			PhoneNumber:      student.Phone,
			GradeId:          student.GradeMaster.ID,
			EnrollmentStatus: upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
			LocationIds:      locIDs,
		},
		SchoolHistories: []*upb.SchoolHistory{
			{
				SchoolId: schoolID,
			},
		},
	}

	return student, createStudentRequest
}

func (helper *ConversationMgmtHelper) createStudentEnrollmentStatus(ctx context.Context, resourcePath, studentID string, locationIDs []string) error {
	ctx2 := interceptors.ContextWithJWTClaims(ctx, &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: resourcePath,
		},
	})

	for _, locID := range locationIDs {
		query := `
			INSERT INTO public.student_enrollment_status_history
			(student_id, location_id, enrollment_status, start_date, end_date, "comment", created_at, updated_at, deleted_at, resource_path, order_id, order_sequence_number)
			VALUES($1, $2, 'STUDENT_ENROLLMENT_STATUS_ENROLLED', timezone('utc'::text, now() - interval '24 hours'), NULL, '', timezone('utc'::text, now()), timezone('utc'::text, now()), NULL, autofillresourcepath(), NULL, NULL);
		`

		_, err := helper.BobDBConn.Exec(ctx2, query, studentID, locID)
		if err != nil {
			return err
		}
	}

	return nil
}

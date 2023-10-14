package helper

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/common"
	"github.com/manabie-com/backend/features/eibanam/communication/entity"
	"github.com/manabie-com/backend/features/eibanam/communication/util"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type CreateStudentsWithSameParentOpt struct {
	StudentsName    []string
	NumberOfStudent int
	ParentName      string
}

func (h *CommunicationHelper) CreateStudent(schoolAdmin *entity.Admin, gradeId int32, locIDs []string, hasParent bool, numOfParent int) (*entity.Student, error) {
	schoolID := int32(schoolAdmin.SchoolIds[0])
	ctxWithToken, cancel := util.ContextWithTokenAndTimeOut(context.Background(), schoolAdmin.Token)
	defer cancel()

	// create student
	student := &entity.Student{}
	randomID := idutil.ULIDNow()
	student.Password = idutil.ULIDNow()
	student.Email = fmt.Sprintf("%v@example.com", randomID)
	student.Name = h.PickName()
	student.Phone = util.RandPhoneNumber()
	student.SchoolID = schoolID
	student.Group = cpb.UserGroup_USER_GROUP_STUDENT.String()

	grade := &entity.Grade{ID: gradeId}
	student.Grade = grade

	createStudentRequest := &upb.CreateStudentRequest{
		SchoolId: student.SchoolID,
		StudentProfile: &upb.CreateStudentRequest_StudentProfile{
			Email:            student.Email,
			Password:         student.Password,
			Name:             student.Name,
			CountryCode:      cpb.Country_COUNTRY_VN,
			PhoneNumber:      student.Phone,
			Grade:            int32(student.Grade.ID),
			EnrollmentStatus: upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
			LocationIds:      locIDs,
		},
	}

	createStudentResponse, err := upb.NewUserModifierServiceClient(h.userMgmtGrpcConn).CreateStudent(ctxWithToken, createStudentRequest)
	if err != nil {
		return nil, err
	}
	student.ID = createStudentResponse.StudentProfile.Student.UserProfile.UserId

	// if request has parents, create parent info
	if hasParent {
		parentProfiles := make([]*upb.CreateParentsAndAssignToStudentRequest_ParentProfile, numOfParent)
		for index := range parentProfiles {
			parent := entity.User{
				Name:     h.PickName(),
				Email:    fmt.Sprintf("%v.parent%v@example.com", randomID, index),
				Phone:    util.RandPhoneNumber(),
				Password: idutil.ULIDNow(),
				Group:    cpb.UserGroup_USER_GROUP_PARENT.String(),
			}
			student.Parents = append(student.Parents, &parent)

			// reassign entity eibanam to parent
			parentProfiles[index] = &upb.CreateParentsAndAssignToStudentRequest_ParentProfile{
				Name:         parent.Name,
				CountryCode:  cpb.Country_COUNTRY_VN,
				PhoneNumber:  parent.Phone,
				Email:        parent.Email,
				Relationship: upb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER,
				Password:     parent.Password,
			}
		}

		createParentsAndAssignToStudent, err := upb.NewUserModifierServiceClient(h.userMgmtGrpcConn).CreateParentsAndAssignToStudent(
			ctxWithToken,
			&upb.CreateParentsAndAssignToStudentRequest{
				SchoolId: schoolID,
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
			student.Parents[index].ID = createParentsAndAssignToStudent.
				GetParentProfiles()[index].
				GetParent().
				GetUserProfile().
				GetUserId()
		}
	}

	return student, nil
}

func (h *CommunicationHelper) RemoveParentFromStudent(admin *entity.Admin, student *entity.Student, parent *entity.User) error {
	ctx, cancel := util.ContextWithTokenAndTimeOut(context.Background(), admin.Token)
	defer cancel()
	_, err := upb.NewUserModifierServiceClient(h.userMgmtGrpcConn).RemoveParentFromStudent(ctx, &upb.RemoveParentFromStudentRequest{
		StudentId: student.ID,
		ParentId:  parent.ID,
	})
	return err
}

func (h *CommunicationHelper) UpdateStudentWithParent(admin *entity.Admin, student *entity.Student) error {
	ctx, cancel := util.ContextWithTokenAndTimeOut(context.Background(), admin.Token)
	defer cancel()

	parentProfiles := make([]*upb.UpdateParentsAndFamilyRelationshipRequest_ParentProfile, 0)

	for _, p := range student.Parents {
		parentProfiles = append(parentProfiles, &upb.UpdateParentsAndFamilyRelationshipRequest_ParentProfile{
			Id:           p.ID,
			Email:        p.Email,
			Relationship: upb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER,
		})
	}

	_, err := upb.NewUserModifierServiceClient(h.userMgmtGrpcConn).UpdateParentsAndFamilyRelationship(ctx, &upb.UpdateParentsAndFamilyRelationshipRequest{
		SchoolId:       student.SchoolID,
		StudentId:      student.ID,
		ParentProfiles: parentProfiles,
	})

	if err != nil {
		return fmt.Errorf("error when call upb.UpdateParentsAndFamilyRelationship: %w", err)
	}

	return nil
}
func (h *CommunicationHelper) AddCourseToStudentWithLocation(admin *entity.Admin, student *entity.Student, courses []*entity.Course, locationIDs []string) error {
	ctx, cancel := util.ContextWithTokenAndTimeOut(context.Background(), admin.Token)
	defer cancel()

	for _, course := range courses {
		_, err := upb.NewUserModifierServiceClient(h.userMgmtGrpcConn).UpsertStudentCoursePackage(ctx, &upb.UpsertStudentCoursePackageRequest{
			StudentId: student.ID,
			StudentPackageProfiles: []*upb.UpsertStudentCoursePackageRequest_StudentPackageProfile{
				{
					Id: &upb.UpsertStudentCoursePackageRequest_StudentPackageProfile_CourseId{
						CourseId: course.ID,
					},
					StartTime:   timestamppb.New(time.Now().Add(time.Hour * -20)),
					EndTime:     timestamppb.New(time.Now().Add(time.Hour * 24 * 10)),
					LocationIds: locationIDs,
				},
			},
		})
		if err != nil {
			return fmt.Errorf("school admin unable to add course to student: %w", err)
		}
	}

	student.Courses = courses

	return nil
}

func (h *CommunicationHelper) AddCourseToStudent(admin *entity.Admin, student *entity.Student, courses []*entity.Course) error {
	return h.AddCourseToStudentWithLocation(admin, student, courses, nil)
}

// Deprecated: please use functions from StatefulHelper instead
func (h *CommunicationHelper) CreateStudentsWithSameParent(ctx context.Context, schoolAdminToken string, locIDs []string, optP *CreateStudentsWithSameParentOpt) ([]*upb.Student, *upb.Parent, error) {
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

	authedCtx := h.Suite.CtxWithAuthToken(ctx, schoolAdminToken)
	firstStudent, parent, err := h.Suite.CreateStudentWithParent(authedCtx, locIDs, &common.CreateStudentWithParentOpt{
		StudentName: opt.StudentsName[0],
		ParentName:  opt.ParentName,
	})

	if err != nil {
		return nil, nil, err
	}

	ret := make([]*upb.Student, 0)
	ret = append(ret, firstStudent)

	for i := 0; i < opt.NumberOfStudent-1; i++ {
		student, err := h.Suite.CreateStudent(authedCtx, locIDs, &common.CreateStudentOpt{
			Name: opt.StudentsName[i+1],
		})
		if err != nil {
			return nil, nil, err
		}
		err = h.Suite.UpdateStudentParent(authedCtx, student.UserProfile.UserId, parent.UserProfile.UserId, parent.UserProfile.Email)
		if err != nil {
			return nil, nil, err
		}

		ret = append(ret, student)
	}

	return ret, parent, nil
}

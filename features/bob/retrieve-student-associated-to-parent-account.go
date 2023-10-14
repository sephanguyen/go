package bob

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	ypb "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"
)

type fetchedStudent struct {
	UserID    string
	Country   string
	UpdatedAt time.Time
}

func (s *suite) createHandsomeFatherAsAParentAndTheRelationshipWithHisChildrenWhoreStudentsAtManabie(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	num := rand.Int63()
	profile := &ypb.CreateUserProfile{
		Name:        fmt.Sprintf("create_user+%d", num),
		PhoneNumber: fmt.Sprintf("+848%d", num),
		Email:       fmt.Sprintf("create_user+%d@gmail.com", num),
		Country:     cpb.Country_COUNTRY_VN,
		Avatar:      fmt.Sprintf("http://valid-user+%d", num),
		Grade:       int32(rand.Intn(3) + 10),
	}

	profiles := []*ypb.CreateUserProfile{
		profile,
	}

	userGroup := cpb.UserGroup_USER_GROUP_PARENT
	schoolID := int64(1)

	stepState.Request = &ypb.CreateUserRequest{
		Users:        profiles,
		UserGroup:    userGroup,
		SchoolId:     schoolID,
		Organization: strconv.Itoa(int(schoolID)),
	}
	req := stepState.Request.(*ypb.CreateUserRequest)

	stepState.Response, stepState.ResponseErr = ypb.NewUserModifierServiceClient(s.YsConn).CreateUser(s.signedCtx(ctx), req)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr

	}

	resp := stepState.Response.(*ypb.CreateUserResponse)
	stepState.CurrentParentID = resp.Users[0].UserId

	ss := make([]*ypb.AssignToParentRequest_AssignParent, 0)
	for _, studentID := range stepState.OtherStudentIDs {
		t := &ypb.AssignToParentRequest_AssignParent{
			StudentId:    studentID,
			ParentId:     resp.Users[0].UserId,
			Relationship: ypb.FamilyRelationship_FAMILY_RELATIONSHIP_GUARDIAN,
		}

		ss = append(ss, t)
	}

	stepState.Response, stepState.ResponseErr = ypb.NewUserModifierServiceClient(s.YsConn).AssignToParent(s.signedCtx(ctx), &ypb.AssignToParentRequest{AssignParents: ss})
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) multipleStudentsProfileInDB(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	for i := 0; i < 2; i++ {
		if ctx, err := s.anOtherStudentProfileInDB(ctx); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) aSignedInParent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var err error
	ctx, err = s.aSignedIn(ctx, "parent")
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	stepState.CurrentUserID = stepState.CurrentParentID
	stepState.CurrentUserGroup = constant.UserGroupParent
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) retrieveStudentsProfilesAssociatedToParentAccount(ctx context.Context) (context.Context, error) {
	ctx = interceptors.ContextWithUserGroup(ctx, cpb.UserGroup_USER_GROUP_PARENT.String())
	stepState := StepStateFromContext(ctx)
	stepState.Response, stepState.ResponseErr = bpb.NewStudentReaderServiceClient(s.Conn).
		RetrieveStudentAssociatedToParentAccount(helper.GRPCContext(ctx, "token", stepState.AuthToken), &bpb.RetrieveStudentAssociatedToParentAccountRequest{})
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr

	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) fetchedStudentsExactlyAssociatedToParent(ctx context.Context) (context.Context, error) {
	beautyQuery := `
		SELECT
			u.user_id AS user_id,
			u.country AS country,
			sp.updated_at AS updated_at
		FROM users AS u
		JOIN public.students AS s ON u.user_id = s.student_id AND s.deleted_at IS NULL
		JOIN student_parents AS sp ON sp.student_id = s.student_id AND sp.parent_id = $1 AND sp.deleted_at IS NULL
		JOIN parents AS p ON p.parent_id = sp.parent_id AND p.deleted_at IS NULL
		LEFT OUTER JOIN apple_users AS au ON au.user_id = u.user_id
		WHERE u.deleted_at IS NULL
		ORDER BY  sp.updated_at ASC
	`
	fetchedStudents := make([]*fetchedStudent, 0)
	stepState := StepStateFromContext(ctx)

	rows, err := s.DB.Query(ctx, beautyQuery, &stepState.CurrentParentID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	defer rows.Close()
	for rows.Next() {
		s := new(fetchedStudent)
		if err := rows.Scan(&s.UserID, &s.Country, &s.UpdatedAt); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		fetchedStudents = append(fetchedStudents, s)
	}

	for i := 0; i < len(fetchedStudents)-1; i++ {
		if fetchedStudents[i].UpdatedAt.After(fetchedStudents[i+1].UpdatedAt) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected return order")
		}
	}

	count := 0
	resp := stepState.Response.(*bpb.RetrieveStudentAssociatedToParentAccountResponse).Profiles
	for _, r := range resp {
		for _, fs := range fetchedStudents {
			if s.validateStudent(r, fs) {
				count++
				break
			}
		}
	}

	if count == len(resp) && len(resp) == len(fetchedStudents) {
		return StepStateToContext(ctx, stepState), nil
	} else {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected number of response return")
	}
}

func (s *suite) validateStudent(resp *cpb.BasicProfile, student *fetchedStudent) bool {
	if resp.UserId != student.UserID {
		return false
	}
	if resp.Country.String() != student.Country {
		return false
	}

	return true
}

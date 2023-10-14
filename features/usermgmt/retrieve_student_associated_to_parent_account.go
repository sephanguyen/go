package usermgmt

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgtype"
)

func (s *suite) createFatherAndMotherAsAParentAndTheRelationshipWithTheirChildren(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	permissionRole := schoolAdminType

	// create student account
	ctx, err := s.onlyStudentInfo(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to get student info: %v", err)
	}
	ctx, err = s.createNewStudentAccount(ctx, permissionRole)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to create student account: %v", err)
	}
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to create student account: %v", stepState.ResponseErr)
	}

	// create father and mother accounts of the student
	fatherID := newID()
	motherID := newID()
	profiles := []*pb.CreateParentsAndAssignToStudentRequest_ParentProfile{
		{
			Name:         fmt.Sprintf("user-%v", fatherID),
			CountryCode:  cpb.Country_COUNTRY_VN,
			PhoneNumber:  fmt.Sprintf("phone-number-%v", fatherID),
			Email:        fmt.Sprintf("%v@example.com", fatherID),
			Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER,
			Password:     fmt.Sprintf("password-%v", fatherID),
		},
		{
			Name:         fmt.Sprintf("user-%v", motherID),
			CountryCode:  cpb.Country_COUNTRY_VN,
			PhoneNumber:  fmt.Sprintf("phone-number-%v", motherID),
			Email:        fmt.Sprintf("%v@example.com", motherID),
			Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_MOTHER,
			Password:     fmt.Sprintf("password-%v", motherID),
		},
	}
	stepState.Request = &pb.CreateParentsAndAssignToStudentRequest{
		SchoolId:       constants.ManabieSchool,
		StudentId:      stepState.CurrentStudentID,
		ParentProfiles: profiles,
	}
	ctx, err = s.createNewParents(ctx, permissionRole)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to create parents: %v", err)
	}

	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to create parents: %v", stepState.ResponseErr)
	}

	// get parent ids
	for _, parent := range stepState.Response.(*pb.CreateParentsAndAssignToStudentResponse).ParentProfiles {
		stepState.ParentIDs = append(stepState.ParentIDs, parent.Parent.UserProfile.UserId)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) removeRelationshipOfStudent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stmt := `UPDATE student_parents set deleted_at = NOW() WHERE student_id = $1`
	// Remove relationship of student
	if _, err := s.BobPostgresDB.Exec(ctx, stmt, stepState.CurrentStudentID); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("err remove relationship of student: %w", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) retrieveStudentsProfilesParentAccount(ctx context.Context) (*pb.RetrieveStudentAssociatedToParentAccountResponse, error) {
	ctx = interceptors.ContextWithUserGroup(ctx, cpb.UserGroup_USER_GROUP_PARENT.String())
	req := &pb.RetrieveStudentAssociatedToParentAccountRequest{}

	return pb.NewUserReaderServiceClient(s.UserMgmtConn).RetrieveStudentAssociatedToParentAccount(contextWithToken(ctx), req)
}

func (s *suite) retrieveStudentsProfilesEachParentAccount(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	profiles := make([]*cpb.BasicProfile, 0)

	for _, parentId := range stepState.ParentIDs {
		token, err := s.generateExchangeToken(parentId, constant.UserGroupParent)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to sign in parent: %v", err)
		}
		stepState.AuthToken = token

		resp, err := s.retrieveStudentsProfilesParentAccount(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve student associated to parent account %s: %v", parentId, err)
		}

		profiles = append(profiles, resp.Profiles...)
	}

	stepState.UserProfiles = profiles
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) fetchedStudentsExactlyAssociatedToParent(ctx context.Context) (context.Context, error) {
	query := `
		SELECT
			u.user_id AS user_id,
			u.country AS country,
			u.first_name AS first_name,
			u.last_name AS last_name
		FROM users AS u

		JOIN public.students AS s ON 
			u.user_id = s.student_id
				AND
			s.deleted_at IS NULL

		JOIN student_parents AS sp ON
			sp.student_id = s.student_id
				AND 
			sp.parent_id = ANY ($1)
				AND
			sp.deleted_at IS NULL

		JOIN parents AS p ON
			p.parent_id = sp.parent_id
				AND
			p.deleted_at IS NULL

		LEFT OUTER JOIN apple_users AS au ON
			au.user_id = u.user_id

		WHERE
			u.deleted_at IS NULL
		ORDER BY
			sp.created_at ASC
	`

	type fetchedStudent struct {
		UserID    string
		Country   string
		FirstName string
		LastName  string
	}

	// fetchedStudents := make([]*fetchedStudent, 0)
	countFetched := 0
	fetchedStudents := make(map[string]*fetchedStudent)
	stepState := StepStateFromContext(ctx)

	parentIDs := pgtype.TextArray{}
	if err := parentIDs.Set(stepState.ParentIDs); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to set parent ids: %v", err)
	}

	rows, err := s.BobPostgresDB.Query(ctx, query, &parentIDs)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to query database: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		fetched := new(fetchedStudent)
		if err := rows.Scan(&fetched.UserID, &fetched.Country, &fetched.FirstName, &fetched.LastName); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to scan row: %v", err)
		}
		fetchedStudents[fetched.UserID] = fetched
		countFetched++
	}

	if rows.Err() != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to scan row: %v", err)
	}

	// check if the fetched students exactly match the expected students
	count := 0
	resps := stepState.UserProfiles
	for _, resp := range resps {
		fetchedStudent, ok := fetchedStudents[resp.UserId]
		if !ok {
			return nil, fmt.Errorf("failed to find student %s in fetched students", resp.UserId)
		}
		if fetchedStudent.Country != resp.Country.String() {
			continue
		}
		count++
	}

	if !(count == len(resps) && len(resps) == countFetched) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected number of response return")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) sameStudentProfiles(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	firstProfile := stepState.UserProfiles[0]
	for _, profile := range stepState.UserProfiles[1:] {
		if !func() bool {
			if firstProfile.UserId != profile.UserId {
				return false
			}

			if firstProfile.Country.String() != profile.Country.String() {
				return false
			}

			return true
		}() {
			return StepStateToContext(ctx, stepState), fmt.Errorf("s.StudentProfiles: student profiles are not the same")
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) noStudentsProfilesAreFetched(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if len(stepState.UserProfiles) != 0 {
		return ctx, fmt.Errorf("unexpected number of student profiles are fetched")
	}
	return StepStateToContext(ctx, stepState), nil
}

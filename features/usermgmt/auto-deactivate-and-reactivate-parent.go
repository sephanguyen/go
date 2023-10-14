package usermgmt

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	http_port "github.com/manabie-com/backend/internal/usermgmt/modules/user/port/http"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"

	"github.com/pkg/errors"
)

func (s *suite) createsParentByOpenAPI(ctx context.Context, numberOfStudents int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := generateUpsertParentOpenAPIReq(stepState.StudentEmails[:numberOfStudents])
	stepState.Request = req
	stepState.Users = []*entity.LegacyUser{
		{
			ExternalUserID: database.Text(req.Parents[0].ExternalUserID().String()),
			Email:          database.Text(req.Parents[0].Email().String()),
			FirstName:      database.Text(req.Parents[0].FirstName().String()),
			LastName:       database.Text(req.Parents[0].LastName().String()),
		},
	}
	ctx, err := s.externalServiceCallUpsertParentsAPI(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) addMoreStudentToParent(ctx context.Context, condition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	_, err := s.createStudentsByOpenAPI(ctx, 1, condition, "enrollment_status_histories")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	currentParents := stepState.Users
	req := generateUpsertParentOpenAPIReq(stepState.StudentEmails)
	for idx, parent := range req.Parents {
		parent.ExternalUserIDAttr = field.NewString(currentParents[idx].ExternalUserID.String)
		parent.EmailAttr = field.NewString(currentParents[idx].Email.String)
		parent.FirstNameAttr = field.NewString(currentParents[idx].FirstName.String)
		parent.LastNameAttr = field.NewString(currentParents[idx].LastName.String)
	}
	stepState.Request = req

	_, err = s.externalServiceCallUpsertParentsAPI(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return ctx, nil
}

func (s *suite) removeStudentFromParent(ctx context.Context, numberOfStudents int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	currentParents := stepState.Users
	req := generateUpsertParentOpenAPIReq(stepState.StudentEmails[:len(stepState.StudentEmails)-numberOfStudents])
	for idx, parent := range req.Parents {
		parent.ExternalUserIDAttr = field.NewString(currentParents[idx].ExternalUserID.String)
		parent.EmailAttr = field.NewString(currentParents[idx].Email.String)
		parent.FirstNameAttr = field.NewString(currentParents[idx].FirstName.String)
		parent.LastNameAttr = field.NewString(currentParents[idx].LastName.String)
	}
	stepState.Request = req

	_, err := s.externalServiceCallUpsertParentsAPI(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return ctx, nil
}

func (s *suite) validationParentActivation(ctx context.Context, parentActivation string) (context.Context, error) {
	req := s.Request.(http_port.UpsertParentRequest)
	resp := s.Response.(http_port.ResponseErrors)

	if resp.Code != 20000 {
		return ctx, fmt.Errorf("message: %s, code: %d", resp.Message, resp.Code)
	}

	parentsResponse, ok := resp.Data.([]interface{})
	if !ok {
		return ctx, fmt.Errorf("resp.Data.([]map[string]interface{}) is invalid")
	}

	for i := range req.Parents {
		parentResponse, ok := parentsResponse[i].(map[string]interface{})
		if !ok {
			return ctx, fmt.Errorf("parentsResponse[i].(map[string]interface{}) is invalid")
		}

		parentID, ok := parentResponse["user_id"].(string)
		if !ok {
			return ctx, fmt.Errorf("cannot get user_id from parentResponse")
		}

		activationStatusStmt := `SELECT deactivated_at FROM users WHERE user_id = $1`
		row := s.BobDBTrace.QueryRow(ctx, activationStatusStmt, parentID)
		parentDeactivatedAt := field.NewNullTime()
		if err := row.Scan(&parentDeactivatedAt); err != nil {
			return ctx, err
		}

		if parentActivation == activated && field.IsPresent(parentDeactivatedAt) {
			return ctx, fmt.Errorf("expect parent is %s but got %s", activated, deactivated)
		}
		if parentActivation == deactivated && !field.IsPresent(parentDeactivatedAt) {
			return ctx, fmt.Errorf("expect parent is %s but got %s", deactivated, activated)
		}

		studentActivationStmt := `SELECT deactivated_at FROM users u 
		INNER JOIN student_parents sp ON u.user_id = sp.student_id
		WHERE sp.parent_id = $1 AND sp.deleted_at IS NULL`

		rows, err := s.BobDBTrace.Query(ctx, studentActivationStmt, parentID)
		if err != nil {
			return ctx, errors.Wrap(err, "query student's deactivated_at parentID")
		}

		studentsDeactivatedAt := []field.Time{}
		for rows.Next() {
			parentDeactivatedAt := field.NewNullTime()

			err := rows.Scan(&parentDeactivatedAt)
			if err != nil {
				return ctx, errors.WithMessage(err, "rows.Scan student's deactivated_at")
			}
			studentsDeactivatedAt = append(studentsDeactivatedAt, parentDeactivatedAt)
		}

		isAllChildrenAreDeactivated := true
		for _, deactivatedAt := range studentsDeactivatedAt {
			if !field.IsPresent(deactivatedAt) {
				isAllChildrenAreDeactivated = false
				break
			}
		}

		if isAllChildrenAreDeactivated != field.IsPresent(parentDeactivatedAt) {
			return ctx, fmt.Errorf("parent activation is incorrect")
		}
	}
	return ctx, nil
}

func generateUpsertParentOpenAPIReq(studentEmails []string) http_port.UpsertParentRequest {
	id := idutil.ULIDNow()
	childrenAttrs := []http_port.ParentChildrenPayload{}

	for _, studentEmail := range studentEmails {
		childrenAttrs = append(childrenAttrs, http_port.ParentChildrenPayload{
			StudentEmailAttr: field.NewString(studentEmail),
			RelationshipAttr: field.NewInt32(1),
		})
	}

	request := http_port.UpsertParentRequest{
		Parents: []http_port.ParentProfile{
			{
				ExternalUserIDAttr: field.NewString(fmt.Sprintf("external-user-id-parent %s", id)),
				EmailAttr:          field.NewString(fmt.Sprintf("email-%s@example.com", id)),
				UserNameAttr:       field.NewString(fmt.Sprintf("username-by-email-format-%s@example.com", id)),
				FirstNameAttr:      field.NewString(fmt.Sprintf("parent first name %s", id)),
				LastNameAttr:       field.NewString(fmt.Sprintf("parent last name %s", id)),
				ChildrenAttr:       childrenAttrs,
			},
		},
	}

	return request
}

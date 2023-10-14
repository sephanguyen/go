package eureka

import "context"

func (s *suite) anotherSchoolAdminLogins(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.AnotherSchoolAdminID = stepState.SchoolAdminID
	stepState.AnotherSchoolAdminToken = stepState.SchoolAdminToken

	ctx, err := s.logins(ctx, "school admin")
	return StepStateToContext(ctx, stepState), err
}

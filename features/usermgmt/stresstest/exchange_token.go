package stresstest

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/common"
)

// Scenario_ExchangeTokenWithValidAuthenticationToken simulate
//Scenario: exchange token with valid authentication token
// Given an other student profile in DB
// And a valid authentication token with ID already exist in DB
// When a user exchange token
// Then our system need to do return valid token

func (s *Suite) Scenario_ExchangeTokenWithValidAuthenticationToken(ctx context.Context) error {
	ctx = common.StepStateToContext(ctx, s.userSuite.CommonSuite.StepState)
	err := s.ASignedInAsAccounts(ctx)
	if err != nil {
		return fmt.Errorf("ASignedInAsSchoolAdmin: %w", err)
	}

	return nil
}

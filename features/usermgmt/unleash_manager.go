package usermgmt

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/features/usermgmt/unleash"

	"github.com/pkg/errors"
)

func (s *suite) aScenarioRequiresWithCorrespondingStatuses(ctx context.Context, featureFlagNamesStr string, featureFlagStatusesStr string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	featureFlagNames := strings.Split(featureFlagNamesStr, ",")
	featureFlagStatuses := strings.Split(featureFlagStatusesStr, ",")

	switch {
	case len(featureFlagNames) < 1:
		return StepStateToContext(ctx, stepState), errors.New("expect number of feature flag names larger than 0")
	case len(featureFlagStatuses) < 1:
		return StepStateToContext(ctx, stepState), errors.New("expect number of feature flag statuses larger than 0")
	case len(featureFlagNames) != len(featureFlagStatuses):
		return StepStateToContext(ctx, stepState), errors.New("number of feature flag names and feature flag statuses must be equal")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) mustBeLockedAndHaveCorrespondingStatuses(ctx context.Context, featureFlagNamesStr string, featureFlagStatusesStr string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	featureFlagNames := strings.Split(featureFlagNamesStr, ",")
	featureFlagStatuses := strings.Split(featureFlagStatusesStr, ",")

	unleashClient := unleash.NewDefaultClient(s.UnleashSuite.UnleashSrvAddr, s.UnleashSuite.UnleashAPIKey, s.UnleashSuite.UnleashLocalAdminAPIKey)

	for i, featureFlagName := range featureFlagNames {
		featureFlagName = strings.TrimSpace(featureFlagName)
		featureFlagStatus := strings.TrimSpace(featureFlagStatuses[i])

		correct, err := unleashClient.IsFeatureToggleCorrect(ctx, featureFlagName, unleash.ToggleChoice(featureFlagStatus))
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		if !correct {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expect %s feature flag's status: %s", featureFlagName, featureFlagStatus)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

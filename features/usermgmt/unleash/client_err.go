package unleash

import "fmt"

type ErrFeatureFlagNotFound struct {
	FeatureFlagName string
}

func (err ErrFeatureFlagNotFound) Error() string {
	return fmt.Sprintf(`feature flag: "%s" not found`, err.FeatureFlagName)
}

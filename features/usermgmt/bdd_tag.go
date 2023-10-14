package usermgmt

import (
	"sort"
	"strings"

	"github.com/manabie-com/backend/features/usermgmt/unleash"

	"github.com/pkg/errors"
)

type FeatureFlagTag struct {
	Name         string
	ToggleChoice unleash.ToggleChoice
}

func ParseTags(tags ...string) ([]FeatureFlagTag, error) {
	featureFlagTags := make([]FeatureFlagTag, 0, len(tags))

	for _, tag := range tags {
		if strings.HasPrefix(tag, "@feature_flag") {
			featureFlagTag, err := ParseFeatureFlagTag(tag)
			if err != nil {
				return nil, err
			}
			featureFlagTags = append(featureFlagTags, *featureFlagTag)
		}
	}

	sort.SliceStable(featureFlagTags, func(i, j int) bool {
		return featureFlagTags[i].Name < featureFlagTags[j].Name
	})

	return featureFlagTags, nil
}

func ParseFeatureFlagTag(tag string) (*FeatureFlagTag, error) {
	splitTexts := strings.Split(tag, ":")

	if len(splitTexts) != 3 {
		return nil, errors.New("feature flag tag is malformed")
	}

	switch toggleChoice := strings.ToLower(splitTexts[2]); toggleChoice {
	case "enable", "disable":
		featureFlagTag := &FeatureFlagTag{
			Name:         splitTexts[1],
			ToggleChoice: unleash.ToggleChoice(toggleChoice),
		}
		return featureFlagTag, nil
	default:
		return nil, errors.New("feature flag's toggle choice is invalid")
	}
}

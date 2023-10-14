package tiertest

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/sliceutils"

	"github.com/cucumber/gherkin-go/v19"
	"github.com/cucumber/messages-go/v16"
)

type Tier int

const (
	TierUnknown Tier = iota
	TierMinor
	TierMajor
	TierCritical
	TierBlocker
)

var (
	TierMap = map[string]Tier{
		"minor":    TierMinor,
		"major":    TierMajor,
		"critical": TierCritical,
		"blocker":  TierBlocker,
	}

	TierDefault = TierMinor
)

func ignoreTagFunc(tags []*messages.Tag) bool {
	for _, t := range tags {
		if t.Name[1:] == "wip" || t.Name[1:] == "quarantined" {
			return true
		}
	}
	return false
}

func countScenariosInFolderWithTier(dir string, tier Tier) (int, int, error) {
	var (
		total   int
		matched int
	)
	err := filepath.Walk(dir, func(p string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if f.IsDir() {
			return nil
		}

		if !strings.HasSuffix(p, ".feature") {
			return nil
		}
		file, err := os.Open(p)
		if err != nil {
			return nil
		}
		defer file.Close()

		fileTotal, fileMatched, err := countScenariosWithTier(file, tier)
		if err != nil {
			return err
		}
		total += fileTotal
		matched += fileMatched
		return nil
	})
	if err != nil {
		return 0, 0, fmt.Errorf("filepath.Walk %s", err)
	}
	return total, matched, nil
}

// we consider a scenario with n Examples as n scenarios
func countScenariosWithTier(reader io.Reader, tier Tier) (int, int, error) {
	gherkinDocument, err := gherkin.ParseGherkinDocument(reader, func() string { return "" })
	if err != nil {
		return 0, 0, fmt.Errorf("gherkin.ParseGherkinDocument %s", err)
	}
	var (
		total   int
		matched int
	)
	// not a valid feature file
	if gherkinDocument.Feature == nil {
		return 0, 0, nil
	}
	globalMatchedTier := tagsMatchTier(gherkinDocument.Feature.Tags, tier)
	for _, c := range gherkinDocument.Feature.Children {
		// Background
		if c.Scenario == nil {
			if c.Background != nil {
				continue
			}
			return 0, 0, fmt.Errorf("unknown err, nil scenario and nil background %v", c)
		}
		if ignoreTagFunc(c.Scenario.Tags) {
			continue
		}

		scenarioCount := 1
		matchedTier := tagsMatchTier(c.Scenario.Tags, tier)

		if len(c.Scenario.Examples) > 0 {
			// mostly we only care about 1st Examples node in a scenario
			example := c.Scenario.Examples[0]
			if len(example.TableBody) > 0 {
				scenarioCount = len(example.TableBody)
			}
		}
		if matchedTier || globalMatchedTier {
			matched += scenarioCount
		}

		total += scenarioCount
	}
	return total, matched, nil
}

// if user assign @critical and @blocker for a scenario, it is considered @blocker
func max(ts []Tier) Tier {
	if len(ts) == 1 {
		return ts[0]
	}
	max := ts[0]
	for _, t := range ts[1:] {
		if t > max {
			max = t
		}
	}
	return max
}

func tagsMatchTier(tags []*messages.Tag, lookupTier Tier) bool {
	tiers := sliceutils.Map(tags, func(t *messages.Tag) Tier {
		tier, exist := TierMap[t.Name[1:]]
		if !exist {
			return TierUnknown
		}
		return tier
	})
	validTiers := sliceutils.Filter(tiers, func(tag Tier) bool {
		return tag != TierUnknown
	})

	var tierInTags Tier
	switch len(validTiers) {
	case 1:
		tierInTags = validTiers[0]
	case 0:
		tierInTags = TierDefault
	default:
		tierInTags = max(validTiers)
	}

	return lookupTier == tierInTags
}

// func checkCoverage(all int, coverage float64, tier Tier) (tooFew, tooMany bool, msg string) {
// 	if tier == TierCritical {
// 		switch {
// 		case all < 10:
// 			return coverage < 50.0, false, "50..100"
// 		default:
// 			return coverage < 10.0, coverage > 20.0, "10..20"
// 		}
// 	}
// 	if tier == TierBlocker {
// 		switch {
// 		case all < 10:
// 			return coverage < 50.0, false, "50..100"
// 		default:
// 			return coverage < 5.0, coverage > 10.0, "5..10"
// 		}
// 	}
// 	panic(fmt.Sprintf("tier %d does not need coverage check", tier))
// }

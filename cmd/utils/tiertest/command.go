package tiertest

import (
	"bufio"
	"fmt"
	"os"
	"regexp"

	"github.com/manabie-com/backend/internal/golibs"

	"github.com/spf13/cobra"
)

var (
	dir  string
	tier string
)

func init() {
	RootCmd.PersistentFlags().StringVar(&dir, "dir", "", "directory of features file")
	RootCmd.PersistentFlags().StringVar(&tier, "tier", "", "tier to calculate")
	RootCmd.AddCommand(PrintFailTestCmd)
}

// RootCmd for coverage
var RootCmd = &cobra.Command{
	Use:   "tiertest [command]",
	Short: "check folder features by tier",
	RunE:  countAndCheckTierCoverage,
}
var PrintFailTestCmd = &cobra.Command{
	Use:   "print-fail-test [command]",
	Short: "Print failed test from gandalf log file",
	RunE:  printFailingTest,
}

func countAndCheckTierCoverage(cmd *cobra.Command, args []string) error {
	lookupTier, exist := TierMap[tier]
	if !exist {
		return fmt.Errorf("invalid tier %s", tier)
	}

	total, matched, err := countScenariosInFolderWithTier(dir, lookupTier)
	if err != nil {
		return err
	}
	if total == 0 {
		return nil
	}
	cov := float64(matched) * 100 / float64(total)
	fmt.Printf("Folder %s has %f coverage for tier %s, %d total scenario, recommended coverage %s\n", dir, cov, tier, total, getTierRecommended(lookupTier))
	return nil
}

func getTierRecommended(t Tier) string {
	switch t {
	case TierCritical:
		return "~10"
	case TierBlocker:
		return "~5"
	}
	panic(fmt.Sprintf("unknown tier %d", t))
}

func printFailingTest(cmd *cobra.Command, args []string) error {
	reg := regexp.MustCompile(`.*(Scenario\:|Scenario Outline\:).*# (.*).feature:([0-9]*).*`)
	file, err := os.Open(args[0])
	if err != nil {
		panic(err)
	}
	scanner := bufio.NewScanner(file)
	pos := map[string][]string{}
	for scanner.Scan() {
		line := scanner.Text()
		if reg.Match([]byte(line)) {
			matches := reg.FindStringSubmatch(line)
			filename := matches[2]
			line := matches[3]
			pos[filename] = append(pos[filename], line)
		}
	}
	for fn := range pos {
		uniqueline := golibs.GetUniqueElementStringArray(pos[fn])
		fmt.Printf("%s %v\n", fn, uniqueline)
	}
	return nil
}

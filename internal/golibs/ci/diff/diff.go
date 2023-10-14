package diff

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/execwrapper"
	"github.com/manabie-com/backend/internal/golibs/logger"

	"gopkg.in/yaml.v3"
)

type Differ struct {
	Force      bool
	PRDesc     string
	PRDescOnly bool
	BaseRef    string
	HeadRef    string
	ConfigPath string
	OutputPath string

	// Squads is the list of squad names that the github.actor belongs to.
	// This allows further customization by using diffRule.EnabledFor and diffRule.DisableFor.
	Squads []string

	// To simplify squad-based trigger logic, we only process with the first valid squad name
	// from Squads. A nil value means no squad-based logic will be triggered.
	targetSquad *string

	outputFile io.Writer
}

// Output is similar to Run, but returns the output instead of
// writing it to --output-path.
func (d *Differ) Output() (string, error) {
	out := &bytes.Buffer{}
	clonedDiffer := *d // clone the Differ so that the original outputFile is not affected.
	clonedDiffer.outputFile = out
	if err := clonedDiffer.Run(); err != nil {
		return "", err
	}
	return out.String(), nil
}

// Run parses the options and determines the requirements for tests.
// It then writes the output to Differ.outputFile.
func (d *Differ) Run() error {
	if err := d.parseOptions(); err != nil {
		return err
	}
	rules, err := d.getConfig()
	if err != nil {
		return err
	}

	// force all requirements
	if d.Force {
		logger.Infof("running in --force mode")
		return rules.forceValues().output(d.outputFile)
	}

	result, err := rules.resultFromPRDesc2(d.PRDesc)
	if err != nil {
		return err
	}

	if d.PRDescOnly {
		logger.Infof("running in --pr-desc-only mode")
		result.setPRDescOnlyMode()
		return result.output(d.outputFile)
	}

	logger.Infof("running in normal mode: extracting requirements from `git diff` and PR description")
	changedFiles, err := execwrapper.GitDiff(d.BaseRef, d.HeadRef)
	if err != nil {
		return fmt.Errorf("failed to run `git diff`: %s", err)
	}
	logger.Debugf("`git diff` returned:\n\t%v", strings.Join(changedFiles, "\n\t"))
	gdResult, err := rules.resultFromFileChanges(changedFiles, d.targetSquad)
	if err != nil {
		return err
	}

	result.combine(*gdResult)
	return result.output(d.outputFile)
}

var (
	errMissingAll        = errors.New("at least one of --pr-desc, --force, or --base-ref/--head-ref must be set")
	errMissingGitHeadRef = errors.New("cannot specify --base-ref without --head-ref")
	errMissingGitBaseRef = errors.New("cannot specify --head-ref without --base-ref")
	errMissingPRDesc     = errors.New("cannot specify --pr-desc-only without --pr-desc")
)

func (d *Differ) parseOptions() error {
	if d.PRDescOnly && d.PRDesc == "" {
		return errMissingPRDesc
	}
	if d.BaseRef == "" && d.HeadRef == "" && d.PRDesc == "" && !d.Force {
		return errMissingAll
	}
	if d.BaseRef == "" && d.HeadRef != "" {
		return errMissingGitBaseRef
	}
	if d.BaseRef != "" && d.HeadRef == "" {
		return errMissingGitHeadRef
	}

	for _, v := range d.Squads {
		if strings.HasPrefix(v, "squad-") {
			if d.isSquadIgnored(v) || d.isSquadSuffixIgnored(v) {
				continue
			}
			v := v
			d.targetSquad = &v
			logger.Infof("squad to activate rules: %s", *d.targetSquad)
			break
		}
	}

	if d.outputFile == nil {
		d.outputFile = os.Stdout
		if d.OutputPath != "" {
			var err error
			d.outputFile, err = os.OpenFile(d.OutputPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
			if err != nil {
				return fmt.Errorf("failed to open output file: %s", err)
			}
		}
	}

	// decode d.PRDesc if it's in base64 format
	dst := make([]byte, base64.StdEncoding.EncodedLen(len(d.PRDesc)))
	n, err := base64.StdEncoding.Decode(dst, []byte(d.PRDesc))
	if err == nil { // decode successfully
		d.PRDesc = string(dst[:n])
	}
	return nil
}

// isSquadIgnored check if squad should never be considered to activate rules.
// Such squads are usually non-functional.
func (d Differ) isSquadIgnored(squad string) bool {
	ignoredSquads := []string{"squad-admin", "squad-release", "squad-ddd", "squad-data"}
	for _, v := range ignoredSquads {
		if v == squad {
			return true
		}
	}
	return false
}

// isSquadSuffixIgnored is similar to isSquadIgnored, but check squad's suffix instead.
func (d Differ) isSquadSuffixIgnored(squad string) bool {
	ignoredSquadsWithSuffix := []string{"-be", "-fe", "-me", "-red", "-green", "-blue", "-purple"}
	for _, v := range ignoredSquadsWithSuffix {
		if strings.HasSuffix(squad, v) {
			return true
		}
	}
	return false
}

func (d Differ) getConfig() (*ruleList, error) {
	data, err := os.ReadFile(d.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read diff config: %s", err)
	}

	dr := &ruleList{}
	if err := yaml.Unmarshal(data, dr); err != nil {
		return nil, fmt.Errorf("failed to parse diff config: %s", err)
	}

	// validate config
	for _, r := range dr.Rules {
		if len(r.EnabledSquads) > 0 && len(r.DisabledSquads) > 0 {
			return nil, fmt.Errorf(`"enabled_for" and "disabled_for" cannot be used together (found in rule %q)`, r.Name)
		}
	}
	return dr, nil
}

package tests

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/execwrapper"
	vr "github.com/manabie-com/backend/internal/golibs/variants"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

type GCPKMSMetadata struct {
	ResourceID string    `yaml:"resource_id"`
	CreatedAt  time.Time `yaml:"created_at"`
	Enc        string    `yaml:"enc"`
}

type SOPSMetadata struct {
	GCPKMS []*GCPKMSMetadata `yaml:"gcp_kms"`
}

type SOPSEncrypted struct {
	SOPSMetadata SOPSMetadata `yaml:"sops"`
}

// TestKMSPathInSOPSFiles ensures that secrets are encrypted using the correct GCP KMS key.
func TestKMSPathInSOPSFiles(t *testing.T) {
	t.Parallel()
	targetDir := execwrapper.Abs("deployments/helm/manabie-all-in-one")
	err := filepath.WalkDir(targetDir, func(fp string, d fs.DirEntry, err error) error {
		require.NoError(t, err, "filepath.WalkDir failed")
		if d.IsDir() {
			return nil
		}
		if strings.Contains(fp, "yugabyte") {
			return nil
		}

		// extract p, e, s from filename
		ok, p, e, s := isSecretFile(fp)
		if !ok {
			return nil
		}

		actualKMSPath, err := extractKMSPathInFile(fp)
		if err != nil {
			return fmt.Errorf("extractKMSPathInFile failed: %s", err)
		}
		expectedKMSPath := getExpectedKMSPath(p, e, s)
		require.Equal(t, expectedKMSPath, actualKMSPath)
		return nil
	})
	require.NoError(t, err)
}

// TestKMSPathInSOPSFiles_NATS is similar to TestKMSPathInSOPSFiles, but for NATS secrets.
func TestKMSPathInSOPSFiles_NATS(t *testing.T) {
	t.Parallel()

	svcNATSSecretPathRe := regexp.MustCompile(`([^/]+)/([^/]+)/([^/]+)_nats\.secrets\.encrypted\.env`)
	getPESFromSvcFilepath := func(fp string) (vr.P, vr.E, vr.S, bool) {
		m := svcNATSSecretPathRe.FindStringSubmatch(fp)
		if len(m) == 0 {
			return vr.PartnerNotDefined, vr.EnvNotDefined, vr.ServiceNotDefined, false
		}
		return vr.ToPartner(m[1]), vr.ToEnv(m[2]), vr.ToService(m[3]), true
	}

	natsSecretPathRe := regexp.MustCompile(`([^/]+)/([^/]+)/(?:nats\.secrets\.conf\.encrypted\.yaml|controller\.seed\.encrypted\.yaml|nats\.secrets\.encrypted\.env)`)
	getEnvFromNATSFilepath := func(fp string) (vr.P, vr.E, bool) {
		m := natsSecretPathRe.FindStringSubmatch(fp)
		if len(m) == 0 {
			return vr.PartnerNotDefined, vr.EnvNotDefined, false
		}
		return vr.ToPartner(m[1]), vr.ToEnv(m[2]), true
	}

	testcase := func(t *testing.T, p vr.P, e vr.E) {
		natsSecretDir := execwrapper.Absf("deployments/helm/platforms/nats-jetstream/secrets/%v/%v", p, e)
		err := filepath.WalkDir(natsSecretDir, func(fp string, d fs.DirEntry, err error) error {
			require.NoError(t, err, "filepath.WalkDir failed")
			if d.IsDir() {
				return nil
			}
			if strings.HasSuffix(fp, "nats.secrets.conf.encrypted.yaml") {
				return nil // skip v1
			}

			p, e, ok := getEnvFromNATSFilepath(fp)
			if ok {
				t.Run(fp, func(t *testing.T) {
					actualKMSPath, err := extractKMSPathInFile(fp)
					require.NoError(t, err)
					expectedKMSPath := getExpectedKMSPath(p, e, vr.ServiceNATSJetstream)
					require.Equal(t, expectedKMSPath, actualKMSPath)
				})
				return nil
			}

			p, e, s, ok := getPESFromSvcFilepath(fp)
			if ok {
				t.Run(fp, func(t *testing.T) {
					actualKMSPath, err := extractKMSPathInFile(fp)
					require.NoError(t, err)
					expectedKMSPath := getExpectedKMSPath(p, e, s)
					require.Equal(t, expectedKMSPath, actualKMSPath)
				})
				return nil
			}

			return fmt.Errorf("unexpected secret filepath: %s", fp)
		})
		require.NoError(t, err)
	}
	vr.Iter(t).SkipE(vr.EnvPreproduction, vr.EnvProduction).IterPE(testcase)
}

// secretPathRe matches string like bob/secrets/aic/prod/bob.secrets.encrypted.yaml.
// It is meant to match filenames of encrypted secret files.
var secretPathRe = regexp.MustCompile(`([^/]+)/secrets/([^/]+)/([^/]+)/([^/_]+)\.secrets\.encrypted\.(?:yaml|yml)$`)

// isSecretFile returns true if fp is a path to an encrypted secret file.
// If yes, it also returns the name of partner, environment, and service, in that order.
func isSecretFile(fp string) (bool, vr.P, vr.E, vr.S) {
	m := secretPathRe.FindStringSubmatch(fp)
	if m == nil {
		return false, vr.PartnerNotDefined, vr.EnvNotDefined, vr.ServiceNotDefined
	}
	return true, vr.ToPartner(m[2]), vr.ToEnv(m[3]), vr.ToService(m[1])
}

func extractKMSPathInFile(fp string) (string, error) {
	if strings.HasSuffix(fp, ".env") {
		data, err := godotenv.Read(fp)
		if err != nil {
			return "", fmt.Errorf("godotenv.Read: %s", err)
		}
		gcpkms, ok := data["sops_gcp_kms__list_0__map_resource_id"]
		if !ok {
			return "", fmt.Errorf(`field "sops_gcp_kms__list_0__map_resource_id" does not exist in env file`)
		}
		return gcpkms, nil
	}

	filecontent, err := os.ReadFile(fp)
	if err != nil {
		return "", fmt.Errorf("os.ReadFile failed: %s", err)
	}

	obj := SOPSEncrypted{}
	if err = yaml.Unmarshal(filecontent, &obj); err != nil {
		return "", fmt.Errorf("yaml.Unmarshal failed: %s", err)
	}
	if len(obj.SOPSMetadata.GCPKMS) == 0 {
		return "", fmt.Errorf(`"gcp_kms" cannot be empty`)
	}
	return obj.SOPSMetadata.GCPKMS[0].ResourceID, nil
}

func getExpectedKMSPath(p vr.P, e vr.E, s vr.S) string {
	project := ""
	location := ""
	keyRings := "backend-services"
	switch e {
	case vr.EnvLocal:
		return "projects/dev-manabie-online/locations/global/keyRings/deployments/cryptoKeys/github-actions"
	case vr.EnvStaging, vr.EnvUAT:
		project = "staging-manabie-online"
		location = "global"
	case vr.EnvProduction, vr.EnvPreproduction:
		project = "student-coach-e1e95"
		location = "asia-northeast1"
	default:
		panic(fmt.Errorf("invalid environment: %v", e))
	}
	// preproduction use kms key same with production
	if e == vr.EnvPreproduction {
		e = vr.EnvProduction
	}
	return fmt.Sprintf("projects/%v/locations/%v/keyRings/%v/cryptoKeys/%v-%v",
		project, location, keyRings, e, s)
}

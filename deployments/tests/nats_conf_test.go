package tests

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/execwrapper"
	skaffoldwrapper "github.com/manabie-com/backend/internal/golibs/execwrapper/skaffold"
	vr "github.com/manabie-com/backend/internal/golibs/variants"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
)

// TestNATSConf ensures that nats.conf will be processed successfully in actual deployments.
//
// For now, it checks that all variables in the config have been properly declared.
func TestNATSConf(t *testing.T) {
	t.Parallel()

	extractNATSConfigMap := func(manifestObjects []interface{}) (string, error) {
		for _, o := range manifestObjects {
			switch v := o.(type) {
			case *corev1.ConfigMap:
				if v.ObjectMeta.Name != "nats-jetstream" {
					continue
				}
				data, ok := v.Data["nats.conf"]
				if !ok {
					return "", errors.New(`failed to find "nats.conf" data inside "nats-jetstream" configmap`)
				}
				return data, nil
			}
		}
		return "", errors.New(`failed to find any "nats-jetstream" configmap`)
	}

	// getAvailableVars returns the list of variables that is provided to nats.conf
	getAvailableVars := func(p vr.P, e vr.E) ([]string, error) {
		dir := execwrapper.Absf("deployments/helm/platforms/nats-jetstream/secrets/%v/%v", p, e)
		res := make([]string, 0, 20)
		err := filepath.WalkDir(dir, func(fp string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			// only applicable to v2 files
			if !strings.HasSuffix(fp, ".secrets.encrypted.env") {
				return nil
			}

			data, err := godotenv.Read(fp)
			if err != nil {
				return fmt.Errorf("godotenv.Read: %s", err)
			}
			for k := range data { // append found keys to the output
				if strings.HasPrefix(k, "sops_") {
					continue
				}
				res = append(res, k)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}

		// these variables are inherited from environment, not configs
		// but only in non-local (where cluster mode is enabled)
		if e != vr.EnvLocal {
			res = append(res, "CLUSTER_ADVERTISE")
		}
		return res, nil
	}

	testfunc := func(t *testing.T, p vr.P, e vr.E) {
		manifestObjects, err := skaffoldwrapper.New().
			E(e).P(p).Filename("skaffold.backbone.yaml").Profile("nats-only").
			CachedRender()
		require.NoError(t, err)

		data, err := extractNATSConfigMap(manifestObjects)
		require.NoError(t, err)

		natsVars := getVariablesInNATSConf(data)
		avaiableNATSVars, err := getAvailableVars(p, e)
		require.NoError(t, err)
		for _, v := range natsVars {
			require.Contains(t, avaiableNATSVars, v, "key %q is not available (have you added secret for that service in nats secret?)", v)
		}
		for _, v := range avaiableNATSVars {
			require.Contains(t, natsVars, v, "variable %q exists but is not referenced in nats.conf", v)
		}
	}

	vr.Iter(t).SkipE(vr.EnvPreproduction).IterPE(testfunc)
}

// TestGetVariablesInNATSConf tests getVariablesInNATSConf function.
func TestGetVariablesInNATSConf(t *testing.T) {
	t.Parallel()

	actual := getVariablesInNATSConf(`
cluster {
	cluster_advertise: $CLUSTER_ADVERTISE
}
accounts {
	A: {
		jetstream: enabled,
		users: [
			{
				nkey: $controller_nkey
			},
			{
				user: "Bob",
				password: $bob_password,
				permissions: {
					publish: {
						allow: [
							"$JS.API.INFO", 
							"$JS.API.STREAM.>", 
							"$JS.API.CONSUMER.*", 
						]
					},
					subscribe: {
						allow: [
							"_INBOX.>",
							"deliver.>"
						]
					}
				}
			},
			{
				user: "Enigma",
				password: $enigma_password,
			},
			{
				user: "Yasuo",
				password: $yasuo_password,
			},
		]
	}
}`)
	expected := []string{"CLUSTER_ADVERTISE", "controller_nkey", "bob_password", "enigma_password", "yasuo_password"}
	require.Equal(t, expected, actual)
}

// getVariablesInNATSConf returns the list of required variables in nats.conf.
func getVariablesInNATSConf(data string) []string {
	re := regexp.MustCompile(`:\s*\$([a-zA-Z\_]+)`)
	m := re.FindAllStringSubmatch(data, -1)
	res := make([]string, 0, len(m))
	for _, v := range m {
		res = append(res, v[1])
	}
	return res
}

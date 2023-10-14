package env

import (
	"fmt"
	"os"
)

// Vars contains the list of user-specific environment variables.
type Vars map[string]string

// ToEnv returns a list of environment variables in key=value format.
// It is intended to be supplied to exec.Command.Env.
func (ev Vars) ToEnv() []string {
	res := []string{}
	for k, v := range ev {
		res = append(res, fmt.Sprintf("%s=%s", k, v))
	}
	res = append(res, os.Environ()...) // last value take precedence
	return res
}

// Default returns the default Vars used in local/CI deployments.
func Default() Vars {
	isCI, ok := os.LookupEnv("CI")
	if !ok {
		isCI = "false"
	}
	cpuLimit := "4"
	memoryLimit := "12g"
	if isCI == "true" {
		cpuLimit = "max"
		memoryLimit = "max"
	}

	useSharedRegistry, ok := os.LookupEnv("USE_SHARED_REGISTRY")
	if !ok {
		useSharedRegistry = "false"
	}
	localRegistryDomain := "localhost:5001"
	artifactRegistryDomain := "localhost:5001"
	if useSharedRegistry == "true" {
		localRegistryDomain = "kind-reg.actions-runner-system.svc"
		artifactRegistryDomain = "asia-southeast1-docker.pkg.dev/student-coach-e1e95/ci"
	}

	return map[string]string{
		"DECRYPT_KEY":                    "9ef85f8fcde4139b88bbbfe5",
		"CPU_LIMIT":                      cpuLimit,
		"MEMORY_LIMIT":                   memoryLimit,
		"ENV":                            "local",
		"ORG":                            "manabie",
		"IMG":                            "asia.gcr.io/student-coach-e1e95/backend",
		"TAG":                            "locally",
		"NAMESPACE":                      "backend",
		"BACKOFFICE_TAG":                 "locally",
		"LEARNER_TAG":                    "locally",
		"TEACHER_TAG":                    "locally",
		"INSTALL_MONITORING_STACKS":      "$false",
		"ELASTIC_NAMESPACE":              "elastic",
		"ELASTIC_RELEASE_NAME":           "elastic",
		"ELASTIC_NAME_OVERRIDE":          "",
		"ELASTIC_CREATE_SERVICE_ACCOUNT": "true",
		"ELASTIC_INIT_INDICES":           "true",
		"APHELIOS_DEPLOYMENT_ENABLED":    "false",
		"REDASH_DEPLOYMENT_ENABLED":      "false",
		"APPSMITH_DEPLOYMENT_ENABLED":    "false",
		"SERVICE_ACCOUNT_EMAIL_SUFFIX":   "",
		"DISABLE_GATEWAY":                "false",
		"LOCAL_REGISTRY_DOMAIN":          localRegistryDomain,
		"ARTIFACT_REGISTRY_DOMAIN":       artifactRegistryDomain,
	}
}

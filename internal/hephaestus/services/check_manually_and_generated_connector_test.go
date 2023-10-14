package services

import (
	"os"
	"path"
	"regexp"
	"sort"
	"strings"
	"testing"

	dplparser "github.com/manabie-com/backend/cmd/utils/data_pipeline_parser"
	"github.com/manabie-com/backend/internal/golibs"
)

var manuallyConnectorDir = "../../../deployments/helm/manabie-all-in-one/charts/hephaestus/connectors/sink"
var dataPipelineDef = "../../../deployments/helm/platforms/kafka-connect/postgresql2postgresql"

var envList = []string{"local", "stag", "uat", "prod"}
var orgList = []string{"e2e", "manabie", "jprep", "aic", "ga", "renseikai", "synersia", "tokyo"}

var skipFiles = []string{
	"bob_to_bob_user_basic_info.json",
	"bob_to_entryexitmgmt_student_entryexit_records.json",
	"bob_to_entryexitmgmt_student_qr.json",
	"calendar_to_bob_date_info.json",
	"fatima_to_elastic_order.json",
	"fatima_to_elastic_order_item.json",
	"fatima_to_elastic_product.json",
}

func TestCheckManuallyConnectorDefineEnvAndOrgMatchGenerateDataPipeline(t *testing.T) {
	es, err := os.ReadDir(manuallyConnectorDir)
	if err != nil {
		t.Error(err)
		return
	}

	des, err := os.ReadDir(dataPipelineDef)
	if err != nil {
		t.Error(err)
		return
	}

	mp := make(map[string]*dplparser.SinkConfig)
	for _, e := range des {
		dpl, err := dplparser.NewDataPipelineParser(path.Join(dataPipelineDef, e.Name()))
		if err != nil {
			t.Error(err)
			return
		}
		for _, pl := range dpl.DataPipelineDef.Datapipelines {
			for _, sink := range pl.Sinks {
				mp[sink.FileName] = sink
			}
		}
	}

	for _, e := range es {
		fileName := e.Name()
		if isSkippedFiles(fileName, skipFiles) {
			continue
		}
		b, err := os.ReadFile(path.Join(manuallyConnectorDir, fileName))
		if err != nil {
			t.Log(err)
			return
		}

		config := string(b)
		envs := getEnvOfManualConnector(config)
		orgs := getOrgOfManualConnector(config, envs)

		sink, ok := mp[fileName]
		if !ok {
			t.Logf("cannot find file %s", fileName)
			continue
		}

		sink.DeployEnvs = removeDorpEnv(sink.DeployEnvs)
		sort.Strings(sink.DeployEnvs)
		sort.Strings(envs)
		if len(sink.DeployEnvs) != len(envs) || !golibs.EqualStringArray(sink.DeployEnvs, envs) {
			t.Errorf("pipeline %s deploy envs not match %s expected %s", sink.FileName, sink.DeployEnvs, envs)
		}

		sort.Strings(sink.DeployOrgs)
		sort.Strings(orgs)
		if len(sink.DeployOrgs) != len(orgs) || !golibs.EqualStringArray(sink.DeployOrgs, orgs) {
			t.Errorf("pipeline %s deploy orgs not match %s expected %s", sink.FileName, sink.DeployOrgs, orgs)
			// return
		}
	}
}

func UpdateOrgsWhenEnvsIsOnlyStaging(orgs, envs []string) []string {
	if len(orgs) == 0 {
		sort.Strings(envs)
		if len(envs) == 2 && envs[0] == "local" && envs[1] == "stag" {
			orgs = []string{"manabie", "jprep"}
		}
		if len(envs) == 3 && envs[0] == "local" && envs[1] == "stag" && envs[2] == "uat" {
			orgs = []string{"manabie", "jprep"}
		}
	}
	return orgs
}

func removeDorpEnv(envs []string) []string {
	// remove dorp env
	for i := 0; i < len(envs); i++ {
		if envs[i] == "dorp" {
			envs = append(envs[:i], envs[i+1:]...)
			break
		}
	}

	return envs
}

func getEnvs(config string) []string {
	envRaw := strings.Split(config, "\n")[0]
	rg := regexp.MustCompile(`eq "(\w+)+" .Values.global.environment`)
	deployEnvs := make([]string, 0)
	for _, res := range rg.FindAllStringSubmatch(envRaw, -1) {
		if len(res) > 1 {
			deployEnvs = append(deployEnvs, res[1])
		}
	}
	return deployEnvs
}
func getOrgs(config string) []string {
	deployOrgs := make([]string, 0)
	for _, line := range strings.Split(config, "\n") {
		orgRaw := line
		rg := regexp.MustCompile(`eq "(\w+)+" .Values.global.vendor`)
		for _, res := range rg.FindAllStringSubmatch(orgRaw, -1) {
			if len(res) > 1 {
				deployOrgs = append(deployOrgs, res[1])
			}
		}
		if len(deployOrgs) > 0 {
			break
		}
	}
	return deployOrgs
}

func getExceptEnvs(config string) []string {
	envRaw := strings.Split(config, "\n")[0]
	rg := regexp.MustCompile(`ne "(\w+)+" .Values.global.environment`)
	exceptEnv := make([]string, 0)
	for _, res := range rg.FindAllStringSubmatch(envRaw, -1) {
		if len(res) > 1 {
			exceptEnv = append(exceptEnv, res[1])
		}
	}
	return exceptEnv
}

func getExceptOrgs(config string) []string {
	orgRaw := strings.Split(config, "\n")[0]
	rg := regexp.MustCompile(`ne "(\w+)+" .Values.global.vendor`)
	exceptOrgs := make([]string, 0)
	for _, res := range rg.FindAllStringSubmatch(orgRaw, -1) {
		if len(res) > 1 {
			exceptOrgs = append(exceptOrgs, res[1])
		}
	}
	return exceptOrgs
}

func getEnvOfManualConnector(config string) []string {
	envs := getEnvs(string(config))
	notEnvs := getExceptEnvs(string(config))
	if len(notEnvs) > 0 {
		for i := 0; i < len(envList); i++ {
			if !golibs.InArrayString(envList[i], notEnvs) {
				envs = append(envs, envList[i])
			}
		}
	}

	if len(envs) == 0 {
		// means all environments
		envs = envList
	}
	return envs
}

func getOrgOfManualConnector(config string, envs []string) []string {
	orgs := getOrgs(config)
	notOrgs := getExceptOrgs(config)
	if len(notOrgs) > 0 {
		for i := 0; i < len(orgList); i++ {
			if !golibs.InArrayString(orgList[i], notOrgs) {
				orgs = append(orgs, orgList[i])
			}
		}
	}

	orgs = UpdateOrgsWhenEnvsIsOnlyStaging(orgs, envs)

	if len(orgs) == 0 {
		// means all orgs
		orgs = orgList
	}
	return orgs
}

func isSkippedFiles(fileName string, skipFiles []string) bool {
	if golibs.InArrayString(fileName, skipFiles) {
		return true
	}
	return false
}

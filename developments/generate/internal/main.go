package internal

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/manabie-com/backend/developments/generate/configs"
	"github.com/manabie-com/backend/developments/generate/internal/template"
	"github.com/manabie-com/backend/developments/generate/internal/utils"

	stringy "github.com/gobeam/stringy"
)

var (
	regexMatchStep       = regexp.MustCompile(`steps := map\[string]interface{}{[\S\s]*?}`)
	regexMatchExpr       = regexp.MustCompile("\\^[\\S\\s]*?\\`")
	regexMatchService    = regexp.MustCompile(`service .*{`)
	regexMatchAllService = regexp.MustCompile(`service[\S\s]*?}`)

	eureka   = "eureka"
	syllabus = "syllabus"
)

func RunGenAll(cfg *configs.Config) {
	var err error
	var service string
	baseDir, _ := os.Getwd()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("No caller information")
	}
	pkgDir := path.Dir(filename)
	serviceCamel := stringy.New(cfg.ChildService).CamelCase()
	funcCamel := stringy.New(strings.Join([]string{cfg.Method, cfg.Entity}, "_")).CamelCase()

	var bddFilePath string
	switch cfg.Service {
	case eureka:
		bddFilePath = fmt.Sprintf("%s/features/%s/bdd.go", baseDir, cfg.Service)
	case syllabus:
		bddFilePath = fmt.Sprintf("%s/features/%s/%s/steps.go", baseDir, cfg.Service, cfg.FolderEntity)
	default:
		bddFilePath = fmt.Sprintf("%s/features/%s/bdd_steps.go", baseDir, cfg.Service)
	}
	// process bdd
	if cfg.Service == syllabus {
		RunGenBddV2(cfg)
		service = eureka
	} else {
		runProcessBdd(cfg, bddFilePath, pkgDir, baseDir)
		service = cfg.Service
	}

	// process proto
	protoFilePath := fmt.Sprintf("%s/proto/%s/v1/%s.proto", baseDir, service, cfg.ChildService)

	var protoContent string
	briefService := string([]rune(cfg.Service)[0]) + "pb"
	b, err := os.ReadFile(protoFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			protoContent = fmt.Sprintf(template.AllProtoTemplate,
				cfg.Service,
				briefService,
				fmt.Sprintf(template.ProtoTemplate, funcCamel),
				serviceCamel,
				fmt.Sprintf(template.ProtoFuncTemplate, funcCamel),
			)
		} else {
			panic(err)
		}
	} else {
		protoContent = string(b)
		matchService := string(regexMatchService.Find(b))
		matchAllService := string(regexMatchAllService.Find(b))
		protoContent = strings.ReplaceAll(protoContent, matchService, fmt.Sprintf(template.ProtoTemplate, funcCamel)+matchService)
		protoContent = strings.ReplaceAll(
			protoContent,
			matchAllService,
			matchAllService[:len(matchAllService)-1]+fmt.Sprintf(template.ProtoFuncTemplate, funcCamel),
		)
	}
	if err := ioutil.WriteFile(protoFilePath, []byte(protoContent), fs.ModePerm); err != nil {
		panic(err)
	}

	// process + mock service
	serviceFilePath := fmt.Sprintf("%s/internal/%s/services/%s.go", baseDir, service, cfg.ChildService)
	var serviceContent string
	if _, err := os.ReadFile(serviceFilePath); err != nil {
		if !os.IsNotExist(err) {
			panic(err)
		}
		funcContent := fmt.Sprintf(template.ServiceFuncTemplate, serviceCamel, funcCamel, briefService)
		serviceContent = fmt.Sprintf(template.AllServiceTemplate, service, briefService, funcContent, serviceCamel)
	} else {
		serviceContent = fmt.Sprintf(template.ServiceFuncTemplate, serviceCamel, funcCamel, briefService)
	}
	f, err := os.OpenFile(serviceFilePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, fs.ModePerm)
	if err != nil {
		panic(err)
	}
	if _, err := f.WriteString(serviceContent); err != nil {
		panic(err)
	}
	defer f.Close()
}

func RunGenBdd(cfg *configs.Config) {
	baseDir, _ := os.Getwd()

	var bddFilePath string
	switch cfg.Service {
	case eureka:
		bddFilePath = fmt.Sprintf("%s/features/%s/bdd.go", baseDir, cfg.Service)
	default:
		bddFilePath = fmt.Sprintf("%s/features/%s/bdd_steps.go", baseDir, cfg.Service)
	}

	// process bdd
	b, err := os.ReadFile(bddFilePath)
	if err != nil {
		panic(err)
	}
	bddDeclaration := string(regexMatchStep.Find(b))
	m := make(map[string]bool)
	exprStrs := regexMatchExpr.FindAllString(bddDeclaration, -1)
	for _, exprStr := range exprStrs {
		expr := exprStr[:len(exprStr)-1]
		m[expr] = true
	}
	bddFileContent := string(b)
	files, err := os.ReadDir(fmt.Sprintf("%s/features/%s", baseDir, cfg.Service))
	if err != nil {
		panic(err)
	}
	var stepTexts = []string{}
	for _, file := range files {
		if filepath.Ext(file.Name()) == utils.FeatureExt {
			stepTexts = utils.WriteSuiteFiles(
				fmt.Sprintf("%s/features/%s/%s", baseDir, cfg.Service, file.Name()),
				cfg.Service,
				m,
				stepTexts,
			)
		}
	}

	newBddDeclaration := bddDeclaration[:len(bddDeclaration)-1] + strings.Join(stepTexts, "") + "}"
	bddContent := strings.ReplaceAll(bddFileContent, bddDeclaration, newBddDeclaration)

	if err := ioutil.WriteFile(bddFilePath, []byte(bddContent), fs.ModePerm); err != nil {
		panic(err)
	}
}

func RunGenBddV2(cfg *configs.Config) {
	baseDir, _ := os.Getwd()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("No caller information")
	}

	snakeString := strings.Join(append([]string{cfg.Method}, strings.Split(cfg.Entity, " ")...), "_")
	fileFeature := fmt.Sprintf("%s/features/%s/%s/%s.feature", baseDir, cfg.Service, cfg.FolderEntity, snakeString)
	if _, err := os.ReadFile(fileFeature); err == nil {
		fmt.Println("file already exists")
		return
	}
	pkgDir := path.Dir(filename)
	var bddFilePath string
	switch cfg.Service {
	case eureka:
		bddFilePath = fmt.Sprintf("%s/features/%s/bdd.go", baseDir, cfg.Service)
	case syllabus:
		bddFilePath = fmt.Sprintf("%s/features/%s/%s/steps.go", baseDir, cfg.Service, cfg.FolderEntity)
	default:
		bddFilePath = fmt.Sprintf("%s/features/%s/bdd_steps.go", baseDir, cfg.Service)
	}

	entityFilePath := fmt.Sprintf("%s/features/%s", baseDir, cfg.Service)
	files, err := ioutil.ReadDir(entityFilePath)
	folderEntityNames := make([]string, 0, len(files))
	check := false
	if err != nil {
		panic(err)
	}
	for _, f := range files {
		if f.IsDir() {
			folderEntityNames = append(folderEntityNames, f.Name())
		}
	}
	for _, folder := range folderEntityNames {
		if cfg.FolderEntity == folder {
			check = true
		}
	}
	if !check {
		if err := os.MkdirAll(entityFilePath+"/"+cfg.FolderEntity, os.ModePerm); err != nil {
			panic(err)
		}
		if _, err := os.Create(bddFilePath); err != nil {
			panic(err)
		}
		if _, err := os.Create(fmt.Sprintf("%s/features/%s/%s/bdd.go", baseDir, cfg.Service, cfg.FolderEntity)); err != nil {
			panic(err)
		}
		utils.WriteFileBdd(fmt.Sprintf("%s/features/%s/%s/bdd.go", baseDir, cfg.Service, cfg.FolderEntity), cfg.FolderEntity)
		utils.WriteFileSteps(fmt.Sprintf("%s/features/%s/%s/steps.go", baseDir, cfg.Service, cfg.FolderEntity), cfg.FolderEntity)
		InitStepSyllabus(bddFilePath, cfg.FolderEntity)
		fmt.Println(" please add initsteps to the corresponding *.go file ")
	}

	b, err := os.ReadFile(bddFilePath)
	if err != nil {
		panic(err)
	}
	bddDeclaration := string(regexMatchStep.Find(b))
	m := make(map[string]bool)
	exprStrs := regexMatchExpr.FindAllString(bddDeclaration, -1)
	for _, exprStr := range exprStrs {
		expr := exprStr[:len(exprStr)-1]
		m[expr] = true
	}
	bddFileContent := string(b)

	b, _ = os.ReadFile(fmt.Sprintf("%s/template/%s/%s.feature", pkgDir, cfg.Service, cfg.Method))
	str := fmt.Sprintf(string(b), cfg.Entity, cfg.FolderEntity)
	snakeStr := strings.Join(append([]string{cfg.Method}, strings.Split(cfg.Entity, " ")...), "_")
	newFileFeature := fmt.Sprintf("%s/features/%s/%s/%s.feature", baseDir, cfg.Service, cfg.FolderEntity, snakeStr)
	if err := ioutil.WriteFile(newFileFeature, []byte(str), fs.ModePerm); err != nil {
		panic(err)
	}
	stepTexts := []string{}
	stepTexts = utils.WriteSuiteFilesSyllabus(newFileFeature, cfg.FolderEntity, m, stepTexts)
	newBddDeclaration := bddDeclaration[:len(bddDeclaration)-1] + strings.Join(stepTexts, "") + "}"
	bddContent := strings.ReplaceAll(bddFileContent, bddDeclaration, newBddDeclaration)
	if err := ioutil.WriteFile(bddFilePath, []byte(bddContent), fs.ModePerm); err != nil {
		panic(err)
	}
}

func runProcessBdd(cfg *configs.Config, bddFilePath, pkgDir, baseDir string) {
	b, err := os.ReadFile(bddFilePath)
	if err != nil {
		panic(err)
	}
	bddDeclaration := string(regexMatchStep.Find(b))
	m := make(map[string]bool)
	exprStrs := regexMatchExpr.FindAllString(bddDeclaration, -1)
	for _, exprStr := range exprStrs {
		expr := exprStr[:len(exprStr)-1]
		m[expr] = true
	}
	bddFileContent := string(b)

	b, _ = os.ReadFile(fmt.Sprintf("%s/template/%s/%s.feature", pkgDir, cfg.Service, cfg.Method))
	str := fmt.Sprintf(string(b), cfg.Entity)
	snakeStr := strings.Join(append([]string{cfg.Method}, strings.Split(cfg.Entity, " ")...), "_")
	newFileFeature := fmt.Sprintf("%s/features/%s/%s.feature", baseDir, cfg.Service, snakeStr)
	if err := ioutil.WriteFile(newFileFeature, []byte(str), fs.ModePerm); err != nil {
		panic(err)
	}
	stepTexts := []string{}

	stepTexts = utils.WriteSuiteFiles(newFileFeature, cfg.Service, m, stepTexts)
	newBddDeclaration := bddDeclaration[:len(bddDeclaration)-1] + strings.Join(stepTexts, "") + "}"
	bddContent := strings.ReplaceAll(bddFileContent, bddDeclaration, newBddDeclaration)

	if err := ioutil.WriteFile(bddFilePath, []byte(bddContent), fs.ModePerm); err != nil {
		panic(err)
	}
}

func InitStepSyllabus(bddFilePath, subFolder string) {
	b, err := os.ReadFile(bddFilePath)
	if err != nil {
		panic(err)
	}
	bddDeclaration := string(regexMatchStep.Find(b))
	m := make(map[string]bool)
	exprStrs := regexMatchExpr.FindAllString(bddDeclaration, -1)
	for _, exprStr := range exprStrs {
		expr := exprStr[:len(exprStr)-1]
		m[expr] = true
	}
	bddFileContent := string(b)
	stepTexts := []string{}
	stepTexts = append(stepTexts, "\n`"+"<"+subFolder+">"+"a signed in \"([^\"]*)\"$"+"`: s."+"aSignedIn"+",\n")
	stepTexts = append(stepTexts, "`"+"<"+subFolder+">"+"returns \"([^\"]*)\" status code$"+"`: s."+"returnsStatusCode"+",\n")
	newBddDeclaration := bddDeclaration[:len(bddDeclaration)-1] + strings.Join(stepTexts, "") + "}"
	bddContent := strings.ReplaceAll(bddFileContent, bddDeclaration, newBddDeclaration)
	if err := ioutil.WriteFile(bddFilePath, []byte(bddContent), fs.ModePerm); err != nil {
		panic(err)
	}
}

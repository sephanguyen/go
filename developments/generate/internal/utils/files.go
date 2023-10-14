package utils

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"

	"github.com/manabie-com/backend/developments/generate/internal/parser"
	"github.com/manabie-com/backend/developments/generate/internal/template"

	"github.com/cucumber/messages-go/v16"
)

const FeatureExt = ".feature"

func WriteSuiteFiles(filePath string, service string, stepMap map[string]bool, stepTexts []string) []string {
	suiteFileHeader := fmt.Sprintf(template.PackageTemplate, service)
	features, err := parser.ParseFeatures([]string{filePath})
	if err != nil {
		panic(err)
	}
	funcTexts := []string{}
	for _, f := range features {
		if f.GherkinDocument.Feature == nil {
			return stepTexts
		}
		for _, children := range f.GherkinDocument.Feature.Children {
			steps := []*messages.Step{}
			if children.Scenario != nil {
				steps = children.Scenario.Steps
			}
			if children.Background != nil {
				steps = append(steps, children.Background.Steps...)
			}
			for _, step := range steps {
				expr, name := GetExprAndFuncName(step.Text)
				if ok1, ok2 := stepMap[expr], stepMap[expr[:len(expr)-1]]; !ok1 && !ok2 {
					args := []string{"ctx context.Context"}
					argsStr := GetArgs(expr)
					if argsStr != "" {
						args = append(args, argsStr)
					}
					funcTexts = append(funcTexts, fmt.Sprintf(template.FuncTemplate, name, strings.Join(args, ",")))
					stepTexts = append(stepTexts, "`"+expr+"`: s."+name+",\n")
					stepMap[expr] = true
				}
			}
		}
	}
	if len(funcTexts) > 0 {
		suiteFilePath := filePath[:len(filePath)-len(FeatureExt)] + ".go"
		var content string
		if _, err := os.ReadFile(suiteFilePath); err != nil {
			if !os.IsNotExist(err) {
				panic(err)
			}
			content = suiteFileHeader + strings.Join(funcTexts, "")
		} else {
			content = strings.Join(funcTexts, "")
		}
		f, err := os.OpenFile(suiteFilePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, fs.ModePerm)
		if err != nil {
			panic(err)
		}
		if _, err := f.WriteString(content); err != nil {
			panic(err)
		}
		defer f.Close()
	}
	return stepTexts
}

func WriteSuiteFilesSyllabus(filePath string, folderEntity string, stepMap map[string]bool, stepTexts []string) []string {
	suiteFileHeader := fmt.Sprintf(template.SyllabusPackageTemplate, folderEntity)
	features, err := parser.ParseFeatures([]string{filePath})
	if err != nil {
		panic(err)
	}
	funcTexts := []string{}
	for _, f := range features {
		if f.GherkinDocument.Feature == nil {
			return stepTexts
		}
		for _, children := range f.GherkinDocument.Feature.Children {
			steps := []*messages.Step{}
			if children.Scenario != nil {
				steps = children.Scenario.Steps
			}
			if children.Background != nil {
				steps = append(steps, children.Background.Steps...)
			}
			for _, step := range steps {
				expr, name := GetExprAndFuncName(step.Text)
				if !strings.Contains(name, "aSignedIn") && !strings.Contains(name, "returnsStatusCode") {
					if ok1, ok2 := stepMap[expr], stepMap[expr[:len(expr)-1]]; !ok1 && !ok2 {
						args := []string{"ctx context.Context"}
						argsStr := GetArgs(expr)
						if argsStr != "" {
							args = append(args, argsStr)
						}
						funcTexts = append(funcTexts, fmt.Sprintf(template.SyllabusFuncTemplate, name, strings.Join(args, ",")))
						stepTexts = append(stepTexts, "`"+expr+"`: s."+name+",\n")
						stepMap[expr] = true
					}
				}
			}
		}
	}
	if len(funcTexts) > 0 {
		suiteFilePath := filePath[:len(filePath)-len(FeatureExt)] + ".go"
		var content string
		if _, err := os.ReadFile(suiteFilePath); err != nil {
			if !os.IsNotExist(err) {
				panic(err)
			}
			content = suiteFileHeader + strings.Join(funcTexts, "")
		} else {
			content = strings.Join(funcTexts, "")
		}
		f, err := os.OpenFile(suiteFilePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, fs.ModePerm)
		if err != nil {
			panic(err)
		}
		if _, err := f.WriteString(content); err != nil {
			panic(err)
		}
		defer f.Close()
	}
	return stepTexts
}

func WriteFileBdd(filePath, folderEntity string) []string {
	if _, err := os.ReadFile(filePath); err != nil {
		if !os.IsNotExist(err) {
			panic(err)
		}
	}
	funcTexts := []string{}
	funcTexts = append(funcTexts, fmt.Sprintf(template.SyllabusBddTemplate, folderEntity))
	content := strings.Join(funcTexts, "")

	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, fs.ModePerm)
	if err != nil {
		panic(err)
	}
	if _, err := f.WriteString(content); err != nil {
		panic(err)
	}
	defer f.Close()
	return funcTexts
}

func WriteFileSteps(filePath, folderEntity string) []string {
	if _, err := os.ReadFile(filePath); err != nil {
		if !os.IsNotExist(err) {
			panic(err)
		}
	}
	funcTexts := []string{}
	funcTexts = append(funcTexts, fmt.Sprintf(template.SyllabusStepsTemplate, folderEntity))
	content := strings.Join(funcTexts, "")

	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, fs.ModePerm)
	if err != nil {
		panic(err)
	}
	if _, err := f.WriteString(content); err != nil {
		panic(err)
	}
	defer f.Close()
	return funcTexts
}

func IsEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err // Either not empty or error, suits both cases
}

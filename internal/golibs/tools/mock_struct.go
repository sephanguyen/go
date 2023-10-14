package tools

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"regexp"
	"strings"
	"text/template"

	"golang.org/x/tools/imports"
)

// GenMockStructs works identically to MockRepository, but with less input arguments.
// A key of input m is the path to a package. Values of m is the list of structs in such path.
func GenMockStructs(m map[string][]interface{}) error {
	for path, names := range m {
		for _, name := range names {
			err := genMockStruct(path, name)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

const mockPkgPrefix = "mock_"

func genMockStruct(path string, targetStruct interface{}) error {
	if !strings.HasPrefix(path, "internal/") {
		return fmt.Errorf("only paths starting with \"internal/\" is allowed (got %q)", path)
	}

	// Get necessary names for the generation
	pkgName, structName := pkgNameAndStructName(targetStruct)
	mockPkgName := mockPkgPrefix + pkgName
	serviceName := serviceNameFrom(path)

	// Generate the mock content
	builder := generateMockContent(targetStruct, mockPkgName, serviceName)
	content, err := imports.Process("", []byte(builder.String()), &imports.Options{Comments: true})
	if err != nil {
		return fmt.Errorf("failed to process %s: %v", path, err)
	}

	// Write generated content to file
	outDir := mockPathFrom(path)
	err = os.MkdirAll(outDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create directory %s: %v", outDir, err)
	}

	outFileName := outDir + "/" + snakecase(structName) + ".go"
	err = os.WriteFile(outFileName, content, 0o644) //nolint:gosec
	if err != nil {
		return fmt.Errorf("failed to write %s: %v", outFileName, err)
	}
	log.Printf("Generated mock at %s", outFileName)
	return nil
}

func pkgNameAndStructName(v interface{}) (string, string) {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	names := strings.Split(t.String(), ".")
	if len(names) != 2 {
		panic("failed to extract pkgname and structname of " + t.String())
	}
	return names[0], names[1]
}

// serviceNameFrom assumes that path has the form `internal/<serviceName>/...`.
func serviceNameFrom(path string) string {
	if !strings.HasPrefix(path, "internal/") {
		panic(fmt.Errorf("path %q is malformed, expected \"internal/...\"", path))
	}
	return strings.SplitN(path, "/", 3)[1]
}

// mockPathFrom assumes that path has the form `internal/<serviceName>/...`.
func mockPathFrom(path string) string {
	if !strings.HasPrefix(path, "internal/") {
		panic(fmt.Errorf("path %q is malformed, expected \"internal/...\"", path))
	}
	return strings.Replace(path, "internal/", "mock/", 1)
}

// referenced from github.com/vektra/mockery, which in turn taken
// from http://stackoverflow.com/questions/1175208/elegant-python-function-to-convert-camelcase-to-camel-caseo
func snakecase(caseName string) string {
	rxp1 := regexp.MustCompile("(.)([A-Z][a-z]+)")
	s1 := rxp1.ReplaceAllString(caseName, "${1}_${2}")
	rxp2 := regexp.MustCompile("([a-z0-9])([A-Z])")
	return strings.ToLower(rxp2.ReplaceAllString(s1, "${1}_${2}"))
}

// Use GenMockStructs for a more simplified version with less input arguments.
func MockRepository(pkgOut, outDir, main string, repos map[string]interface{}) {
	os.MkdirAll(outDir, os.ModePerm)

	for name, repo := range repos {
		builder := generateMockContent(repo, pkgOut, main)

		res, err := imports.Process("", []byte(builder.String()), &imports.Options{Comments: true})
		if err != nil {
			panic(err)
		}

		err = ioutil.WriteFile(outDir+"/"+name+".go", res, 0o644) //nolint:gosec
		if err != nil {
			log.Println(err)
			panic(err)
		}
		log.Print("Mock gen " + outDir + "/" + name + ".go")
	}
}

func generateMockContent(in interface{}, pkgName, main string) strings.Builder {
	typ := reflect.TypeOf(in)
	mockStructName := "Mock" + typ.Elem().Name()

	builder := strings.Builder{}
	builder.WriteString(headerLine)
	builder.WriteString(fmt.Sprintf("package %s\n", pkgName))
	builder.WriteString(generateImportBlock(main))
	builder.WriteString(fmt.Sprintf(`
type %s struct {
	mock.Mock
}`, mockStructName))

	for i := 0; i < typ.NumMethod(); i++ {
		method := typ.Method(i)
		methodType := method.Type.String()[4:len(method.Type.String())]

		count := 0

		var (
			startAt, endAt        int
			inputArgs, returnArgs string
		)
		for i, char := range methodType {
			if char == '(' {
				count++
				if count == 1 {
					startAt = i + 1
				}
			}

			if char == ')' {
				count--
			}

			if count == 0 {
				endAt = i

				returnArgs = strings.Trim(methodType[endAt+1:], " ")
				inputArgs = methodType[startAt:endAt]
				break
			}
		}

		args := strings.Split(inputArgs, ",")

		inputArgs = ""

		argsName := []string{}
		argsNameAndType := []string{}
		// skip arg1 because this is receiver
		for i, argType := range args[1:] {
			argName := fmt.Sprintf("arg%d", i+1)
			argsNameAndType = append(argsNameAndType, argName+" "+argType)
			argsName = append(argsName, argName)
		}

		preReturnGroup := ""
		returnGroup := []string{}
		parts := strings.Split(returnArgs, ",")
		for i, arg := range parts {
			arg = strings.Trim(arg, "( )")

			if arg == "" {
				continue
			} else if arg == "error" {
				returnGroup = append(returnGroup, fmt.Sprintf("args.Error(%d)", i))
			} else {
				returnGroup = append(returnGroup, fmt.Sprintf("args.Get(%d).(%s)", i, arg))
			}
		}

		returnArgsString := "return " + strings.Join(returnGroup, ", ")

		if len(returnGroup) > 0 {
			preReturnGroup = fmt.Sprintf("args := r.Called(%s)", strings.Join(argsName, ", "))
		} else {
			preReturnGroup = fmt.Sprintf("_ = r.Called(%s)", strings.Join(argsName, ", "))
		}

		for i, arg := range returnGroup {
			if strings.Contains(arg, "[]") || strings.Contains(arg, "*") {
				returnGroup[i] = "nil"
				preReturnGroup += fmt.Sprintf(`

	if args.Get(%d) == nil {
		return %s
	}`, i, strings.Join(returnGroup, ", "))

			}
		}

		builder.WriteString(fmt.Sprintf(`
func (r *%s) %s (%s) %s {
	%s
	%s
}
`, mockStructName, typ.Method(i).Name, strings.Join(argsNameAndType, ", "), returnArgs, preReturnGroup, returnArgsString))
	}
	return builder
}

const headerLine = "// Code generated by mockgen. DO NOT EDIT.\n"

var defaultImportPaths = []string{
	`"github.com/stretchr/testify/mock"`,
	`"github.com/jackc/pgtype"`,
	`"github.com/jackc/pgx/v4"`,
	`"github.com/googleapis/gax-go/v2"`,
	`"google.golang.org/genproto/googleapis/cloud/texttospeech/v1"`,
	``, // empty line to separate internal vs 3rd-party imports
	`pb "github.com/manabie-com/backend/pkg/genproto/{{.Service}}"`,
	`"github.com/manabie-com/backend/pkg/manabuf/{{.Service}}/v1"`,
	`"github.com/manabie-com/backend/internal/golibs/database"`,
	`"github.com/manabie-com/backend/internal/{{.Service}}/entities"`,
	`"github.com/manabie-com/backend/internal/{{.Service}}/repositories"`,
}

var (
	importPaths        = []string{}
	removedImportPaths = []string{}
)

func init() {
	ResetImports()
}

// AddImportPath adds path to the import block for the generated mock files.
func AddImport(path string) {
	importPaths = append(importPaths, fmt.Sprintf("%q", path))
}

// AddImportWithPkgAlias adds path with package alias to the import block used in the generated mock files.
func AddImportWithPkgAlias(path, pkgAlias string) {
	importPaths = append(importPaths, fmt.Sprintf("%s %q", pkgAlias, path))
}

// RemoveImport removes path from the import block used in the generated mock files.
func RemoveImport(path string) {
	removedImportPaths = append(removedImportPaths, path)
}

// RemoveImportWithPkgAlias removes path with package alias from the import block used in the generated mock files.
func RemoveImportWithPkgAlias(path, pkgAlias string) {
	removedImportPaths = append(removedImportPaths, fmt.Sprintf("%s %q", pkgAlias, path))
}

// ResetImports resets the import block back to the default version.
func ResetImports() {
	importPaths = make([]string, len(defaultImportPaths))
	copy(importPaths, defaultImportPaths)
}

// ClearImports remove everything from the import block.
func ClearImports() {
	importPaths = []string{}
}

func generateImportBlock(serviceName string) string {
	rawBlock := fmt.Sprintf(`
import (
	%s
)
`, strings.Join(importPaths, "\n\t"))
	t := template.Must(template.New("importBlock").Parse(rawBlock))

	sb := &strings.Builder{}
	err := t.Execute(sb, map[string]string{"Service": serviceName})
	if err != nil {
		panic(fmt.Errorf("failed to execute import block template: %s", err))
	}

	res := sb.String()
	for _, p := range removedImportPaths {
		res = strings.ReplaceAll(res, formatImportLine(p), "")
	}
	return res
}

func formatImportLine(path string) string {
	return fmt.Sprintf("\t\"%s\"\n", path)
}

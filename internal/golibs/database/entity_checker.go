package database

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/stringutil"
)

const (
	entity   = "entity"
	entities = "entities"
)

var (
	mEntitiesMethods []method
	mEntityMethods   []method
)

func init() {
	mEntityMethods = listMethods(entity)
	mEntitiesMethods = listMethods(entities)
}

type validateEntity struct {
	entitiesMethods map[string][]method
	count           int
}

type method struct {
	name    string
	params  []string
	results []string
}

func CheckEntity(path string) (int, error) {
	return checkEntities(path, entity)
}

func CheckEntities(path string) (int, error) {
	return checkEntities(path, entities)
}

func checkEntities(path string, eType string) (int, error) {
	packageMap, err := parseDirectory(path)
	if err != nil {
		return 0, err
	}

	validator := new(validateEntity)
	for _, p := range packageMap {
		for fileName, content := range p.Files {
			// exclude test file
			if strings.Contains(fileName, "_test") {
				continue
			}

			validator.entitiesMethods = make(map[string][]method)
			ast.Inspect(content, func(n ast.Node) bool {
				switch node := n.(type) {
				case *ast.FuncDecl:
					validator.processNodeAsFunction(node)
				}
				return true
			})
			validator.countEntity(getEntitiesMethods(eType))
		}
	}

	return validator.count, nil
}

func (v *validateEntity) processNodeAsFunction(node *ast.FuncDecl) {
	if node.Recv != nil {
		for _, recv := range node.Recv.List {
			entityName := ""

			switch typeExpr := recv.Type.(type) {
			case *ast.Ident:
				entityName = typeExpr.Name
			case *ast.StarExpr:
				switch typeExpr2 := typeExpr.X.(type) {
				case *ast.Ident:
					entityName = typeExpr2.Name
				}
			}

			if entityName == "" {
				continue
			}

			temp := method{
				name:    node.Name.Name,
				params:  funcParams(node.Type),
				results: funcResults(node.Type),
			}

			v.entitiesMethods[entityName] = append(v.entitiesMethods[entityName], temp)
		}
	}
}

func (v *validateEntity) countEntity(enInterfaceMethods []method) {
	for _, lsMethods := range v.entitiesMethods {
		if len(lsMethods) < len(enInterfaceMethods) {
			continue
		}

		matchedCount := 0
		for _, iMethod := range enInterfaceMethods {
			for _, enMethod := range lsMethods {
				if compareMethod(iMethod, enMethod) {
					matchedCount++
				}
			}
		}

		if matchedCount == len(enInterfaceMethods) {
			v.count++
		}
	}
}

func parseDirectory(path string) (map[string]*ast.Package, error) {
	fset := token.NewFileSet()
	return parser.ParseDir(fset, path, nil, 0)
}

func listMethods(eType string) []method {
	methods := make([]method, 0)
	var rType reflect.Type
	switch eType {
	case entity:
		rType = reflect.TypeOf((*Entity)(nil)).Elem()
	case entities:
		rType = reflect.TypeOf((*Entities)(nil)).Elem()
	}

	for i := 0; i < rType.NumMethod(); i++ {
		methods = append(methods, getMethodParamsAndResults(rType.Method(i)))
	}

	return methods
}

func getEntitiesMethods(eType string) []method {
	switch eType {
	case entity:
		return mEntityMethods
	case entities:
		return mEntitiesMethods
	default:
		return nil
	}
}

func getMethodParamsAndResults(rMethod reflect.Method) (method method) {
	fStr := rMethod.Type.String()
	oParenthesesIdx := strings.Index(fStr, "(")
	if oParenthesesIdx < 0 {
		return
	}

	cParenthesesIdx := strings.Index(fStr, ")")
	if cParenthesesIdx < 0 {
		return
	}

	replacer := strings.NewReplacer("(", "", ")", "", " ", "")
	params := replacer.Replace(fStr[oParenthesesIdx:cParenthesesIdx])
	result := replacer.Replace(fStr[cParenthesesIdx:])

	method.name = rMethod.Name
	if params != "" {
		method.params = strings.Split(replacer.Replace(fStr[oParenthesesIdx:cParenthesesIdx]), ",")
	}
	if result != "" {
		method.results = strings.Split(replacer.Replace(fStr[cParenthesesIdx:]), ",")
	}

	return
}

func funcParams(fType *ast.FuncType) []string {
	params := []string{}
	if fType.Params != nil {
		for _, param := range fType.Params.List {
			switch expr := param.Type.(type) {
			case *ast.Ident:
				params = append(params, expr.Name)
			}
		}
	}
	return params
}

func funcResults(fType *ast.FuncType) []string {
	results := []string{}
	if fType.Results != nil {
		for _, result := range fType.Results.List {
			// get result type, such as: string, []interface, database.Entities, []database.Entity
			switch expr := result.Type.(type) {
			case *ast.Ident:
				results = append(results, expr.Name)
			case *ast.ArrayType:
				name := ""
				switch typeExpr := expr.Elt.(type) {
				case *ast.Ident:
					name = fmt.Sprintf("[]%s", typeExpr.Name)
				case *ast.InterfaceType:
					name = "[]interface{}"
				case *ast.SelectorExpr:
					switch typeExpr2 := typeExpr.X.(type) {
					case *ast.Ident:
						name = fmt.Sprintf("[]%s.%s", typeExpr2.Name, typeExpr.Sel.Name)
					}
				}
				results = append(results, name)
			case *ast.SelectorExpr:
				name := ""
				switch typeExpr := expr.X.(type) {
				case *ast.Ident:
					name = fmt.Sprintf("%s.%s", typeExpr.Name, expr.Sel.Name)
				}
				results = append(results, name)
			}
		}
	}

	return results
}

func compareMethod(iMethod, enMethod method) bool {
	if iMethod.name != enMethod.name {
		return false
	}
	if !stringutil.SliceEqual(iMethod.params, enMethod.params) {
		return false
	}
	if !stringutil.SliceEqual(iMethod.results, enMethod.results) {
		return false
	}

	return true
}

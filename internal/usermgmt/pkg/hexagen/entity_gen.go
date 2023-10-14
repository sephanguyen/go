package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/manabie-com/backend/internal/usermgmt/pkg/utils"
)

func genEntImpl(typeNamesText string, output string, outpkg string, args ...string) {
	typeNames := strings.Split(typeNamesText, ",")

	var tags []string
	if len(*buildTags) > 0 {
		tags = strings.Split(*buildTags, ",")
	}

	// We accept either one directory or a list of files. Which do we have?
	if len(args) == 0 {
		// Default: process whole package in current directory.
		args = []string{"."}
	}

	var outputDir string
	if len(args) == 1 && isDirectory(args[0]) {
		outputDir = args[0]
	} else {
		if len(tags) != 0 {
			log.Fatal("-tags option applies only to directories, not when files are specified")
		}
		outputDir = filepath.Dir(args[0])
	}

	g := Generator{
		trimPrefix:  *trimprefix,
		lineComment: *linecomment,
	}

	g.addPackages(ParsePackage(args, tags))

	if output == "" {
		output = outputDir
	}

	var importPaths []string

	// Generate implementation
	var generatedImplBuffer bytes.Buffer
	for _, typeName := range typeNames {
		buffer, requiredImportPaths := generateEntImpl(g.pkgs, typeName)
		if _, err := buffer.WriteTo(&generatedImplBuffer); err != nil {
			panic(err)
		}
		importPaths = append(importPaths, requiredImportPaths...)
	}

	if outpkg == "" {
		outpkg = g.findPackageName(typeNames[0]) // g.pkg.name
	}

	// Write header and package clauses
	if _, err := g.buf.Write(headerAndPackageClause(typeNamesText, output, outpkg, args)); err != nil {
		panic(err)
	}

	// Write import paths
	importPathsBuffer, err := BufferFromImportPaths(importPaths)
	if err != nil {
		panic(err)
	}

	if _, err := g.buf.Write(importPathsBuffer); err != nil {
		panic(err)
	}

	// Write all implementations
	if _, err := generatedImplBuffer.WriteTo(&g.buf); err != nil {
		panic(err)
	}
	// Format the output.
	src := gofmt(&g.buf)

	var dest string
	// Write to file.
	for _, pkg := range g.pkgs {
		for _, file := range pkg.files {
			if file.HasInterfaceTypeName(typeNames[0]) {
				fileName := strings.Split(filepath.Base(file.absolutePath), ".")[0]
				baseName := fmt.Sprintf("%s_generated_impl.go", fileName)
				// baseName := fmt.Sprintf("%s_impl.go", typeNames[0])
				dest = filepath.Join(output, strings.ToLower(baseName))
			}
		}
	}

	fmt.Println("output:", output)
	fmt.Println("dest:", dest)

	if err := os.WriteFile(dest, src, 0600); err != nil {
		log.Fatalf("writing output: %s", err)
	}
}

func printf(buffer *bytes.Buffer, format string, args ...interface{}) {
	_, err := fmt.Fprintf(buffer, format, args...)
	if err != nil {
		panic(err)
	}
}

func generateEntImpl(packages []*Package, typeName string) (*bytes.Buffer, []string) {
	var (
		buffer              bytes.Buffer
		requiredImportPaths = []string{
			`github.com/manabie-com/backend/internal/usermgmt/pkg/field`,
			`github.com/pkg/errors`,
		}
	)

	interfaceDecls := make([]*InterfaceDecl, 0, 100)

	for _, pkg := range packages {
		for _, file := range pkg.files {
			ast.Inspect(file.file, file.findInterfaceType(typeName))
			interfaceDecls = append(interfaceDecls, file.InterfaceDecls...)
		}
	}

	if len(interfaceDecls) == 0 {
		log.Fatalf(`no interface declarations defined for type "%s"`, typeName)
	}

	// Generate Null Entity
	for _, interfaceDecl := range interfaceDecls {
		entityName := fmt.Sprintf("Null%s", interfaceDecl.Name)

		printf(&buffer, "// This statement will fail to compile if *%s ever stops matching the interface.\n", entityName)
		printf(&buffer, "var _ %s = (*%s)(nil)\n", interfaceDecl.Name, entityName)
		printf(&buffer, "type %s struct {\n", entityName)
		for _, ident := range interfaceDecl.Idents {
			printf(&buffer, "\t%s\n", fmt.Sprintf("Null%s", ident.Name))
		}
		printf(&buffer, "}\n")

		for _, field := range interfaceDecl.Fields {
			funcType, ok := field.IsFuncType()
			if !ok {
				fmt.Printf("%T\n", field.Type)
				fmt.Printf("%+v\n", field)
				fmt.Printf("%+v\n", field.Type)
				panic("not func type")
			}

			printf(&buffer, "func (%s %s) %s() %s {\n", utils.LowerCaseFirstLetter(entityName), entityName, field.Name(), strings.Join(funcType.SelectorExprs().Names(), ", "))
			printf(&buffer, "\treturn field.NewNull%s()", utils.UpperCaseFirstLetter(funcType.SelectorExprs()[0].SelectorExprSelName()))
			printf(&buffer, "}\n")
		}
	}

	// Generate function to compare two entities
	for _, interfaceDecl := range interfaceDecls {
		funcName := fmt.Sprintf("Compare%sValues", interfaceDecl.Name)
		argName1 := utils.LowerCaseFirstLetter(interfaceDecl.Name) + "1"
		argName2 := utils.LowerCaseFirstLetter(interfaceDecl.Name) + "2"

		comparableFieldTypeName := interfaceDecl.Name + "Field"

		printf(&buffer, "\n")
		printf(&buffer, "//%s compare values of two %s entities\n", funcName, interfaceDecl.Name)
		printf(&buffer, "func %s(%s %s, %s %s, fieldsToCompare ...%s) error {\n", funcName, argName1, interfaceDecl.Name, argName2, interfaceDecl.Name, comparableFieldTypeName)
		printf(&buffer, "// By default, compare all fields if number of fields to compare is 0\n")
		printf(&buffer, "if len(fieldsToCompare) < 1 {\n")
		printf(&buffer, "fieldsToCompare = []%s{\n", comparableFieldTypeName)
		for _, interfaceField := range interfaceDecl.Fields {
			comparableFieldMethodName := comparableFieldTypeName + interfaceField.Name()
			printf(&buffer, "\t%s,\n", comparableFieldMethodName)
		}
		printf(&buffer, "\t}\n")
		printf(&buffer, "}\n")
		printf(&buffer, "\n")

		for _, ident := range interfaceDecl.Idents {
			embeddedFuncName := fmt.Sprintf("Compare%sValues", ident.Name)
			printf(&buffer, "\t%s(%s, %s)\n", embeddedFuncName, argName1, argName2)
		}
		printf(&buffer, "\n")

		printf(&buffer, "\tfor _, fieldToCompare := range fieldsToCompare {\n")
		printf(&buffer, "\t\tswitch (fieldToCompare) {\n")
		for _, interfaceField := range interfaceDecl.Fields {
			comparableFieldMethodName := comparableFieldTypeName + interfaceField.Name()
			funcType, ok := interfaceField.IsFuncType()
			if !ok {
				panic("not func type")
			}
			printf(&buffer, "\t\t\tcase %s:\n", comparableFieldMethodName)
			printf(&buffer, "\t\t\tif (%s.%s().%s() != %s.%s().%s()){\n", argName1, interfaceField.Name(), strings.Split(funcType.SelectorExprs().Names()[0], ".")[1], argName2, interfaceField.Name(), strings.Split(funcType.SelectorExprs().Names()[0], ".")[1])
			printf(&buffer, "\t\treturn errors.New(\"%s is not equal\")\n", interfaceField.Name())
			printf(&buffer, "\t\t}\n")
		}
		printf(&buffer, "\t\t}")
		printf(&buffer, "\t}\n")
		printf(&buffer, "\n")
		printf(&buffer, "\treturn nil\n")
		printf(&buffer, "}\n")
	}

	for _, interfaceDecl := range interfaceDecls {
		sliceName := fmt.Sprintf("%ss", interfaceDecl.Name)

		printf(&buffer, "// %s represents for a slice of %s\n", sliceName, interfaceDecl.Name)
		printf(&buffer, "type %s []%s\n", sliceName, interfaceDecl.Name)
		/*g.Printf("}\n")
		g.Printf("//This statement will fail to compile if *%s ever stops matching the interface.\n", interfaceDecl.Name)
		g.Printf("var _ entity.User = (*User)(nil)\n")*/

		for _, interfaceField := range interfaceDecl.Fields {
			funcType, ok := interfaceField.IsFuncType()
			if !ok {
				panic("not func type")
			}
			methodName := fmt.Sprintf("%ss", interfaceField.Name())
			methodResult := strings.Join(funcType.SelectorExprs().Names(), ", ")
			receiver := strings.ToLower(sliceName)
			if len(funcType.SelectorExprs()) > 1 {
				methodResult = fmt.Sprintf("(%s)", methodResult)
			}
			printf(&buffer, "func (%s %s) %s() %ss {\n", receiver, sliceName, methodName, methodResult)
			printf(&buffer, "\t%s := make([]%s, 0, len(%s))\n", utils.LowerCaseFirstLetter(methodName), methodResult, receiver)
			printf(&buffer, "\tfor _, %s := range %s {\n", strings.ToLower(interfaceDecl.Name), receiver)
			printf(&buffer, "\t\t%s = append(%s, %s.%s())", utils.LowerCaseFirstLetter(methodName), utils.LowerCaseFirstLetter(methodName), strings.ToLower(interfaceDecl.Name), interfaceField.Name())
			printf(&buffer, "\t}\n")
			printf(&buffer, "\treturn %s", utils.LowerCaseFirstLetter(methodName))
			printf(&buffer, "}\n")
			printf(&buffer, "\n")
		}
	}

	for _, interfaceDecl := range interfaceDecls {
		exportedEntName := interfaceDecl.Name
		unexportedEntName := strings.ToLower(interfaceDecl.Name)
		nullEntityName := fmt.Sprintf("Null%s", interfaceDecl.Name)
		printf(&buffer, "type %s struct {\n", unexportedEntName)
		printf(&buffer, "\t%s\n", nullEntityName)
		printf(&buffer, "\n")
		for _, interfaceField := range interfaceDecl.Fields {
			fieldName := utils.LowerCaseFirstLetter(interfaceField.Name())
			funcType, ok := interfaceField.IsFuncType()
			if !ok {
				panic("not func type")
			}
			printf(&buffer, "\t%s\t%s\n", fieldName, funcType.SelectorExprs().Names()[0])
		}
		printf(&buffer, "}\n")

		for _, interfaceField := range interfaceDecl.Fields {
			fieldName := utils.LowerCaseFirstLetter(interfaceField.Name())
			funcType, ok := interfaceField.IsFuncType()
			if !ok {
				panic("not func type")
			}
			printf(&buffer, "func (%s *%s) %s() %s {\n", unexportedEntName, unexportedEntName, interfaceField.Name(), strings.Join(funcType.SelectorExprs().Names(), ", "))
			printf(&buffer, "\treturn %s.%s", unexportedEntName, fieldName)
			printf(&buffer, "}\n")
		}

		printf(&buffer, "\n")

		printf(&buffer, "type %sOption interface {\n", exportedEntName)
		printf(&buffer, "\tapply(*%s)\n", unexportedEntName)
		printf(&buffer, "}\n")

		printf(&buffer, "type %sOption func(*%s)\n", unexportedEntName, unexportedEntName)
		printf(&buffer, "func (%sOption %sOption) apply(%s *%s) {\n", unexportedEntName, unexportedEntName, unexportedEntName, unexportedEntName)
		printf(&buffer, "\t%sOption(%s)\n", unexportedEntName, unexportedEntName)
		printf(&buffer, "}\n")

		printf(&buffer, "\n")

		printf(&buffer, "type %sFieldsImpl struct{}\n", exportedEntName)
		printf(&buffer, "var %sFields = %sFieldsImpl{}\n", exportedEntName, exportedEntName)

		printf(&buffer, "func (%sFieldsImpl %sFieldsImpl) From(src%s %s) %sOption {\n", unexportedEntName, exportedEntName, exportedEntName, exportedEntName, exportedEntName)
		printf(&buffer, "\treturn %sOption(func(%s *%s) {\n", unexportedEntName, unexportedEntName, unexportedEntName)
		for _, interfaceField := range interfaceDecl.Fields {
			fieldName := utils.LowerCaseFirstLetter(interfaceField.Name())
			_, ok := interfaceField.IsFuncType()
			if !ok {
				panic("not func type")
			}
			printf(&buffer, "\t\t%s.%s = src%s.%s()\n", unexportedEntName, fieldName, exportedEntName, interfaceField.Name())
		}
		printf(&buffer, "\t})\n")
		printf(&buffer, "}\n")

		for _, interfaceField := range interfaceDecl.Fields {
			fieldName := utils.LowerCaseFirstLetter(interfaceField.Name())
			funcType, ok := interfaceField.IsFuncType()
			if !ok {
				panic("not func type")
			}
			printf(&buffer, "func (%sFieldsImpl %sFieldsImpl) %s(%s %s) %sOption {\n", unexportedEntName, exportedEntName, interfaceField.Name(), fieldName, funcType.SelectorExprs().Names()[0], exportedEntName)
			printf(&buffer, "\treturn %sOption(func(%s *%s) {\n", unexportedEntName, unexportedEntName, unexportedEntName)
			printf(&buffer, "\t\t%s.%s = %s\n", unexportedEntName, fieldName, fieldName)
			printf(&buffer, "\t})\n")
			printf(&buffer, "}\n")
		}

		printf(&buffer, "\n")
		printf(&buffer, "func New%s(options ...%sOption) %s {\n", exportedEntName, exportedEntName, exportedEntName)
		printf(&buffer, "\t%s := &%s{}\n", unexportedEntName, unexportedEntName)
		printf(&buffer, "\tfor _, option := range options {\n")
		printf(&buffer, "\t\toption.apply(%s)\n", unexportedEntName)
		printf(&buffer, "\t}\n")
		printf(&buffer, "\treturn %s\n", unexportedEntName)
		printf(&buffer, "}\n")
	}

	return &buffer, requiredImportPaths
}

package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"log"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/tools/go/packages"
)

var version = "v0.1.0"

var (
	trimprefix  = flag.String("trimprefix", "", "trim the `prefix` from the generated constant names")
	linecomment = flag.Bool("linecomment", false, "use line comment text as printed text when present")
	buildTags   = flag.String("tags", "", "comma-separated list of build tags to apply")

	typeNames string
	output    string
	outpkg    string
)

func main() {
	var hexagenCmd = &cobra.Command{
		Use:     "hexagen",
		Short:   "hexagen is a code generator for abstractions of hexagon architecture",
		Long:    `hexagen is a code generator for abstractions of hexagon architecture`,
		Version: version,
		Run: func(cmd *cobra.Command, args []string) {
			// Do Stuff Here
		},
	}

	entImplCmd := &cobra.Command{
		Use:   "ent-impl",
		Short: "generate entity impl from entity abstraction",
		Long:  `generate entity impl from entity abstraction`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Hexagen %s\n", version)
			fmt.Printf("hexagen generating ent impl for %s...\n", typeNames)
			genEntImpl(typeNames, output, outpkg, args...)
			fmt.Printf("hexagen genterated entity implementation for %s\n", typeNames)
		},
	}
	entImplCmd.Flags().StringVarP(&typeNames, "type", "t", "", "")
	entImplCmd.Flags().StringVarP(&output, "output", "o", ".", "")
	entImplCmd.Flags().StringVarP(&outpkg, "outpkg", "p", "", "")
	if err := entImplCmd.MarkFlagRequired("type"); err != nil {
		panic(err)
	}

	hexagenCmd.AddCommand(entImplCmd)

	if err := hexagenCmd.Execute(); err != nil {
		log.Panic(err)
	}
}

type Package struct {
	name  string
	path  string
	files []*File

	legacyPkg *packages.Package
}

// File holds a single parsed file and associated data.
type File struct {
	absolutePath string
	pkg          *Package  // Package to which this file belongs.
	file         *ast.File // Parsed AST.

	trimPrefix  string
	lineComment bool

	InterfaceDecls []*InterfaceDecl
}

func (file *File) HasInterfaceTypeName(interfaceTypeName string) bool {
	for _, interfaceDecl := range file.InterfaceDecls {
		if interfaceDecl.Name == interfaceTypeName {
			return true
		}
	}
	return false
}

type InterfaceType ast.InterfaceType

func (interfaceType *InterfaceType) ListOfMethods() Fields {
	fields := make(Fields, len(interfaceType.Methods.List))
	for i := range interfaceType.Methods.List {
		fields[i] = (*Field)(interfaceType.Methods.List[i])
	}
	return fields
}
func (interfaceType *InterfaceType) InterfaceMethods() {
	interfaceType.ListOfMethods()
}

type Field ast.Field
type Fields []*Field

func (field *Field) Name() string {
	return field.Names[0].Name
}
func (field *Field) IsFuncType() (*FuncOfInterface, bool) {
	funcType, ok := field.Type.(*ast.FuncType)
	if !ok {
		return nil, false
	}
	return (*FuncOfInterface)(funcType), true
}

type FuncOfInterface ast.FuncType
type FuncsOfInterface []*FuncOfInterface

func (funcOfInterface *FuncOfInterface) SelectorExprs() SelectorExprs {
	funcResults := make(SelectorExprs, 0)
	for _, methodResult := range funcOfInterface.Results.List {
		methodSelectorExpr, _ := methodResult.Type.(*ast.SelectorExpr)
		funcResults = append(funcResults, (*SelectorExpr)(methodSelectorExpr))
	}
	return funcResults
}

type SelectorExpr ast.SelectorExpr
type SelectorExprs []*SelectorExpr

func (selectorExpr *SelectorExpr) SelectorExprXName() string {
	funcSelectorExprX, _ := selectorExpr.X.(*ast.Ident)
	return funcSelectorExprX.Name
}
func (selectorExpr *SelectorExpr) SelectorExprSelName() string {
	return selectorExpr.Sel.Name
}

func (funcResults SelectorExprs) Names() []string {
	names := make([]string, 0)
	for _, funcResult := range funcResults {
		names = append(names, funcResult.SelectorExprXName()+"."+funcResult.SelectorExprSelName())
	}
	return names
}

// genDecl processes one declaration clause.
func (file *File) findInterfaceType(typeName string) func(node ast.Node) bool {
	return func(node ast.Node) bool {
		genDecl, ok := node.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			// We only care about type declarations.
			// ex:
			// 	type User interface {}
			// 	type User struct {}
			return true
		}

		// Loop over the elements of the declaration. Each element is a ValueSpec:
		// a list of names possibly followed by a type, possibly followed by values.
		// If the type and value are both missing, we carry down the type (and value,
		// but the "go/types" package takes care of that).*/

		// interfaceDecls := make([]*InterfaceDecl, 0, 10)

		for _, spec := range genDecl.Specs {
			typeSpec := spec.(*ast.TypeSpec) // Guaranteed to succeed as this is TypeSpec.

			if typeSpec.Name.Name != typeName {
				return false
			}

			if typeSpec.Type != nil {
				interfaceType, ok := typeSpec.Type.(*ast.InterfaceType)
				if !ok {
					continue
				}

				interfaceDecl := &InterfaceDecl{
					Name: typeSpec.Name.Name,
				}

				// Interface's methods
				for _, method := range interfaceType.Methods.List {
					if ident, ok := method.Type.(*ast.Ident); ok {
						// interfaceDecl.Idents = append(interfaceDecl.Idents, (*Ident)(ident))

						internalPkg := &Package{
							name:      file.pkg.name,
							path:      file.pkg.path,
							files:     file.pkg.files,
							legacyPkg: file.pkg.legacyPkg,
						}

						for _, importedFile := range file.pkg.files {
							absolutePath := importedFile.absolutePath
							if strings.HasSuffix(absolutePath, "_generated_impl.go") {
								continue
							}
							f := &File{
								absolutePath:   absolutePath,
								file:           importedFile.file,
								pkg:            internalPkg,
								InterfaceDecls: make([]*InterfaceDecl, 0),
							}
							internalPkg.files = append(internalPkg.files, f)

							ast.Inspect(f.file, f.findInterfaceType(ident.Name))

							for _, embeddedInterfaceDecl := range f.InterfaceDecls {
								interfaceDecl.Fields = append(interfaceDecl.Fields, embeddedInterfaceDecl.Fields...)
							}
						}
						continue
					}
					if selectorExpr, ok := method.Type.(*ast.SelectorExpr); ok {
						interfaceDecl.FuncResults = append(interfaceDecl.FuncResults, (*SelectorExpr)(selectorExpr))

						for key, pkg := range file.pkg.legacyPkg.Imports {
							if !strings.HasSuffix(key, (*SelectorExpr)(selectorExpr).SelectorExprXName()) {
								continue
							}

							internalPkg := &Package{
								name:      pkg.Name,
								path:      pkg.PkgPath,
								files:     make([]*File, 0, len(pkg.Syntax)),
								legacyPkg: pkg,
							}

							for _, importedFile := range pkg.Syntax {
								absolutePath := pkg.Fset.File(importedFile.Pos()).Name()
								if strings.HasSuffix(absolutePath, "_generated_impl.go") {
									continue
								}
								f := &File{
									absolutePath:   absolutePath,
									file:           importedFile,
									pkg:            internalPkg,
									InterfaceDecls: make([]*InterfaceDecl, 0),
								}
								internalPkg.files = append(internalPkg.files, f)

								ast.Inspect(f.file, f.findInterfaceType((*SelectorExpr)(selectorExpr).SelectorExprSelName()))

								for _, embeddedInterfaceDecl := range f.InterfaceDecls {
									interfaceDecl.Fields = append(interfaceDecl.Fields, embeddedInterfaceDecl.Fields...)
								}
							}
						}
						continue
					}

					interfaceDecl.Fields = append(interfaceDecl.Fields, (*Field)(method))
				}
				file.InterfaceDecls = append(file.InterfaceDecls, interfaceDecl)
			}
		}

		return false
	}
}

type InterfaceDecl struct {
	Name        string
	Idents      Idents
	Fields      Fields
	FuncResults SelectorExprs
}

type Ident ast.Ident
type Idents []*Ident

type Generator struct {
	buf  bytes.Buffer // Accumulated output.
	pkgs []*Package   // Package we are scanning.

	trimPrefix  string
	lineComment bool
}

// addPackages adds a type checked Package and its syntax files to the generator.
func (g *Generator) addPackages(packages []*packages.Package) {
	for _, pkg := range packages {
		internalPkg := &Package{
			name:      pkg.Name,
			path:      pkg.PkgPath,
			files:     make([]*File, 0, len(pkg.Syntax)),
			legacyPkg: pkg,
		}

		for _, file := range pkg.Syntax {
			absolutePath := pkg.Fset.File(file.Pos()).Name()
			if strings.HasSuffix(absolutePath, "_generated_impl.go") {
				continue
			}
			f := &File{
				absolutePath: absolutePath,
				file:         file,
				pkg:          internalPkg,
				trimPrefix:   g.trimPrefix,
				lineComment:  g.lineComment,
			}
			internalPkg.files = append(internalPkg.files, f)
		}
		g.pkgs = append(g.pkgs, internalPkg)
	}
}

func (g *Generator) findPackageName(interfaceTypeName string) string {
	var pkgName string
	for _, pkg := range g.pkgs {
		for _, file := range pkg.files {
			if file.HasInterfaceTypeName(interfaceTypeName) {
				pkgName = pkg.name
			}
		}
	}
	return pkgName
}

// ParsePackage analyzes the single package constructed from the patterns and tags.
// ParsePackage exits if there is an error.
func ParsePackage(patterns []string, tags []string) []*packages.Package {
	cfg := &packages.Config{
		Mode:       packages.NeedName | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedSyntax | packages.NeedImports | packages.NeedDeps,
		Tests:      false,
		BuildFlags: []string{fmt.Sprintf("-tags=%s", strings.Join(tags, " "))},
	}
	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		log.Fatal(err)
	}
	return pkgs
}

// gofmt returns the gofmt-ed contents of the Generator's buffer.
func gofmt(buffer *bytes.Buffer) []byte {
	src, err := format.Source(buffer.Bytes())
	if err != nil {
		// Should never happen, but can arise when developing this code.
		// The user can compile the output to see the error.
		log.Printf("warning: internal error: invalid Go generated: %s", err)
		log.Printf("warning: compile the package to analyze the error")
		return buffer.Bytes()
	}
	return src
}

func headerAndPackageClause(typeName string, output string, outpkg string, args []string) []byte {
	var buffer bytes.Buffer
	printf(&buffer, "// Code generated by \"hexagen --type=%s --output=%s %s\" (%s); DO NOT EDIT.\n", typeName, output, strings.Join(args, " "), version)
	printf(&buffer, "\n")
	printf(&buffer, "package %s", outpkg)
	printf(&buffer, "\n")
	return buffer.Bytes()
}

package analysisgo

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

	"go.uber.org/multierr"
	"golang.org/x/tools/go/packages"
)

type Package struct {
	*packages.Package
}

func (p *Package) IsTest() bool {
	return strings.HasSuffix(p.Name, "_test")
}

type PackageOption func(cf *packages.Config) error

func WithDir(path string) PackageOption {
	return func(cf *packages.Config) error {
		cf.Dir = path
		return nil
	}
}

func WithFileContents(files map[string][]byte) PackageOption {
	return func(cf *packages.Config) error {
		cf.Overlay = files
		return nil
	}
}

func NewPackages(opts ...PackageOption) ([]*Package, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax | packages.NeedImports | packages.NeedTypes | packages.NeedTypesInfo,
		ParseFile: func(fset *token.FileSet, filename string, src []byte) (*ast.File, error) {
			return parser.ParseFile(fset, filename, src, parser.ParseComments)
		},
	}
	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, err
		}
	}

	pkgs, err := packages.Load(cfg)
	if err != nil {
		return nil, fmt.Errorf("packages.Load: %w", err)
	}

	var errs []error
	packages.Visit(pkgs, nil, func(pkg *packages.Package) {
		for _, err = range pkg.Errors {
			errs = append(errs, err)
		}
	})
	if len(errs) != 0 {
		return nil, fmt.Errorf("got error when parse package: %w", multierr.Combine(errs...))
	}

	res := make([]*Package, 0, len(pkgs))
	for _, pkg := range pkgs {
		res = append(res, &Package{Package: pkg})
	}

	return res, nil
}

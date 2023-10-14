package analysisgo

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

type GoFile struct {
	path string
	src  any
	fs   *token.FileSet
	f    *ast.File
}

type InputSource func(f *GoFile) error

func WithFile(path string) InputSource {
	return func(f *GoFile) error {
		f.path = path
		return nil
	}
}

// WithSource need src not nil
func WithSource(src any, path string) InputSource {
	return func(f *GoFile) error {
		if src == nil {
			return fmt.Errorf("src should not is null")
		}
		f.path = path
		f.src = src
		return nil
	}
}

func NewGoFile(input InputSource) (*GoFile, error) {
	f := &GoFile{}
	err := input(f)
	if err != nil {
		return nil, err
	}

	fset := token.NewFileSet()
	f.fs = fset
	f.f, err = parser.ParseFile(fset, f.path, f.src, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("could not parse go file at path \"%s\": %w", f.path, err)
	}

	return f, nil
}

type Iterator interface {
	hasNext() bool
}

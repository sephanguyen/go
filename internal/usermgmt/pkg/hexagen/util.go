package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"

	"golang.org/x/tools/go/packages"
)

// isDirectory reports whether the named file is a directory.
func isDirectory(name string) bool {
	info, err := os.Stat(name)
	if err != nil {
		log.Fatal(err)
	}
	return info.IsDir()
}

func SortImportPaths(importPaths []string) ([]string, []string, []string, error) {
	stdPkgs := make(map[string]struct{})

	if pkgs, err := packages.Load(nil, "std"); err == nil {
		for _, p := range pkgs {
			stdPkgs[p.PkgPath] = struct{}{}
		}
	} else {
		return nil, nil, nil, err
	}

	stdImportPaths := make([]string, 0, len(importPaths))
	internalImportPaths := make([]string, 0, len(importPaths))
	externalImportPaths := make([]string, 0, len(importPaths))

	for _, importPath := range importPaths {
		_, isStdPkg := stdPkgs[importPath]

		switch {
		case isStdPkg:
			stdImportPaths = append(stdImportPaths, importPath)
		case strings.HasPrefix(strings.Trim(importPath, `"`), `github.com/manabie-com/backend`):
			internalImportPaths = append(internalImportPaths, importPath)
		default:
			externalImportPaths = append(externalImportPaths, importPath)
		}
	}
	return stdImportPaths, internalImportPaths, externalImportPaths, nil
}

func BufferFromImportPaths(importPaths []string) ([]byte, error) {
	var buffer bytes.Buffer

	// Write import paths
	if len(importPaths) == 1 {
		printf(&buffer, "import %s", importPaths[0])
	} else {
		stdImportPaths, internalImportPaths, externalImportPaths, err := SortImportPaths(importPaths)
		if err != nil {
			return nil, err
		}

		printf(&buffer, "import (\n")

		for _, stdImportPath := range stdImportPaths {
			printf(&buffer, fmt.Sprintf("\t\"%s\"\n", stdImportPath))
		}

		if len(stdImportPaths) > 0 && len(internalImportPaths) > 0 {
			printf(&buffer, "\n")
		}
		for _, internalImportPath := range internalImportPaths {
			printf(&buffer, fmt.Sprintf("\t\"%s\"\n", internalImportPath))
		}

		if len(internalImportPaths) > 0 && len(externalImportPaths) > 0 {
			printf(&buffer, "\n")
		}
		for _, externalImportPath := range externalImportPaths {
			printf(&buffer, fmt.Sprintf("\t\"%s\"\n", externalImportPath))
		}
		printf(&buffer, ")\n")
	}

	return buffer.Bytes(), nil
}

package automation

import (
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/manabie-com/backend/deployments/decl/scripts/automation/defs"

	"github.com/pkg/errors"
)

type SQLRunner struct {
	srcFile        string
	dstDir         string
	customizations []sqlCustomization
}

func NewSQL() *SQLRunner {
	return &SQLRunner{}
}

func (r *SQLRunner) From(fpath string) *SQLRunner {
	r.srcFile = fpath
	return r
}

func (r *SQLRunner) To(dirpath string) *SQLRunner {
	r.dstDir = dirpath
	return r
}

// Customize allows adding extra content into files whose names
// contain the provided suffix namesuffix.
// For examples:
//   - namesuffix=".sql" will match all files
//   - namesuffix="bob.sql" will match 1.bob.sql, 2.bob.sql, and so on
func (r *SQLRunner) Customize(filesuffix string, stmts ...string) *SQLRunner {
	r.customizations = append(r.customizations, sqlCustomization{filesuffix: filesuffix, stmts: stmts})
	return r
}

func (r *SQLRunner) Run() error {
	svcs, err := defs.NewServicesFrom(r.srcFile)
	if err != nil {
		return err
	}
	if err := svcs.GenerateSQL(r.dstDir); err != nil {
		return errors.Wrap(err, "Services.GenerateSQL")
	}
	if err := r.applyCustomizations(r.dstDir); err != nil {
		return errors.Wrap(err, "applyCustomizations")
	}
	return nil
}

func (r SQLRunner) applyCustomizations(dir string) error {
	filelist, err := r.getSQLFilesIn(dir)
	if err != nil {
		return errors.Wrap(err, "getSQLFilesIn")
	}
	for _, c := range r.customizations {
		for _, filename := range filelist {
			if strings.HasSuffix(filename, c.filesuffix) {
				if err := c.apply(filepath.Join(dir, filename)); err != nil {
					return errors.Wrap(err, "failed to apply sql customization")
				}
			}
		}
	}

	return nil
}

var sqlRe = regexp.MustCompile(`^\d+\.\w+\.sql$`) // matches "1234.some_db_name.sql"

func (r SQLRunner) getSQLFilesIn(dir string) ([]string, error) {
	ls, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	out := []string{}
	for _, v := range ls {
		if v.IsDir() {
			continue
		}

		fname := v.Name()
		if sqlRe.MatchString(fname) {
			out = append(out, fname)
		}
	}
	return out, nil
}

type sqlCustomization struct {
	filesuffix string
	stmts      []string
}

func (s *sqlCustomization) apply(fp string) error {
	f, err := os.OpenFile(fp, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return errors.Wrap(err, "os.OpenFile")
	}
	defer f.Close()
	return s.writeTo(f)
}

func (s *sqlCustomization) writeTo(out io.Writer) error {
	for _, stmt := range s.stmts {
		_, err := out.Write([]byte(stmt + "\n"))
		if err != nil {
			return errors.Wrap(err, "io.Writer.Write")
		}
	}
	return nil
}

package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestMain(m *testing.M) {
	if os.Getenv("HEXAGEN_TEST_IS_HEXAGEN") != "" {
		main()
		os.Exit(0)
	}

	// Inform subprocesses that they should run the cmd/stringer main instead of
	// running tests. It's a close approximation to building and running the real
	// command, and much less complicated and expensive to build and clean up.

	if err := os.Setenv("HEXAGEN_TEST_IS_HEXAGEN", "1"); err != nil {
		panic(err)
	}

	os.Exit(m.Run())
}

func TestEndToEnd(t *testing.T) {
	hexagen := hexagenPath(t)

	tmpDir := "./tmp/e2e"
	if err := os.MkdirAll(tmpDir, os.ModePerm); err != nil {
		t.Fatal(err)
	}

	// Read the testdata directory.
	srcDir := "testdata/e2e"
	fd, err := os.Open(srcDir)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = fd.Close()
	}()

	names, err := fd.Readdirnames(-1)
	if err != nil {
		t.Fatalf("Readdirnames: %s", err)
	}

	// Generate, compile, and run the test programs.
	for _, name := range names {
		fmt.Println(name)
		if !strings.HasSuffix(name, ".go") {
			t.Errorf("%s is not a Go file", name)
			continue
		}
		hexagenCompileAndRun(t, tmpDir, hexagen, strings.ToLower(typeName(name)), srcDir, name)
	}
}

// a type name for stringer. use the last component of the file name with the .go
func typeName(fname string) string {
	// file names are known to be ascii and end .go
	base := path.Base(fname)
	return fmt.Sprintf("%c%s", base[0]+'A'-'a', base[1:len(base)-len(".go")])
}

var exe struct {
	path string
	err  error
	once sync.Once
}

func hexagenPath(t *testing.T) string {
	exe.once.Do(func() {
		exe.path, exe.err = os.Executable()
	})
	if exe.err != nil {
		t.Fatal(exe.err)
	}
	return exe.path
}

// hexagenCompileAndRun runs stringer for the named file and compiles and
// runs the target binary in directory dir. That binary will panic if the String method is incorrect.
func hexagenCompileAndRun(t *testing.T, destDir, stringer, typeName, srcDir string, fileName string) {
	t.Helper()
	t.Logf("run: %s %s\n", fileName, typeName)
	source := filepath.Join(destDir, path.Base(fileName))
	err := copy(source, filepath.Join(srcDir, fileName))
	if err != nil {
		t.Fatalf("copying file to temporary directory: %s", err)
	}
	stringSource := filepath.Join(destDir, typeName+"_generated_impl.go")
	/*// Run stringer in temporary directory.
	err = run(stringer, "ent-impl", "--type", utils.UpperCaseFirstLetter(typeName), "--output", strings.TrimSuffix(stringSource, "/user_generated_impl.go"), "../../modules/user/core/valueobj/valueobj.go", source)
	if err != nil {
		t.Fatal(errors.Wrap(err, "run()"))
	}*/

	if err := os.Setenv("GOPATH", ""); err != nil {
		t.Fatal(err)
	}
	// Build new executable
	err = run("go", "build", "-ldflags", fmt.Sprintf("-X main.version=%s", version+"-test"), "-o", destDir)
	if err != nil {
		t.Fatal(err)
	}

	// Run the binary in the temporary directory.
	err = run("go", "generate", "./...")
	if err != nil {
		t.Fatal(err)
	}

	// Run the binary in the temporary directory.
	err = run("go", "run", stringSource, source)
	if err != nil {
		t.Fatal(err)
	}
}

// copy copies the from file to the to file.
func copy(to, from string) error {
	toFd, err := os.Create(to)
	if err != nil {
		return err
	}
	defer func() {
		_ = toFd.Close()
	}()

	fromFd, err := os.Open(from)
	if err != nil {
		return err
	}
	defer func() {
		_ = fromFd.Close()
	}()

	_, err = io.Copy(toFd, fromFd)
	return err
}

// run runs a single command and returns an error if it does not succeed.
// os/exec should have this function, to be honest.
func run(name string, arg ...string) error {
	return runInDir(".", name, arg...)
}

// runInDir runs a single command in directory dir and returns an error if
// it does not succeed.
func runInDir(dir, name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "GO111MODULE=auto")
	return cmd.Run()
}

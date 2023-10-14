package execwrapper

import (
	"os"
	"os/exec"
)

func GoCommand(args ...string) (*exec.Cmd, error) {
	goExe, err := LookPath("go")
	if err != nil {
		return nil, err
	}
	return exec.Command(goExe, args...), nil
}

func GoTest(pkg string, args ...string) error {
	cmdWithArgs := append([]string{"test", pkg}, args...)
	testcmd, err := GoCommand(cmdWithArgs...)
	if err != nil {
		return err
	}
	testcmd.Stdout = os.Stdout
	testcmd.Stderr = os.Stderr
	err = testcmd.Run()
	return err
}

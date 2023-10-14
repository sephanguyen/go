package execwrapper

import (
	"os"
	"os/exec"
)

func CloudSQLProxyCommand(args ...string) error {
	cloudSQLProxyExe, err := LookPath("cloud_sql_proxy")
	if err != nil {
		return err
	}
	cmd := exec.Command(cloudSQLProxyExe, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

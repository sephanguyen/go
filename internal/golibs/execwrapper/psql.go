package execwrapper

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
)

func PsqlCommand(args ...string) (*exec.Cmd, error) {
	psqlExe, err := LookPath("psql")
	if err != nil {
		return nil, err
	}
	return exec.Command(psqlExe, args...), nil
}

func Psql(host, port, username, password string) error {
	cmd, err := PsqlCommand(
		"-h", host,
		"-p", port,
		"-U", username,
		"-d", "postgres", // prevent error "db does not exist"
	)
	if err != nil {
		return err
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if password != "" {
		if err = os.Setenv("PGPASSWORD", password); err != nil {
			return fmt.Errorf("failed to set PGPASSWORD: %w", err)
		}
	}

	signal.Ignore()
	return cmd.Run()
}

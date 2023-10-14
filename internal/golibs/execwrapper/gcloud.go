package execwrapper

import (
	"os"
	"os/exec"
	"strings"
)

func GCloudCommand(args ...string) (*exec.Cmd, error) {
	gcloudExe, err := LookPath("gcloud")
	if err != nil {
		return nil, err
	}
	return exec.Command(gcloudExe, args...), nil
}

func GCloudGetAccount() (string, error) {
	getCmd, err := GCloudCommand("config", "get-value", "account")
	if err != nil {
		return "", err
	}
	getCmd.Stderr = os.Stderr
	output, err := getCmd.Output()
	return firstLine(output), err
}

func GCloudPrintAccessTokenOf(impersonatedServiceAccount string) (string, error) {
	cmd, err := GCloudCommand(
		"auth", "print-access-token", "--quiet",
		"--impersonate-service-account", impersonatedServiceAccount,
	)
	if err != nil {
		return "", err
	}
	cmd.Stderr = os.Stderr
	output, err := cmd.Output()
	return strings.Trim(string(output), "\n"), err
}

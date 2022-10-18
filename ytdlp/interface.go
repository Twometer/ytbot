package ytdlp

import (
	"log"
	"os/exec"
	"strings"
)

func CheckForUpdates() error {
	res, err := runYtdl("-U")
	if err != nil {
		return err
	}

	ver, err := GetVersion()
	if err != nil {
		return err
	}

	if strings.Contains(res, "is up to date") {
		log.Printf("yt-dlp is up to date with version %s\n", ver)
	} else {
		log.Printf("yt-dlp was updated to version %s\n", ver)
	}

	return nil
}

func GetVersion() (string, error) {
	ver, err := runYtdl("--version")
	return strings.TrimSpace(ver), err
}

func GetStreamUrl() (string, error) {
	return "", nil
}

func runYtdl(args ...string) (string, error) {
	cmd := exec.Command(getExecutablePath(), args...)

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(output), nil
}

package ytdlp

import (
	"errors"
	"log"
	"net/url"
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

func GetStreamUrl(ytUrl string) (string, error) {
	result, err := runYtdl("-g", ytUrl)
	if err != nil {
		return "", err
	}

	urls := strings.Split(result, "\n")
	validUrls := make([]*url.URL, 0)
	for _, urlStr := range urls {
		urlObj, err := url.Parse(urlStr)
		if err == nil {
			validUrls = append(validUrls, urlObj)
		}
	}

	if len(validUrls) == 0 {
		return "", errors.New("could not resolve YouTube video")
	} else if len(validUrls) == 1 {
		return validUrls[0].String(), nil
	} else {
		for _, candidate := range validUrls {
			if strings.HasPrefix(candidate.Query().Get("mime"), "audio") {
				return candidate.String(), nil
			}
		}

		log.Println("warn: no audio url found, returning best effort url")
		return validUrls[0].String(), nil
	}
}

func runYtdl(args ...string) (string, error) {
	cmd := exec.Command(getExecutablePath(), args...)

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(output), nil
}

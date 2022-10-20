package ytdlp

import (
	"errors"
	"go.uber.org/zap"
	"net/url"
	"os/exec"
	"strings"
)

func CheckForUpdates() error {
	prevVer, err := GetVersion()
	if err != nil {
		return err
	}

	_, err = runYtdl("-U")
	if err != nil {
		return err
	}

	curVer, err := GetVersion()
	if err != nil {
		return err
	}

	if curVer == prevVer {
		zap.S().Infow("yt-dlp is at the latest version", "version", curVer)
	} else {
		zap.S().Infow("yt-dlp was updated to a new version", "curVer", curVer, "prevVer", prevVer)
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

		zap.S().Warnw("No download URL with audio mimetype was found, returning best effort.", "urlCandidates", validUrls)
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

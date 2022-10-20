package ytdlp

import (
	"errors"
	"go.uber.org/zap"
	"io"
	"net/http"
	"os"
)

func EnsurePresent() error {
	zap.S().Debugln("Checking if yt-dlp is present")

	downloadPath := getExecutablePath()
	zap.S().Debugln("Path for yt-dlp was loaded", "path", downloadPath)

	_, err := os.Stat(downloadPath)
	if errors.Is(err, os.ErrNotExist) {
		zap.S().Debugln("yt-dlp is not present, downloading")
		err = download(downloadPath)
		if err != nil {
			return errors.New("Failed to download yt-dlp: " + err.Error())
		}
	} else {
		zap.S().Debugln("yt-dlp is present")
	}

	zap.S().Debugln("Validating yt-dlp executable")
	validate()

	return nil
}

func download(path string) error {
	downloadUrl := getExecutableUrl()
	zap.S().Infoln("Downloading yt-dlp executable", "downloadUrl", downloadUrl)

	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(downloadUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func validate() {
	ver, err := GetVersion()
	if err != nil {
		zap.S().Fatalw("Failed to get version from yt-dlp. Executable may be corrupted.", "error", err)
	} else {
		zap.S().Infow("Found valid yt-dlp executable", "version", ver)
	}
}

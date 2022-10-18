package ytdlp

import (
	"errors"
	"io"
	"log"
	"net/http"
	"os"
)

func EnsurePresent() error {
	log.Println("Checking if yt-dlp is present...")

	downloadPath := getExecutablePath()
	log.Printf("Path for yt-dlp is `%s`\n", downloadPath)

	_, err := os.Stat(downloadPath)
	if errors.Is(err, os.ErrNotExist) {
		log.Println("yt-dlp executable is not present")
		err = download(downloadPath)
		if err != nil {
			return errors.New("Failed to download yt-dlp: " + err.Error())
		}
	} else {
		log.Println("yt-dlp executable is present")
	}

	log.Println("Checking yt-dlp executable...")
	validate()

	return nil
}

func download(path string) error {
	downloadUrl := getExecutableUrl()
	log.Printf("Downloading yt-dlp from `%s`...\n", downloadUrl)

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
		log.Fatalln("Executable may be corrupted: failed to get version from yt-dlp:", err)
	} else {
		log.Printf("Found valid yt-dlp with version %s\n", ver)
	}
}

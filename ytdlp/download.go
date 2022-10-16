package ytdlp

import (
	"errors"
	"log"
	"os"
)

func EnsurePresent() {
	log.Println("Checking if yt-dlp is present")

	downloadPath := getDownloadPath()
	log.Printf(" The path for yt-dlp is `%s`\n", downloadPath)

	_, err := os.Stat(downloadPath)
	if errors.Is(err, os.ErrNotExist) {
		log.Println("Not present, downloading...")
		download(downloadPath)
	} else {
		log.Println("Present, validating...")
		validate(downloadPath)
	}
}

func download(path string) {
	downloadUrl := getDownloadUrl()
	log.Printf(" Downloading from `%s`...\n", downloadUrl)
}

func validate(path string) {
	ver, err := GetVersion(path)
	if err != nil {
		log.Fatalln("Failed to get version from yt-dlp:", err)
	}

	log.Printf("Found valid yt-dlp with version %s", ver)
}

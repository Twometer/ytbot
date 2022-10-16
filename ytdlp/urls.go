package ytdlp

import (
	"log"
	"os"
	"path"
	"runtime"
)

const baseUrl = "https://github.com/yt-dlp/yt-dlp/releases/latest/download/"
const windowsFileName = "yt-dlp.exe"
const linuxFileName = "yt-dlp"

func getFileName() string {
	switch runtime.GOOS {
	case "windows":
		return windowsFileName
	case "linux":
		return linuxFileName
	default:
		log.Fatalf("Unsupported operating system `%s`", runtime.GOOS)
	}
	return ""
}

func getDownloadUrl() string {
	return baseUrl + getFileName()
}

func getDownloadPath() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatalln("Failed to get current working directory:", err)
	}

	return path.Join(dir, getFileName())
}

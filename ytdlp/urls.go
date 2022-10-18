package ytdlp

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
)

const baseUrl = "https://github.com/yt-dlp/yt-dlp/releases/latest/download/"
const windowsFileName = "yt-dlp.exe"
const linuxFileName = "yt-dlp"

func getExecutableFileName() string {
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

func getExecutableUrl() string {
	return baseUrl + getExecutableFileName()
}

func getExecutablePath() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatalln("Failed to get current working directory:", err)
	}

	return filepath.Join(dir, getExecutableFileName())
}

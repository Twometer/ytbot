package main

import (
	"log"
	"ytbot/ytdlp"
)

func initYtdl() {
	err := ytdlp.EnsurePresent()
	if err != nil {
		log.Fatalln("failed to ensure valid yt-dlp:", err)
	}

	err = ytdlp.CheckForUpdates()
	if err != nil {
		log.Fatalln("failed to check for updates:", err)
	}
}

func main() {
	log.Println(">> Starting YTBot <<")
	initYtdl()
}

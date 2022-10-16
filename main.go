package main

import (
	"log"
	"ytbot/ytdlp"
)

func main() {
	log.Println(">> Starting YTBot <<")
	ytdlp.EnsurePresent()
}

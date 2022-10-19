package core

import (
	"log"
	"ytbot/discord"
	"ytbot/ytapi"
)

func init() {
	RegisterCommand("ping", PingCommand)
	RegisterCommand("play", PlayCommand)
	RegisterCommand("skip", SkipCommand)
	RegisterCommand("stop", StopCommand)
	RegisterCommand("leave", StopCommand)
	RegisterCommand("move", MoveCommand)
	RegisterCommand("clear", ClearCommand)
}

func PingCommand(cmd discord.CommandBuffer, client *discord.Client) {
	client.ReplyMessage(cmd.Message, "Pong! "+cmd.GetStringAll())
}

func PlayCommand(cmd discord.CommandBuffer, client *discord.Client) {
	query := cmd.GetStringAll()
	if len(query) == 0 {
		client.ReplyMessage(cmd.Message, "A search query is required")
		return
	}

	items, err := ytapi.LoadMediaItems(query)
	if err != nil {
		client.ReplyMessage(cmd.Message, "An error occurred while loading the media items")
		log.Println("error loading media items:", err)
		return
	}

	log.Println("Added", len(items), "items to list")
}

func SkipCommand(cmd discord.CommandBuffer, client *discord.Client) {

}

func StopCommand(cmd discord.CommandBuffer, client *discord.Client) {

}

func MoveCommand(cmd discord.CommandBuffer, client *discord.Client) {

}

func ClearCommand(cmd discord.CommandBuffer, client *discord.Client) {

}

package core

import (
	"log"
	"strconv"
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

	// todo add to playlist

	if len(items) == 0 {
		client.ReplyMessage(cmd.Message, "No Results")
	} else if len(items) == 1 {
		client.ReplyMessage(cmd.Message, "Added `"+items[0].Name+"` to queue")
	} else {
		client.ReplyMessage(cmd.Message, "Added **"+strconv.Itoa(len(items))+"** items to queue")
	}
}

func SkipCommand(cmd discord.CommandBuffer, client *discord.Client) {

}

func StopCommand(cmd discord.CommandBuffer, client *discord.Client) {

}

func MoveCommand(cmd discord.CommandBuffer, client *discord.Client) {

}

func ClearCommand(cmd discord.CommandBuffer, client *discord.Client) {

}

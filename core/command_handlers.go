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
	statusMsg := client.ReplyMessage(cmd.Message, ":arrows_counterclockwise: Searching...")

	query := cmd.GetStringAll()
	if len(query) == 0 {
		client.EditMessage(statusMsg, ":x: A search query or YouTube link is required")
		return
	}

	items, err := ytapi.LoadMediaItems(query)
	if err != nil {
		client.EditMessage(statusMsg, ":x: An error occurred while running the search")
		log.Println("error loading media items:", err)
		return
	}

	// todo add to playlist

	if len(items) == 0 {
		client.EditMessage(statusMsg, ":x: No Results")
	} else if len(items) == 1 {
		client.EditMessage(statusMsg, ":white_check_mark: Added `"+items[0].Name+"` to queue")
	} else {
		client.EditMessage(statusMsg, ":white_check_mark: Added **"+strconv.Itoa(len(items))+"** items to queue")
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

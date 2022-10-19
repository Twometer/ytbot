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
	RegisterCommand("remove", RemoveCommand)
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

	botState := GetBotState(cmd.Message)
	for _, item := range items {
		botState.Queue = append(botState.Queue, item)
	}

	if len(items) == 0 {
		client.EditMessage(statusMsg, ":x: No Results")
	} else if len(items) == 1 {
		client.EditMessage(statusMsg, ":white_check_mark: Added `"+items[0].Name+"` to queue")
	} else {
		client.EditMessage(statusMsg, ":white_check_mark: Added **"+strconv.Itoa(len(items))+"** items to queue")
	}

	// todo trigger playback
}

func SkipCommand(cmd discord.CommandBuffer, client *discord.Client) {

}

func StopCommand(cmd discord.CommandBuffer, client *discord.Client) {

}

func MoveCommand(cmd discord.CommandBuffer, client *discord.Client) {
	oldIdx := cmd.GetInt() - 1
	newIdx := cmd.GetInt() - 1

	botState := GetBotState(cmd.Message)

	newQueue := make([]ytapi.MediaItem, 0)
	newQueue = append(newQueue, botState.Queue[:newIdx]...)
	newQueue = append(newQueue, botState.Queue[oldIdx])
	newQueue = append(newQueue, botState.Queue[newIdx:oldIdx]...)
	newQueue = append(newQueue, botState.Queue[oldIdx:]...)

	botState.Queue = newQueue

	client.ReplyMessage(cmd.Message, ":white_check_mark: Moved item #"+strconv.Itoa(oldIdx+1)+" to #"+strconv.Itoa(newIdx+1))
}

func ClearCommand(cmd discord.CommandBuffer, client *discord.Client) {
	GetBotState(cmd.Message).Queue = nil
	client.ReplyMessage(cmd.Message, ":white_check_mark: Queue was cleared")
}

func RemoveCommand(cmd discord.CommandBuffer, client *discord.Client) {
	index := cmd.GetInt() - 1
	botState := GetBotState(cmd.Message)
	if index < 0 || index >= len(botState.Queue) {
		client.ReplyMessage(cmd.Message, ":x: There is no item with that index")
		return
	}

	item := botState.Queue[index]

	botState.Queue = append(botState.Queue[:index], botState.Queue[index+1:]...)
	client.ReplyMessage(cmd.Message, ":white_check_mark: Item `"+item.Name+"` at position #"+strconv.Itoa(index+1)+" was removed.")
}

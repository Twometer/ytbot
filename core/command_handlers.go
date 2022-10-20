package core

import (
	"log"
	"strconv"
	"strings"
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
	RegisterCommand("queue", QueueCommand)
}

func PingCommand(cmd discord.CommandBuffer, client *discord.Client) {
	client.ReplyMessage(cmd.Message, "Pong! "+cmd.GetStringAll())
}

func PlayCommand(cmd discord.CommandBuffer, client *discord.Client) {
	voiceState, inVoiceChannel := client.Guilds[cmd.Message.GuildId].VoiceStates[cmd.Message.Author.Id]
	if !inVoiceChannel {
		client.ReplyMessage(cmd.Message, EmojiFailed+"You are not in a voice channel")
		return
	}

	statusMsg := client.ReplyMessage(cmd.Message, EmojiLoading+"Searching...")

	query := strings.TrimSpace(cmd.GetStringAll())
	if len(query) == 0 {
		client.EditMessage(statusMsg, EmojiFailed+"A search query or YouTube link is required")
		return
	}

	items, err := ytapi.LoadMediaItems(query)
	if err != nil {
		client.EditMessage(statusMsg, EmojiFailed+"An error occurred while connecting to YouTube")
		log.Println("Failed to load media items:", err)
		return
	}

	botState := GetBotState(cmd.Message)
	for _, item := range items {
		botState.Queue = append(botState.Queue, item)
	}

	if len(items) == 0 {
		client.EditMessage(statusMsg, EmojiFailed+"No results for `"+query+"`")
	} else if len(items) == 1 {
		client.EditMessage(statusMsg, EmojiSuccess+"Added `"+items[0].Name+"` to queue")
	} else {
		client.EditMessage(statusMsg, EmojiSuccess+"Added **"+strconv.Itoa(len(items))+" items** to queue")
	}

	voiceClient := client.GetVoiceClient(cmd.Message.GuildId)
	if voiceClient == nil || !voiceClient.IsPlaying() {
		log.Println("Triggering playback because voice client is idle")
		playNext(cmd, client, voiceState.GuildId, voiceState.ChannelId)
	}
}

func SkipCommand(cmd discord.CommandBuffer, client *discord.Client) {
	if voiceState, ok := client.Guilds[cmd.Message.GuildId].VoiceStates[cmd.Message.Author.Id]; ok {
		playNext(cmd, client, voiceState.GuildId, voiceState.ChannelId)
	} else {
		client.ReplyMessage(cmd.Message, EmojiFailed+"You are not in a voice channel")
	}
}

func StopCommand(cmd discord.CommandBuffer, client *discord.Client) {
	client.LeaveVoiceChannel(cmd.Message.GuildId)
	botState := GetBotState(cmd.Message)
	botState.Queue = nil
	botState.Encoder.Stop()
	client.ReplyMessage(cmd.Message, EmojiStop+"Stopped playback and left the voice channel")
}

func MoveCommand(cmd discord.CommandBuffer, client *discord.Client) {
	oldIdx := cmd.GetIntOrDefault(-1) - 1
	newIdx := cmd.GetIntOrDefault(-1) - 1

	botState := GetBotState(cmd.Message)

	if oldIdx < 0 || oldIdx >= len(botState.Queue) || newIdx < 0 || newIdx >= len(botState.Queue) {
		client.ReplyMessage(cmd.Message, EmojiFailed+"There is no item at that position")
		return
	}

	newQueue := make([]ytapi.MediaItem, 0)
	newQueue = append(newQueue, botState.Queue[:newIdx]...)
	newQueue = append(newQueue, botState.Queue[oldIdx])
	newQueue = append(newQueue, botState.Queue[newIdx:oldIdx]...)
	newQueue = append(newQueue, botState.Queue[oldIdx+1:]...)

	botState.Queue = newQueue

	client.ReplyMessage(cmd.Message, EmojiSuccess+"Moved item #"+strconv.Itoa(oldIdx+1)+" to #"+strconv.Itoa(newIdx+1))
}

func ClearCommand(cmd discord.CommandBuffer, client *discord.Client) {
	GetBotState(cmd.Message).Queue = nil
	client.ReplyMessage(cmd.Message, EmojiSuccess+"Queue was cleared")
}

func RemoveCommand(cmd discord.CommandBuffer, client *discord.Client) {
	index := cmd.GetInt() - 1
	botState := GetBotState(cmd.Message)
	if index < 0 || index >= len(botState.Queue) {
		client.ReplyMessage(cmd.Message, EmojiFailed+"There is no item with that index")
		return
	}

	item := botState.Queue[index]

	botState.Queue = append(botState.Queue[:index], botState.Queue[index+1:]...)
	client.ReplyMessage(cmd.Message, EmojiSuccess+"Item `"+item.Name+"` at position #"+strconv.Itoa(index+1)+" was removed.")
}

func QueueCommand(cmd discord.CommandBuffer, client *discord.Client) {
	queue := GetBotState(cmd.Message).Queue
	pageIdx := cmd.GetIntOrDefault(1) - 1
	offset := pageIdx * 10

	if len(queue) == 0 {
		client.ReplyMessage(cmd.Message, EmojiNeutral+"The queue is empty")
	} else {
		var lines []string

		rangeMin := max(offset, 0)
		rangeMax := min(offset+10, len(queue))

		if rangeMax <= 0 || rangeMin >= len(queue) {
			client.ReplyMessage(cmd.Message, EmojiFailed+"There is no page "+strconv.Itoa(pageIdx+1))
			return
		}

		for idx, item := range queue[rangeMin:rangeMax] {
			lines = append(lines, "**#"+strconv.Itoa(idx+1+offset)+"**: `"+item.Name+"`")
		}

		client.ReplyMessage(cmd.Message, "__Playback queue (page "+strconv.Itoa(pageIdx+1)+")__\n"+strings.Join(lines, "\n"))
	}

}

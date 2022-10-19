package core

import (
	"log"
	"strconv"
	"ytbot/codec"
	"ytbot/discord"
	"ytbot/ytapi"
	"ytbot/ytdlp"
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
	voiceState, inVoiceChannel := client.Guilds[cmd.Message.GuildId].VoiceStates[cmd.Message.Author.Id]
	if !inVoiceChannel {
		client.ReplyMessage(cmd.Message, EmojiFailed+"You are not in a voice channel")
		return
	}

	statusMsg := client.ReplyMessage(cmd.Message, EmojiLoading+"Searching...")

	query := cmd.GetStringAll()
	if len(query) == 0 {
		client.EditMessage(statusMsg, EmojiFailed+"A search query or YouTube link is required")
		return
	}

	items, err := ytapi.LoadMediaItems(query)
	if err != nil {
		client.EditMessage(statusMsg, EmojiFailed+"An error occurred while running the search")
		log.Println("error loading media items:", err)
		return
	}

	botState := GetBotState(cmd.Message)
	for _, item := range items {
		botState.Queue = append(botState.Queue, item)
	}

	if len(items) == 0 {
		client.EditMessage(statusMsg, EmojiFailed+"No Results")
	} else if len(items) == 1 {
		client.EditMessage(statusMsg, EmojiSuccess+"Added `"+items[0].Name+"` to queue")
	} else {
		client.EditMessage(statusMsg, EmojiSuccess+"Added **"+strconv.Itoa(len(items))+"** items to queue")
	}

	voiceClient := client.GetVoiceClient(cmd.Message.GuildId)
	if voiceClient == nil || !voiceClient.IsPlaying() {
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

func playNext(cmd discord.CommandBuffer, client *discord.Client, guildId string, channelId string) {
	state := GetBotState(cmd.Message)
	if len(state.Queue) == 0 {
		log.Println("playback queue is empty")
		return
	}

	nextSong := state.Queue[0]
	state.Queue = state.Queue[1:]

	statusMsg := client.ReplyMessage(cmd.Message, EmojiLoading+"Preparing to play `"+nextSong.Name+"`...")

	log.Println("Fetching youtube url for", nextSong.Name+"...")
	url, err := ytdlp.GetStreamUrl(nextSong.Url)
	if err != nil {
		log.Println("Failed to get YouTube stream url:", err)
		client.EditMessage(statusMsg, EmojiFailed+"Failed to get YouTube stream URL")
		return
	}

	log.Println("Starting Discord playback...")
	voiceClient, err := client.JoinVoiceChannel(guildId, channelId)
	if err != nil {
		log.Println("Failed to join voice channel:", err)
		client.EditMessage(statusMsg, EmojiFailed+"Failed to join voice channel")
		return
	}

	if voiceClient.IsPlaying() {
		if state.Encoder != nil {
			state.Encoder.Stop()
			log.Println("current encoder stopped")
		} else {
			log.Println("Error: Voice client is playing but encoder is not present")
			client.EditMessage(statusMsg, EmojiFailed+"Failed to stop current playback")
			return
		}
	}

	if !voiceClient.IsReady() {
		log.Println("Waiting for VoiceClient to be ready...")
		for event := range voiceClient.Events {
			if event == discord.VoiceEventReady {
				break
			}
		}
	}

	log.Println("Starting encoder...")
	state.Encoder = codec.NewEncoder(url, voiceClient.VoiceStream)
	err = state.Encoder.Start()
	if err != nil {
		log.Println("Failed to start encoder:", err)
		client.EditMessage(statusMsg, EmojiFailed+"Failed to start audio stream")
		return
	}

	client.EditMessage(statusMsg, EmojiSuccess+"Now playing: `"+nextSong.Name+"`.")
	log.Println("New song started playing (hopefully)")

	go func() {
		log.Println("Waiting for current playback to finish ...")
		for event := range voiceClient.Events {
			if event == discord.VoiceEventFinished {
				log.Println("Playback finished gracefully, starting next one")
				go playNext(cmd, client, guildId, channelId)
				return
			} else if event == discord.VoiceEventError {
				log.Println("Playback finished with error, sending error message")
				client.ReplyMessage(statusMsg, EmojiFailed+"Something went wrong during playback")
				return
			} else if event == discord.VoiceEventStopped {
				log.Println("Playback was stopped")
				return
			}
		}
	}()
}

func StopCommand(cmd discord.CommandBuffer, client *discord.Client) {
	client.LeaveVoiceChannel(cmd.Message.GuildId)
	botState := GetBotState(cmd.Message)
	botState.Queue = nil
	botState.Encoder.Stop()
	client.ReplyMessage(cmd.Message, EmojiSuccess+"Stopped playback and left the voice channel")
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

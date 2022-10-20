package core

import (
	"log"
	"ytbot/codec"
	"ytbot/discord"
	"ytbot/ytdlp"
)

func playNext(cmd discord.CommandBuffer, client *discord.Client, guildId string, channelId string) {
	state := GetBotState(cmd.Message)
	if len(state.Queue) == 0 {
		log.Println("Playback queue is empty, exiting")
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
			log.Println("Current audio encoder stopped")
		} else {
			log.Println("Error: Voice client is playing but encoder is not present")
			client.EditMessage(statusMsg, EmojiFailed+"Failed to stop current playback")
			return
		}
	}

	if !voiceClient.IsReady() {
		log.Println("Waiting for voice client to become ready...")
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

	client.EditMessage(statusMsg, EmojiPlay+"Now playing: `"+nextSong.Name+"`.")
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
				client.LeaveVoiceChannel(cmd.Message.GuildId)
				return
			} else if event == discord.VoiceEventStopped {
				log.Println("Playback was stopped")
				return
			}
		}
	}()
}

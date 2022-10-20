package core

import (
	"go.uber.org/zap"
	"ytbot/codec"
	"ytbot/discord"
	"ytbot/ytdlp"
)

func playNext(cmd discord.CommandBuffer, client *discord.Client, guildId string, channelId string) {
	state := GetBotState(cmd.Message)
	if len(state.Queue) == 0 {
		zap.S().Debugln("Playback queue is empty, exiting from playNext()")
		return
	}

	nextSong := state.Queue[0]
	state.Queue = state.Queue[1:]

	statusMsg := client.ReplyMessage(cmd.Message, EmojiLoading+"Preparing to play `"+nextSong.Name+"`...")

	zap.S().Debugw("Fetching YouTube streaming URL", "mediaName", nextSong.Name, "mediaUrl", nextSong.Url)
	url, err := ytdlp.GetStreamUrl(nextSong.Url)
	if err != nil {
		zap.S().Errorw("Failed to get YouTube streaming URL", "mediaName", nextSong.Name, "error", err)
		client.EditMessage(statusMsg, EmojiFailed+"Failed to get YouTube stream URL")
		return
	}

	zap.S().Debugln("Joining voice channel")
	voiceClient, err := client.JoinVoiceChannel(guildId, channelId)
	if err != nil {
		zap.S().Errorw("Failed to join voice channel", "guildId", guildId, "channelId", channelId, "error", err)
		client.EditMessage(statusMsg, EmojiFailed+"Failed to join voice channel")
		return
	}

	if voiceClient.IsPlaying() {
		if state.Encoder != nil {
			state.Encoder.Stop()
			zap.S().Debugln("Current audio encoder stopped to make space for new playback")
		} else {
			zap.S().Errorln("Failed to stop playback because voice client is playing, but encoder was not found")
			client.EditMessage(statusMsg, EmojiFailed+"Failed to stop current playback")
			return
		}
	}

	if !voiceClient.IsReady() {
		zap.S().Debugln("Waiting for voice client to become ready")
		for event := range voiceClient.Events {
			if event == discord.VoiceEventReady {
				break
			}
		}
	}

	zap.S().Debugw("Starting encoder for a media item", "mediaName", nextSong.Name)
	state.Encoder = codec.NewEncoder(url, voiceClient.VoiceStream)
	err = state.Encoder.Start()
	if err != nil {
		zap.S().Errorw("Failed to start encoder for a media item", "mediaName", nextSong.Name)
		client.EditMessage(statusMsg, EmojiFailed+"Failed to start audio stream")
		return
	}

	client.EditMessage(statusMsg, EmojiPlay+"Now playing: `"+nextSong.Name+"`.")
	zap.S().Infow("A new media item started playing", "guildId", guildId, "mediaName", nextSong.Name)

	go func() {
		zap.S().Debugln("Waiting for playback to finish")
		for event := range voiceClient.Events {
			if event == discord.VoiceEventFinished {
				zap.S().Debugln("Playback finished gracefully, starting next one")
				go playNext(cmd, client, guildId, channelId)
				return
			} else if event == discord.VoiceEventError {
				zap.S().Warnw("Playback finished with error, sending error message", "mediaName", nextSong.Name)
				client.ReplyMessage(statusMsg, EmojiFailed+"Something went wrong during playback")
				client.LeaveVoiceChannel(cmd.Message.GuildId)
				return
			} else if event == discord.VoiceEventStopped {
				zap.S().Debugln("Playback was stopped, not starting next one")
				return
			}
		}
	}()
}

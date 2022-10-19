package main

import (
	"log"
	"ytbot/codec"
	"ytbot/config"
	"ytbot/discord"
	"ytbot/ytdlp"
)

func main() {
	log.Println(">> Starting YTBot <<")
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	err := ytdlp.EnsurePresent()
	if err != nil {
		log.Fatalln("failed to ensure valid yt-dlp:", err)
	}

	err = ytdlp.CheckForUpdates()
	if err != nil {
		log.Fatalln("failed to check for yt-dlp updates:", err)
	}

	discordClient := discord.NewClient(config.GetString(config.KeyAuthToken), '.')
	discordClient.AddIntent(discord.IntentGuilds)
	discordClient.AddIntent(discord.IntentVoiceStates)
	discordClient.AddIntent(discord.IntentMessages)
	discordClient.AddIntent(discord.IntentMessageContent)

	err = discordClient.Start()
	if err != nil {
		log.Fatalln("failed to start Discord client:", err)
	}

	for cmd := range discordClient.Commands {
		name := cmd.GetString()
		log.Printf("Handling command %s", name)
		if name == "ping" {
			discordClient.ReplyMessage(cmd.Message, "Pong!")
		} else if name == "join" {
			voiceState, ok := discordClient.Guilds[cmd.Message.GuildId].VoiceStates[cmd.Message.Author.Id]
			if ok {
				client, err := discordClient.JoinVoiceChannel(voiceState)
				if err != nil {
					log.Println("Failed to join voice channel:", err)
				}

				<-client.Ready
				log.Println("Voice client ready")

				encoder := codec.NewEncoder("https://data.twometer.de/video/crab_rave.mp4", client.VoiceStream)
				err = encoder.Start()
				if err != nil {
					log.Println("Failed to start encoder:", err)
				}
			} else {
				discordClient.ReplyMessage(cmd.Message, "You are not in a voice channel")
			}
		}
	}
}

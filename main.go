package main

import (
	"go.uber.org/zap"
	"ytbot/config"
	"ytbot/core"
	"ytbot/discord"
	"ytbot/ytdlp"
)

func main() {
	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)

	zap.S().Infoln("Starting YTBot")

	err := ytdlp.EnsurePresent()
	if err != nil {
		zap.S().Fatalw("Failed to ensure a valid yt-dlp is present",
			"error", err,
		)
	}

	err = ytdlp.CheckForUpdates()
	if err != nil {
		zap.S().Fatalw("Failed to check for yt-dlp updates",
			"error", err,
		)
	}

	discordClient := discord.NewClient(config.GetString(config.KeyAuthToken), '.')
	discordClient.AddIntent(discord.IntentGuilds)
	discordClient.AddIntent(discord.IntentVoiceStates)
	discordClient.AddIntent(discord.IntentMessages)
	discordClient.AddIntent(discord.IntentMessageContent)

	zap.S().Debugln("Starting Discord client")
	err = discordClient.Start()
	if err != nil {
		zap.S().Fatalw("Failed to start Discord client",
			"error", err,
		)
	}

	zap.S().Debugln("Starting command handler")
	for cmd := range discordClient.Commands {
		core.HandleCommand(cmd, discordClient)
	}
}

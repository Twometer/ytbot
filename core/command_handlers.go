package core

import "ytbot/discord"

func init() {
	RegisterCommand("ping", PingCommand)
}

func PingCommand(cmd discord.CommandBuffer, client *discord.Client) {
	client.ReplyMessage(cmd.Message, "Pong! "+cmd.GetStringAll())
}

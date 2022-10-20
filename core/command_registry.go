package core

import (
	"log"
	"strings"
	"ytbot/discord"
)

type CommandHandler = func(cmd discord.CommandBuffer, client *discord.Client)

var commands = make(map[string]CommandHandler)

func RegisterCommand(name string, handler CommandHandler) {
	commands[strings.ToLower(name)] = handler
}

func HandleCommand(cmd discord.CommandBuffer, client *discord.Client) {
	name := cmd.GetString()
	handler, ok := commands[strings.ToLower(name)]
	log.Println("Handling command " + name)
	if ok {
		handler(cmd, client)
	} else {
		client.ReplyMessage(cmd.Message, "Unknown command `"+name+"`")
	}
}

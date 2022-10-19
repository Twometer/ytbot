package core

import (
	"ytbot/codec"
	"ytbot/discord"
	"ytbot/ytapi"
)

type BotState struct {
	Queue   []ytapi.MediaItem
	Encoder *codec.Encoder
}

var botStates = make(map[string]*BotState)

func GetBotState(msg discord.Message) *BotState {
	if botState, ok := botStates[msg.GuildId]; ok {
		return botState
	} else {
		botState := &BotState{}
		botStates[msg.GuildId] = botState
		return botState
	}
}

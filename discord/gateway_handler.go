package discord

import (
	"log"
	"time"
)

func (client *Client) handleMessage(in WsMessageIn) {
	switch in.Opcode {
	case GatewayOpDispatch:
		client.sequence = in.Sequence
		client.handleEvent(in)
	case GatewayOpHello:
		var message GatewayHelloMessage
		in.Unmarshal(&message)
		client.startHeartbeat(time.Millisecond * time.Duration(message.HeartbeatInterval))
	case GatewayOpInvalidSession:
		log.Fatalln("gateway client session was invalidated")
	}
}

func (client *Client) handleEvent(in WsMessageIn) {
	switch in.Type {
	case GatewayEventReady:
		var message GatewayReadyMessage
		in.Unmarshal(&message)
		log.Printf("Logged in as %s#%s\n", message.User.Username, message.User.Discriminator)
		client.userId = message.User.Id
	case GatewayEventGuildCreate:
		var message GatewayGuildCreateMessage
		in.Unmarshal(&message)

		voiceStateMap := make(map[string]VoiceState)
		for _, state := range message.VoiceStates {
			state.GuildId = message.Id
			voiceStateMap[state.UserId] = state
		}

		client.Guilds[message.Id] = GuildState{
			Id:          message.Id,
			Name:        message.Name,
			VoiceStates: voiceStateMap,
		}
	case GatewayEventVoiceStateUpdate:
		var state VoiceState
		in.Unmarshal(&state)

		client.Guilds[state.GuildId].VoiceStates[state.UserId] = state
	case GatewayEventMessageCreate:
		var message Message
		in.Unmarshal(&message)

		if message.Content[0] == client.cmdPrefix {
			client.Commands <- NewCommandBuffer(message)
		}
	case GatewayEventVoiceServerUpdate:
		var voiceServer VoiceServer
		in.Unmarshal(&voiceServer)

		client.VoiceServers <- voiceServer
	default:
		log.Println("Unhandled event:", in.String())
	}
}

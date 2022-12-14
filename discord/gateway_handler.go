package discord

import (
	"go.uber.org/zap"
	"time"
)

func (client *Client) handleMessage(in WsMessageIn) {
	if in.Data == nil {
		return
	}

	switch in.Opcode {
	case GatewayOpDispatch:
		client.sequence = in.Sequence
		client.handleEvent(in)
	case GatewayOpHello:
		var message GatewayHelloMessage
		in.Unmarshal(&message)
		client.ws.StartHeartbeat(time.Millisecond*time.Duration(message.HeartbeatInterval), func() WsMessageOut {
			return WsMessageOut{Opcode: GatewayOpHeartbeat, Data: client.sequence}
		})
	case GatewayOpInvalidSession:
		zap.S().Fatalln("The session is invalid")
	case GatewayOpReconnect:
		zap.S().Debugln("A reconnect was requested by the gateway")
		client.ws.Reconnect()
	}
}

func (client *Client) handleEvent(in WsMessageIn) {
	switch in.Type {
	case GatewayEventReady:
		var message GatewayReadyMessage
		in.Unmarshal(&message)
		zap.S().Infof("Logged in as %s#%s", message.User.Username, message.User.Discriminator)
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

		if len(message.Content) > 0 && message.Content[0] == client.cmdPrefix {
			client.Commands <- NewCommandBuffer(message)
		}
	case GatewayEventMessageUpdate:
		// ignore
	case GatewayEventVoiceServerUpdate:
		var voiceServer VoiceServer
		in.Unmarshal(&voiceServer)

		client.VoiceServers <- voiceServer
	default:
		zap.S().Debugw("Unhandled gateway event", "event", in.String())
	}
}

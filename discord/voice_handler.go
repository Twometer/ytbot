package discord

import (
	"go.uber.org/zap"
	"time"
	"ytbot/discord/utils"
)

func (vc *VoiceClient) handleMessage(message WsMessageIn) {
	if message.Data == nil {
		return
	}

	switch message.Opcode {
	case VoiceOpHello:
		var msg VoiceHelloMessage
		message.Unmarshal(&msg)
		vc.ws.StartHeartbeat(time.Millisecond*time.Duration(msg.HeartbeatInterval), func() WsMessageOut {
			return WsMessageOut{Opcode: VoiceOpHeartbeat, Data: utils.NewNonce()}
		})
	case VoiceOpReady:
		var msg VoiceReadyMessage
		message.Unmarshal(&msg)
		err := vc.createVoiceStream(msg)
		if err != nil {
			zap.S().Errorw("Failed to create a voice stream", "error", err)
			vc.Events <- VoiceEventError
		}
	case VoiceOpSessionDesc:
		var msg VoiceSessionDescriptionMessage
		message.Unmarshal(&msg)
		vc.VoiceStream.FinishSetup(msg.SecretKey)
		vc.Events <- VoiceEventReady
		vc.ready = true
	case VoiceOpHeartbeatAck:
	default:
		zap.S().Debugw("Unhandled voice event", "event", message.String())
	}
}

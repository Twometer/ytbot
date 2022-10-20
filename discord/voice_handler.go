package discord

import (
	"log"
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
			log.Println("Failed to init voice stream:", err)
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
		log.Println("Unhandled voice message:", message.String())
	}
}

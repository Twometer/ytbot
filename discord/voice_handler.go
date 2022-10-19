package discord

import (
	"log"
	"time"
	"ytbot/discord/utils"
)

func (vc *VoiceClient) handleMessage(message WsMessageIn) {
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
		err := vc.initVoiceStream(msg)
		if err != nil {
			log.Println("Failed to init voice stream:", err)
		}
	case VoiceOpSessionDesc:
		var msg VoiceSessionDescriptionMessage
		message.Unmarshal(&msg)
		vc.VoiceStream.FinishSetup(msg.SecretKey)
		vc.Ready <- true
	case VoiceOpHeartbeatAck:
	default:
		log.Println("unhandled voice message:", message.String())
	}
}

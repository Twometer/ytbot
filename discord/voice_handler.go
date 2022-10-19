package discord

import (
	"log"
	"time"
)

func (vc *VoiceClient) handleMessage(message WsMessageIn) {
	switch message.Opcode {
	case VoiceOpHello:
		var msg VoiceHelloMessage
		message.Unmarshal(&msg)
		vc.startHeartbeat(time.Millisecond * time.Duration(msg.HeartbeatInterval))
	case VoiceOpReady:
		var msg VoiceReadyMessage
		message.Unmarshal(&msg)
		vc.initVoiceStream(msg)
	case VoiceOpSessionDesc:
		var msg VoiceSessionDescriptionMessage
		message.Unmarshal(&msg)
		vc.VoiceStream.FinishSetup(msg.SecretKey)
	case VoiceOpHeartbeatAck:
	// ignore
	default:
		log.Println("unhandled voice message:", message.String())
	}
}

func (vc *VoiceClient) startHeartbeat(interval time.Duration) {
	if vc.heartbeat == nil {
		vc.heartbeat = time.NewTicker(interval)
	} else {
		vc.heartbeat.Reset(interval)
	}

	go func() {
		vc.sendHeartbeat()
		for range vc.heartbeat.C {
			vc.sendHeartbeat()
		}
	}()
}

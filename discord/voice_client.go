package discord

import (
	"errors"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
)

const preferredEncryptionMode = "xsalsa20_poly1305"

// VoiceClient represents a WebSocket connection to the voice gateway. It manages an associated VoiceStream
type VoiceClient struct {
	VoiceStream *VoiceStream
	Events      chan VoiceEvent

	ws        *WebSocket
	userId    string
	sessionId string
	server    VoiceServer
	ready     bool
}

func NewVoiceClient(userId string, sessionId string, server VoiceServer) *VoiceClient {
	return &VoiceClient{
		userId:    userId,
		sessionId: sessionId,
		server:    server,
		Events:    make(chan VoiceEvent, 25),
	}
}

func (vc *VoiceClient) start() error {
	ws, err := OpenWebSocket("wss://"+vc.server.Endpoint+"?v=4", "Voice", false)
	if err != nil {
		return err
	}
	vc.ws = ws

	vc.sendIdentify()
	go vc.handlerLoop()

	return nil
}

func (vc *VoiceClient) IsPlaying() bool {
	return vc.VoiceStream != nil && vc.VoiceStream.playing
}

func (vc *VoiceClient) IsReady() bool {
	return vc.ready
}

func (vc *VoiceClient) handlerLoop() {
	defer zap.S().Debugln("Voice handler loop exited")
	for {
		select {
		case msg := <-vc.ws.MessagesIn:
			vc.handleMessage(msg)
		case event := <-vc.ws.Events:
			if event == WsEventError {
				vc.Events <- VoiceEventError
				return
			}
		}
	}
}

func (vc *VoiceClient) createVoiceStream(msg VoiceReadyMessage) error {
	zap.S().Debugln("Connected to voice gateway. Initializing voice stream")

	if !slices.Contains(msg.Modes, preferredEncryptionMode) {
		return errors.New("remote does not support preferred encryption mode: " + preferredEncryptionMode)
	}

	stream := NewVoiceStream(vc, msg.Ip, msg.Port, msg.Ssrc)
	err := stream.BeginSetup()
	if err != nil {
		return err
	}
	vc.VoiceStream = stream

	vc.sendSelectProtocol()

	return nil
}

func (vc *VoiceClient) sendIdentify() {
	vc.ws.Send(VoiceOpIdentify, VoiceIdentifyMessage{
		ServerId:  vc.server.GuildId,
		UserId:    vc.userId,
		SessionId: vc.sessionId,
		Token:     vc.server.Token,
	})
}

func (vc *VoiceClient) sendSpeaking(speaking bool) {
	speakingValue := 0
	if speaking {
		speakingValue = 1
	}

	vc.ws.Send(VoiceOpSpeaking, VoiceSpeakingMessage{
		Speaking: speakingValue,
		Delay:    0,
		Ssrc:     vc.VoiceStream.Ssrc,
	})
}

func (vc *VoiceClient) sendSelectProtocol() {
	vc.ws.Send(VoiceOpSelectProtocol, VoiceSelectProtocolMessage{
		Protocol: "udp",
		Data: ProtocolData{
			Address: vc.VoiceStream.LocalIp,
			Port:    vc.VoiceStream.LocalPort,
			Mode:    preferredEncryptionMode,
		},
	})
}

func (vc *VoiceClient) Close() {
	if vc.VoiceStream != nil {
		vc.VoiceStream.Close()
	}
	vc.ws.Close()
}

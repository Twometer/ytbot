package discord

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/websocket"
	"golang.org/x/exp/slices"
	"log"
	"time"
	"ytbot/discord/utils"
)

const preferredEncryptionMode = "xsalsa20_poly1305"

type VoiceClient struct {
	VoiceStream *VoiceStream
	userId      string
	sessionId   string
	server      VoiceServer
	conn        *websocket.Conn
	heartbeat   *time.Ticker
}

func NewVoiceClient(userId string, sessionId string, server VoiceServer) VoiceClient {
	return VoiceClient{
		userId:    userId,
		sessionId: sessionId,
		server:    server,
		conn:      nil,
	}
}

func (vc *VoiceClient) Start() error {
	conn, _, err := websocket.DefaultDialer.Dial("wss://"+vc.server.Endpoint+"?v=4", nil)
	if err != nil {
		return err
	}
	vc.conn = conn

	err = vc.sendIdentify()
	if err != nil {
		return err
	}

	go vc.receiveLoop()

	return nil
}

func (vc *VoiceClient) receiveLoop() {
	for {
		_, data, err := vc.conn.ReadMessage()
		if err != nil {
			log.Fatalln("failed to read from WebSocket:", err)
		}

		var message WsMessageIn
		err = json.Unmarshal(data, &message)
		if err != nil {
			log.Println("error: failed to decode JSON:", err)
			continue
		}

		vc.handleMessage(message)
	}
}

func (vc *VoiceClient) sendIdentify() error {
	return vc.sendMessage(VoiceOpIdentify, VoiceIdentifyMessage{
		ServerId:  vc.server.GuildId,
		UserId:    vc.userId,
		SessionId: vc.sessionId,
		Token:     vc.server.Token,
	})
}

func (vc *VoiceClient) sendHeartbeat() error {
	return vc.sendMessage(VoiceOpHeartbeat, utils.NewNonce())
}

func (vc *VoiceClient) sendMessage(opcode VoiceOp, data interface{}) error {
	return vc.conn.WriteJSON(WsMessageOut{
		Opcode: opcode,
		Data:   data,
	})
}

func (vc *VoiceClient) initVoiceStream(msg VoiceReadyMessage) error {
	log.Println("Connected to voice gateway, initializing voice stream...")

	if !slices.Contains(msg.Modes, preferredEncryptionMode) {
		return errors.New("remote does not support preferred encryption mode: " + preferredEncryptionMode)
	}

	stream := NewVoiceStream(msg.Ip, msg.Port, msg.Ssrc)

	err := stream.BeginSetup()
	if err != nil {
		return err
	}

	err = vc.sendMessage(VoiceOpSelectProtocol, VoiceSelectProtocolMessage{
		Protocol: "udp",
		Data: ProtocolData{
			Address: stream.LocalIp,
			Port:    stream.LocalPort,
			Mode:    preferredEncryptionMode,
		},
	})
	if err != nil {
		return err
	}

	vc.VoiceStream = stream
	return nil
}

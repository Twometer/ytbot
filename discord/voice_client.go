package discord

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"time"
	"ytbot/discord/utils"
)

type VoiceClient struct {
	Stream    *VoiceStream
	userId    string
	sessionId string
	server    VoiceServer
	conn      *websocket.Conn
	heartbeat *time.Ticker
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
			log.Fatalln("failed to decode JSON:", err)
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

func (vc *VoiceClient) startVoiceStream(msg VoiceReadyMessage) {
	log.Println("Connected to voice gateway, opening voice stream...")
	vc.Stream = NewVoiceStream(msg.Ip, msg.Port, msg.Ssrc)
	err := vc.Stream.BeginSetup()
	if err != nil {
		log.Println("Failed to open voice stream:", err)
		return
	}

	err = vc.sendMessage(VoiceOpSelectProtocol, VoiceSelectProtocolMessage{
		Protocol: "udp",
		Data: ProtocolData{
			Address: vc.Stream.LocalIp,
			Port:    vc.Stream.LocalPort,
			Mode:    "xsalsa20_poly1305_lite",
		},
	})
	if err != nil {
		log.Println("Failed to select voice protocol:", err)
		return
	}
}

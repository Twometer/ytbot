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
	Ready       chan interface{}
	sendQueue   chan WsMessageOut
}

func NewVoiceClient(userId string, sessionId string, server VoiceServer) *VoiceClient {
	return &VoiceClient{
		userId:    userId,
		sessionId: sessionId,
		server:    server,
		conn:      nil,
		Ready:     make(chan interface{}),
		sendQueue: make(chan WsMessageOut),
	}
}

func (vc *VoiceClient) start() error {
	conn, _, err := websocket.DefaultDialer.Dial("wss://"+vc.server.Endpoint+"?v=4", nil)
	if err != nil {
		return err
	}
	vc.conn = conn

	go vc.receiveLoop()
	go vc.sendLoop()

	vc.sendIdentify()

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

func (vc *VoiceClient) sendLoop() {
	for messageOut := range vc.sendQueue {
		err := vc.conn.WriteJSON(messageOut)
		if err != nil {
			log.Println("Failed to dispatch message:", messageOut)
		}
	}
}

func (vc *VoiceClient) sendIdentify() {
	vc.enqueueMessage(VoiceOpIdentify, VoiceIdentifyMessage{
		ServerId:  vc.server.GuildId,
		UserId:    vc.userId,
		SessionId: vc.sessionId,
		Token:     vc.server.Token,
	})
}

func (vc *VoiceClient) sendSpeaking(speaking bool) {
	speakingValue := 0
	if speaking {
		speakingValue = 5
	}

	vc.enqueueMessage(VoiceOpSpeaking, VoiceSpeakingMessage{
		Speaking: speakingValue,
		Delay:    0,
		Ssrc:     vc.VoiceStream.Ssrc,
	})
}

func (vc *VoiceClient) sendHeartbeat() {
	vc.enqueueMessage(VoiceOpHeartbeat, utils.NewNonce())
}

func (vc *VoiceClient) enqueueMessage(opcode VoiceOp, data interface{}) {
	vc.sendQueue <- WsMessageOut{
		Opcode: opcode,
		Data:   data,
	}
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

	vc.enqueueMessage(VoiceOpSelectProtocol, VoiceSelectProtocolMessage{
		Protocol: "udp",
		Data: ProtocolData{
			Address: stream.LocalIp,
			Port:    stream.LocalPort,
			Mode:    preferredEncryptionMode,
		},
	})
	vc.VoiceStream = stream

	go func() {
		for state := range stream.StateChanges {
			log.Println("New Stream State:", state)
			if state == StatePlaying {
				vc.sendSpeaking(true)
			} else {
				vc.sendSpeaking(false)
			}
		}
	}()

	return nil
}

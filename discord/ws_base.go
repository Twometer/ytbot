package discord

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"time"
)

type HeartbeatProvider = func() WsMessageOut

type WebSocket struct {
	MessagesOut chan WsMessageOut
	MessagesIn  chan WsMessageIn

	conn              *websocket.Conn
	heartbeat         *time.Ticker
	heartbeatProvider HeartbeatProvider
	closeChan         chan bool
}

func OpenWebSocket(url string) (*WebSocket, error) {
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}

	ws := &WebSocket{
		MessagesOut: make(chan WsMessageOut, 25),
		MessagesIn:  make(chan WsMessageIn, 25),
		conn:        conn,
		closeChan:   make(chan bool),
	}

	go ws.runReceiveLoop()
	go ws.runSendLoop()

	return ws, nil
}

func (ws *WebSocket) Send(opcode int, data interface{}) {
	ws.MessagesOut <- WsMessageOut{
		Opcode: opcode,
		Data:   data,
	}
}

func (ws *WebSocket) Receive() WsMessageIn {
	return <-ws.MessagesIn
}

func (ws *WebSocket) StartHeartbeat(interval time.Duration, provider HeartbeatProvider) {
	if ws.heartbeat == nil {
		ws.heartbeat = time.NewTicker(interval)
	} else {
		ws.heartbeat.Reset(interval)
	}
	ws.heartbeatProvider = provider
}

func (ws *WebSocket) Close() {
	if ws.heartbeat != nil {
		ws.heartbeat.Stop()
	}

	err := ws.conn.Close()
	if err != nil {
		log.Println("Failed to close WebSocket gracefully:", err)
	}

	close(ws.closeChan)
	close(ws.MessagesOut)
	close(ws.MessagesIn)
}

func (ws *WebSocket) runSendLoop() {
	for {
		if ws.heartbeat != nil {
			select {
			case msg := <-ws.MessagesOut:
				ws.sendMessage(msg)
			case <-ws.heartbeat.C:
				ws.sendHeartbeat()
			case <-ws.closeChan:
				return
			}
		} else {
			select {
			case msg := <-ws.MessagesOut:
				ws.sendMessage(msg)
			case <-ws.closeChan:
				return
			}
		}
	}
}

func (ws *WebSocket) sendHeartbeat() {
	ws.sendMessage(ws.heartbeatProvider())
}

func (ws *WebSocket) sendMessage(msg WsMessageOut) {
	err := ws.conn.WriteJSON(msg)
	if err != nil {
		log.Println("failed to write WebSocket message:", err)
	}
}

func (ws *WebSocket) runReceiveLoop() {
	for {
		msgType, data, err := ws.conn.ReadMessage()
		if err != nil {
			log.Println("failed to read WebSocket message:", err)
			return
		}

		if msgType != websocket.TextMessage {
			log.Fatalln("received a message that was not text from WebSocket")
		}

		var message WsMessageIn
		err = json.Unmarshal(data, &message)

		if err != nil {
			log.Println("error: failed to decode JSON:", err)
			continue
		}

		ws.MessagesIn <- message
	}
}

package discord

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"sync"
	"time"
)

type WsEvent = int

const (
	WsEventOpen  = 1
	WsEventClose = 2
	WsEventError = 3
)

type HeartbeatProvider = func() WsMessageOut

type WebSocket struct {
	MessagesOut chan WsMessageOut
	MessagesIn  chan WsMessageIn
	Events      chan WsEvent

	conn              *websocket.Conn
	heartbeat         *time.Ticker
	heartbeatProvider HeartbeatProvider
	closeChan         chan bool
	closeMutex        sync.Mutex
	closed            bool
}

func OpenWebSocket(url string) (*WebSocket, error) {
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}

	ws := &WebSocket{
		MessagesOut: make(chan WsMessageOut, 25),
		MessagesIn:  make(chan WsMessageIn, 25),
		Events:      make(chan WsEvent, 25),
		conn:        conn,
		closeChan:   make(chan bool),
		closeMutex:  sync.Mutex{},
	}

	go ws.runReceiveLoop()
	go ws.runSendLoop()

	return ws, nil
}

func (ws *WebSocket) Send(opcode int, data interface{}) {
	if ws.closed {
		log.Println("attempt to send on closed WebSocket")
		return
	}

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
	ws.closeMutex.Lock()
	defer ws.closeMutex.Unlock()

	if ws.closed {
		log.Println("WebSocket is already closed")
		return
	}

	if ws.heartbeat != nil {
		ws.heartbeat.Stop()
	}

	close(ws.closeChan)
	close(ws.MessagesOut)
	close(ws.MessagesIn)

	err := ws.conn.Close()
	if err != nil {
		log.Println("Failed to close WebSocket gracefully:", err)
	}

	ws.closed = true
}

func (ws *WebSocket) sendHeartbeat() {
	ws.sendMessage(ws.heartbeatProvider())
}

func (ws *WebSocket) sendMessage(msg WsMessageOut) {
	err := ws.conn.WriteJSON(msg)
	if err != nil {
		log.Println("failed to write WebSocket message:", err)
		ws.Events <- WsEventError
	}
}

func (ws *WebSocket) runSendLoop() {
	//defer log.Println("WS send loop terminated")
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

func (ws *WebSocket) runReceiveLoop() {
	//defer log.Println("WS recv loop terminated")
	for {
		msgType, data, err := ws.conn.ReadMessage()
		if err != nil {
			log.Println("failed to read WebSocket message:", err)
			ws.Events <- WsEventError
			return
		}

		if msgType != websocket.TextMessage {
			log.Fatalln("received a message that was not text from WebSocket")
		}

		var message WsMessageIn
		err = json.Unmarshal(data, &message)

		if err != nil {
			log.Println("error: failed to decode JSON:", err)
			ws.Events <- WsEventError
			return
		}

		ws.MessagesIn <- message
	}
}

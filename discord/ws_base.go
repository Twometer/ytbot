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
	Url           string
	Name          string
	MessagesOut   chan WsMessageOut
	MessagesIn    chan WsMessageIn
	Events        chan WsEvent
	ReconnectFunc func()

	conn              *websocket.Conn
	heartbeat         *time.Ticker
	heartbeatProvider HeartbeatProvider
	closeChan         chan bool
	closeMutex        sync.Mutex
	closed            bool
	autoReconnect     bool
}

func OpenWebSocket(url string, name string, autoReconnect bool) (*WebSocket, error) {
	ws := &WebSocket{
		Name:          name,
		Url:           url,
		autoReconnect: autoReconnect,
	}

	err := ws.connect()

	return ws, err
}

func (ws *WebSocket) connect() error {
	log.Println(ws.Name + ": Dialing...")
	conn, _, err := websocket.DefaultDialer.Dial(ws.Url, nil)
	if err != nil {
		return err
	}

	ws.MessagesOut = make(chan WsMessageOut, 25)
	ws.MessagesIn = make(chan WsMessageIn, 25)
	ws.Events = make(chan WsEvent, 25)
	ws.conn = conn
	ws.closeChan = make(chan bool)
	ws.closed = false

	go ws.runReceiveLoop()
	go ws.runSendLoop()

	return nil
}

func (ws *WebSocket) Send(opcode int, data interface{}) {
	if ws.closed {
		log.Println(ws.Name + ": attempt to send on closed WebSocket")
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
		log.Println(ws.Name + ": WebSocket is already closed")
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
		log.Println(ws.Name+": Failed to close WebSocket gracefully:", err)
	}

	ws.closed = true
}

func (ws *WebSocket) sendHeartbeat() {
	ws.sendMessage(ws.heartbeatProvider())
}

func (ws *WebSocket) sendMessage(msg WsMessageOut) {
	err := ws.conn.WriteJSON(msg)
	if err != nil {
		log.Println(ws.Name+": failed to write WebSocket message:", err)
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
	defer func() {
		if !ws.closed && ws.autoReconnect {
			log.Println(ws.Name + ": disconnected unexpectedly, reconnecting in 5 seconds...")
			go func() {
				time.Sleep(5 * time.Second)
				ws.Reconnect()
			}()
		}
	}()
	for {
		msgType, data, err := ws.conn.ReadMessage()
		if err != nil {
			log.Println(ws.Name+": failed to read WebSocket message:", err)
			ws.Events <- WsEventError
			return
		}

		if msgType != websocket.TextMessage {
			log.Fatalln(ws.Name + ": received a non-text message from WebSocket")
		}

		var message WsMessageIn
		err = json.Unmarshal(data, &message)

		if err != nil {
			log.Println(ws.Name+": failed to decode JSON:", err)
			ws.Events <- WsEventError
			return
		}

		ws.MessagesIn <- message
	}
}

func (ws *WebSocket) Reconnect() {
	ws.Close()
	err := ws.connect()
	if err != nil {
		log.Println(ws.Name + ": Failed to reconnect.")
		return
	}

	if ws.ReconnectFunc != nil {
		ws.ReconnectFunc()
	}
}

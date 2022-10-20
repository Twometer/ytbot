package discord

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"sync"
	"time"
)

type WsEvent = int

//goland:noinspection ALL
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
	zap.S().Debugw("Dialing WebSocket", "name", ws.Name, "url", ws.Url)
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
		zap.S().Warnw("Attempted to send on closed WebSocket", "name", ws.Name)
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
		zap.S().Warnw("Attempted to close WebSocket twice", "name", ws.Name)
		return
	}

	ws.closed = true

	if ws.heartbeat != nil {
		ws.heartbeat.Stop()
	}

	close(ws.closeChan)
	close(ws.MessagesOut)
	close(ws.MessagesIn)

	err := ws.conn.Close()
	if err != nil {
		zap.S().Warnw("Failed to close WebSocket gracefully", "name", ws.Name, "error", err)
	}
}

func (ws *WebSocket) sendHeartbeat() error {
	return ws.sendMessage(ws.heartbeatProvider())
}

func (ws *WebSocket) sendMessage(msg WsMessageOut) error {
	err := ws.conn.WriteJSON(msg)
	if err != nil {
		if ws.closed {
			return errors.New(ws.Name + ": WebSocket was closed")
		} else {
			zap.S().Errorw("Failed to write WebSocket message", "name", ws.Name, "error", err)
			ws.Events <- WsEventError
		}
	}
	return nil
}

func (ws *WebSocket) runSendLoop() {
	defer zap.S().Debugln("WebSocket sending loop exited")
	for {
		var err error = nil
		if ws.heartbeat != nil {
			select {
			case msg := <-ws.MessagesOut:
				err = ws.sendMessage(msg)
			case <-ws.heartbeat.C:
				err = ws.sendHeartbeat()
			case <-ws.closeChan:
				return
			}
		} else {
			select {
			case msg := <-ws.MessagesOut:
				err = ws.sendMessage(msg)
			case <-ws.closeChan:
				return
			}
		}

		if err != nil {
			zap.S().Errorw("Send loop encountered an error. Exiting.", "name", ws.Name, "error", err)
			return
		}
	}
}

func (ws *WebSocket) runReceiveLoop() {
	defer func() {
		if !ws.closed && ws.autoReconnect {
			zap.S().Warnw("Connection closed unexpectedly. Reconnecting after 5 seconds.", "name", ws.Name)
			go func() {
				time.Sleep(5 * time.Second)
				ws.Reconnect()
			}()
		}
	}()
	for {
		msgType, data, err := ws.conn.ReadMessage()
		if ws.closed {
			zap.S().Debugw("WebSocket closed, receive loop is stopping.", "name", ws.Name)
			return
		}
		if err != nil {
			zap.S().Errorw("Failed to read from WebSocket", "name", ws.Name, "error", err)
			ws.Events <- WsEventError
			return
		}

		if msgType != websocket.TextMessage {
			zap.S().Errorw("Received non-text message from WebSocket", "name", ws.Name, "messageType", msgType)
			continue
		}

		var message WsMessageIn
		err = json.Unmarshal(data, &message)

		if err != nil {
			zap.S().Errorw("Failed to decode JSON message", "name", ws.Name, "error", err)
			ws.Events <- WsEventError
			continue
		}

		ws.MessagesIn <- message
	}
}

func (ws *WebSocket) Reconnect() {
	ws.Close()
	err := ws.connect()
	if err != nil {
		zap.S().Errorw("Failed to reconnect WebSocket", "name", ws.Name, "error", err)
		return
	}

	if ws.ReconnectFunc != nil {
		ws.ReconnectFunc()
	}
}

package discord

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"runtime"
	"time"
	"ytbot/discord/utils"
)

const gatewayUrl = "wss://gateway.discord.gg/?encoding=json&v=10"
const apiUrl = "https://discord.com/api/v10"

type Client struct {
	authToken string
	conn      *websocket.Conn
	intents   int32
	sequence  int
	cmdPrefix byte
	heartbeat *time.Ticker

	Guilds   map[string]GuildState
	Commands chan CommandBuffer
}

func NewClient(token string, commandPrefix byte) *Client {
	return &Client{
		authToken: "Bot " + token,
		Guilds:    make(map[string]GuildState),
		cmdPrefix: commandPrefix,
		Commands:  make(chan CommandBuffer, 25),
	}
}

func (client *Client) Start() error {
	log.Println("Connecting to Discord...")
	conn, _, err := websocket.DefaultDialer.Dial(gatewayUrl, nil)
	if err != nil {
		return err
	}

	client.conn = conn

	log.Println("Logging in...")
	err = client.sendIdentify()
	if err != nil {
		return err
	}

	go client.receiveLoop()

	return nil
}

func (client *Client) AddIntent(intent Intent) {
	client.intents |= intent
}

func (client *Client) ReplyMessage(message Message, content string) {
	client.PostMessage(Message{
		Content:   content,
		ChannelId: message.ChannelId,
	})
}

func (client *Client) EditMessage(message Message, newContent string) {

}

func (client *Client) SendMessage(channel string, content string) Message {
	message := Message{
		Content:   content,
		ChannelId: channel,
	}
	client.PostMessage(message)
	return message
}

func (client *Client) PostMessage(message Message) {
	url := fmt.Sprintf("%s/channels/%s/messages", apiUrl, message.ChannelId)
	message.Nonce = utils.NewNonce()
	err := utils.HttpPost(url, client.authToken, message)
	if err != nil {
		log.Println("Failed to send message:", err)
	}
}

func (client *Client) sendMessage(opcode GatewayOp, data interface{}) error {
	return client.conn.WriteJSON(WsMessageOut{
		Opcode: opcode,
		Data:   data,
	})
}

func (client *Client) sendIdentify() error {
	return client.sendMessage(GatewayOpIdentify, IdentifyPayload{
		Token:   client.authToken,
		Intents: client.intents,
		Properties: IdentifyProperties{
			OperatingSystem: runtime.GOOS,
			Browser:         "neko",
			Device:          "neko",
		},
	})
}

func (client *Client) sendHeartbeat() error {
	return client.sendMessage(GatewayOpHeartbeat, client.sequence)
}

func (client *Client) receiveLoop() {
	for {
		_, data, err := client.conn.ReadMessage()
		if err != nil {
			log.Fatalln("failed to read from WebSocket:", err)
		}

		var message WsMessageIn
		err = json.Unmarshal(data, &message)
		if err != nil {
			log.Fatalln("failed to decode JSON:", err)
		}

		client.handleMessage(message)
	}
}

func (client *Client) startHeartbeat(interval time.Duration) {
	if client.heartbeat == nil {
		client.heartbeat = time.NewTicker(interval)
	} else {
		client.heartbeat.Reset(interval)
	}

	go func() {
		client.sendHeartbeat()
		for range client.heartbeat.C {
			client.sendHeartbeat()
		}
	}()
}

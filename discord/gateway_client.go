package discord

import (
	"errors"
	"fmt"
	"log"
	"runtime"
	"ytbot/discord/utils"
)

const gatewayUrl = "wss://gateway.discord.gg/?encoding=json&v=10"
const apiUrl = "https://discord.com/api/v10"

type Client struct {
	ws        *WebSocket
	authToken string
	intents   int32
	sequence  int
	cmdPrefix byte
	userId    string

	Guilds       map[string]GuildState
	Commands     chan CommandBuffer
	VoiceServers chan VoiceServer
}

func NewClient(token string, commandPrefix byte) *Client {
	return &Client{
		authToken:    "Bot " + token,
		Guilds:       make(map[string]GuildState),
		cmdPrefix:    commandPrefix,
		Commands:     make(chan CommandBuffer, 25),
		VoiceServers: make(chan VoiceServer),
	}
}

func (client *Client) AddIntent(intent Intent) {
	client.intents |= intent
}

func (client *Client) Start() error {
	log.Println("Connecting to Discord...")

	ws, err := OpenWebSocket(gatewayUrl)
	if err != nil {
		return err
	}
	client.ws = ws

	client.sendIdentify()
	go client.handlerLoop()

	return nil
}

func (client *Client) ReplyMessage(message Message, content string) {
	client.PostMessage(Message{
		Content:   content,
		ChannelId: message.ChannelId,
	})
}

func (client *Client) EditMessage(message Message, newContent string) {
	panic("todo")
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

func (client *Client) JoinVoiceChannel(state VoiceState) (*VoiceClient, error) {
	log.Printf("Joining %s, %s", state.ChannelId, state.GuildId)

	client.ws.Send(GatewayOpVoiceStateUpdate, VoiceState{
		ChannelId: state.ChannelId,
		GuildId:   state.GuildId,
		SelfVideo: false,
		SelfMute:  false,
		SelfDeaf:  true,
	})

	log.Println("Waiting for voice gateway...")
	voiceServer := <-client.VoiceServers

	ownVoiceState, ok := client.Guilds[state.GuildId].VoiceStates[client.userId]
	if !ok {
		return nil, errors.New("could not get own voice state")
	}

	log.Printf("Connecting to voice gateway at `%s`...\n", voiceServer.Endpoint)
	voiceClient := NewVoiceClient(client.userId, ownVoiceState.SessionId, voiceServer)
	err := voiceClient.start()
	if err != nil {
		return nil, err
	}

	guild := client.Guilds[state.GuildId]
	guild.VoiceClient = voiceClient

	return guild.VoiceClient, nil
}

func (client *Client) sendIdentify() {
	client.ws.Send(GatewayOpIdentify, IdentifyPayload{
		Token:   client.authToken,
		Intents: client.intents,
		Properties: IdentifyProperties{
			OperatingSystem: runtime.GOOS,
			Browser:         "neko",
			Device:          "neko",
		},
	})
}

func (client *Client) handlerLoop() {
	for message := range client.ws.MessagesIn {
		client.handleMessage(message)
	}
}

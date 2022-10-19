package discord

import (
	"errors"
	"fmt"
	"github.com/buger/jsonparser"
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

func (client *Client) ReplyMessage(message Message, content string) Message {
	return client.PostMessage(Message{
		Content:   content,
		ChannelId: message.ChannelId,
	})
}

func (client *Client) EditMessage(message Message, newContent string) {
	url := fmt.Sprintf("%s/channels/%s/messages/%s", apiUrl, message.ChannelId, message.Id)
	message.Content = newContent
	_, err := utils.HttpSend("PATCH", url, client.authToken, message)
	if err != nil {
		log.Println("Failed to edit message:", err)
	}
}

func (client *Client) SendMessage(channel string, content string) Message {
	return client.PostMessage(Message{
		Content:   content,
		ChannelId: channel,
	})
}

func (client *Client) PostMessage(message Message) Message {
	url := fmt.Sprintf("%s/channels/%s/messages", apiUrl, message.ChannelId)
	message.Nonce = utils.NewNonce()
	resp, err := utils.HttpSend("POST", url, client.authToken, message)
	if err != nil {
		log.Println("Failed to send message:", err)
	}
	msg, _ := jsonparser.GetString(resp, "id")
	message.Id = msg
	return message
}

func (client *Client) JoinVoiceChannel(guildId string, channelId string) (*VoiceClient, error) {
	// Find guild
	guild, ok := client.Guilds[guildId]
	if !ok {
		return nil, errors.New("tried to join invalid guild")
	}

	// Check if already there?
	if guild.VoiceClient != nil {
		return guild.VoiceClient, nil
	}

	// Join channel
	client.ws.Send(GatewayOpVoiceStateUpdate, VoiceState{
		GuildId:   guildId,
		ChannelId: channelId,
		SelfVideo: false,
		SelfMute:  false,
		SelfDeaf:  true,
	})

	// Acquire voice server
	log.Println("Waiting for voice gateway...")
	voiceServer := <-client.VoiceServers

	// Get own voice session
	ownVoiceState, ok := guild.VoiceStates[client.userId]
	if !ok {
		return nil, errors.New("could not get own voice state")
	}

	// Create voice client
	log.Printf("Connecting to voice gateway at `%s`...\n", voiceServer.Endpoint)
	voiceClient := NewVoiceClient(client.userId, ownVoiceState.SessionId, voiceServer)
	err := voiceClient.start()
	if err != nil {
		return nil, err
	}

	// Save voice client to guild
	guild.VoiceClient = voiceClient
	client.Guilds[guildId] = guild

	return voiceClient, nil
}

func (client *Client) LeaveVoiceChannel(guildId string) {
	guild, ok := client.Guilds[guildId]
	if !ok {
		log.Println("failed to find guild while leaving")
		return
	}

	guild.VoiceClient.Close()
	client.ws.Send(GatewayOpVoiceStateUpdate, VoiceStateLeave{
		GuildId: guildId,
	})

	guild.VoiceClient = nil
	client.Guilds[guildId] = guild
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
	//defer log.Println("Gateway handler loop has terminated")
	for message := range client.ws.MessagesIn {
		client.handleMessage(message)
	}
}

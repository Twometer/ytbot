package discord

// GatewayOp identifies Discord gateway opcodes
type GatewayOp = int

//goland:noinspection GoUnusedConst
const (
	GatewayOpDispatch            = 0
	GatewayOpHeartbeat           = 1
	GatewayOpIdentify            = 2
	GatewayOpPresenceUpdate      = 3
	GatewayOpVoiceStateUpdate    = 4
	GatewayOpResume              = 6
	GatewayOpReconnect           = 7
	GatewayOpRequestGuildMembers = 8
	GatewayOpInvalidSession      = 9
	GatewayOpHello               = 10
	GatewayOpHeartbeatAck        = 11
)

// GatewayEvent identifies Discord gateway event types
type GatewayEvent = string

//goland:noinspection GoUnusedConst
const (
	GatewayEventMessageCreate     = "MESSAGE_CREATE"
	GatewayEventReady             = "READY"
	GatewayEventGuildCreate       = "GUILD_CREATE"
	GatewayEventVoiceStateUpdate  = "VOICE_STATE_UPDATE"
	GatewayEventVoiceServerUpdate = "VOICE_SERVER_UPDATE"
)

// Intent identifies bitflags for Discord bot intents
type Intent = int32

//goland:noinspection GoUnusedConst
const (
	IntentGuilds                = 1 << 0
	IntentMembers               = 1 << 1
	IntentBans                  = 1 << 2
	IntentEmojis                = 1 << 3
	IntentIntegrations          = 1 << 4
	IntentWebhooks              = 1 << 5
	IntentInvites               = 1 << 6
	IntentVoiceStates           = 1 << 7
	IntentPresences             = 1 << 8
	IntentMessages              = 1 << 9
	IntentMessageReactions      = 1 << 10
	IntentMessageTyping         = 1 << 11
	IntentDirectMessages        = 1 << 12
	IntentDirectMessageReaction = 1 << 13
	IntentDirectMessageTyping   = 1 << 14
	IntentMessageContent        = 1 << 15
	IntentScheduledEvents       = 1 << 16
	IntentAutoModConfig         = 1 << 20
	IntentAutoModExec           = 1 << 21
)

type IdentifyPayload struct {
	Token      string             `json:"token"`
	Intents    int32              `json:"intents"`
	Properties IdentifyProperties `json:"properties"`
}

type IdentifyProperties struct {
	OperatingSystem string `json:"os"`
	Browser         string `json:"browser"`
	Device          string `json:"device"`
}

type GatewayHelloMessage struct {
	HeartbeatInterval int `json:"heartbeat_interval"`
}

type GatewayReadyMessage struct {
	User User `json:"user"`
}

type GatewayGuildCreateMessage struct {
	Id          string       `json:"id"`
	Name        string       `json:"name"`
	VoiceStates []VoiceState `json:"voice_states"`
}

type User struct {
	Id            string `json:"id"`
	Username      string `json:"username"`
	Discriminator string `json:"discriminator"`
}

type Message struct {
	Id        string `json:"id"`
	Content   string `json:"content"`
	Author    User   `json:"author"`
	GuildId   string `json:"guild_id"`
	ChannelId string `json:"channel_id"`
	Nonce     string `json:"nonce"`
}

type VoiceState struct {
	UserId    string `json:"user_id"`
	ChannelId string `json:"channel_id"`
	GuildId   string `json:"guild_id"`
	SessionId string `json:"session_id"`
	SelfVideo bool   `json:"self_video"`
	SelfMute  bool   `json:"self_mute"`
	SelfDeaf  bool   `json:"self_deaf"`
	Mute      bool   `json:"mute"`
	Deaf      bool   `json:"deaf"`
}

type GuildState struct {
	Id          string
	Name        string
	VoiceStates map[string]VoiceState
}

type VoiceServer struct {
	Token    string `json:"token"`
	GuildId  string `json:"guild_id"`
	Endpoint string `json:"endpoint"`
}

package discord

type VoiceOp = int

//goland:noinspection GoUnusedConst
const (
	VoiceOpIdentify         = 0
	VoiceOpSelectProtocol   = 1
	VoiceOpReady            = 2
	VoiceOpHeartbeat        = 3
	VoiceOpSessionDesc      = 4
	VoiceOpSpeaking         = 5
	VoiceOpHeartbeatAck     = 6
	VoiceOpResume           = 7
	VoiceOpHello            = 8
	VoiceOpResumed          = 9
	VoiceOpClientDisconnect = 13
)

type VoiceIdentifyMessage struct {
	ServerId  string `json:"server_id"`
	UserId    string `json:"user_id"`
	SessionId string `json:"session_id"`
	Token     string `json:"token"`
}

type VoiceHelloMessage struct {
	HeartbeatInterval float32 `json:"heartbeat_interval"`
}

type VoiceReadyMessage struct {
	Ssrc  uint32   `json:"ssrc"`
	Port  int      `json:"port"`
	Ip    string   `json:"ip"`
	Modes []string `json:"modes"`
}

type VoiceSelectProtocolMessage struct {
	Protocol string       `json:"protocol"`
	Data     ProtocolData `json:"data"`
}

type ProtocolData struct {
	Address string `json:"address"`
	Port    int    `json:"port"`
	Mode    string `json:"mode"`
}

type VoiceSessionDescriptionMessage struct {
	SecretKey []byte `json:"secret_key"`
}

type VoiceSpeakingMessage struct {
	Speaking int    `json:"speaking"`
	Delay    int    `json:"delay"`
	Ssrc     uint32 `json:"ssrc"`
}

type VoiceEvent int

//goland:noinspection GoUnusedConst
const (
	VoiceEventReady    = 1
	VoiceEventPlaying  = 2
	VoiceEventFinished = 3
	VoiceEventStopped  = 4
	VoiceEventError    = 5
)

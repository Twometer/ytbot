package discord

type VoiceOp = int

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
	Ssrc  int      `json:"ssrc"`
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

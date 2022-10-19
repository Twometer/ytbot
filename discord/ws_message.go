package discord

import (
	"encoding/json"
	"log"
	"strconv"
)

// WsMessageOut models an outgoing WebSocket message in Discord's format.
// It takes any Go object in its Data field, which is later serialized to JSON
type WsMessageOut struct {
	Opcode int         `json:"op"`
	Data   interface{} `json:"d"`
}

// WsMessageIn is an incoming Discord WebSocket message that represents the Data field
// as raw JSON, since the type of the Data field can only be determined after the Opcode
// and Type are known.
type WsMessageIn struct {
	Opcode   int              `json:"op"`
	Sequence int              `json:"s"`
	Type     string           `json:"t"`
	Data     *json.RawMessage `json:"d"`
}

// Unmarshal deserializes the Data field in the WsMessageIn into an object
func (msg *WsMessageIn) Unmarshal(target interface{}) {
	err := json.Unmarshal(*msg.Data, target)
	if err != nil {
		log.Fatalln("failed to unmarshal message:", err)
	}
}

// String creates a string summary of the WsMessageIn
func (msg *WsMessageIn) String() string {
	dataStr := "nil"
	if msg.Data != nil {
		dataStr = string(*msg.Data)
	}
	return "op=" + strconv.Itoa(msg.Opcode) + "; data=" + dataStr + "; t=" + msg.Type
}

// String creates a string summary of the WsMessageOut
func (msg *WsMessageOut) String() (string, error) {
	data, err := json.Marshal(msg.Data)
	if err != nil {
		return "", err
	}
	return "op=" + strconv.Itoa(msg.Opcode) + "; data=" + string(data), err
}

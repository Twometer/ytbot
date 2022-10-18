package discord

import (
	"strconv"
	"strings"
)

type CommandBuffer struct {
	Message Message
	index   int
	parts   []string
}

func NewCommandBuffer(message Message) CommandBuffer {
	return CommandBuffer{
		Message: message,
		parts:   strings.Split(message.Content[1:], " "),
	}
}

func (buf *CommandBuffer) GetInt() int {
	i, _ := strconv.Atoi(buf.parts[buf.index])
	buf.index++
	return i
}

func (buf *CommandBuffer) GetString() string {
	str := buf.parts[buf.index]
	buf.index++
	return str
}

func (buf *CommandBuffer) GetStringAll() string {
	return strings.Join(buf.parts[buf.index:], " ")
}

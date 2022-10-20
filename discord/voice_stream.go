package discord

import (
	"bytes"
	"encoding/binary"
	"errors"
	"go.uber.org/zap"
	"golang.org/x/crypto/nacl/secretbox"
	"net"
)

// VoiceStream represents the UDP connection that does the actual voice streaming
type VoiceStream struct {
	RemoteIp   string
	RemotePort int

	LocalIp   string
	LocalPort int

	Ssrc uint32

	playing  bool
	parent   *VoiceClient
	conn     *net.UDPConn
	key      []byte
	sequence uint16
}

func NewVoiceStream(parent *VoiceClient, ip string, port int, ssrc uint32) *VoiceStream {
	return &VoiceStream{
		parent:     parent,
		RemoteIp:   ip,
		RemotePort: port,
		Ssrc:       ssrc,
	}
}

func (stream *VoiceStream) BeginSetup() error {
	addr := net.UDPAddr{
		Port: stream.RemotePort,
		IP:   net.ParseIP(stream.RemoteIp),
	}
	conn, err := net.DialUDP("udp", nil, &addr)
	if err != nil {
		return err
	}
	stream.conn = conn

	err = stream.discoverLocalIp()
	if err != nil {
		return err
	}

	return nil
}

func (stream *VoiceStream) FinishSetup(key []byte) {
	stream.key = key
	zap.S().Infow("Voice stream finished initialization")
}

func (stream *VoiceStream) SendOpusFrame(timestamp uint32, frame []byte) error {
	if stream.key == nil {
		return errors.New("voice stream is not initialized")
	}

	sequence := stream.nextSequence()
	packetBuffer := bytes.NewBuffer(make([]byte, 0))

	// RTP Header
	_ = binary.Write(packetBuffer, binary.BigEndian, uint8(0x80))
	_ = binary.Write(packetBuffer, binary.BigEndian, uint8(0x78))
	_ = binary.Write(packetBuffer, binary.BigEndian, sequence)
	_ = binary.Write(packetBuffer, binary.BigEndian, timestamp)
	_ = binary.Write(packetBuffer, binary.BigEndian, stream.Ssrc)

	// Encrypted audio data
	encryptedFrame := stream.encryptAudio(frame, packetBuffer.Bytes()[:12])
	packetBuffer.Write(encryptedFrame)

	// Send
	_, err := stream.conn.Write(packetBuffer.Bytes())
	return err
}

func (stream *VoiceStream) OnBegin() {
	stream.parent.Events <- VoiceEventPlaying
	stream.parent.sendSpeaking(true)
	stream.playing = true
}

func (stream *VoiceStream) OnFinished() {
	stream.parent.Events <- VoiceEventFinished
	stream.parent.sendSpeaking(false)
	stream.playing = false
}

func (stream *VoiceStream) OnStopped() {
	stream.parent.Events <- VoiceEventStopped
	stream.parent.sendSpeaking(false)
	stream.playing = false
}

func (stream *VoiceStream) OnFailed() {
	stream.parent.Events <- VoiceEventError
	stream.parent.sendSpeaking(false)
	stream.playing = false
}

func (stream *VoiceStream) encryptAudio(audioFrame []byte, nonceBytes []byte) []byte {
	var secretKey [32]byte
	copy(secretKey[:], stream.key)

	var nonce [24]byte
	copy(nonce[:12], nonceBytes)

	encryptedFrame := secretbox.Seal(make([]byte, 0), audioFrame, &nonce, &secretKey)
	return encryptedFrame
}

func (stream *VoiceStream) nextSequence() uint16 {
	stream.sequence++
	return stream.sequence
}

func (stream *VoiceStream) discoverLocalIp() error {
	reqBuf := bytes.NewBuffer(make([]byte, 0))
	_ = binary.Write(reqBuf, binary.BigEndian, uint16(1))
	_ = binary.Write(reqBuf, binary.BigEndian, uint16(70))
	_ = binary.Write(reqBuf, binary.BigEndian, stream.Ssrc)
	reqBuf.Write(make([]byte, 64))
	_ = binary.Write(reqBuf, binary.BigEndian, uint16(stream.RemotePort))

	_, err := stream.conn.Write(reqBuf.Bytes())
	if err != nil {
		return err
	}

	respData := make([]byte, 74)
	_, _, err = stream.conn.ReadFromUDP(respData)
	if err != nil {
		return err
	}

	var resp struct {
		msgType uint16
		msgLen  uint16
		ssrc    uint32
		ip      string
		port    uint16
	}
	respBuf := bytes.NewBuffer(respData)
	_ = binary.Read(respBuf, binary.BigEndian, &resp.msgType)
	_ = binary.Read(respBuf, binary.BigEndian, &resp.msgLen)
	_ = binary.Read(respBuf, binary.BigEndian, &resp.ssrc)
	ip, _ := respBuf.ReadBytes(0)
	resp.ip = string(ip[:len(ip)-1])
	respBuf.Next(64 - len(ip))
	_ = binary.Read(respBuf, binary.BigEndian, &resp.port)

	stream.LocalIp = resp.ip
	stream.LocalPort = int(resp.port)

	zap.S().Debugw("IP discovery finished successfully", "ip", resp.ip, "port", resp.port)
	return nil
}

func (stream *VoiceStream) Close() {
	if stream.conn == nil {
		return
	}

	err := stream.conn.Close()
	if err != nil {
		zap.S().Warnw("Gould not gracefully shut down the voice stream", "error", err)
	}
}

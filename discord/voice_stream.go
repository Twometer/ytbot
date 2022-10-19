package discord

import (
	"bytes"
	"encoding/binary"
	"errors"
	"golang.org/x/crypto/nacl/secretbox"
	"log"
	"net"
)

const payloadType byte = 0x78

const StatePlaying = 2
const StateStopped = 3

type VoiceStream struct {
	RemoteIp   string
	RemotePort int

	LocalIp   string
	LocalPort int

	Ssrc uint32

	conn         *net.UDPConn
	key          []byte
	sequence     uint16
	Ready        chan interface{}
	StateChanges chan int
	ts           uint32
}

func NewVoiceStream(ip string, port int, ssrc uint32) *VoiceStream {
	return &VoiceStream{
		RemoteIp:     ip,
		RemotePort:   port,
		Ssrc:         ssrc,
		Ready:        make(chan interface{}),
		StateChanges: make(chan int),
	}
}

func (vc *VoiceStream) BeginSetup() error {
	addr := net.UDPAddr{
		Port: vc.RemotePort,
		IP:   net.ParseIP(vc.RemoteIp),
	}
	conn, err := net.DialUDP("udp", nil, &addr)
	if err != nil {
		return err
	}
	vc.conn = conn

	err = vc.discoverLocalIp()
	if err != nil {
		return err
	}

	return nil
}

func (vc *VoiceStream) FinishSetup(key []byte) {
	vc.key = key
	vc.Ready <- true
	log.Println("Voice streaming connection established!")
}

func (vc *VoiceStream) SendOpusFrame(frame []byte) error {
	if vc.key == nil {
		return errors.New("voice stream is not initialized")
	}

	sequence := vc.nextSequence()
	timestamp := vc.ts
	vc.ts += (48000 / 100) * 2
	packetBuffer := bytes.NewBuffer(make([]byte, 0))

	// Discord Header
	_ = binary.Write(packetBuffer, binary.BigEndian, uint8(0x80))
	_ = binary.Write(packetBuffer, binary.BigEndian, payloadType)
	_ = binary.Write(packetBuffer, binary.BigEndian, sequence)
	_ = binary.Write(packetBuffer, binary.BigEndian, timestamp)
	_ = binary.Write(packetBuffer, binary.BigEndian, vc.Ssrc)

	// Encrypted audio data
	encryptedFrame := vc.encryptAudio(frame, packetBuffer.Bytes()[:12])
	packetBuffer.Write(encryptedFrame)

	// Send
	_, err := vc.conn.Write(packetBuffer.Bytes())
	return err
}

func (vc *VoiceStream) OnPlayingStateChanged(playing bool) {
	if playing {
		vc.StateChanges <- StatePlaying
	} else {
		vc.StateChanges <- StateStopped
	}
}

func (vc *VoiceStream) encryptAudio(audioFrame []byte, nonceBytes []byte) []byte {
	var secretKey [32]byte
	copy(secretKey[:], vc.key)

	var nonce [24]byte
	copy(nonce[:12], nonceBytes)

	encryptedFrame := secretbox.Seal(make([]byte, 0), audioFrame, &nonce, &secretKey)
	return encryptedFrame
}

func (vc *VoiceStream) nextSequence() uint16 {
	vc.sequence++
	return vc.sequence
}

func (vc *VoiceStream) discoverLocalIp() error {
	reqBuf := bytes.NewBuffer(make([]byte, 0))
	_ = binary.Write(reqBuf, binary.BigEndian, uint16(1))
	_ = binary.Write(reqBuf, binary.BigEndian, uint16(70))
	_ = binary.Write(reqBuf, binary.BigEndian, vc.Ssrc)
	reqBuf.Write(make([]byte, 64))
	_ = binary.Write(reqBuf, binary.BigEndian, uint16(vc.RemotePort))

	_, err := vc.conn.Write(reqBuf.Bytes())
	if err != nil {
		return err
	}

	respData := make([]byte, 74)
	_, _, err = vc.conn.ReadFromUDP(respData)
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
	resp.ip = string(ip)
	respBuf.Next(64 - len(ip))
	_ = binary.Read(respBuf, binary.BigEndian, &resp.port)

	vc.LocalIp = resp.ip
	vc.LocalPort = int(resp.port)

	log.Printf("IP Discovery completed: %s:%d\n", resp.ip, resp.port)

	return nil
}

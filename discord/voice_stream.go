package discord

import (
	"bytes"
	"encoding/binary"
	"log"
	"net"
)

type VoiceStream struct {
	RemoteIp   string
	RemotePort int

	LocalIp   string
	LocalPort int

	Ssrc int

	conn *net.UDPConn
	key  []byte
}

func NewVoiceStream(ip string, port int, ssrc int) *VoiceStream {
	return &VoiceStream{
		RemoteIp:   ip,
		RemotePort: port,
		Ssrc:       ssrc,
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

	err = vc.doIpDiscovery()
	if err != nil {
		return err
	}

	return nil
}

func (vc *VoiceStream) FinishSetup(key []byte) {
	vc.key = key
	log.Println("Voice connection established!")
}

func (vc *VoiceStream) SendRawFrame(frame []byte) error {
	reqBuf := bytes.NewBuffer(make([]byte, 0))
	binary.Write(reqBuf, binary.BigEndian, uint8(0x80))
	binary.Write(reqBuf, binary.BigEndian, uint8(0x78))
	binary.Write(reqBuf, binary.BigEndian, uint16(0)) // todo sequence
	binary.Write(reqBuf, binary.BigEndian, uint32(0)) // todo timestamp
	binary.Write(reqBuf, binary.BigEndian, uint32(vc.Ssrc))
	reqBuf.Write(frame) // todo encryption

	_, err := vc.conn.Write(reqBuf.Bytes())
	return err
}

func (vc *VoiceStream) doIpDiscovery() error {
	reqBuf := bytes.NewBuffer(make([]byte, 0))
	binary.Write(reqBuf, binary.BigEndian, uint16(1))
	binary.Write(reqBuf, binary.BigEndian, uint16(70))
	binary.Write(reqBuf, binary.BigEndian, uint32(vc.Ssrc))
	reqBuf.Write(make([]byte, 64))
	binary.Write(reqBuf, binary.BigEndian, uint16(vc.RemotePort))

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
	binary.Read(respBuf, binary.BigEndian, &resp.msgType)
	binary.Read(respBuf, binary.BigEndian, &resp.msgLen)
	binary.Read(respBuf, binary.BigEndian, &resp.ssrc)
	ip, _ := respBuf.ReadBytes(0)
	resp.ip = string(ip)
	respBuf.Next(64 - len(ip))
	binary.Read(respBuf, binary.BigEndian, &resp.port)

	vc.LocalIp = resp.ip
	vc.LocalPort = int(resp.port)

	log.Printf("IP Discovery completed: %s:%d\n", resp.ip, resp.port)

	return nil
}

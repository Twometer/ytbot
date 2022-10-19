package codec

import (
	"github.com/pion/webrtc/v3/pkg/media/oggreader"
	"io"
	"log"
	"time"
	"ytbot/config"
)

type Encoder struct {
	ffmpeg   Ffmpeg
	sink     AudioSink
	stopChan chan interface{}
}

type AudioSink interface {
	OnBegin()
	OnFinished()
	OnStopped()
	OnFailed()
	SendOpusFrame(timestamp uint32, frame []byte) error
}

type audioFrame struct {
	data   []byte
	header *oggreader.OggPageHeader
}

func NewEncoder(url string, sink AudioSink) *Encoder {
	return &Encoder{
		ffmpeg: Ffmpeg{
			Executable: config.GetString(config.KeyFfmpegLocation),
			SourceUrl:  url,
			OutputStreams: []OutputStream{
				{
					Number: 1,
					Config: "-c:a libopus -b:a 48K -vn -page_duration 20000 -f ogg",
				},
			},
		},
		sink:     sink,
		stopChan: make(chan interface{}),
	}
}

func (encoder *Encoder) Start() error {
	err := encoder.ffmpeg.Start()
	if err != nil {
		return err
	}

	oggReader, _, err := oggreader.NewWith(encoder.ffmpeg.Stdout)
	if err != nil {
		return err
	}

	ticker := time.NewTicker(20 * time.Millisecond)
	encoder.sink.OnBegin()

	pageChan := make(chan audioFrame, 300000)

	go func() {
		for {
			select {
			case <-encoder.stopChan:
				log.Println("Audio buffer stopped")
				return
			default:
				pageData, pageHeader, err := oggReader.ParseNextPage()
				if err == io.EOF {
					log.Println("Audio buffer finished")
					pageChan <- audioFrame{
						data:   nil,
						header: nil,
					}
					return
				} else if err != nil {
					log.Println("Audio buffer failed:", err)
					return
				}

				pageChan <- audioFrame{
					data:   pageData,
					header: pageHeader,
				}
			}
		}
	}()

	go func() {
		defer encoder.ffmpeg.Stop()
		for {
			select {
			case <-ticker.C:
				page := <-pageChan
				if page.header == nil && page.data == nil {
					log.Println("Audio playback finished")
					encoder.sink.OnFinished()
					return
				}
				err = encoder.sink.SendOpusFrame(uint32(page.header.GranulePosition), page.data)
				if err != nil {
					log.Println("Failed to write audio frame:", err)
					encoder.sink.OnFailed()
					return
				}
			case <-encoder.stopChan:
				log.Println("Audio playback stopped")
				encoder.sink.OnStopped()
				return
			}
		}
	}()

	return nil
}

func (encoder *Encoder) Stop() {
	close(encoder.stopChan)
}

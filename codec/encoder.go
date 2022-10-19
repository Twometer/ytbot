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
	OnPlayingStateChanged(playing bool)

	SendOpusFrame(frame []byte) error
}

func NewEncoder(url string, sink AudioSink) Encoder {
	return Encoder{
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
	encoder.sink.OnPlayingStateChanged(true)

	go func() {
		for {
			select {
			case <-ticker.C:
				pageData, _, err := oggReader.ParseNextPage()
				if err == io.EOF {
					log.Println("Audio playback finished")
					encoder.sink.OnPlayingStateChanged(false)
					return
				} else if err != nil {
					log.Println("Audio decoder failed:", err)
					encoder.sink.OnPlayingStateChanged(false)
					return
				}

				err = encoder.sink.SendOpusFrame(pageData)
				if err != nil {
					log.Println("Failed to write audio frame:", err)
				}
			case <-encoder.stopChan:
				log.Println("Audio playback stopped")
				encoder.sink.OnPlayingStateChanged(false)
				return
			}
		}
	}()

	return nil
}

func (encoder *Encoder) Stop() {
	close(encoder.stopChan)
}

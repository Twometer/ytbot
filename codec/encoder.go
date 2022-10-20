package codec

import (
	"github.com/pion/webrtc/v3/pkg/media/oggreader"
	"go.uber.org/zap"
	"io"
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
		zap.S().Debugln("Audio buffer is starting")
		for {
			select {
			case <-encoder.stopChan:
				zap.S().Debugln("Audio buffering was stopped")
				return
			default:
				pageData, pageHeader, err := oggReader.ParseNextPage()
				if err == io.EOF {
					zap.S().Debugln("Audio buffering completed")
					pageChan <- audioFrame{
						data:   nil,
						header: nil,
					}
					return
				} else if err != nil {
					zap.S().Warnw("Audio buffering failed", "error", err)
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
		zap.S().Debugln("Audio streamer is starting")
		defer encoder.ffmpeg.Stop()
		for {
			select {
			case <-ticker.C:
				page := <-pageChan
				if page.header == nil && page.data == nil {
					zap.S().Debugln("Audio streaming completed")
					encoder.sink.OnFinished()
					return
				}
				err = encoder.sink.SendOpusFrame(uint32(page.header.GranulePosition), page.data)
				if err != nil {
					zap.S().Warnw("Failed to write audio frame to stream", "error", err)
					encoder.sink.OnFailed()
					return
				}
			case <-encoder.stopChan:
				zap.S().Debugln("Audio streaming was stopped")
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

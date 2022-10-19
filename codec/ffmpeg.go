package codec

import (
	"io"
	"os/exec"
	"strconv"
	"strings"
)

type Ffmpeg struct {
	Executable    string
	SourceUrl     string
	OutputStreams []OutputStream
	Command       *exec.Cmd
	Stdout        io.Reader
	Stderr        io.Reader
}

type OutputStream struct {
	Number int
	Config string
}

func (ffmpeg *Ffmpeg) Start() error {
	cmd := exec.Command(ffmpeg.Executable, ffmpeg.buildArguments()...)
	ffmpeg.Command = cmd

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	ffmpeg.Stdout = stdout

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	ffmpeg.Stderr = stderr

	return cmd.Start()
}

func (ffmpeg *Ffmpeg) buildArguments() []string {
	arguments := []string{"-loglevel", "error", "-i", ffmpeg.SourceUrl}

	for _, stream := range ffmpeg.OutputStreams {
		streamArgs := strings.Split(stream.Config, " ")
		streamArgs = append(streamArgs, "pipe:"+strconv.Itoa(stream.Number))
		arguments = append(arguments, streamArgs...)
	}

	return arguments
}

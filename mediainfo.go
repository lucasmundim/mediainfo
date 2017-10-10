// Lists the streams and some codec details of a media file
//
// Tested with
//
// $ go run examples/mediainfo/medianfo.go --input=https://bintray.com/imkira/go-libav/download_file?file_path=sample_iPod.m4v
//
// stream 0: eng aac audio, 2 channels, 44100 Hz
// stream 1: h264 video, 320x240

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/imkira/go-libav/avcodec"
	"github.com/imkira/go-libav/avformat"
	"github.com/imkira/go-libav/avutil"
)

var initFileName string
var inputFileName string

func init() {
	flag.StringVar(&initFileName, "init", "", "init file to probe")
	flag.StringVar(&inputFileName, "input", "", "source file to probe")
	flag.Parse()
}

func main() {
	if len(initFileName) == 0 {
		log.Fatalf("Missing --init=file\n")
	}

	if len(inputFileName) == 0 {
		log.Fatalf("Missing --input=file\n")
	}

	tempFile, err := ioutil.TempFile(os.TempDir(), "prefix")
	defer os.Remove(tempFile.Name())

	var data []byte
	initData, err := ioutil.ReadFile(initFileName)
	if err != nil {
		panic(err)
	}
	data = append(data, initData...)

	inputData, err := ioutil.ReadFile(inputFileName)
	if err != nil {
		panic(err)
	}
	data = append(data, inputData...)

	if _, err = tempFile.Write(data); err != nil {
		panic(err)
	}

	// open format (container) context
	decFmt, err := avformat.NewContextForInput()
	if err != nil {
		log.Fatalf("Failed to open input context: %v", err)
	}

	// set some options for opening file
	options := avutil.NewDictionary()
	defer options.Free()

	// open file for decoding
	if err := decFmt.OpenInput(tempFile.Name(), nil, options); err != nil {
		log.Fatalf("Failed to open input file: %v", err)
	}
	defer decFmt.CloseInput()

	// initialize context with stream information
	if err := decFmt.FindStreamInfo(nil); err != nil {
		log.Fatalf("Failed to find stream info: %v", err)
	}

	// show stream info
	for _, stream := range decFmt.Streams() {
		language := stream.MetaData().Get("language")
		streamCtx := stream.CodecContext()
		codecID := streamCtx.CodecID()
		descriptor := avcodec.CodecDescriptorByID(codecID)
		switch streamCtx.CodecType() {
		case avutil.MediaTypeVideo:
			width := streamCtx.Width()
			height := streamCtx.Height()

			fmt.Printf("streamIndex: %d\n", stream.Index())
			fmt.Printf("codec: %s\n", descriptor.Name())
			fmt.Printf("resolution: %dx%d\n", width, height)
			fmt.Printf("pts: %d\n", stream.StartTime())
			fmt.Printf("duration: %d\n", (stream.RawDuration()-stream.StartTime())/int64(stream.TimeBase().Denominator()))
		case avutil.MediaTypeAudio:
			channels := streamCtx.Channels()
			sampleRate := streamCtx.SampleRate()
			fmt.Printf("stream %d: %s %s audio, %d channels, %d Hz\n",
				stream.Index(),
				language,
				descriptor.Name(),
				channels,
				sampleRate)
		case avutil.MediaTypeSubtitle:
			fmt.Printf("stream %d: %s %s subtitle\n",
				stream.Index(),
				language,
				descriptor.Name())
		}
	}
}

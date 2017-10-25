// Lists the streams and some codec details of a media file
//
// Tested with
//
// $ go run examples/mediainfo/medianfo.go --input=https://bintray.com/imkira/go-libav/download_file?file_path=sample_iPod.m4v
// $ GODEBUG=cgocheck=0 go run mediainfo.go --init fixtures/sample_video_1_init.mp4 --input=fixtures/sample_video_1_1.mp4
//
// stream 0: eng aac audio, 2 channels, 44100 Hz
// stream 1: h264 video, 320x240

package main

//#include <string.h>
//#include <libavutil/avutil.h>
//#include <libavutil/avstring.h>
//#include <libavcodec/avcodec.h>
//#include <libavformat/avformat.h>
//#include <libavformat/avio.h>
//
// extern int readPacket(void *p0, uint8_t *p1, int p2);
//
// #cgo pkg-config: libavformat libavutil
import "C"

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"unsafe"

	"github.com/imkira/go-libav/avcodec"
	"github.com/imkira/go-libav/avformat"
	"github.com/imkira/go-libav/avutil"
)

var initFileName, inputFileName string

type bufferData struct {
	ptr  *C.uint8_t
	size C.size_t
}

func min(x, y C.int) C.int {
	if x < y {
		return x
	}
	return y
}

//export readPacket
func readPacket(opaque unsafe.Pointer, buffer *C.uint8_t, bufferSize C.int) C.int {
	fmt.Printf("opaque:%p buf:%p buf_size:%d\n", opaque, buffer, bufferSize)

	bd := (*bufferData)(opaque)
	bufferSize = min(bufferSize, (C.int)(bd.size))

	fmt.Printf("ptr:%p size:%d updated_buf_size:%d\n", bd.ptr, bd.size, bufferSize)

	C.memcpy(unsafe.Pointer(buffer), unsafe.Pointer(bd.ptr), (C.size_t)(bufferSize))
	bd.ptr = (*C.uint8_t)(unsafe.Pointer(uintptr(unsafe.Pointer(bd.ptr)) + uintptr(bufferSize)))
	bd.size -= (C.size_t)(bufferSize)
	return bufferSize
}

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

	avutil.SetLogLevel(avutil.LogLevelDebug)

	var buffer []byte
	initData, err := ioutil.ReadFile(initFileName)
	if err != nil {
		panic(err)
	}
	buffer = append(buffer, initData...)

	inputData, err := ioutil.ReadFile(inputFileName)
	if err != nil {
		panic(err)
	}
	buffer = append(buffer, inputData...)

	// open format (container) context
	decFmt, err := avformat.NewContextForInput()
	if err != nil {
		log.Fatalf("Failed to open input context: %v", err)
	}

	// set some options for opening file
	options := avutil.NewDictionary()
	defer options.Free()

	var bufferSize C.int
	bufferSize = C.int(4096)
	readBufferSize := C.size_t(bufferSize)
	readExchangeArea := C.av_malloc(readBufferSize)
	// defer C.av_free(unsafe.Pointer(readExchangeArea))

	var bd bufferData
	bd.ptr = (*C.uint8_t)(unsafe.Pointer(&buffer[0]))
	bd.size = C.size_t(len(buffer))

	cCtx := C.avio_alloc_context((*C.uchar)(unsafe.Pointer(readExchangeArea)), bufferSize, 0, unsafe.Pointer(&bd), (*[0]byte)(C.readPacket), nil, nil)
	defer C.av_free(unsafe.Pointer(cCtx))
	ioCtx := avformat.NewIOContextFromC(unsafe.Pointer(cCtx))
	// defer C.av_free(unsafe.Pointer(ioCtx))
	decFmt.SetIOContext(ioCtx)

	// open file for decoding
	if err := decFmt.OpenInput("", nil, options); err != nil {
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
			fmt.Printf("bitrate: %d\n", streamCtx.BitRate())
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

	C.av_dump_format((*C.AVFormatContext)(unsafe.Pointer(decFmt.CAVFormatContext)), C.int(0), C.CString(""), C.int(0))
}

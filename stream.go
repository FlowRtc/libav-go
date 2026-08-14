package libav

/*
	#cgo pkg-config: libavcodec libavformat libavutil
	#include <libavcodec/packet.h>
	#include <libavcodec/defs.h>
	#include <libavcodec/codec_par.h>
	#include <libavcodec/avcodec.h>
	#include <libavformat/avformat.h>
	#include <libavutil/avutil.h>
	#include <libavutil/rational.h>
*/
import (
	"C"
)
import "unsafe"

type Stream struct {
	inner *C.AVStream
}

func CreateAVStream(codecParameters CodecParameters, num, den, streamIndex int) (Stream, bool) {
	params := (*C.AVCodecParameters)(codecParameters.inner)
	if params == nil {
		return Stream{}, false
	}
	defer func() {
		params.extradata = nil
		params.extradata_size = 0
		C.avcodec_parameters_free(&params)
	}()

	stream := (*C.AVStream)(C.av_mallocz(C.size_t(unsafe.Sizeof(C.AVStream{}))))
	if stream == nil {
		return Stream{}, false
	}

	stream.codecpar = C.avcodec_parameters_alloc()
	if stream.codecpar == nil {
		C.av_free(unsafe.Pointer(stream))
		return Stream{}, false
	}

	if C.avcodec_parameters_copy(stream.codecpar, params) < 0 {
		C.avcodec_parameters_free(&stream.codecpar)
		C.av_free(unsafe.Pointer(stream))
		return Stream{}, false
	}

	stream.codecpar.codec_tag = 0
	stream.time_base = C.AVRational{num: C.int(num), den: C.int(den)}
	stream.index = C.int(streamIndex)
	return Stream{stream}, true
}

func WrapAVStream(stream unsafe.Pointer) Stream {
	return Stream{
		inner: (*C.AVStream)(stream),
	}
}

func (s *Stream) AVStream() *C.AVStream {
	return s.inner
}
func (s *Stream) Index() int {
	return int(s.inner.index)
}
func (s *Stream) SetIndex(index int) {
	s.inner.index = C.int(index)
}

func (s *Stream) IsCodecTypeVideo() bool {
	return s.inner.codecpar.codec_type == C.AVMEDIA_TYPE_VIDEO
}

func (s *Stream) IsCodecTypeAudio() bool {
	return s.inner.codecpar.codec_type == C.AVMEDIA_TYPE_AUDIO
}

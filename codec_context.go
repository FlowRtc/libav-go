package libav

/*
#cgo pkg-config: libavcodec libavformat libavutil
#include <libavcodec/packet.h>
#include <libavcodec/codec.h>
#include <libavcodec/codec_id.h>
#include <libavcodec/codec_par.h>
#include <libavcodec/avcodec.h>
#include <libavformat/avformat.h>
#include <libavutil/avutil.h>
#include <libavutil/rational.h>
#include <libavutil/error.h>
#include <errno.h>
#include <string.h>

static inline int av_error_eagain(void) { return AVERROR(EAGAIN); }
*/
import "C"

import (
	"errors"
	"unsafe"
)

var ErrCodecNotFound error = errors.New("failed to find requested codec!")
var ErrCodecCtxAlloc error = errors.New("failed to allocate codec context!")

// EAGAIN error code returned by avcodec_send_* / avcodec_receive_* to signal
// that more input is needed / no output is currently available.
var ErrEAGAIN = int(C.av_error_eagain())

// EOF error code returned by avcodec_receive_* when the stream has ended.
const ErrEOF = int(C.AVERROR_EOF)

// MediaType is a Go-native enum of the FFmpeg AVMediaType values.
type MediaType int32

const (
	CodecTypeUnknown MediaType = C.AVMEDIA_TYPE_UNKNOWN
	CodecTypeVideo   MediaType = C.AVMEDIA_TYPE_VIDEO
	CodecTypeAudio   MediaType = C.AVMEDIA_TYPE_AUDIO
)

type CodecContext struct {
	inner *C.AVCodecContext
}

func FindDecoder(id CodecID) (unsafe.Pointer, error) {
	codec := C.avcodec_find_decoder(C.enum_AVCodecID(id))
	if codec == nil {
		return nil, ErrCodecNotFound
	}
	return unsafe.Pointer(codec), nil
}

func FindEncoder(id CodecID) (unsafe.Pointer, error) {
	codec := C.avcodec_find_encoder(C.enum_AVCodecID(id))
	if codec == nil {
		return nil, ErrCodecNotFound
	}
	return unsafe.Pointer(codec), nil
}

// GetCodecType returns the media type (video/audio) for the given codec id.
func GetCodecType(id CodecID) MediaType {
	return MediaType(C.avcodec_get_type(C.enum_AVCodecID(id)))
}

func AllocContext3(codec unsafe.Pointer) (CodecContext, error) {
	ctx := C.avcodec_alloc_context3((*C.AVCodec)(codec))
	if ctx == nil {
		return CodecContext{}, ErrCodecCtxAlloc
	}
	return CodecContext{ctx}, nil
}

func (c CodecContext) Inner() unsafe.Pointer {
	return unsafe.Pointer(c.inner)
}

func (c CodecContext) IsValid() bool {
	return c.inner != nil
}

func (c CodecContext) ParametersToContext(codecpar unsafe.Pointer) int {
	par := (*C.AVCodecParameters)(codecpar)
	return int(C.avcodec_parameters_to_context(c.inner, par))
}

func (c CodecContext) Open(codec unsafe.Pointer) int {
	return int(C.avcodec_open2(c.inner, (*C.AVCodec)(codec), nil))
}

// OpenWithOptions opens the codec context, passing the given options as an
// internal AVDictionary. Returns the raw return code.
func (c CodecContext) OpenWithOptions(codec unsafe.Pointer, opts map[string]string) int {
	var dict *C.AVDictionary = nil
	for key, val := range opts {
		C.av_dict_set(&dict, C.CString(key), C.CString(val), 0)
	}
	ret := int(C.avcodec_open2(c.inner, (*C.AVCodec)(codec), &dict))
	C.av_dict_free(&dict)
	return ret
}

func (c CodecContext) SendPacket(pkt unsafe.Pointer) int {
	return int(C.avcodec_send_packet(c.inner, (*C.AVPacket)(pkt)))
}

func (c CodecContext) SendFrame(f unsafe.Pointer) int {
	return int(C.avcodec_send_frame(c.inner, (*C.AVFrame)(f)))
}

func (c CodecContext) ReceiveFrame(f *Frame) int {
	if f == nil || f.inner == nil {
		panic("packet is nil bru we can't read for shit")
	}
	return int(C.avcodec_receive_frame(c.inner, f.inner))
}

func (c CodecContext) ReceivePacket(p *Packet) int {
	if p == nil || p.inner == nil {
		panic("packet is nil bru we can't read for shit")
	}
	return int(C.avcodec_receive_packet(c.inner, p.inner))
}

func (c *CodecContext) Free() {
	if c == nil || c.inner == nil {
		return
	}
	C.avcodec_free_context(&c.inner)
}

func (c CodecContext) Extradata() []byte {
	if c.inner == nil || c.inner.extradata == nil || c.inner.extradata_size <= 0 {
		return nil
	}
	return unsafe.Slice((*byte)(c.inner.extradata), int(c.inner.extradata_size))
}

func (c CodecContext) SetCodecID(id CodecID) {
	c.inner.codec_id = C.enum_AVCodecID(id)
}

func (c CodecContext) SetCodecType(t MediaType) {
	c.inner.codec_type = C.enum_AVMediaType(t)
}

func (c CodecContext) SetSampleRate(rate int) {
	c.inner.sample_rate = C.int(rate)
}

func (c CodecContext) SetBitRate(bitrate int) {
	c.inner.bit_rate = C.int64_t(bitrate)
}

func (c CodecContext) SetSampleFormat(fmt SampleFormat) {
	c.inner.sample_fmt = C.enum_AVSampleFormat(fmt)
}

func (c CodecContext) SetTimeBase(tb Rational) {
	c.inner.time_base = tb.AVRational()
}

func (c CodecContext) SetChannelLayoutFrom(par CodecParameters) {
	C.av_channel_layout_copy(&c.inner.ch_layout, &par.inner.ch_layout)
}

func (c CodecContext) SetGlobalHeader() {
	c.inner.flags |= C.AV_CODEC_FLAG_GLOBAL_HEADER
}

func (c CodecContext) SetHeight(height int) {
	c.inner.height = C.int(height)
}

func (c CodecContext) SetWidth(width int) {
	c.inner.width = C.int(width)
}

// PixelFormat is a Go-native enum of the FFmpeg AVPixelFormat values.
type PixelFormat int32

const (
	PixelFmtYUV420P PixelFormat = C.AV_PIX_FMT_YUV420P
)

// SetPixelFormat sets the codec context pixel format.
func (c CodecContext) SetPixelFormat(fmt PixelFormat) {
	c.inner.pix_fmt = C.enum_AVPixelFormat(fmt)
}

func (c CodecContext) SetGOP(gop int) {
	c.inner.gop_size = C.int(gop)
}

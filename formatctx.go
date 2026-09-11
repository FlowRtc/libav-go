package libav

/*
#cgo pkg-config: libavcodec libavformat libavutil
#include <libavcodec/packet.h>
#include <libavcodec/codec_par.h>
#include <libavcodec/avcodec.h>
#include <libavformat/avformat.h>
#include <libavutil/rational.h>
#include <string.h>
*/
import "C"

import "unsafe"

type AVFormatContext struct {
	inner *C.AVFormatContext
}

func NewAVFormatContext() (AVFormatContext, error) {
	ctx := C.avformat_alloc_context()

	if ctx == nil {
		return AVFormatContext{}, ErrOOM
	}

	return AVFormatContext{ctx}, nil

}
func (ctx AVFormatContext) Inner() *C.AVFormatContext {
	return ctx.inner
}
func (ctx AVFormatContext) SetPB(ioctx AVIOContext) {
	ctx.inner.pb = ioctx.inner

}
func (ctx AVFormatContext) GetFlags() int {
	return int(ctx.inner.flags)
}

func (ctx AVFormatContext) SetFlags(flag int) {
	ctx.inner.flags = C.int(flag)
}
func (ctx AVFormatContext) NB_Streams() int {
	return int(ctx.inner.nb_streams)
}

func (ctx AVFormatContext) AVFormatOpenInput(url unsafe.Pointer, fmt unsafe.Pointer, options unsafe.Pointer) int {
	return int(C.avformat_open_input(&ctx.inner, nil, (*C.AVInputFormat)(fmt), (**C.AVDictionary)(options)))
}

func (ctx AVFormatContext) FindStreamInfo(options unsafe.Pointer) int {
	return (int)(C.avformat_find_stream_info(ctx.inner, (**C.AVDictionary)(options)))
}

func (ctx AVFormatContext) Streams(index int) Stream {
	stream := *(**C.AVStream)(unsafe.Pointer(
		uintptr(unsafe.Pointer(ctx.inner.streams)) +
			uintptr(index)*unsafe.Sizeof(uintptr(0)),
	))

	return WrapAVStream(unsafe.Pointer(stream))
}

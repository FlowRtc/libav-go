package libav

/*
#cgo pkg-config: libavcodec libavformat libavutil
#include <libavcodec/packet.h>
#include <libavcodec/codec_par.h>
#include <libavcodec/avcodec.h>
#include <libavformat/avformat.h>
#include <libavutil/rational.h>
#include <libavutil/mem.h>
#include <stdint.h>
#include <string.h>

static AVIOContext *avio_alloc_context_from_ptrs(
    unsigned char *buffer,
    int buffer_size,
    int write_flag,
    void *opaque,
    void *read_packet,
    void *write_packet,
    void *seek) {
	return avio_alloc_context(
	    buffer,
	    buffer_size,
	    write_flag,
	    opaque,
	    (int (*)(void *, uint8_t *, int))read_packet,
	    (int (*)(void *, const uint8_t *, int))write_packet,
	    (int64_t (*)(void *, int64_t, int))seek);
}
*/
import "C"

import "unsafe"

type AVIOContext struct {
	inner *C.AVIOContext
}

func NewAVIOContext(buffer []byte, size int, writeFlag int, userData unsafe.Pointer, readPacket unsafe.Pointer, writePacket unsafe.Pointer, seek unsafe.Pointer) (AVIOContext, error) {
	if size <= 0 {
		size = len(buffer)
	}
	if size <= 0 {
		return AVIOContext{}, ErrOOM
	}

	cbuf := (*C.uchar)(C.av_malloc(C.size_t(size)))
	if cbuf == nil {
		return AVIOContext{}, ErrOOM
	}

	ctx := C.avio_alloc_context_from_ptrs(
		cbuf,
		C.int(size),
		C.int(writeFlag),
		userData,
		readPacket,
		writePacket,
		seek,
	)
	if ctx == nil {
		C.av_free(unsafe.Pointer(cbuf))
		return AVIOContext{}, ErrOOM
	}

	return AVIOContext{ctx}, nil
}

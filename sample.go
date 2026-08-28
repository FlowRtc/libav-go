package libav

/*
#include <libavutil/samplefmt.h>
*/
import "C"

import "unsafe"

// SampleFormat is a Go-native enum of the FFmpeg AVSampleFormat values. It is
// used by the public API so that callers never need to reference cgo types.
type SampleFormat int32

const (
	SampleFmtU8   SampleFormat = C.AV_SAMPLE_FMT_U8
	SampleFmtS16  SampleFormat = C.AV_SAMPLE_FMT_S16
	SampleFmtS32  SampleFormat = C.AV_SAMPLE_FMT_S32
	SampleFmtFLT  SampleFormat = C.AV_SAMPLE_FMT_FLT
	SampleFmtDBL  SampleFormat = C.AV_SAMPLE_FMT_DBL
	SampleFmtU8P  SampleFormat = C.AV_SAMPLE_FMT_U8P
	SampleFmtS16P SampleFormat = C.AV_SAMPLE_FMT_S16P
	SampleFmtS32P SampleFormat = C.AV_SAMPLE_FMT_S32P
	SampleFmtFLTP SampleFormat = C.AV_SAMPLE_FMT_FLTP
	SampleFmtDBLP SampleFormat = C.AV_SAMPLE_FMT_DBLP
)

// SampleData is a Go-native representation of FFmpeg planar sample buffers: a
// slice holding one pointer per channel plane. It replaces the previous public
// **C.uint8_t so that callers never need cgo.
type SampleData []unsafe.Pointer

// cptr converts the Go-native plane pointers back into a **C.uint8_t suitable
// for the FFmpeg C APIs. The returned pointer is only valid while data lives.
func (d SampleData) cptr8() **C.uint8_t {
	if len(d) == 0 {
		return nil
	}
	return (**C.uint8_t)(unsafe.Pointer(&d[0]))
}

// toSampleData builds a SampleData slice over a C **C.uint8_t plane array of
// length nbChannels. The returned slice shares storage with the C array.
func toSampleData(planes **C.uint8_t, nbChannels int) SampleData {
	if planes == nil {
		return nil
	}
	return unsafe.Slice((*unsafe.Pointer)(unsafe.Pointer(planes)), nbChannels)
}

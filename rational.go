package libav

/*
#cgo pkg-config: libavcodec libavformat libavutil
#include <libavcodec/packet.h>
#include <libavcodec/codec_par.h>
#include <libavcodec/avcodec.h>
#include <libavformat/avformat.h>
#include <libavutil/rational.h>
#include <libavutil/error.h>
#include <string.h>

static inline const char *av_err2str_wrapper(int errnum) { return av_err2str(AVERROR(errnum)); }
*/
import "C"

import "unsafe"

type Rational struct {
	inner C.AVRational
}

func AVRational(r unsafe.Pointer) Rational {
	rat := (*C.AVRational)(r)
	return Rational{
		*rat,
	}
}

func NewRational(num, den int) Rational {
	return Rational{
		C.AVRational{
			C.int(num), C.int(den),
		},
	}
}
func (r Rational) AVRational() C.AVRational {
	return r.inner
}

func (r Rational) SetNum(n int) {
	r.inner.num = C.int(n)
}

func (r Rational) SetDen(d int) {
	r.inner.den = C.int(d)
}
func (r Rational) Num() int {
	return int(r.inner.num)
}

func (r Rational) Den() int {
	return int(r.inner.den)
}

func AvErrorString(errnum int) string {
	return C.GoString(C.av_err2str_wrapper((C.int)(errnum)))
}

// a * b / c
func RescaleQ(a int64, b, c Rational) int64 {
	res := C.av_rescale_q(C.int64_t(a), b.AVRational(), c.AVRational())
	return int64(res)
}

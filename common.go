package libav

/*
#include <libavcodec/packet.h>
#include <libavcodec/codec_par.h>
#include <libavcodec/avcodec.h>
#include <libavformat/avformat.h>
#include <libavutil/rational.h>
#include <libavutil/error.h>
#include <string.h>
*/
import "C"
import "unsafe"

const NoPTSValue = int64(C.AV_NOPTS_VALUE)

type CodecID int

const (
	CodecIdH264 CodecID = C.AV_CODEC_ID_H264
	CodecIdH265         = C.AV_CODEC_ID_HEVC
	CodecIdVP8          = C.AV_CODEC_ID_VP8
	CodecIdVP9          = C.AV_CODEC_ID_VP9
	CodecIdAV1          = C.AV_CODEC_ID_AV1
	CodecIdOpus         = C.AV_CODEC_ID_OPUS
	CodecIdAAC          = C.AV_CODEC_ID_AAC
)

// AudioFormat is a Go-native alias for SampleFormat, retained for backward
// compatibility with the original SetAudio signature.
type AudioFormat = SampleFormat

const AudioFmtS16 AudioFormat = SampleFmtS16
const AudioFmtFltp AudioFormat = SampleFmtFLTP

func GetCodecName(codec_id uint32) string {
	return C.GoString(C.avcodec_get_name(codec_id))
}

func OptionsToAVDict(options map[string]string) unsafe.Pointer {
	var dict *C.AVDictionary = nil
	for it_index, it := range options {
		C.av_dict_set(&dict, C.CString(it_index), C.CString(it), 0)
	}
	return unsafe.Pointer(dict)
}

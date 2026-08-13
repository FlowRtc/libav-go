package libav

/*
#cgo pkg-config: libavutil libavcodec libswresample

#include <stdlib.h>
#include "resampler.h"
#include <libavutil/samplefmt.h>
*/
import "C"

import (
	"errors"
)

type AudioInfo struct {
	SampleFmt  C.enum_AVSampleFormat
	SampleRate int
	Channels   int
	frameSize  int
}

type Resampler struct {
	c C.AudioResampler
}

type AudioFifo struct {
	c C.AudioFifo
}

func DefaultOpusInfo() AudioInfo {
	// todo remove this and parse the opus frame to know the info bru
	return AudioInfo{
		SampleFmt:  C.AV_SAMPLE_FMT_FLTP,
		SampleRate: 48000,
		Channels:   2,
	}
}

func DefaultAACInfo() AudioInfo {
	return AudioInfo{
		SampleFmt:  C.AV_SAMPLE_FMT_FLTP,
		SampleRate: 48000,
		Channels:   2,
	}
}

func NewAudioInfo(
	sampleFmt C.enum_AVSampleFormat,
	sampleRate int,
	channels int,
) AudioInfo {
	return AudioInfo{
		SampleFmt:  sampleFmt,
		SampleRate: sampleRate,
		Channels:   channels,
	}
}
func makeAudioInfo(info AudioInfo) C.AudioInfo {
	var cinfo C.AudioInfo

	cinfo.sample_fmt = info.SampleFmt
	cinfo.sample_rate = C.int(info.SampleRate)

	C.av_channel_layout_default(&cinfo.ch_layout, C.int(info.Channels))

	return cinfo
}

func NewResampler(input, output AudioInfo) (Resampler, error) {
	cin := makeAudioInfo(input)
	cout := makeAudioInfo(output)
	defer func() {
		C.av_channel_layout_uninit(&cin.ch_layout)
		C.av_channel_layout_uninit(&cout.ch_layout)
	}()

	r := Resampler{}

	if !bool(C.audio_resampler_init(&r.c, cin, cout)) {
		return Resampler{}, errors.New("audio_resampler_init failed")
	}

	return r, nil
}

func (r *Resampler) Resample(input **C.uint8_t, samples int) error {
	if !bool(C.ar_resample(
		&r.c,
		input,
		C.size_t(samples),
	)) {
		return errors.New("resample failed")
	}

	return nil
}

func (r *Resampler) Samples() **C.uint8_t {
	return C.ar_get_samples(&r.c)
}

func NewFIFO(info AudioInfo, min_samples int) AudioFifo {
	cinfo := makeAudioInfo(info)

	f := AudioFifo{
		c: C.audio_fifo_init(&cinfo, C.size_t(min_samples)),
	}

	C.av_channel_layout_uninit(&cinfo.ch_layout)

	return f
}

func (f *AudioFifo) Write(data **C.uint8_t, samples int) error {
	if data == nil {
		return nil
	}

	if !bool(C.af_write_sample(
		&f.c,
		data,
		C.int(samples),
	)) {
		return errors.New("fifo write failed")
	}

	return nil
}

func (f *AudioFifo) ReadFrame() (Frame, bool) {
	frame := Frame{
		inner: C.af_read_frame(&f.c),
	}
	return frame, frame.inner != nil
}

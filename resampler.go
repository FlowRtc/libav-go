package libav

/*
#cgo pkg-config: libavutil libavcodec libswresample
#include <stdlib.h>
#include <libavutil/samplefmt.h>
#include <libavutil/audio_fifo.h>
#include <libavutil/channel_layout.h>
#include <libavutil/mem.h>
#include <libavutil/frame.h>
#include <libswresample/swresample.h>
*/
import "C"

import (
	"errors"
	"fmt"
	"unsafe"
)

// AudioInfo describes audio format parameters used to create resamplers and
// FIFOs. It mirrors the old C AudioInfo struct but is fully a Go type.
type AudioInfo struct {
	SampleFmt  SampleFormat
	SampleRate int
	Channels   int
	FrameSize  int
}

func DefaultOpusInfo() AudioInfo {
	// todo remove this and parse the opus frame to know the info bru
	return AudioInfo{
		SampleFmt:  SampleFmtFLTP,
		SampleRate: 48000,
		Channels:   2,
		FrameSize:  960,
	}
}

func DefaultAACInfo() AudioInfo {
	return AudioInfo{
		SampleFmt:  SampleFmtFLTP,
		SampleRate: 48000,
		Channels:   2,
		FrameSize:  1024,
	}
}

func NewAudioInfo(
	sampleFmt SampleFormat,
	sampleRate int,
	channels int,
) AudioInfo {
	return AudioInfo{
		SampleFmt:  sampleFmt,
		SampleRate: sampleRate,
		Channels:   channels,
	}
}

// Resampler is a pure-Go rewrite of the former resampler.c AudioResampler.
// It wraps a raw SwrContext plus a reusable output sample buffer.
type Resampler struct {
	swr        *C.SwrContext
	outputInfo AudioInfo
	samples    **C.uint8_t
	maxSamples int
}

// ResampleErr is returned on any resampling failure.
var ResampleErr = errors.New("resample failed")

func NewResampler(input, output AudioInfo) (Resampler, error) {
	r := Resampler{outputInfo: output}

	inLayout := C.AVChannelLayout{}
	C.av_channel_layout_default(&inLayout, C.int(input.Channels))
	outLayout := C.AVChannelLayout{}
	C.av_channel_layout_default(&outLayout, C.int(output.Channels))

	errorcode := int(C.swr_alloc_set_opts2(
		&r.swr,
		&outLayout, C.enum_AVSampleFormat(output.SampleFmt), C.int(output.SampleRate),
		&inLayout, C.enum_AVSampleFormat(input.SampleFmt), C.int(input.SampleRate),
		0, nil,
	))

	C.av_channel_layout_uninit(&inLayout)
	C.av_channel_layout_uninit(&outLayout)

	if errorcode < 0 {
		return Resampler{}, fmt.Errorf("could not allocate resample context: %s", AvErrorString(errorcode))
	}

	if errorcode = int(C.swr_init(r.swr)); errorcode < 0 {
		defer r.Free()
		return Resampler{}, fmt.Errorf("could not open resample context: %s", AvErrorString(errorcode))
	}

	return r, nil
}

// reallocSamples mirrors ar_realloc_samples: grows the output buffer so it can
// hold frame_size samples per channel.
func (r *Resampler) reallocSamples(frameSize int) bool {
	if r.maxSamples >= frameSize {
		return true
	}
	if r.samples != nil {
		C.av_freep(unsafe.Pointer(r.samples))
		C.av_freep(unsafe.Pointer(&r.samples))
	}
	var linesize C.int
	err := int(C.av_samples_alloc_array_and_samples(
		&r.samples,
		&linesize,
		C.int(r.outputInfo.Channels),
		C.int(frameSize),
		C.enum_AVSampleFormat(r.outputInfo.SampleFmt),
		0,
	))
	if err < 0 {
		return false
	}
	r.maxSamples = frameSize
	return true
}

func (r *Resampler) Resample(input SampleData, samples int) error {
	if !r.reallocSamples(samples) {
		return errors.New("could not allocate converted input samples")
	}

	err := int(C.swr_convert(
		r.swr,
		r.samples,
		C.int(r.maxSamples),
		input.cptr8(),
		C.int(r.maxSamples),
	))
	if err < 0 {
		return fmt.Errorf("could not convert input samples: %s", AvErrorString(err))
	}
	return nil
}

func (r *Resampler) Samples() SampleData {
	if r.samples == nil {
		return nil
	}
	return toSampleData(r.samples, r.outputInfo.Channels)
}

// Free releases the swr context and any allocated output buffer.
func (r *Resampler) Free() {
	if r.swr != nil {
		C.swr_free(&r.swr)
		r.swr = nil
	}
	if r.samples != nil {
		C.av_freep(unsafe.Pointer(r.samples))
		C.av_freep(unsafe.Pointer(&r.samples))
		r.samples = nil
	}
	r.maxSamples = 0
}

// AudioFifo is a pure-Go rewrite of the former resampler.c AudioFifo. It wraps
// a raw AVAudioFifo plus a reusable AVFrame that acts as the read target.
type AudioFifo struct {
	fifo       *C.AVAudioFifo
	minSamples int
	outputInfo AudioInfo
	frame      Frame
}

// NewFIFO mirrors audio_fifo_init.
func NewFIFO(info AudioInfo, minSamples int) AudioFifo {
	layout := C.AVChannelLayout{}
	C.av_channel_layout_default(&layout, C.int(info.Channels))

	const audioFifoSize = 4096 * 2
	fifo := C.av_audio_fifo_alloc(C.enum_AVSampleFormat(info.SampleFmt), C.int(info.Channels), C.int(audioFifoSize))
	if fifo == nil {
		C.av_channel_layout_uninit(&layout)
		return AudioFifo{}
	}

	frame, err := NewFrame()
	if err != nil {
		C.av_audio_fifo_free(fifo)
		C.av_channel_layout_uninit(&layout)
		return AudioFifo{}
	}

	// needed because av_frame_get_buffer and the audio fifo require us to
	// allocate the buffer.
	frame.inner.nb_samples = C.int(audioFifoSize)
	frame.inner.format = C.int(info.SampleFmt)
	frame.inner.ch_layout = layout

	if C.av_frame_get_buffer(frame.inner, 0) != 0 {
		frame.Free()
		C.av_audio_fifo_free(fifo)
		return AudioFifo{}
	}

	return AudioFifo{
		fifo:       fifo,
		minSamples: minSamples,
		outputInfo: info,
		frame:      frame,
	}
}

func (f *AudioFifo) Write(data SampleData, samples int) error {
	if len(data) == 0 {
		return nil
	}
	bytesWrote := int(C.av_audio_fifo_write(f.fifo, (*unsafe.Pointer)(unsafe.Pointer(data.cptr8())), C.int(samples)))
	if bytesWrote < samples {
		return fmt.Errorf("could not write data to FIFO %s", AvErrorString(bytesWrote))
	}
	return nil
}

func (f *AudioFifo) ReadFrame() (Frame, bool) {
	if f.fifo == nil {
		return Frame{}, false
	}
	if int(C.av_audio_fifo_size(f.fifo)) < f.minSamples {
		return Frame{}, false
	}

	samples := int(C.av_audio_fifo_read(f.fifo, (*unsafe.Pointer)(unsafe.Pointer(&f.frame.inner.data[0])), C.int(f.minSamples)))
	if samples < 0 {
		return Frame{}, false
	}

	f.frame.inner.nb_samples = C.int(samples)

	return f.frame, true
}

// Free releases the fifo and its reusable frame.
func (f *AudioFifo) Free() {
	if f.fifo != nil {
		C.av_audio_fifo_free(f.fifo)
		f.fifo = nil
	}
	if f.frame.inner != nil {
		f.frame.Free()
	}
	f.minSamples = 0
}

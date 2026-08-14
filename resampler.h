#pragma once

#include <stdbool.h>
#include <stddef.h>
#include <stdint.h>

#include <libavcodec/avcodec.h>
#include <libavutil/audio_fifo.h>
#include <libavutil/channel_layout.h>
#include <libavutil/samplefmt.h>
#include <libswresample/swresample.h>


typedef struct {
    enum AVSampleFormat sample_fmt;
    AVChannelLayout ch_layout;
    int sample_rate;
} AudioInfo;

typedef struct {
    SwrContext *resample_context;
    AudioInfo output_info;
    uint8_t **samples;
    size_t max_samples;
} AudioResampler;

typedef struct {
    AVAudioFifo *fifo;
    size_t min_samples;
    AudioInfo out_info;
    AVFrame *frame;
} AudioFifo;

/* Audio Resampler */

bool audio_resampler_init(AudioResampler *ar,
                          AudioInfo input_info,
                          AudioInfo output_info);

bool ar_realloc_samples(AudioResampler *ar,
                        int frame_size);

bool ar_resample(AudioResampler *ar,
                 const uint8_t **input,
                 size_t nb_samples);

uint8_t **ar_get_samples(AudioResampler *ar);

/* Audio FIFO */

AudioFifo audio_fifo_init(AudioInfo *output_codec_context, size_t min_samples);

bool af_write_sample(AudioFifo *af,
                     uint8_t **data,
                     int nb_samples);

AVFrame *af_read_frame(AudioFifo *af);



#include "resampler.h"
#include "utils.h"
#include <libavformat/avformat.h>
#include <libavformat/avio.h>
#include <libavutil/error.h>
#include <libavutil/mem.h>
#include <sched.h>
#include <stdbool.h>
#include <stdint.h>
#include <stdio.h>

#include <libavcodec/avcodec.h>

#include <libavutil/audio_fifo.h>
#include <libavutil/avassert.h>
#include <libavutil/avstring.h>
#include <libavutil/channel_layout.h>
#include <libavutil/frame.h>
#include <libavutil/opt.h>

#include <libswresample/swresample.h>

bool audio_resampler_init(AudioResampler *ar, AudioInfo input_info,
                          AudioInfo output_info) {
  int error = swr_alloc_set_opts2(
      &ar->resample_context, &output_info.ch_layout, output_info.sample_fmt,
      output_info.sample_rate, &input_info.ch_layout, input_info.sample_fmt,
      input_info.sample_rate, 0, NULL);
  if (error < 0) {
    _LOG_ERROR("Could not allocate resample context: %s", av_err2str(error));
    return false;
  }
  // av_assert0(output_info.sample_rate ==
  // input_info.sample_rate); // AAC should be 48khz

  if ((error = swr_init(ar->resample_context)) < 0) {
    _LOG_ERROR("Could not open resample context: %s", av_err2str(error));
    swr_free(&ar->resample_context);
    return false;
  }

  ar->output_info = output_info;
  return true;
}

bool ar_realloc_samples(AudioResampler *ar, int frame_size) {
  if (ar->max_samples < frame_size) {
    uint8_t ***samples = &ar->samples;
    if (*samples != NULL) {
      av_freep(&(*samples)[0]);
      av_freep(&(*samples));
    }
    int err = av_samples_alloc_array_and_samples(
        samples, NULL, ar->output_info.ch_layout.nb_channels, frame_size,
        ar->output_info.sample_fmt, 0);

    if (err < 0) {
      _LOG_ERROR("Could not allocate converted input samples: %s",
                 av_err2str(err));
      return false;
    }
    ar->max_samples = frame_size;
  }

  return true;
}

bool ar_resample(AudioResampler *ar, uint8_t const **input, size_t nb_samples) {
  bool ok = ar_realloc_samples(ar, nb_samples);
  if (!ok) return false;

  int err = swr_convert(ar->resample_context, ar->samples, ar->max_samples,
                        input, ar->max_samples);
  if (err < 0) {
    _LOG_ERROR("could not convert input samples: %s", av_err2str(err));
    return false;
  }
  return true;
}

uint8_t **ar_get_samples(AudioResampler *ar) { return ar->samples; }

AudioFifo audio_fifo_init(AudioInfo *out, size_t min_samples) {
#define AUDIO_FIFO_SIZE_TYP_SHI (4096 * 2)
  AVAudioFifo *fifo = NULL;
  fifo = av_audio_fifo_alloc(out->sample_fmt, out->ch_layout.nb_channels,
                             AUDIO_FIFO_SIZE_TYP_SHI);
  assert(fifo != NULL && "Could not allocate FIFO");

  AVFrame *frame = av_frame_alloc();
  assert(frame != NULL && "failed to allocate avframe!");

  // todo change this to 1024 or smth idk
  // max is AUDIO_FIFO_SIZE_TYP_SHI but this is only for
  // allocation
  frame->nb_samples = AUDIO_FIFO_SIZE_TYP_SHI;
  frame->format = out->sample_fmt;
  frame->ch_layout = out->ch_layout;
  frame->pts = 0;

  int err = av_frame_get_buffer(frame, 0);

  AudioFifo f = {
      .fifo = fifo,
      .min_samples = min_samples,
      .out_info = *out,
      .frame = frame,
  };

  return f;
}

bool af_write_sample(AudioFifo *af, uint8_t **data, const int nb_samples) {
  int bytes_wrote = av_audio_fifo_write(af->fifo, (void **)data, nb_samples);
  if (bytes_wrote < nb_samples) {
    _LOG_ERROR("Could not write data to FIFO %s", av_err2str(bytes_wrote));
    return false;
  }
  return true;
}

AVFrame *af_read_frame(AudioFifo *af) {
  if (av_audio_fifo_size(af->fifo) < af->min_samples) return NULL;

  AVFrame *frame = af->frame;

  // todo use smth like avcodec_fill_audio_frame but that will
  // require custom fifo impl instead of this
  int samples =
      av_audio_fifo_read(af->fifo, (void **)frame->data, af->min_samples);

  if (samples < 0) {
    _LOG_ERROR("failed to read data from fifo! %s", av_err2str(samples));
    return NULL;
  }

  frame->nb_samples = samples;

  return frame;
}

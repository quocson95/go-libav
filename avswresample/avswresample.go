package avswresample

// #include <libavutil/mathematics.h>
// #include <libswresample/swresample.h>
// #include <libavutil/avutil.h>
// #include <libavutil/frame.h>
// #define MAX_AUDIO_FRAME_SIZE 192000
// struct AudioFrameInfo {
// 	float t;
// 	float tincr;
// 	float tincr2;
// } AudioFrameInfo;
// static void go_av_swresample_free(SwrContext *swrCtx) {
// 	swr_free(&swrCtx);
// }
// static double go_sin(double x){
//	double sign=1;
//	if (x<0){
//		sign=-1.0;
//		x=-x;
//	}
//	if (x > 360)
//		x = x - (int)(x/360)*360;
//	x*=M_PI/180.0;
//	double res=0;
//	double term=x;
//	int k=1;
//	while (res+term!=res){
//		res+=term;
//		k+=2;
//		term*=-x*x/k/(k-1);
//	}
//
//	return sign*res;
//}
//static struct AudioFrameInfo go_get_audio_frame(AVFrame *frame, int nb_channels, float t, float tincr, float tincr2)
//{
// int j, i, v;
// struct AudioFrameInfo audioFrameInfo;
// audioFrameInfo.t = t;
// audioFrameInfo.tincr = tincr;
// audioFrameInfo.tincr2 = tincr2;
// int16_t *q;
// q = (int16_t*)frame->data[0];
// for (j = 0; j < frame->nb_samples; j++) {
// 	v = (int)(go_sin(audioFrameInfo.t) * 10000);
// 	for (i = 0; i < nb_channels; i++) {
// 		*q++ = v;
// 	}
// 	audioFrameInfo.t     += audioFrameInfo.tincr;
// 	audioFrameInfo.tincr += audioFrameInfo.tincr2;
// }
// return audioFrameInfo;
//}
// #cgo pkg-config: libswresample libavutil
import "C"
import (
	"errors"
	"github.com/baohavan/go-libav/avcodec"
	"github.com/baohavan/go-libav/avutil"
	"unsafe"
)

type SwrContext struct {
	CAVSwrContext uintptr
}

func NewSwrContext(inputCtx *avcodec.Context, outputCtx *avcodec.Context) (*SwrContext, error) {
	swrCtxOut := SwrContext{}
	outputChannels, _ := avutil.FindDefaultChannelLayout(outputCtx.Channels())
	inputChannels, _ := avutil.FindDefaultChannelLayout(inputCtx.Channels())
	swrCtxOut.CAVSwrContext = uintptr(unsafe.Pointer(C.swr_alloc_set_opts((*C.SwrContext)(C.NULL),
		(C.int64_t)(outputChannels),
		(C.enum_AVSampleFormat)(outputCtx.SampleFormat()), (C.int)(outputCtx.SampleRate()),
		(C.int64_t)(inputChannels),
		(C.enum_AVSampleFormat)(inputCtx.SampleFormat()), (C.int)(inputCtx.SampleRate()),
		0, C.NULL)))

	if swrCtxOut.CAVSwrContext == 0 {
		return nil, errors.New("Could not allocate swresample context\n")
	}

	return &swrCtxOut, nil
}

func (swr *SwrContext) Init() error {
	if (int)(C.swr_init((*C.SwrContext)(unsafe.Pointer(swr.CAVSwrContext)))) < 0 {
		return errors.New("Could not init swresample context\n")
	}
	return nil
}

func (swr *SwrContext) Free() {
	C.go_av_swresample_free((*C.SwrContext)(unsafe.Pointer(swr.CAVSwrContext)))
	swr.CAVSwrContext = 0
}

func (swr *SwrContext) Close() {
	C.swr_close((*C.SwrContext)(unsafe.Pointer(swr.CAVSwrContext)))
}

func (swr *SwrContext) SwrConvert(frame *avutil.Frame, frameBuffer *avutil.Frame) error {
	errCode := C.swr_convert((*C.SwrContext)(unsafe.Pointer(swr.CAVSwrContext)), (**C.uchar)(frameBuffer.ExtendedData()),
		(C.int)(frameBuffer.NumberOfSamples()),
		(**C.uint8_t)(frame.ExtendedData()), (C.int)(frame.NumberOfSamples()))
	if (int)(errCode) < 0 {
		return avutil.NewErrorFromCode((avutil.ErrorCode)(errCode))
	}
	return nil
}

func Get_Audio_Frame(frame *avutil.Frame, nb_channels int, t float64, tincr float64, tincr2 float64) (float64, float64, float64) {
	var audioFrameInfo C.struct_AudioFrameInfo
	audioFrameInfo = C.go_get_audio_frame((*C.AVFrame)(unsafe.Pointer(frame.CAVFrame)), (C.int)(nb_channels),
		C.float(t), C.float(tincr), C.float(tincr2))
	return float64(audioFrameInfo.t), float64(audioFrameInfo.tincr), float64(audioFrameInfo.tincr2)
}

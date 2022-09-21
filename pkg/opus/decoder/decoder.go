package decoder

/*
#cgo pkg-config: opus
#include <opus.h>

int
bridge_decoder_get_last_packet_duration(OpusDecoder *st, opus_int32 *samples)
{
	return opus_decoder_ctl(st, OPUS_GET_LAST_PACKET_DURATION(samples));
}
*/
import "C"

import (
	"errors"
	"fmt"
	"unsafe"
)

type Decoder struct {
	p *C.struct_OpusDecoder
	// Same purpose as encoder struct
	mem         []byte
	sample_rate int
	channels    int
	opus_data   []byte
}

// NewDecoder allocates a new Opus decoder and initializes it with the
// appropriate parameters. All related memory is managed by the Go GC.
func NewOpusDecoder(sample_rate int, channels int) (*Decoder, error) {
	var dec Decoder
	err := dec.Init(sample_rate, channels)
	if err != nil {
		return nil, err
	}
	return &dec, nil
}

func (dec *Decoder) Init(sample_rate int, channels int) error {
	if dec.p != nil {
		return fmt.Errorf("opus decoder already initialized")
	}
	if channels != 1 && channels != 2 {
		return fmt.Errorf("Number of channels must be 1 or 2: %d", channels)
	}
	size := C.opus_decoder_get_size(C.int(channels))
	dec.sample_rate = sample_rate
	dec.channels = channels
	dec.mem = make([]byte, size)
	fmt.Println("decode init size:", size)
	dec.p = (*C.OpusDecoder)(unsafe.Pointer(&dec.mem[0]))
	errno := C.opus_decoder_init(
		dec.p,
		C.opus_int32(sample_rate),
		C.int(channels))
	if errno != 0 {
		return errors.New("errno")
	}
	return nil
}
func (dec *Decoder) SetOpusData(data []byte) error {
	dec.opus_data = data // *(*[]byte)(unsafe.Pointer(&data))
	return nil
}

//这里做一个fifo，wirte在on.track中调用，read在play中调用
func (dec *Decoder) Read(pcm []byte) (int, error) {
	if dec.p == nil {
		return 0, fmt.Errorf("opus decoder uninitialized")
	}
	//fmt.Println("2:", len(dec.opus_data)) //, &dec.opus_data)
	if len(dec.opus_data) == 0 {
		return 0, fmt.Errorf("opus: no data supplied")
	}
	if len(pcm) == 0 {
		return 0, fmt.Errorf("opus: target buffer empty")
	}
	if cap(pcm)%dec.channels != 0 {
		return 0, fmt.Errorf("opus: target buffer capacity must be multiple of channels")
	}
	n := int(C.opus_decode(
		dec.p,
		(*C.uchar)(&dec.opus_data[0]),
		C.opus_int32(len(dec.opus_data)),
		(*C.opus_int16)((*int16)(unsafe.Pointer(&pcm[0]))),
		C.int((cap(pcm)/dec.channels)/2),
		0))
	if n < 0 {
		return 0, errors.New("n<0")
	}
	return n * 2, nil
}

func (dec *Decoder) Write(pcm []byte) (int, error) {
	dec.opus_data = pcm
	length := len(dec.opus_data)
	//fmt.Printf("lenght:%d\n", length)
	return length, nil
}

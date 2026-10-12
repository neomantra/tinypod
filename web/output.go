package web

import (
	"github.com/NimbleMarkets/booba-shim/audio"
)

// Output plays tinyplayer's stereo frames through an audio.Player.
type Output struct {
	Player  audio.Player
	samples []int16
}

// NewOutput plays through Web Audio, or keeps time silently where there is
// none, so the progress bar and the LEDs follow the song either way.
func NewOutput() *Output {
	p, err := audio.New(audio.Options{})
	if err != nil {
		p = audio.NewSilent()
	}
	return &Output{Player: p}
}

func (o *Output) SetSampleFrequency(freq uint32) error {
	return o.Player.SetSampleRate(int(freq), 2)
}

func (o *Output) Enable(on bool) {
	o.Player.Enable(on)
}

// WriteStereo unpacks the frames, left in the low half and right in the high
// half, into interleaved samples.
func (o *Output) WriteStereo(b []uint32) (int, error) {
	o.samples = o.samples[:0]
	for _, f := range b {
		o.samples = append(o.samples, int16(f), int16(f>>16))
	}
	_, err := o.Player.Write(o.samples)
	return len(b), err
}

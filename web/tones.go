package web

import (
	"encoding/binary"
	"math"
)

const (
	toneRate = 22050
	noteTime = 0.3
)

// note frequencies in Hz
const (
	c4   = 261.63
	d4   = 293.66
	e4   = 329.63
	f4   = 349.23
	g4   = 392.00
	a4   = 440.00
	b4   = 493.88
	c5   = 523.25
	e5   = 659.25
	g5   = 783.99
	rest = 0
)

// Tones returns a few short songs made of sine notes, so the player has
// something to play when there are no WAV files.
func Tones() Memory {
	scale := []float64{c4, d4, e4, f4, g4, a4, b4, c5, rest, c5, b4, a4, g4, f4, e4, d4, c4, rest}
	arpeggio := []float64{c4, e4, g4, c5, e5, g5, e5, c5, g4, e4, c4, rest}
	chime := []float64{e5, c5, g4, rest, rest, e5, c5, g4, rest, rest}
	return Memory{
		"/Demo Tones/01 Scale.wav":    wavPCM16(toneRate, melody(repeat(scale, 2))),
		"/Demo Tones/02 Arpeggio.wav": wavPCM16(toneRate, melody(repeat(arpeggio, 3))),
		"/Chime.wav":                  wavPCM16(toneRate, melody(chime)),
	}
}

func repeat(notes []float64, n int) []float64 {
	var out []float64
	for range n {
		out = append(out, notes...)
	}
	return out
}

// melody plays each note for noteTime with a fade out.
func melody(notes []float64) []int16 {
	n := int(noteTime * toneRate)
	out := make([]int16, 0, len(notes)*n)
	for _, f := range notes {
		for i := range n {
			t := float64(i) / toneRate
			env := 1 - float64(i)/float64(n)
			v := math.Sin(2*math.Pi*f*t) * env * 0.5
			out = append(out, int16(v*32767))
		}
	}
	return out
}

// wavPCM16 wraps mono 16 bit samples in a WAV file.
func wavPCM16(rate int, samples []int16) []byte {
	le := binary.LittleEndian
	data := 2 * len(samples)
	b := make([]byte, 44+data)
	copy(b, "RIFF")
	le.PutUint32(b[4:], uint32(36+data))
	copy(b[8:], "WAVEfmt ")
	le.PutUint32(b[16:], 16)
	le.PutUint16(b[20:], 1)
	le.PutUint16(b[22:], 1)
	le.PutUint32(b[24:], uint32(rate))
	le.PutUint32(b[28:], uint32(2*rate))
	le.PutUint16(b[32:], 2)
	le.PutUint16(b[34:], 16)
	copy(b[36:], "data")
	le.PutUint32(b[40:], uint32(data))
	for i, s := range samples {
		le.PutUint16(b[44+2*i:], uint16(s))
	}
	return b
}

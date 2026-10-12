package web

import (
	"time"
)

// Silent is a tinyplayer.Output that plays nothing but takes as long as the
// audio would, so the progress bar and the LEDs follow the song.
type Silent struct {
	rate uint32
	next time.Time
}

func (s *Silent) SetSampleFrequency(freq uint32) error {
	s.rate = freq
	return nil
}

func (s *Silent) Enable(bool) {}

func (s *Silent) WriteStereo(b []uint32) (int, error) {
	if s.rate == 0 {
		return len(b), nil
	}
	now := time.Now()
	if s.next.Before(now) {
		s.next = now
	}
	s.next = s.next.Add(time.Duration(len(b)) * time.Second / time.Duration(s.rate))
	time.Sleep(s.next.Sub(now))
	return len(b), nil
}

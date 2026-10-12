package web

import (
	"embed"
	"io/fs"

	"github.com/hybridgroup/tinypod"
)

// Music holds the WAV files in the music folder, which are built into the
// page. Each folder is an album.
//
//go:embed music
var Music embed.FS

// DefaultStorage returns the songs in the music folder, or the demo tones
// when the folder has no WAV files.
func DefaultStorage() tinypod.Storage {
	music, err := fs.Sub(Music, "music")
	if err == nil {
		if st := NewFS(music); HasSongs(st) {
			return st
		}
	}
	return Tones()
}

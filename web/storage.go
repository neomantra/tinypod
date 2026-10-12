package web

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"slices"
	"strings"

	"github.com/hybridgroup/tinypod"
)

// FS is a read only tinypod.Storage over an fs.FS, such as an embed.FS.
type FS struct {
	fsys fs.FS
}

// NewFS returns a Storage that reads songs from fsys.
func NewFS(fsys fs.FS) *FS {
	return &FS{fsys: fsys}
}

// fsPath turns a tinypod path such as /Music/Album into an fs.FS path.
func fsPath(p string) string {
	p = strings.TrimPrefix(p, "/")
	if p == "" {
		return "."
	}
	return p
}

func (f *FS) List(dir string) (songs, dirs []string, err error) {
	entries, err := fs.ReadDir(f.fsys, fsPath(dir))
	if err != nil {
		return nil, nil, err
	}
	for _, e := range entries {
		name := e.Name()
		switch {
		case strings.HasPrefix(name, "."):
		case e.IsDir():
			dirs = append(dirs, name)
		case isWAV(name):
			songs = append(songs, dir+"/"+name)
		}
	}
	slices.Sort(songs)
	slices.Sort(dirs)
	return songs, dirs, nil
}

func (f *FS) Open(path string) (io.ReadCloser, error) {
	return f.fsys.Open(fsPath(path))
}

func (f *FS) Changed() bool { return false }
func (f *FS) Reload() error { return nil }
func (f *FS) USB() bool     { return false }

// HasSongs reports whether storage has at least one song.
func HasSongs(st tinypod.Storage) bool {
	return hasSongs(st, "")
}

func hasSongs(st tinypod.Storage, dir string) bool {
	songs, dirs, err := st.List(dir)
	if err != nil {
		return false
	}
	if len(songs) > 0 {
		return true
	}
	for _, d := range dirs {
		if hasSongs(st, dir+"/"+d) {
			return true
		}
	}
	return false
}

// Memory is a tinypod.Storage of songs held in memory. The keys are paths
// such as /Album/01 Song.wav.
type Memory map[string][]byte

func (m Memory) List(dir string) (songs, dirs []string, err error) {
	found := dir == ""
	for p := range m {
		rest, ok := strings.CutPrefix(p, dir+"/")
		if !ok {
			continue
		}
		found = true
		if i := strings.IndexByte(rest, '/'); i >= 0 {
			if !slices.Contains(dirs, rest[:i]) {
				dirs = append(dirs, rest[:i])
			}
		} else if isWAV(rest) {
			songs = append(songs, p)
		}
	}
	if !found {
		return nil, nil, errors.New("not found: " + dir)
	}
	slices.Sort(songs)
	slices.Sort(dirs)
	return songs, dirs, nil
}

func (m Memory) Open(path string) (io.ReadCloser, error) {
	b, ok := m[path]
	if !ok {
		return nil, errors.New("not found: " + path)
	}
	return io.NopCloser(bytes.NewReader(b)), nil
}

func (m Memory) Changed() bool { return false }
func (m Memory) Reload() error { return nil }
func (m Memory) USB() bool     { return false }

func isWAV(name string) bool {
	return strings.HasSuffix(strings.ToLower(name), ".wav")
}

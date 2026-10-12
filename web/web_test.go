package web

import (
	"encoding/base64"
	"encoding/json"
	"image/color"
	"io"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"testing/fstest"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/NimbleMarkets/go-booba/actions"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/hybridgroup/tinyplayer/wav"

	"github.com/hybridgroup/tinypod"
)

func TestFS(t *testing.T) {
	fsys := fstest.MapFS{
		"README.md":             {Data: []byte("x")},
		"Single.wav":            {Data: []byte("s")},
		"Album/02 B.wav":        {Data: []byte("b")},
		"Album/01 A.WAV":        {Data: []byte("a")},
		"Album/.hidden.wav":     {Data: []byte("h")},
		"Album/Disc 1/03 C.wav": {Data: []byte("c")},
		"Zed/notes.txt":         {Data: []byte("n")},
	}
	st := NewFS(fsys)
	songs, dirs, err := st.List("")
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(songs, []string{"/Single.wav"}) || !slices.Equal(dirs, []string{"Album", "Zed"}) {
		t.Fatalf("root: %v %v", songs, dirs)
	}
	songs, dirs, err = st.List("/Album")
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(songs, []string{"/Album/01 A.WAV", "/Album/02 B.wav"}) || !slices.Equal(dirs, []string{"Disc 1"}) {
		t.Fatalf("album: %v %v", songs, dirs)
	}
	if _, _, err := st.List("/Music"); err == nil {
		t.Fatal("missing folder listed")
	}
	f, err := st.Open("/Album/Disc 1/03 C.wav")
	if err != nil {
		t.Fatal(err)
	}
	b, _ := io.ReadAll(f)
	f.Close()
	if string(b) != "c" {
		t.Fatalf("read %q", b)
	}
	if !HasSongs(st) || HasSongs(NewFS(fstest.MapFS{"Zed/notes.txt": {Data: []byte("n")}})) {
		t.Fatal("HasSongs")
	}
}

func TestMemory(t *testing.T) {
	st := Memory{"/A/1.wav": {1}, "/A/2.wav": {2}, "/b.wav": {3}, "/A/x.txt": {4}}
	songs, dirs, err := st.List("")
	if err != nil || !slices.Equal(songs, []string{"/b.wav"}) || !slices.Equal(dirs, []string{"A"}) {
		t.Fatalf("root: %v %v %v", songs, dirs, err)
	}
	songs, dirs, err = st.List("/A")
	if err != nil || !slices.Equal(songs, []string{"/A/1.wav", "/A/2.wav"}) || len(dirs) != 0 {
		t.Fatalf("A: %v %v %v", songs, dirs, err)
	}
	if _, _, err := st.List("/Music"); err == nil {
		t.Fatal("missing folder listed")
	}
	if _, err := st.Open("/nope.wav"); err == nil {
		t.Fatal("opened missing song")
	}
}

func TestTones(t *testing.T) {
	st := Tones()
	if !HasSongs(st) {
		t.Fatal("no tones")
	}
	for path := range st {
		f, err := st.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		var dec wav.Decoder
		if err := dec.Reset(f); err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		if fm := dec.Format(); fm.SampleRate != toneRate || fm.Channels != 1 {
			t.Fatalf("%s: format %+v", path, fm)
		}
		if dec.Frames() < toneRate {
			t.Fatalf("%s: only %d frames", path, dec.Frames())
		}
	}
}

func TestSilentPaces(t *testing.T) {
	var out Silent
	out.SetSampleFrequency(8000)
	start := time.Now()
	for range 4 {
		out.WriteStereo(make([]uint32, 200))
	}
	if el := time.Since(start); el < 80*time.Millisecond {
		t.Fatalf("800 frames at 8 kHz took %v", el)
	}
}

type display struct{ frames int }

// counting counts the audio buffers written to a Silent output.
type counting struct {
	Silent
	n atomic.Int32
}

func (c *counting) WriteStereo(b []uint32) (int, error) {
	c.n.Add(1)
	return c.Silent.WriteStereo(b)
}

func (*display) Size() (int16, int16)              { return 128, 64 }
func (*display) SetPixel(int16, int16, color.RGBA) {}
func (d *display) Display() error                  { d.frames++; return nil }

func TestModelKeys(t *testing.T) {
	m := NewModel(tinypod.Config{}, DefaultKeys())
	for _, tc := range []struct {
		msg tea.KeyPressMsg
		ev  tinypod.Event
	}{
		{tea.KeyPressMsg{Code: tea.KeyUp}, tinypod.Up},
		{tea.KeyPressMsg{Code: tea.KeyDown}, tinypod.Down},
		{tea.KeyPressMsg{Code: tea.KeyLeft}, tinypod.Left},
		{tea.KeyPressMsg{Code: tea.KeyRight}, tinypod.Right},
		{tea.KeyPressMsg{Code: tea.KeyEnter}, tinypod.Select},
		{tea.KeyPressMsg{Code: tea.KeyEscape}, tinypod.Back},
		{tea.KeyPressMsg{Code: tea.KeySpace, Text: " "}, tinypod.PlayPause},
		{tea.KeyPressMsg{Code: '+', Text: "+"}, tinypod.VolumeUp},
		{tea.KeyPressMsg{Code: '-', Text: "-"}, tinypod.VolumeDown},
		{tea.KeyPressMsg{Code: 'n', Text: "n"}, tinypod.Next},
		{tea.KeyPressMsg{Code: 'p', Text: "p"}, tinypod.Previous},
		{tea.KeyPressMsg{Code: 'x', Text: "x"}, tinypod.None},
	} {
		m.Update(tc.msg)
		if ev := m.in.Poll(); ev != tc.ev {
			t.Errorf("%s: event %v, want %v", tc.msg.String(), ev, tc.ev)
		}
	}
	if m.in.Poll() != tinypod.None {
		t.Fatal("event left over")
	}
}

// TestRun drives tinypod through the model with the demo tones and checks
// that it draws and plays.
func TestRun(t *testing.T) {
	d := &display{}
	out := &counting{}
	wrapped, m, err := New(tinypod.Config{Display: d, Theme: tinypod.MonoTheme(), Output: out, Storage: Tones()})
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan tea.Msg, 1)
	go func() { done <- m.Init()() }()
	wait := func(cond func() bool) {
		for i := 0; i < 300 && !cond(); i++ {
			time.Sleep(10 * time.Millisecond)
		}
		if !cond() {
			t.Fatal("timed out")
		}
	}
	wait(func() bool { return d.frames > 0 })
	frames := d.frames
	wrapped.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	wait(func() bool { return d.frames > frames })
	for _, k := range []rune{tea.KeyEnter, tea.KeyEnter, tea.KeyEnter} {
		wrapped.Update(tea.KeyPressMsg{Code: k})
		time.Sleep(50 * time.Millisecond)
	}
	wait(func() bool { return out.n.Load() > 0 })
	select {
	case msg := <-done:
		t.Fatalf("tinypod stopped: %v", msg)
	default:
	}
}

func TestActions(t *testing.T) {
	wrapped, _, err := New(tinypod.Config{Display: &display{}, Theme: tinypod.MonoTheme(), Output: &Silent{}, Storage: Tones()})
	if err != nil {
		t.Fatal(err)
	}
	_, cmd := wrapped.Update(uv.UnknownOscEvent(actions.QueryProgramStatus))
	s := snapshot(t, cmd)
	if len(s.Short) != 5 || len(s.Full) != 3 || len(s.Actions) != 11 {
		t.Fatalf("short %d full %d actions %d", len(s.Short), len(s.Full), len(s.Actions))
	}
	var labels []string
	for _, a := range s.Actions {
		labels = append(labels, a.Label)
	}
	for _, want := range []string{"Up", "Down", "Select", "Back", "Play/Pause", "Vol +", "Next"} {
		if !slices.Contains(labels, want) {
			t.Fatalf("no %s in %v", want, labels)
		}
	}
}

// snapshot decodes the OSC 7501 action payload from the publish command.
func snapshot(t *testing.T, cmd tea.Cmd) actions.Snapshot {
	t.Helper()
	if cmd == nil {
		t.Fatal("no action update")
	}
	msg := cmd()
	if batch, ok := msg.(tea.BatchMsg); ok {
		for _, c := range batch {
			if c != nil {
				return snapshot(t, c)
			}
		}
		t.Fatal("empty batch")
	}
	raw, ok := msg.(tea.RawMsg)
	if !ok {
		t.Fatalf("unexpected message %T", msg)
	}
	_, encoded, ok := strings.Cut(raw.Msg.(string), ":msg=")
	if !ok {
		t.Fatalf("no OSC status in %q", raw.Msg)
	}
	data, err := base64.StdEncoding.DecodeString(strings.TrimSuffix(encoded, "\x1b\\"))
	if err != nil {
		t.Fatal(err)
	}
	_, payload, ok := strings.Cut(string(data), actions.MessageMarker)
	if !ok {
		t.Fatal("no action payload")
	}
	var s actions.Snapshot
	if err := json.Unmarshal([]byte(payload), &s); err != nil {
		t.Fatal(err)
	}
	if err := s.Validate(); err != nil {
		t.Fatal(err)
	}
	return s
}

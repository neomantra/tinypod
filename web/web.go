// Package web runs tinypod in a web browser. tinypod draws on a canvas and
// Booba carries the keyboard input and the touch controls. The Bubble Tea
// model here is a thin shell. It turns key presses into tinypod events and
// publishes the key map as Booba Terminal Actions.
package web

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	browser "nimbleterminal.dev/tinybubbles/web"

	"github.com/hybridgroup/tinypod"
)

// KeyMap is the keyboard for tinypod. It implements help.KeyMap, which Booba
// turns into buttons on the page.
type KeyMap struct {
	Up, Down, Left, Right   key.Binding
	Select, Back, PlayPause key.Binding
	VolumeUp, VolumeDown    key.Binding
	Next, Previous          key.Binding
}

// DefaultKeys uses the arrows, enter, escape and space, with + and - for the
// volume and n and p for the next and previous song.
func DefaultKeys() KeyMap {
	return KeyMap{
		Up:         key.NewBinding(key.WithKeys("up"), key.WithHelp("↑", "Up")),
		Down:       key.NewBinding(key.WithKeys("down"), key.WithHelp("↓", "Down")),
		Left:       key.NewBinding(key.WithKeys("left"), key.WithHelp("←", "Left")),
		Right:      key.NewBinding(key.WithKeys("right"), key.WithHelp("→", "Right")),
		Select:     key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "Select")),
		Back:       key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "Back")),
		PlayPause:  key.NewBinding(key.WithKeys("space"), key.WithHelp("space", "Play/Pause")),
		VolumeUp:   key.NewBinding(key.WithKeys("+"), key.WithHelp("+", "Vol +")),
		VolumeDown: key.NewBinding(key.WithKeys("-"), key.WithHelp("-", "Vol -")),
		Next:       key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "Next")),
		Previous:   key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "Previous")),
	}
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Select, k.Back, k.PlayPause}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		k.ShortHelp(),
		{k.Left, k.Right, k.VolumeUp, k.VolumeDown},
		{k.Previous, k.Next},
	}
}

// Input is a tinypod.Input fed from the Bubble Tea model.
type Input struct {
	ch chan tinypod.Event
}

// NewInput returns an Input that holds a few events until tinypod polls them.
func NewInput() *Input {
	return &Input{ch: make(chan tinypod.Event, 16)}
}

// Send queues an event. It drops the event if tinypod is not keeping up.
func (in *Input) Send(ev tinypod.Event) {
	select {
	case in.ch <- ev:
	default:
	}
}

func (in *Input) Poll() tinypod.Event {
	select {
	case ev := <-in.ch:
		return ev
	default:
		return tinypod.None
	}
}

// doneMsg arrives when tinypod.Run returns, which only happens on an error.
type doneMsg struct{ err error }

// Model runs tinypod in the background and feeds it key presses.
type Model struct {
	keys KeyMap
	in   *Input
	cfg  tinypod.Config
	err  error
}

// NewModel returns a Model for cfg. It replaces cfg.Input with its own input.
func NewModel(cfg tinypod.Config, keys KeyMap) *Model {
	m := &Model{keys: keys, in: NewInput(), cfg: cfg}
	m.cfg.Input = m.in
	return m
}

// Keys returns the key map, for the help and the page controls.
func (m *Model) Keys() KeyMap { return m.keys }

// Err returns the error that stopped tinypod, if any.
func (m *Model) Err() error { return m.err }

// Init starts tinypod. Bubble Tea runs the command in its own goroutine.
func (m *Model) Init() tea.Cmd {
	return func() tea.Msg {
		return doneMsg{tinypod.Run(m.cfg)}
	}
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case doneMsg:
		m.err = msg.err
		return m, tea.Quit
	case tea.KeyPressMsg:
		if ev := m.event(msg); ev != tinypod.None {
			m.in.Send(ev)
		}
	}
	return m, nil
}

// event returns the tinypod event for a key press.
func (m *Model) event(msg tea.KeyPressMsg) tinypod.Event {
	k := m.keys
	for _, b := range [...]struct {
		key.Binding
		ev tinypod.Event
	}{
		{k.Up, tinypod.Up}, {k.Down, tinypod.Down}, {k.Left, tinypod.Left}, {k.Right, tinypod.Right},
		{k.Select, tinypod.Select}, {k.Back, tinypod.Back}, {k.PlayPause, tinypod.PlayPause},
		{k.VolumeUp, tinypod.VolumeUp}, {k.VolumeDown, tinypod.VolumeDown},
		{k.Next, tinypod.Next}, {k.Previous, tinypod.Previous},
	} {
		if key.Matches(msg, b.Binding) {
			return b.ev
		}
	}
	return tinypod.None
}

// View is empty. tinypod draws on the display, not in the terminal.
func (m *Model) View() tea.View {
	return tea.NewView("")
}

// New returns the Model wrapped so Booba shows its keys as page controls.
func New(cfg tinypod.Config) (tea.Model, *Model, error) {
	m := NewModel(cfg, DefaultKeys())
	wrapped, err := browser.WrapActions(m, "tinypod", func(model tea.Model) help.KeyMap {
		return model.(*Model).Keys()
	}, nil)
	if err != nil {
		return nil, nil, err
	}
	return wrapped, m, nil
}

module github.com/hybridgroup/tinypod/web

go 1.26.8

require (
	charm.land/bubbles/v2 v2.1.0
	charm.land/bubbletea/v2 v2.1.0
	github.com/NimbleMarkets/go-booba v0.9.1
	github.com/charmbracelet/ultraviolet v0.0.0-20261008173134-6b8d4baf91b4
	github.com/hybridgroup/tinyplayer v0.0.0-20261011093738-3dddba4996d2
	github.com/hybridgroup/tinypod v0.0.0
	nimbleterminal.dev/tinybubbles/web v0.1.1
)

require (
	charm.land/lipgloss/v2 v2.0.6 // indirect
	github.com/charmbracelet/colorprofile v0.4.3 // indirect
	github.com/charmbracelet/x/ansi v0.11.9 // indirect
	github.com/charmbracelet/x/term v0.2.2 // indirect
	github.com/charmbracelet/x/termios v0.1.1 // indirect
	github.com/charmbracelet/x/windows v0.2.2 // indirect
	github.com/clipperhouse/displaywidth v0.11.0 // indirect
	github.com/clipperhouse/uax29/v2 v2.7.0 // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/lucasb-eyer/go-colorful v1.4.1 // indirect
	github.com/mattn/go-runewidth v0.0.30 // indirect
	github.com/muesli/cancelreader v0.2.2 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/soypat/fat v0.2.2-0.20261001161623-e5421565dcac // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/xo/terminfo v1.2.0 // indirect
	golang.org/x/sync v0.23.0 // indirect
	golang.org/x/sys v0.48.0 // indirect
	tinygo.org/x/drivers v0.36.1-0.20261010154131-b620b0b7f928 // indirect
	tinygo.org/x/tinyfont v0.7.0 // indirect
)

replace github.com/hybridgroup/tinypod => ..

replace charm.land/bubbletea/v2 => github.com/neomantra/bubbletea/v2 v2.0.0-20261009201241-c397c32a5074

tool github.com/NimbleMarkets/go-booba/cmd/booba-assets

tool nimbleterminal.dev/tinybubbles/web/cmd/tinybubbles-web

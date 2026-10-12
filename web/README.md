# tinypod in the browser

This module runs tinypod on a canvas through the shared
[TinyBubbles browser harness](https://github.com/NimbleTerminal/TinyBubbles/tree/main/web)
and [Booba](https://github.com/NimbleMarkets/go-booba). The buttons come from
its Go help map. Two page LEDs show the device LED output.

To run it locally:

```sh
make -C web serve
```

Open <http://127.0.0.1:8080>. Use `PORT=8089` if port 8080 is occupied.
Pick a display, use the buttons, or focus the canvas for keyboard input.
`make -C web build` writes the static site to `web/build`.
`make -C web test` checks the app model and its WASM main.

The web module pins the harness to a TinyBubbles release in `go.mod`.
It selects Go 1.26.8 and needs Python 3. Node.js is needed only when running the
shared harness's key encoder tests. The root tinypod module is unchanged.

Use standard Go for this app's WASM build. TinyGo 0.43 currently fails in
`tinygo.org/x/drivers/internal/legacy` because its pin configuration helpers
have no WASM implementation. The shared harness itself compiles with TinyGo;
TinyBubbles Hello World checks that path in CI.

## App code

- `web.go` runs tinypod and maps its keys to events.
- `storage.go`, `music.go`, and `tones.go` supply songs.
- `output.go` adapts tinyplayer's output to the booba-shim audio player.
- `ssd1306.go` simulates an SSD1306 panel behind the real ssd1306 driver.
- `cmd/tinypod/main.go` chooses display presets and connects the app to the harness.

The canvas display, LEDs, page, key encoder, actions wrapper, licenses, and build
rules come from the shared package and tool. The Makefile generates assets at
build time; there are no local copies of the harness source files.

Put small WAV files in [music](./music) and rebuild. Without WAV files the app
uses demo tones.

## Simulated SSD1306

The `SSD1306 over I2C` screen runs the same code path as the XIAO board. The
ssd1306 driver and tinypod's chunked `Display` write to `web.Panel`, a
`drivers.I2C` that decodes the control bytes, the column and page windows and
the addressing mode into the panel's 1 bit memory. The page shows that memory
with big dots and gaps, like the OLED. Invert, display off and contrast work
through the driver's commands. Use `?screen=ssd1306` on the URL to pick it.

## Audio

`web.Output` hands tinyplayer's stereo frames to
[`booba-shim/audio`](https://github.com/NimbleMarkets/booba-shim/tree/main/audio),
which plays them through Web Audio. It schedules chunks back to back at the
song's sample rate, lets the browser resample, and blocks once the queue is
300 ms ahead of the audio clock, so the deck keeps time and the progress bar
and LEDs follow the sound. Pause drains the queued sound and suspends the
AudioContext. Browsers only start sound after a click or key press, so the
first button press on the page also unlocks the audio. Where there is no Web
Audio, the silent player from the same package keeps the same timing.

The Bubble Tea WASM replacement stays in `web/go.mod`. Go does not inherit
replacements from dependencies, so this module must pin it explicitly.

Hidden browser tabs throttle timers to about once per second. This slows every
`time.Sleep` in the Go program, including playback and LEDs. Keep the tab visible
when checking timing.

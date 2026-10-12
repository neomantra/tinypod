//go:build js && wasm

package main

import (
	"github.com/hybridgroup/tinypod"
	"github.com/hybridgroup/tinypod/web"
	browser "nimbleterminal.dev/tinybubbles/web"
)

func main() {
	selected := browser.Configure(browser.Config{
		Title: "tinypod", Source: "https://github.com/hybridgroup/tinypod",
		Description: "The TinyGo music player, running in your browser. The same Go code draws on a canvas instead of an OLED.",
		Note:        "Arrows move, Enter opens, Escape goes back, Space plays and pauses. + and - set volume, n and p skip songs. Songs play silently; this preview has no audio output.",
		Screens: []browser.Screen{
			{ID: "color", Label: "Color TFT · 320 × 240", Width: 320, Height: 240},
			{ID: "oled", Label: "Monochrome OLED · 128 × 64", Width: 128, Height: 64, Monochrome: true},
		}, LEDCount: 2,
	})
	theme := tinypod.ColorTheme()
	if selected.Screen.Monochrome {
		theme = tinypod.MonoTheme()
	}
	leds := browser.NewLEDs("led")
	model, m, err := web.New(tinypod.Config{
		Display: browser.NewCanvasDisplay("display", selected.Screen.Width, selected.Screen.Height),
		Theme:   theme, Output: &web.Silent{}, Storage: web.DefaultStorage(), LED: leds, LEDCount: leds.Len(),
	})
	if err != nil {
		browser.Stopped(err)
		select {}
	}
	browser.Run(model, browser.RunOptions{Error: m.Err})
}

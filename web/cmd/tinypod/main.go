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
		Note:        "Arrows move, Enter opens, Escape goes back, Space plays and pauses. + and - set volume, n and p skip songs. Click or press a key once to let the browser start the sound.",
		Screens: []browser.Screen{
			{ID: "color", Label: "Color TFT · 320 × 240", Width: 320, Height: 240},
			{ID: "oled", Label: "Monochrome OLED · 128 × 64", Width: 128, Height: 64, Monochrome: true},
			{ID: "ssd1306", Label: "SSD1306 over I2C · 128 × 64", Width: 128, Height: 64, Monochrome: true},
		}, LEDCount: 2,
	})
	theme := tinypod.ColorTheme()
	if selected.Screen.Monochrome {
		theme = tinypod.MonoTheme()
	}
	w, h := selected.Screen.Width, selected.Screen.Height
	var display tinypod.Display = browser.NewCanvasDisplay("display", w, h)
	if selected.Screen.ID == "ssd1306" {
		// The real ssd1306 driver draws on a simulated panel, shown with
		// big dots and gaps like the OLED.
		scale := int16(web.PanelScale)
		display = web.NewPanelDisplay(browser.NewCanvasDisplay("display", w*scale, h*scale), w, h)
	}
	leds := browser.NewLEDs("led")
	model, m, err := web.New(tinypod.Config{
		Display: display,
		Theme:   theme, Output: web.NewOutput(), Storage: web.DefaultStorage(), LED: leds, LEDCount: leds.Len(),
	})
	if err != nil {
		browser.Stopped(err)
		select {}
	}
	browser.Run(model, browser.RunOptions{Error: m.Err})
}

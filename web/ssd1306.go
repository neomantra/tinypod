package web

import (
	"image/color"

	"github.com/hybridgroup/tinypod"
)

// Panel is a simulated SSD1306. It is a drivers.I2C that decodes what the
// ssd1306 driver sends into the panel's 1 bit memory, 8 pages of 128
// columns with 8 vertical pixels per byte. See the SSD1306 datasheet,
// sections 8.7 and 10.1, https://cdn-shop.adafruit.com/datasheets/SSD1306.pdf.
type Panel struct {
	w, h   int
	mem    [8 * 128]byte
	shown  [8 * 128]byte
	first  bool
	colS   int
	colE   int
	pageS  int
	pageE  int
	col    int
	page   int
	mode   byte
	on     bool
	invert bool
	allOn  bool
	bright uint8

	cmd  byte
	args int
}

// NewPanel returns a panel with the visible size w by h.
func NewPanel(w, h int) *Panel {
	p := &Panel{w: w, h: h, first: true}
	p.colE, p.pageE = 127, 7
	p.bright = 0x7F
	return p
}

// argCount is the number of argument bytes of each multi byte command.
var argCount = map[byte]int{
	0x20: 1, 0x21: 2, 0x22: 2, 0x81: 1, 0x8D: 1, 0xA8: 1, 0xD3: 1, 0xD5: 1, 0xD9: 1, 0xDA: 1, 0xDB: 1,
	0x26: 6, 0x27: 6, 0x29: 5, 0x2A: 5, 0xA3: 2,
}

// Tx decodes one I2C write. The first byte is a control byte. Bit 6 picks
// data or command, and bit 7 means only the next byte belongs to it.
func (p *Panel) Tx(addr uint16, w, r []byte) error {
	for len(w) > 0 {
		ctrl := w[0]
		w = w[1:]
		n := len(w)
		if ctrl&0x80 != 0 {
			n = min(n, 1)
		}
		if ctrl&0x40 != 0 {
			for _, b := range w[:n] {
				p.data(b)
			}
		} else {
			for _, b := range w[:n] {
				p.command(b)
			}
		}
		w = w[n:]
	}
	return nil
}

func (p *Panel) command(b byte) {
	if p.args > 0 {
		p.args--
		p.argument(b)
		return
	}
	p.cmd = b
	p.args = argCount[b]
	switch {
	case b == 0xAE:
		p.on = false
	case b == 0xAF:
		p.on = true
	case b == 0xA6:
		p.invert = false
	case b == 0xA7:
		p.invert = true
	case b == 0xA4:
		p.allOn = false
	case b == 0xA5:
		p.allOn = true
	case b >= 0xB0 && b <= 0xB7:
		p.page = int(b - 0xB0)
	case b < 0x10:
		p.col = p.col&0xF0 | int(b)
	case b < 0x20:
		p.col = p.col&0x0F | int(b-0x10)<<4
	}
}

func (p *Panel) argument(b byte) {
	switch p.cmd {
	case 0x20:
		p.mode = b & 3
	case 0x21:
		if p.args == 1 {
			p.colS = int(b & 0x7F)
			p.col = p.colS
		} else {
			p.colE = int(b & 0x7F)
		}
	case 0x22:
		if p.args == 1 {
			p.pageS = int(b & 7)
			p.page = p.pageS
		} else {
			p.pageE = int(b & 7)
		}
	case 0x81:
		p.bright = b
	}
}

// data writes one byte at the cursor and moves it on, as in the datasheet
// section 10.1.3 for the addressing modes.
func (p *Panel) data(b byte) {
	p.mem[p.page&7*128+p.col&127] = b
	switch p.mode {
	case 1:
		p.page++
		if p.page > p.pageE {
			p.page = p.pageS
			p.col++
			if p.col > p.colE {
				p.col = p.colS
			}
		}
	case 2:
		p.col++
		if p.col > p.colE {
			p.col = p.colS
		}
	default:
		p.col++
		if p.col > p.colE {
			p.col = p.colS
			p.page++
			if p.page > p.pageE {
				p.page = p.pageS
			}
		}
	}
}

// Pixel reports whether the pixel at x, y is lit.
func (p *Panel) Pixel(x, y int) bool {
	if x < 0 || y < 0 || x >= p.w || y >= p.h || !p.on {
		return false
	}
	lit := p.allOn || p.mem[y/8*128+x]>>(y%8)&1 != 0
	return lit != p.invert
}

// On returns the color of a lit pixel at the current contrast.
func (p *Panel) On() color.RGBA {
	b := 160 + int(p.bright)*95/255
	return color.RGBA{uint8(b * 190 / 255), uint8(b * 225 / 255), uint8(b), 255}
}

var off = color.RGBA{8, 10, 14, 255}

// draw copies the panel to d, scale pixels per panel pixel with a dark gap,
// like the dots of an OLED. It only draws the bytes that changed.
func (p *Panel) draw(d tinypod.Display, scale int) {
	on := p.On()
	for i := range p.mem {
		if !p.first && p.mem[i] == p.shown[i] {
			continue
		}
		p.shown[i] = p.mem[i]
		x, page := i%128, i/128
		if x >= p.w || page*8 >= p.h {
			continue
		}
		for bit := 0; bit < 8 && page*8+bit < p.h; bit++ {
			y := page*8 + bit
			c := off
			if p.Pixel(x, y) {
				c = on
			}
			for dy := 0; dy < scale-1; dy++ {
				for dx := 0; dx < scale-1; dx++ {
					d.SetPixel(int16(x*scale+dx), int16(y*scale+dy), c)
				}
			}
		}
	}
	p.first = false
}

// PanelScale is how many display pixels a PanelDisplay uses per panel pixel.
const PanelScale = 4

// PanelDisplay runs the real ssd1306 driver over a simulated Panel and shows
// the panel on out, scaled up by PanelScale. out must be that many times
// larger than the panel.
type PanelDisplay struct {
	*tinypod.SSD1306
	panel *Panel
	out   tinypod.Display
}

// NewPanelDisplay returns a w by h SSD1306 shown on out.
func NewPanelDisplay(out tinypod.Display, w, h int16) *PanelDisplay {
	panel := NewPanel(int(w), int(h))
	d := &PanelDisplay{SSD1306: tinypod.NewSSD1306(panel, 0x3C, w, h), panel: panel, out: out}
	d.fill(off)
	return d
}

func (d *PanelDisplay) fill(c color.RGBA) {
	w, h := d.out.Size()
	for y := int16(0); y < h; y++ {
		for x := int16(0); x < w; x++ {
			d.out.SetPixel(x, y, c)
		}
	}
}

// Panel returns the simulated panel.
func (d *PanelDisplay) Panel() *Panel { return d.panel }

// Display sends the frame over the simulated bus and shows the result.
func (d *PanelDisplay) Display() error {
	if err := d.SSD1306.Display(); err != nil {
		return err
	}
	d.panel.draw(d.out, PanelScale)
	return d.out.Display()
}

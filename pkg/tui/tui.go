package tui

import (
	"os"
	"strings"

	"github.com/fatih/color"
)

var (
	BOLD  = "\033[1m"
	DIM   = "\033[2m"
	RESET = "\033[0m"
)

var (
	RED     = "\033[31m"
	GREEN   = "\033[32m"
	YELLOW  = "\033[33m"
	BLUE    = "\033[34m"
	CYAN    = "\033[36m"
	MAGENTA = "\033[35m"
	WHITE   = "\033[97m"
	BLACK   = "\033[30m"
)

var (
	BACKDARKGRAY  = "\033[100m"
	BACKRED       = "\033[41m"
	BACKGREEN     = "\033[42m"
	BACKYELLOW    = "\033[43m"
	BACKLIGHTBLUE = "\033[104m"
)

var (
	FGRed     = color.FgRed
	FGGreen   = color.FgGreen
	FGYellow  = color.FgYellow
	FGBlue    = color.FgBlue
	FGCyan    = color.FgCyan
	FGMagenta = color.FgMagenta
	FGWhite   = color.FgWhite
	FGBLACK   = color.FgBlack
)

var (
	BGRed       = color.BgRed
	BGGreen     = color.BgGreen
	BGYellow    = color.BgYellow
	BGBlue      = color.BgBlue
	BGCyan      = color.BgCyan
	BGHiRed     = color.BgHiRed
	BGHiCyan    = color.BgHiCyan
	BGDarkGray  = color.BgHiBlack
	BGMagenta   = color.BgMagenta
	BGLightBlue = color.BgHiBlue
	BGWhite     = color.BgWhite
)

var (
	BOLDCOLOR = color.Bold
)

var ctrl = []string{"\x033", "\\e", "\x1b"}

func Style(text string, fg color.Attribute, bg color.Attribute, attrs ...color.Attribute) string {
	var c *color.Color

	if fg == color.FgBlack {
		c = color.RGB(0, 0, 0)
	} else {
		c = color.New(fg)
	}

	if bg == color.BgBlack {
		c = c.Add(color.Attribute(color.BgHiBlack))
	} else {
		c.Add(bg)
	}

	if len(attrs) > 0 {
		c.Add(attrs...)
	}

	return c.Sprint(text)
}

func Color(text string, fg color.Attribute, attrs ...color.Attribute) string {
	var c *color.Color

	if fg == color.FgBlack {
		c = color.RGB(0, 0, 0)
		if len(attrs) > 0 {
			c.Add(attrs...)
		}
	} else {
		allAttrs := append([]color.Attribute{fg}, attrs...)
		c = color.New(allAttrs...)
	}

	return c.Sprint(text)
}

func Fg(text string, fg color.Attribute) string {
	return color.New(fg).Sprint(text)
}

func Bg(text string, bg color.Attribute) string {
	return color.New(bg).Sprint(text)
}

func Wrap(e, s string) string { return e + s + RESET }
func Bold(s string) string    { return Wrap(BOLD, s) }
func Dim(s string) string     { return Wrap(DIM, s) }
func Red(s string) string     { return Wrap(RED, s) }
func Green(s string) string   { return Wrap(GREEN, s) }
func Yellow(s string) string  { return Wrap(YELLOW, s) }
func Blue(s string) string    { return Wrap(BLUE, s) }
func Cyan(s string) string    { return Wrap(CYAN, s) }
func Magenta(s string) string { return Wrap(MAGENTA, s) }
func White(s string) string   { return Wrap(WHITE, s) }
func Black(s string) string   { return Wrap(BLACK, s) }

func Effects() bool {
	term := os.Getenv("TERM")
	return term != "" && term != "dumb"
}

func Disable() {
	color.NoColor = true
	BOLD, DIM, RESET = "", "", ""
	RED, GREEN, YELLOW, BLUE, CYAN, MAGENTA, WHITE, BLACK = "", "", "", "", "", "", "", ""
}

func HasEffect(s string) bool {
	for _, ch := range ctrl {
		if strings.Contains(s, ch) {
			return true
		}
	}
	return false
}

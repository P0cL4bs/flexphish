package cli

import (
	"flexphish/pkg/tui"
	"fmt"
	"runtime"

	"github.com/fatih/color"
)

func PrintBanner() {
	light := color.New(color.BgHiGreen).SprintFunc()
	magenta := color.New(color.FgMagenta).SprintFunc()
	dark := color.New(color.FgHiBlack).SprintFunc()
	logo := []string{
		"        ████        ",
		"   ██████████████   ",
		"  █████▓▓▓▓▓▓█████  ",
		"  ███▓▓▓▓▓▓░░▓▓███  ",
		" ███▓▓▓▓▓▓▓██▓▓▓███ ",
		"████▓▓▓██▓▓██▓▓▓████",
		" ███▓▓▓██▓▓██▓▓▓███ ",
		"  ███▓▓▓████▓▓▓███  ",
		"  █████▓▓▓▓▓▓█████  ",
		"   ██████████████   ",
		"        ████        ",
	}
	coloredLogo := make([]string, len(logo))

	for i, line := range logo {
		coloredLine := ""
		for _, ch := range line {
			switch ch {
			case '░':
				coloredLine += light(string(ch))
			case '▓':
				coloredLine += magenta(string(ch))
			case '█':
				coloredLine += dark(string(ch))
			default:
				coloredLine += string(ch)
			}
		}
		coloredLogo[i] = coloredLine
	}
	appBuild := fmt.Sprintf("[built for %s %s]", runtime.GOOS, runtime.GOARCH)
	info := []string{
		"                                                    ",
		"██████ ▄▄    ▄▄▄▄▄ ▄▄ ▄▄ ▄▄▄▄  ▄▄ ▄▄ ▄▄  ▄▄▄▄ ▄▄ ▄▄",
		"██▄▄   ██    ██▄▄  ▀█▄█▀ ██▄█▀ ██▄██ ██ ███▄▄ ██▄██",
		"██     ██▄▄▄ ██▄▄▄ ██ ██ ██    ██ ██ ██ ▄▄██▀ ██ ██",
		"                                        " + tui.Dim("version") + " " + tui.Color(Version, color.FgGreen, color.Bold) + "",
		"The ultimate Red Team toolkit for phishing operations.",
		"",
		appBuild,
		" by: " + tui.Color("@mh4x0f", color.FgHiYellow, color.Bold) +
			" • PocL4bs Team - " + tui.Color("10 Years", color.FgHiMagenta, color.Bold),
		"",
	}

	offset := 2

	maxLines := len(coloredLogo)
	if len(info)+offset > maxLines {
		maxLines = len(info) + offset
	}

	for i := 0; i < maxLines; i++ {
		var logoLine, infoLine string

		if i < len(logo) {
			logoLine = coloredLogo[i]
		} else {
			logoLine = " "
		}

		if i >= offset && i-offset < len(info) {
			infoLine = info[i-offset]
		} else {
			infoLine = ""
		}

		fmt.Printf("%-20s   %s\n", logoLine, infoLine)
	}
}

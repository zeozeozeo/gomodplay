package main

import (
	"github.com/gdamore/tcell/v2"
)

func drawBox(s tcell.Screen, x1, y1, x2, y2 int) {
	style := tcell.StyleDefault.Background(boxBgColour).Foreground(boxFgColour)

	// Fill background
	for row := y1; row <= y2; row++ {
		for col := x1; col <= x2; col++ {
			s.SetContent(col, row, ' ', nil, style)
		}
	}

	// Draw borders
	for col := x1; col <= x2; col++ {
		s.SetContent(col, y1, '─', nil, style)
		s.SetContent(col, y2, '─', nil, style)
	}

	for row := y1 + 1; row < y2; row++ {
		s.SetContent(x1, row, tcell.RuneVLine, nil, style)
		s.SetContent(x2, row, tcell.RuneVLine, nil, style)
	}

	// Only draw corners if necessary
	if y1 != y2 && x1 != x2 {
		s.SetContent(x1, y1, '╭', nil, style)
		s.SetContent(x2, y1, '╮', nil, style)
		s.SetContent(x1, y2, '╰', nil, style)
		s.SetContent(x2, y2, '╯', nil, style)
	}
}

func drawText(s tcell.Screen, x, y, width, height int, style tcell.Style, text string) {
	xPos := x
	yPos := y
	for _, r := range []rune(text) {
		s.SetContent(xPos, yPos, r, nil, style)
		xPos++
		if xPos > x+width {
			yPos++
			xPos = x
		}
		if yPos > y+height {
			return
		}
	}

	for yPos < y+height {
		for xPos < x+width {
			s.SetContent(xPos, yPos, ' ', nil, style)
			xPos++
		}
		yPos++
		xPos = x
	}
}

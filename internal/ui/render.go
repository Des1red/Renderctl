package ui

import (
	"github.com/gdamore/tcell/v2"
)

const (
	uiBoxWidth     = 80
	uiBoxMargin    = 2
	uiBoxMaxHeight = 24

	maxTextWidth = uiBoxWidth - 6 // left padding + safety
)

// MODE SELECTION SCREEN
func renderModeScreen(s tcell.Screen, styles UIStyles, selected int) {
	s.Clear()

	w, h := s.Size()

	boxW := uiBoxWidth
	boxH := h - (uiBoxMargin * 2)
	if boxH > uiBoxMaxHeight {
		boxH = uiBoxMaxHeight
	}

	x := (w - boxW) / 2
	y := (h - boxH) / 2
	if y < uiBoxMargin {
		y = uiBoxMargin
	}

	drawBox(s, styles.Border, x, y, boxW, boxH)

	drawText(s, styles.Title, x+4, y+2, "renderctl v2")
	drawText(s, styles.Normal, x+4, y+4, "Select execution mode")

	for i, m := range modes {
		style := styles.Normal
		prefix := "  "
		if i == selected {
			prefix = "> "
			style = styles.Select
		}
		drawText(s, style, x+4, y+6+i, prefix+m)
	}

	drawText(s, styles.Dim, x+4, y+boxH-3, "↑ ↓ move    Enter select    q quit")

	s.Show()
}

// INPUT HANDLING

func renderInputScreen(
	s tcell.Screen,
	styles UIStyles,
	fields []Field,
	selected int,
	editMode bool,
	editBuffer string,
) {
	s.Clear()
	w, h := s.Size()

	boxW := uiBoxWidth
	boxH := h - (uiBoxMargin * 2)
	if boxH > uiBoxMaxHeight {
		boxH = uiBoxMaxHeight
	}

	y := (h - boxH) / 2
	if y < uiBoxMargin {
		y = uiBoxMargin
	}

	x := (w - boxW) / 2

	drawBox(s, styles.Border, x, y, boxW, boxH)

	drawText(s, styles.Title, x+3, y+1, "renderctl configuration")
	drawText(s, styles.Dim, x+3, y+3, "↑ ↓ navigate   Enter edit/toggle   Esc back   q quit")

	rowY := y + 5
	selectIdx := 0 // counts ONLY selectable items
	lastSection := ""

	// ----- FIELDS -----
	for i := 0; i < len(fields); i++ {
		f := fields[i]
		section := fieldSection(f.Label)

		if section != "" && section != lastSection {
			drawText(s, styles.Dim, x+3, rowY, "-- "+section+" --")
			rowY++
			lastSection = section
		}

		style := styles.Normal
		if selectIdx == selected {
			if editMode {
				style = styles.Edit
			} else {
				style = styles.Select
			}
		}

		drawFieldRow(
			s,
			styles,
			f,
			x+3,
			rowY,
			selectIdx == selected,
			editMode,
			editBuffer,
			style,
		)

		rowY++
		selectIdx++
	}

	// spacer
	rowY++
	// ----- EXECUTE -----
	style := styles.Normal
	if selectIdx == selected {
		style = styles.Select
	}
	drawText(s, style, x+3, rowY, "[ Execute ]")
	rowY++
	selectIdx++

	// ----- BACK -----
	style = styles.Normal
	if selectIdx == selected {
		style = styles.Select
	}
	drawText(s, style, x+3, rowY, "[ Back to mode selection ]")

	s.Show()
}

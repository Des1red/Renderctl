package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
)

func renderState(
	screen tcell.Screen,
	styles UIStyles,
	state uiState,
	selectedMode int,
	fields []Field,
	selectedField int,
	editMode bool,
	editBuffer string,
	ctx *uiContext,
	confirmSelected int,
) {
	switch state {
	case stateModeSelect:
		renderModeScreen(screen, styles, selectedMode)

	case stateConfig:
		renderInputScreen(
			screen,
			styles,
			ctx.working.Mode,
			fields,
			selectedField,
			editMode,
			editBuffer,
			ctx,
		)

	case stateConfirm:
		renderConfirmScreen(
			screen,
			styles,
			ctx.working.Mode,
			fields,
			confirmSelected,
		)
	}
}

const (
	uiBoxWidth     = 80
	uiBoxMargin    = 2
	uiBoxMaxHeight = 26

	footerHeight = 3 // hint + breathing room

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

		// mode name
		drawText(s, style, x+4, y+6+i, prefix+m)

		// description (dim, offset)
		desc := modeDescription(m)
		if desc != "" {
			drawText(
				s,
				styles.Dim,
				x+22, // spacing between name and description
				y+6+i,
				desc,
			)
		}
	}

	drawText(s, styles.Dim, x+4, y+boxH-3, "↑ ↓ move    Enter select    q quit")

	s.Show()
}

// INPUT HANDLING

func renderInputScreen(
	s tcell.Screen,
	styles UIStyles,
	mode string,
	fields []Field,
	selected int,
	editMode bool,
	editBuffer string,
	ctx *uiContext,
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
	contentMaxY := y + boxH - footerHeight

	x := (w - boxW) / 2

	drawBox(s, styles.Border, x, y, boxW, boxH)

	drawText(
		s,
		styles.Title,
		x+3,
		y+1,
		configHeaderForMode(mode),
	)
	drawText(s, styles.Dim, x+3, y+3, "↑ ↓ navigate   Enter edit/toggle   Esc back")

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

		disabled := isFieldDisabled(f, ctx)

		style := styles.Normal
		if disabled {
			style = styles.Dim
		} else if selectIdx == selected {
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
	execDisabled := isExecuteDisabled(ctx)

	if rowY < contentMaxY {
		style := styles.Normal
		if execDisabled {
			style = styles.Dim
		} else if selectIdx == selected {
			style = styles.Select
		}

		drawText(
			s,
			style,
			x+3,
			rowY,
			fmt.Sprintf("[ %s ]", executeLabelForMode(mode)),
		)
	}
	if execDisabled {
		reason := executeDisableReason(ctx)
		if reason != "" {
			rowY++
			drawText(
				s,
				styles.Dim,
				x+6,
				rowY,
				"↳ "+reason,
			)
		}
	}

	rowY++
	selectIdx++

	// ----- BACK -----
	style := styles.Normal
	if selectIdx == selected {
		style = styles.Select
	}
	if rowY < contentMaxY {
		drawText(s, style, x+3, rowY, "[ Back to mode selection ]")
	}

	s.Show()
}

func renderConfirmScreen(
	s tcell.Screen,
	styles UIStyles,
	mode string,
	fields []Field,
	selected int,
) {
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

	rowY := y + 5

	// dump fields (simple version, ordered later)
	printKV := func(label, value string, valueStyle tcell.Style) {
		drawText(s, styles.Label, x+3, rowY, fmt.Sprintf("%-22s", label))
		drawText(s, styles.Dim, x+26, rowY, "→")
		drawText(s, valueStyle, x+30, rowY, value)
		rowY++
	}

	// --- minimal explicit list (safe & clear) ---
	lastSection := ""
	for _, f := range fields {
		section := fieldSection(f.Label)
		if section != "" && section != lastSection {
			drawText(
				s,
				styles.Dim,
				x+3,
				rowY,
				"-- "+section+" --",
			)
			rowY++
			lastSection = section
		}

		var (
			value string
			style = styles.Normal
		)

		switch f.Type {
		case FieldBool:
			if f.Bool != nil && *f.Bool {
				value = "Yes"
			} else {
				value = "No"
			}

		case FieldString:
			if f.String != nil && *f.String != "" {
				value = *f.String
			} else {
				value = "(not set)"
				style = styles.Dim
			}

		case FieldInt:
			if f.Int != nil && *f.Int != 0 {
				value = fmt.Sprintf("%d", *f.Int)
			} else {
				value = "(default)"
				style = styles.Dim
			}
		}

		printKV(f.Label, value, style)

	}

	// title
	drawText(
		s,
		styles.Title,
		x+3,
		y+1,
		confirmTitleForMode(mode),
	)
	// subtitle
	drawText(
		s,
		styles.Dim,
		x+3,
		y+2,
		confirmSubtitleForMode(mode),
	)
	// ... fields rendering ...

	styleConfirm := styles.Normal
	styleCancel := styles.Normal

	if selected == 0 {
		styleConfirm = styles.Select
	} else {
		styleCancel = styles.Select
	}

	rowY++
	drawText(s, styleConfirm, x+3, rowY, "[ Confirm & Execute ]")
	rowY++
	drawText(s, styleCancel, x+3, rowY, "[ Cancel ]")

	s.Show()
}

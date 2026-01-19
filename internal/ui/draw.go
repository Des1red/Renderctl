package ui

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
)

func drawBox(s tcell.Screen, style tcell.Style, x, y, w, h int) {
	s.SetContent(x, y, '╔', nil, style)
	s.SetContent(x+w-1, y, '╗', nil, style)
	s.SetContent(x, y+h-1, '╚', nil, style)
	s.SetContent(x+w-1, y+h-1, '╝', nil, style)

	for i := 1; i < w-1; i++ {
		s.SetContent(x+i, y, '═', nil, style)
		s.SetContent(x+i, y+h-1, '═', nil, style)
	}

	for i := 1; i < h-1; i++ {
		s.SetContent(x, y+i, '║', nil, style)
	}
}

func drawText(s tcell.Screen, style tcell.Style, x, y int, text string) {
	for i, r := range text {
		s.SetContent(x+i, y, r, nil, style)
	}
}

func clipText(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 1 {
		return s[:max]
	}
	return s[:max-1] + "…"
}

func drawFieldRow(
	s tcell.Screen,
	styles UIStyles,
	f Field,
	x int,
	y int,
	isSelected bool,
	editMode bool,
	editBuffer string,
	style tcell.Style,
) {
	// label
	labelText := fmt.Sprintf("%-22s", f.Label)

	// value
	value := ""
	switch f.Type {
	case FieldBool:
		if *f.Bool {
			value = "Yes"
		} else {
			value = "No"
		}

	case FieldString:
		value = *f.String

	case FieldInt:
		value = fmt.Sprintf("%d", *f.Int)

	case FieldDuration:
		if f.Duration != nil && *f.Duration > 0 {
			value = fmt.Sprintf("%d", int(*f.Duration/time.Second))
		} else {
			value = "0"
		}
	}

	if editMode && isSelected {
		value = editBuffer
	}

	// draw label (dark red)
	drawText(s, styles.Label, x, y, labelText)

	// draw value
	drawText(
		s,
		style,
		x+len(labelText),
		y,
		" : "+clipText(value, maxTextWidth-len(labelText)-3),
	)
}

func drawModesStatic(
	s tcell.Screen,
	styles UIStyles,
	x, y int,
) {
	drawText(s, styles.Normal, x+4, y+4, "Select execution mode")

	for i, m := range modes {
		drawText(
			s,
			styles.Dim, // dim during boot
			x+4,
			y+6+i,
			"  "+m,
		)

		desc := modeDescription(m)
		if desc != "" {
			drawText(
				s,
				styles.Dim,
				x+22,
				y+6+i,
				desc,
			)
		}
	}

	s.Show()
}

// ====================== //
// ANIMATION INFRASTRUCT //
// ====================== //

type drawOp func()

// ====================== //
// BOOT SCREEN SEQUENCE //
// ====================== //

func drawBootScreen(s tcell.Screen, styles UIStyles) {
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

	ops := make(chan drawOp, 256)

	// ---- SINGLE SCREEN OWNER ----
	go func() {
		for op := range ops {
			op()
			s.Show()
		}
	}()

	// ---- PHASE 1: BOX ----
	boxDone := make(chan struct{})
	drawBoxAnimatedAsync(
		ops,
		s,
		styles.Border,
		x, y, boxW, boxH,
		5,
		boxDone,
	)
	<-boxDone

	// ---- PHASE 2: TITLE + MODES (PARALLEL) ----
	titleDone := make(chan struct{})
	modesDone := make(chan struct{})

	drawTextAnimatedLTRAsync(
		ops,
		s,
		styles.Title,
		x+4,
		y+2,
		"renderctl v2",
		30,
		titleDone,
	)

	drawModesAnimatedAsync(
		ops,
		s,
		styles,
		x,
		y,
		5,
		modesDone,
	)

	<-titleDone
	<-modesDone

	// ---- PHASE 3: BORDER SWEEP ----
	sweepDone := make(chan struct{})
	drawClockwiseBorderSweepAsync(
		ops,
		s,
		styles.Normal,
		styles.Border,
		x, y, boxW, boxH,
		4,
		sweepDone,
	)
	<-sweepDone

	close(ops)
}

// ====================== //
// ANIMATION HELPERS     //
// ====================== //

func drawTextAnimatedLTRAsync(
	out chan<- drawOp,
	s tcell.Screen,
	style tcell.Style,
	x, y int,
	text string,
	delayMs int,
	done chan<- struct{},
) {
	go func() {
		for i, r := range text {
			out <- func(cx int, rr rune) drawOp {
				return func() {
					s.SetContent(cx, y, rr, nil, style)
				}
			}(x+i, r)

			time.Sleep(time.Duration(delayMs) * time.Millisecond)
		}
		close(done)
	}()
}

func drawBoxAnimatedAsync(
	out chan<- drawOp,
	s tcell.Screen,
	style tcell.Style,
	x, y, w, h int,
	delayMs int,
	done chan<- struct{},
) {
	go func() {
		// top
		for i := 1; i < w-1; i++ {
			out <- func(ix int) drawOp {
				return func() {
					s.SetContent(x+ix, y, '═', nil, style)
				}
			}(i)
			time.Sleep(time.Duration(delayMs) * time.Millisecond)
		}

		// bottom
		for i := 1; i < w-1; i++ {
			out <- func(ix int) drawOp {
				return func() {
					s.SetContent(x+ix, y+h-1, '═', nil, style)
				}
			}(i)
			time.Sleep(time.Duration(delayMs) * time.Millisecond)
		}

		// left
		for i := 1; i < h-1; i++ {
			out <- func(iy int) drawOp {
				return func() {
					s.SetContent(x, y+iy, '║', nil, style)
				}
			}(i)
			time.Sleep(time.Duration(delayMs) * time.Millisecond)
		}

		// corners
		out <- func() drawOp {
			return func() {
				s.SetContent(x, y, '╔', nil, style)
				s.SetContent(x+w-1, y, '╗', nil, style)
				s.SetContent(x, y+h-1, '╚', nil, style)
				s.SetContent(x+w-1, y+h-1, '╝', nil, style)
			}
		}()

		close(done)
	}()
}

func drawModesAnimatedAsync(
	out chan<- drawOp,
	s tcell.Screen,
	styles UIStyles,
	x, y int,
	delayMs int,
	done chan<- struct{},
) {
	go func() {
		out <- func() {
			drawText(s, styles.Normal, x+4, y+4, "Select execution mode")
		}

		time.Sleep(80 * time.Millisecond)

		for i, m := range modes {
			lineY := y + 6 + i

			text := "  " + m
			for j, r := range text {
				out <- func(cx, cy int, rr rune) drawOp {
					return func() {
						s.SetContent(cx, cy, rr, nil, styles.Dim)

					}
				}(x+4+j, lineY, r)

				time.Sleep(time.Duration(delayMs) * time.Millisecond)
			}

			desc := modeDescription(m)
			for j, r := range desc {
				out <- func(cx, cy int, rr rune) drawOp {
					return func() {
						s.SetContent(cx, cy, rr, nil, styles.Dim)

					}
				}(x+22+j, lineY, r)

				time.Sleep(time.Duration(delayMs) * time.Millisecond)
			}
		}

		close(done)
	}()
}

func drawClockwiseBorderSweepAsync(
	out chan<- drawOp,
	s tcell.Screen,
	highlight tcell.Style,
	normal tcell.Style,
	x, y, w, h int,
	delayMs int,
	done chan<- struct{},
) {
	go func() {
		type cell struct {
			x, y int
			ch   rune
		}

		const trailLen = 4
		var path []cell

		path = append(path, cell{x + w - 1, y + h - 1, '╝'})
		for i := w - 2; i > 0; i-- {
			path = append(path, cell{x + i, y + h - 1, '═'})
		}
		path = append(path, cell{x, y + h - 1, '╚'})
		for i := h - 2; i > 0; i-- {
			path = append(path, cell{x, y + i, '║'})
		}
		path = append(path, cell{x, y, '╔'})
		for i := 1; i < w-1; i++ {
			path = append(path, cell{x + i, y, '═'})
		}
		path = append(path, cell{x + w - 1, y, '╗'})

		for i := 0; i < len(path); i++ {
			c := path[i]

			out <- func(c cell, i int) drawOp {
				return func() {
					s.SetContent(c.x, c.y, c.ch, nil, highlight)
					if i >= trailLen {
						p := path[i-trailLen]
						s.SetContent(p.x, p.y, p.ch, nil, normal)
					}
				}
			}(c, i)

			time.Sleep(time.Duration(delayMs) * time.Millisecond)
		}

		// cleanup
		for i := len(path) - trailLen; i < len(path); i++ {
			if i >= 0 {
				p := path[i]
				out <- func(p cell) drawOp {
					return func() {
						s.SetContent(p.x, p.y, p.ch, nil, normal)
					}
				}(p)
			}
		}

		close(done)
	}()
}

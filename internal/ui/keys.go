package ui

import (
	"fmt"
	"os"
	"renderctl/internal/models"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
)

func handleModeSelectKey(
	ev *tcell.EventKey,
	cfg *models.Config,
	modes []string,
	selectedMode *int,
	state *uiState,
	fields *[]Field,
	selectedField *int,
	editMode *bool,
	editBuffer *string,
	screen tcell.Screen,
) {
	// hard quit
	if ev.Key() == tcell.KeyRune && ev.Rune() == 'q' {
		screen.Fini()
		os.Exit(0)
	}

	switch ev.Key() {
	case tcell.KeyUp:
		*selectedMode = (*selectedMode - 1 + len(modes)) % len(modes)

	case tcell.KeyDown:
		*selectedMode = (*selectedMode + 1) % len(modes)

	case tcell.KeyEnter:
		selected := strings.ToLower(modes[*selectedMode])

		cfg.Mode = selected

		*fields = buildFieldsForMode(cfg, cfg.Mode)
		*selectedField = 0
		*editMode = false
		*editBuffer = ""
		*state = stateConfig

	case tcell.KeyEscape:
		// do nothing, stay here
	}
}

func handleConfigKey(
	ev *tcell.EventKey,
	cfg *models.Config,
	fields []Field,
	selectedField *int,
	editMode *bool,
	editBuffer *string,
	state *uiState,
	screen tcell.Screen,
) {
	// hard quit
	if ev.Key() == tcell.KeyRune && ev.Rune() == 'q' {
		screen.Fini()
		os.Exit(0)
	}

	// ----- EDIT MODE -----
	if *editMode {
		switch ev.Key() {
		case tcell.KeyEnter:
			f := fields[*selectedField]
			if f.Type == FieldString {
				*f.String = *editBuffer
			} else if f.Type == FieldInt {
				if v, err := strconv.Atoi(*editBuffer); err == nil {
					*f.Int = v
				}
			}
			*editMode = false

		case tcell.KeyEscape:
			*editMode = false

		case tcell.KeyBackspace, tcell.KeyBackspace2:
			if len(*editBuffer) > 0 {
				*editBuffer = (*editBuffer)[:len(*editBuffer)-1]
			}

		case tcell.KeyRune:
			*editBuffer += string(ev.Rune())
		}
		return
	}

	// ----- NORMAL MODE -----
	switch ev.Key() {
	case tcell.KeyUp:
		if *selectedField > 0 {
			*selectedField--
		}

	case tcell.KeyDown:
		if *selectedField < len(fields)+1 {
			*selectedField++
		}

	case tcell.KeyEnter:

		// Execute
		if *selectedField == len(fields) {
			*state = stateExit
			return
		}

		// Back
		if *selectedField == len(fields)+1 {
			*selectedField = 0
			*editMode = false
			*editBuffer = ""
			*state = stateModeSelect
			return
		}

		// Normal field
		f := fields[*selectedField]
		switch f.Type {
		case FieldBool:
			*f.Bool = !*f.Bool
		case FieldString:
			*editMode = true
			*editBuffer = *f.String
		case FieldInt:
			*editMode = true
			*editBuffer = fmt.Sprintf("%d", *f.Int)
		}

	case tcell.KeyEscape:
		*selectedField = 0
		*editMode = false
		*editBuffer = ""
		*state = stateModeSelect
	}
}

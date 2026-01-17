package ui

import (
	"renderctl/internal/models"

	"github.com/gdamore/tcell/v2"
)

var modes = []string{
	"Auto",
	"Stream",
	"Scan",
	"Manual",
	"Cache",
}

func Run(cfg *models.Config) {
	screen, err := tcell.NewScreen()
	if err != nil {
		return
	}
	if err := screen.Init(); err != nil {
		return
	}
	defer screen.Fini()

	styles := defaultStyles()

	state := stateModeSelect
	selectedMode := 0

	var fields []Field
	selectedField := 0
	editMode := false
	editBuffer := ""

	for state != stateExit {
		switch state {
		case stateModeSelect:
			renderModeScreen(screen, styles, selectedMode)

		case stateConfig:
			renderInputScreen(
				screen,
				styles,
				fields,
				selectedField,
				editMode,
				editBuffer,
			)
		}

		ev := screen.PollEvent()
		switch ev := ev.(type) {

		case *tcell.EventKey:
			if state == stateModeSelect {
				handleModeSelectKey(
					ev, cfg, modes,
					&selectedMode, &state,
					&fields, &selectedField,
					&editMode, &editBuffer,
					screen,
				)
			} else {
				handleConfigKey(
					ev, cfg, fields,
					&selectedField,
					&editMode,
					&editBuffer,
					&state,
					screen,
				)
			}

		case *tcell.EventResize:
			screen.Sync()
		}
	}
	screen.Fini()
	return
}

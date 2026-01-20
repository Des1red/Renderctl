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

	// UI context
	ctx := &uiContext{cfg: cfg}
	ctx.bootSkip = make(chan struct{})
	ctx.bootDoneCh = make(chan struct{})

	state := stateBoot
	confirmSelected := 0
	selectedMode := 0

	var fields []Field
	selectedField := 0
	editMode := false
	editBuffer := ""

	for state != stateExit {

		renderState(
			screen,
			styles,
			state,
			selectedMode,
			fields,
			selectedField,
			editMode,
			editBuffer,
			ctx,
			confirmSelected,
		)

		if state == stateBoot {
			// poll events asynchronously so animation can finish OR be skipped
			evCh := make(chan tcell.Event, 1)

			go func() {
				evCh <- screen.PollEvent()
			}()

			select {
			case <-ctx.bootDoneCh:
				state = stateModeSelect
				continue

			case ev := <-evCh:
				if kev, ok := ev.(*tcell.EventKey); ok {
					handleKeyEvent(
						kev,
						screen,
						styles,
						ctx,
						&state,
						&selectedMode,
						&fields,
						&selectedField,
						&editMode,
						&editBuffer,
						&confirmSelected,
					)
				}
				continue
			}
		}

		// ---- normal mode ----
		ev := screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			handleKeyEvent(
				ev,
				screen,
				styles,
				ctx,
				&state,
				&selectedMode,
				&fields,
				&selectedField,
				&editMode,
				&editBuffer,
				&confirmSelected,
			)

		case *tcell.EventResize:
			screen.Sync()
		}
	}

}

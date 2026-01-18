package ui

import (
	"strings"

	"github.com/gdamore/tcell/v2"
)

type UIStyles struct {
	Title  tcell.Style
	Normal tcell.Style
	Label  tcell.Style
	Select tcell.Style
	Edit   tcell.Style
	Dim    tcell.Style
	Border tcell.Style
}

func defaultStyles() UIStyles {
	base := tcell.StyleDefault

	return UIStyles{
		Title:  base.Foreground(tcell.NewRGBColor(120, 200, 255)).Bold(true),
		Normal: base.Foreground(tcell.NewRGBColor(220, 220, 220)),
		Select: base.Foreground(tcell.NewRGBColor(0, 255, 180)).Bold(true),
		Label:  base.Foreground(tcell.NewRGBColor(180, 180, 250)),
		Edit: base.
			Foreground(tcell.NewRGBColor(190, 120, 255)). // purple
			Background(tcell.NewRGBColor(0, 120, 140)).   // cyan-ish bg
			Bold(true),
		Dim:    base.Foreground(tcell.NewRGBColor(130, 130, 130)),
		Border: base.Foreground(tcell.NewRGBColor(90, 140, 180)),
	}
}

func fieldSection(label string) string {
	switch label {
	case "TV IP", "TV Port", "TV Path", "TV Vendor":
		return "TV"

	case "Local file", "Local IP", "Local dir", "Serve port":
		return "Local"

	case "Subnet", "Deep search", "SSDP discovery":
		return "Scan"

	case "Probe only", "Auto cache", "Use cache", "Select cache index":
		return "Cache"

	case "List cache", "Forget cache":
		return "Cache"

	default:
		return ""
	}
}

func modeDescription(mode string) string {
	switch strings.ToLower(mode) {
	case "auto":
		return "     Automatically discover and stream media"
	case "stream":
		return " Stream a local file to a known TV"
	case "scan":
		return "    Discover TVs on the local network"
	case "manual":
		return "Use manually specified TV parameters"
	case "cache":
		return "  Inspect or modify cached devices"
	default:
		return ""
	}
}

func configHeaderForMode(mode string) string {
	switch mode {
	case "scan":
		return "Scan configuration"
	case "stream":
		return "Stream configuration"
	case "manual":
		return "Manual configuration"
	case "cache":
		return "Cache configuration"
	case "auto":
		return "Automatic configuration"
	default:
		return "Configuration"
	}
}

func executeLabelForMode(mode string) string {
	switch mode {
	case "scan":
		return "Execute scan"
	case "stream":
		return "Start stream"
	case "manual":
		return "Execute manual"
	case "cache":
		return "Manage cache"
	case "auto":
		return "Run automatic"
	default:
		return "Execute"
	}
}

func confirmTitleForMode(mode string) string {
	switch mode {
	case "scan":
		return "Confirm scan execution"
	case "stream":
		return "Confirm stream execution"
	case "manual":
		return "Confirm manual execution"
	case "cache":
		return "Confirm cache operation"
	case "auto":
		return "Confirm automatic execution"
	default:
		return "Confirm execution"
	}
}

func confirmSubtitleForMode(mode string) string {
	switch mode {
	case "scan":
		return "Discover TVs on the selected subnet"
	case "stream":
		return "Stream the selected media to the TV"
	case "manual":
		return "Use manual TV connection parameters"
	case "cache":
		return "Inspect or modify cached devices"
	case "auto":
		return "Automatically discover and stream media"
	default:
		return ""
	}
}

func executeDisableReason(ctx *uiContext) string {
	switch ctx.working.Mode {

	case "auto":
		if ctx.working.TIP == "" {
			return "TV IP is required"
		}
		if ctx.working.LIP == "" {
			return "Local IP is required"
		}
		if !ctx.working.ProbeOnly && ctx.working.LFile == "" {
			return "Local file required unless probing"
		}

	case "stream", "manual":
		if ctx.working.TIP == "" {
			return "TV IP is required"
		}
		if ctx.working.LIP == "" {
			return "Local IP is required"
		}
		if ctx.working.LFile == "" {
			return "Local file is required"
		}
	}

	return ""
}

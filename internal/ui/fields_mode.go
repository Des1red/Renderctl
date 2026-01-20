package ui

import "renderctl/internal/models"

func buildFieldsForMode(cfg *models.Config, mode string) []Field {
	switch mode {

	case "scan":
		return []Field{
			{Label: "Subnet", Type: FieldString, String: &cfg.Subnet},
			{Label: "Deep search", Type: FieldBool, Bool: &cfg.DeepSearch},
			{Label: "SSDP discovery", Type: FieldBool, Bool: &cfg.Discover},
			{Label: "SSDP timeout (seconds)", Type: FieldDuration, Duration: &cfg.SSDPTimeout},
			{Label: "TV IP", Type: FieldString, String: &cfg.TIP},
			{Label: "Local IP", Type: FieldString, String: &cfg.LIP},
			{Label: "Auto cache", Type: FieldBool, Bool: &cfg.AutoCache},
			{Label: "Use cache", Type: FieldBool, Bool: &cfg.UseCache},
		}

	case "auto":
		return []Field{
			{Label: "TV IP", Type: FieldString, String: &cfg.TIP},
			{Label: "TV Port", Type: FieldString, String: &cfg.TPort},
			{Label: "TV Vendor", Type: FieldString, String: &cfg.TVVendor},

			{Label: "Local file", Type: FieldString, String: &cfg.LFile},
			{Label: "Local IP", Type: FieldString, String: &cfg.LIP},
			{Label: "Local dir", Type: FieldString, String: &cfg.LDir},
			{Label: "Serve port", Type: FieldString, String: &cfg.ServePort},

			// auto-only
			{Label: "SSDP discovery", Type: FieldBool, Bool: &cfg.Discover},
			{Label: "SSDP timeout (seconds)", Type: FieldDuration, Duration: &cfg.SSDPTimeout},
			{Label: "Probe only", Type: FieldBool, Bool: &cfg.ProbeOnly},

			{Label: "Auto cache", Type: FieldBool, Bool: &cfg.AutoCache},
			{Label: "Use cache", Type: FieldBool, Bool: &cfg.UseCache},
			{Label: "Select cache index", Type: FieldInt, Int: &cfg.SelectCache},
		}

	case "stream":
		return []Field{
			{Label: "TV IP", Type: FieldString, String: &cfg.TIP},
			{Label: "TV Port", Type: FieldString, String: &cfg.TPort},
			{Label: "TV Vendor", Type: FieldString, String: &cfg.TVVendor},

			{Label: "Local file", Type: FieldString, String: &cfg.LFile},
			{Label: "Local IP", Type: FieldString, String: &cfg.LIP},
			{Label: "Local dir", Type: FieldString, String: &cfg.LDir},
			{Label: "Serve port", Type: FieldString, String: &cfg.ServePort},

			{Label: "Auto cache", Type: FieldBool, Bool: &cfg.AutoCache},
			{Label: "Use cache", Type: FieldBool, Bool: &cfg.UseCache},
			{Label: "Select cache index", Type: FieldInt, Int: &cfg.SelectCache},
		}

	case "manual":
		return []Field{
			{Label: "Local file", Type: FieldString, String: &cfg.LFile},
			{Label: "Local IP", Type: FieldString, String: &cfg.LIP},
			{Label: "Local dir", Type: FieldString, String: &cfg.LDir},
			{Label: "Serve port", Type: FieldString, String: &cfg.ServePort},

			{Label: "TV IP", Type: FieldString, String: &cfg.TIP},
			{Label: "TV Port", Type: FieldString, String: &cfg.TPort},
			{Label: "TV Path", Type: FieldString, String: &cfg.TPath},
			{Label: "TV Vendor", Type: FieldString, String: &cfg.TVVendor},

			{Label: "Auto cache", Type: FieldBool, Bool: &cfg.AutoCache},
			{Label: "Use cache", Type: FieldBool, Bool: &cfg.UseCache},
			{Label: "Select cache index", Type: FieldInt, Int: &cfg.SelectCache},
		}

	case "cache":
		return []Field{
			{Label: "List cache", Type: FieldBool, Bool: &cfg.ListCache},
			{Label: "Forget cache", Type: FieldString, String: &cfg.ForgetCache},
		}
	}

	return nil
}

func isFieldDisabled(f Field, ctx *uiContext) bool {
	switch f.Label {

	// cache logic
	case "Auto cache", "Select cache index":
		return !ctx.working.UseCache

	case "TV Port", "TV Path", "TV Vendor", "TV IP":
		return ctx.working.SelectCache >= 0
	// SSDP timeout only makes sense if SSDP is enabled
	case "SSDP timeout (seconds)":
		return !ctx.working.Discover

	// Local file disabled ONLY when auto + probe-only
	case "Local file":
		return ctx.working.Mode == "auto" && ctx.working.ProbeOnly
	// Local IP disabled when scan + SSDP enabled
	case "Local IP":
		return ctx.working.Mode == "scan" && !ctx.working.Discover

	}

	return false
}

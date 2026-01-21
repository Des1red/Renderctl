package avtransport

import (
	"encoding/xml"
	"net/http"
)

type scpd struct {
	ActionList struct {
		Actions []struct {
			Name string `xml:"name"`
		} `xml:"action"`
	} `xml:"actionList"`
}

func FetchActions(scpdURL string) (map[string]bool, error) {
	resp, err := http.Get(scpdURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var doc scpd
	if err := xml.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return nil, err
	}

	actions := make(map[string]bool)
	for _, a := range doc.ActionList.Actions {
		actions[a.Name] = true
	}

	return actions, nil
}

var safeActions = []string{
	// --- Status / state ---
	"GetTransportInfo",
	"GetPositionInfo",
	"GetMediaInfo",
	"GetDeviceCapabilities",

	// --- Settings (read-only) ---
	"GetTransportSettings",

	// --- Control (reversible / safe) ---
	"Stop",
	"Pause",
}

func ValidateActions(target Target) map[string]bool {
	valid := make(map[string]bool)

	for _, action := range safeActions {
		body := `<u:` + action + ` xmlns:u="urn:schemas-upnp-org:service:AVTransport:1">
			<InstanceID>0</InstanceID>
		</u:` + action + `>`

		ok := probeSOAPEndpoint(target.ControlURL, body)
		valid[action] = ok
	}

	return valid
}

package identity

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"time"
	"tvctrl/logger"
)

func Enrich(baseURL string, timeout time.Duration) (*Info, error) {
	client := http.Client{Timeout: timeout}

	paths := []string{
		// standard / generic
		"/device.xml",
		"/rootDesc.xml",
		"/description.xml",
		"/desc.xml",

		// UPnP common
		"/upnp/device.xml",
		"/upnp/devicedesc.xml",
		"/upnp/description.xml",
		"/upnp/desc.xml",

		// MediaRenderer / DMR
		"/dmr/device.xml",
		"/dmr/description.xml",
		"/dmr/desc.xml",
		"/MediaRenderer/device.xml",
		"/MediaRenderer/description.xml",
		"/MediaRenderer/desc.xml",

		// Samsung
		"/smp/device.xml",
		"/smp/description.xml",
		"/smp/desc.xml",
		"/AllShare/device.xml",
		"/AllShare/description.xml",

		// LG / webOS
		"/webos/device.xml",
		"/webos/description.xml",
		"/webos/desc.xml",

		// Sony / Android TV
		"/sony/device.xml",
		"/sony/description.xml",
		"/AV/device.xml",
		"/AV/description.xml",

		// Chromecast / Android-style
		"/setup/eureka_info",
		"/ssdp/device-desc.xml",

		// fallback guesses (cheap, last resort)
		"/renderer/device.xml",
		"/renderer/description.xml",
	}

	logger.Notify("Starting identity descriptor probing (%d paths)", len(paths))
	for i, p := range paths {
		fullURL := baseURL + p
		logger.Info("Identity probe [%d/%d]: %s", i+1, len(paths), fullURL)
		resp, err := client.Get(fullURL)
		if err != nil || resp.StatusCode != 200 {
			continue
		}

		var d struct {
			Device struct {
				FriendlyName string `xml:"friendlyName"`
				Manufacturer string `xml:"manufacturer"`
				ModelName    string `xml:"modelName"`
				ModelNumber  string `xml:"modelNumber"`
				UDN          string `xml:"UDN"`
				Presentation string `xml:"presentationURL"`
			} `xml:"device"`
		}

		if err := xml.NewDecoder(resp.Body).Decode(&d); err != nil {
			continue
		}
		resp.Body.Close()

		logger.Success("Identity descriptor found at: %s", fullURL)
		return &Info{
			FriendlyName: d.Device.FriendlyName,
			Manufacturer: d.Device.Manufacturer,
			ModelName:    d.Device.ModelName,
			ModelNumber:  d.Device.ModelNumber,
			UDN:          d.Device.UDN,
			Presentation: d.Device.Presentation,
		}, nil
	}
	logger.Success("Identity probing completed — ")
	return nil, fmt.Errorf("no descriptor found")
}

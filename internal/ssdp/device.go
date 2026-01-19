package ssdp

import (
	"encoding/xml"
	"net/http"
	"net/url"
	"renderctl/logger"
	"strings"
)

type DeviceDescription struct {
	Device struct {
		Manufacturer string `xml:"manufacturer"`
		ModelName    string `xml:"modelName"`
		UDN          string `xml:"UDN"`

		ServiceList struct {
			Services []struct {
				ServiceType string `xml:"serviceType"`
				ControlURL  string `xml:"controlURL"`
				SCPDURL     string `xml:"SCPDURL"`
			} `xml:"service"`
		} `xml:"serviceList"`
	} `xml:"device"`
}

type DetectedTV struct {
	IP         string
	Port       string
	Vendor     string
	ControlURL string

	AVTransportSCPD       string
	ConnectionManagerCtrl string

	UDN string
}

func FetchAndDetect(location string) (*DetectedTV, error) {
	u, err := url.Parse(location)
	if err != nil {
		return nil, err
	}
	logger.Notify("SSDP LOCATION: %s", location)

	resp, err := http.Get(location)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var dd DeviceDescription
	if err := xml.NewDecoder(resp.Body).Decode(&dd); err != nil {
		return nil, err
	}

	vendor := detectVendor(dd.Device.Manufacturer)

	var avTransportCtrl string
	var avTransportSCPD string
	var connMgrCtrl string

	for _, s := range dd.Device.ServiceList.Services {
		switch {
		case strings.Contains(s.ServiceType, "service:AVTransport"):
			avTransportCtrl = s.ControlURL
			avTransportSCPD = s.SCPDURL

		case strings.Contains(s.ServiceType, "service:ConnectionManager"):
			connMgrCtrl = s.ControlURL
		}
	}

	// Build absolute URLs
	base := u.Scheme + "://" + u.Host

	fix := func(p string) string {
		if p == "" {
			return ""
		}
		if strings.HasPrefix(p, "http://") || strings.HasPrefix(p, "https://") {
			return p
		}
		if !strings.HasPrefix(p, "/") {
			p = "/" + p
		}
		return base + p
	}

	port := u.Port()
	if port == "" {
		port = "80"
	}
	logger.Info("Detected AVTransport ControlURL: %s", fix(avTransportCtrl))
	logger.Info("Detected AVTransport SCPDURL : %s", fix(avTransportSCPD))
	logger.Info("Detected ConnMgr ControlURL  : %s", fix(connMgrCtrl))

	return &DetectedTV{
		IP:         u.Hostname(),
		Port:       port,
		Vendor:     vendor,
		ControlURL: fix(avTransportCtrl),

		AVTransportSCPD:       fix(avTransportSCPD),
		ConnectionManagerCtrl: fix(connMgrCtrl),
		UDN:                   dd.Device.UDN,
	}, nil

}

func detectVendor(m string) string {
	m = strings.ToLower(m)
	switch {
	case strings.Contains(m, "samsung"):
		return "samsung"
	case strings.Contains(m, "lg"):
		return "lg"
	case strings.Contains(m, "sony"):
		return "sony"
	case strings.Contains(m, "philips"):
		return "philips"
	default:
		return "generic"
	}
}

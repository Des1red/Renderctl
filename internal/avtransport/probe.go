package avtransport

import (
	"errors"
	"fmt"
	"time"
)

var probePorts = []string{
	"9197", // Samsung AVTransport
	"7678", // Samsung AllShare
	"8187",
	"9119",
	"8080",
}

var probePaths = []string{
	"/dmr/upnp/control/AVTransport1",
	"/upnp/control/AVTransport",
	"/MediaRenderer/AVTransport/Control",
	"/AVTransport/control",
}

func Probe(ip string, timeout time.Duration) (*Target, error) {
	deadline := time.Now().Add(timeout)

	for _, port := range probePorts {
		for _, path := range probePaths {
			if time.Now().After(deadline) {
				return nil, errors.New("AVTransport probe timed out")
			}

			controlURL := fmt.Sprintf("http://%s:%s%s", ip, port, path)

			ok := soapProbe(controlURL, "")
			if ok {
				return &Target{
					ControlURL: controlURL,
				}, nil
			}
		}
	}

	return nil, errors.New("no AVTransport endpoint found")
}

package avtransport

import (
	"fmt"
	"html"
	"renderctl/logger"
	"time"
)

type Target struct {
	ControlURL string
	MediaURL   string
}

func Run(t Target, meta string) {
	controlURL := t.ControlURL

	setAVTransportBody := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
	<SOAP-ENV:Envelope 
	    xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/" 
	    SOAP-ENV:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">
	  <SOAP-ENV:Body>
	    <u:SetAVTransportURI xmlns:u="urn:schemas-upnp-org:service:AVTransport:1">
	      <InstanceID>0</InstanceID>
	      <CurrentURI>%s</CurrentURI>
	      <CurrentURIMetaData>%s</CurrentURIMetaData>
	    </u:SetAVTransportURI>
	  </SOAP-ENV:Body>
	</SOAP-ENV:Envelope>`, t.MediaURL, html.EscapeString(meta))

	playBody := `<?xml version="1.0" encoding="utf-8"?>
	<SOAP-ENV:Envelope 
	    xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/" 
	    SOAP-ENV:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">
	  <SOAP-ENV:Body>
	    <u:Play xmlns:u="urn:schemas-upnp-org:service:AVTransport:1">
	      <InstanceID>0</InstanceID>
	      <Speed>1</Speed>
	    </u:Play>
	  </SOAP-ENV:Body>
	</SOAP-ENV:Envelope>`

	stopBody := `<?xml version="1.0" encoding="utf-8"?>
	<SOAP-ENV:Envelope 
	  xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/" 
	  SOAP-ENV:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">
	<SOAP-ENV:Body>
	<u:Stop xmlns:u="urn:schemas-upnp-org:service:AVTransport:1">
	<InstanceID>0</InstanceID>
	</u:Stop>
	</SOAP-ENV:Body>
	</SOAP-ENV:Envelope>`

	// 0) STOP (Samsung quirk)
	_ = soapRequest(
		controlURL,
		stopBody,
		`"urn:schemas-upnp-org:service:AVTransport:1#Stop"`,
	)

	time.Sleep(150 * time.Millisecond)
	// 1) Set media URI
	err := soapRequest(
		controlURL,
		setAVTransportBody,
		`"urn:schemas-upnp-org:service:AVTransport:1#SetAVTransportURI"`,
	)
	if err != nil {
		logger.Fatal("%v", err)
		return
	}

	// 2) Play
	err = soapRequest(
		controlURL,
		playBody,
		`"urn:schemas-upnp-org:service:AVTransport:1#Play"`,
	)
	if err != nil {
		logger.Fatal("%v", err)
	}
}

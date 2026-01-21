package avtransport

import (
	"bytes"
	"net/http"
	"time"
)

const probeSOAP = `<?xml version="1.0" encoding="utf-8"?>
<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/"
            s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">
  <s:Body>
    <u:GetTransportInfo xmlns:u="urn:schemas-upnp-org:service:AVTransport:1">
      <InstanceID>0</InstanceID>
    </u:GetTransportInfo>
  </s:Body>
</s:Envelope>`

func dynamicBody(body string) string {
	reqBody := `<?xml version="1.0" encoding="utf-8"?>
	<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/"
				s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">
	  <s:Body>` + body + `</s:Body>
	</s:Envelope>`

	return reqBody
}

func probeSOAPEndpoint(controlURL string, body string) bool {
	if body == "" {
		body = probeSOAP
	} else {
		body = dynamicBody(body)
	}

	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	req, err := http.NewRequest("POST", controlURL, bytes.NewBufferString(body))
	if err != nil {
		return false
	}

	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	req.Header.Set("SOAPAction", `"urn:schemas-upnp-org:service:AVTransport:1#GetTransportInfo"`)

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	// ANY SOAP response means this is a valid AVTransport endpoint
	if resp.StatusCode == 200 || resp.StatusCode == 500 {
		return true
	}

	return false
}

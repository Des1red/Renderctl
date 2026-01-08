package avtransport

import (
	"bytes"
	"encoding/xml"
	"net/http"
	"strings"
	"time"
)

type protocolInfoResp struct {
	Sink string `xml:"Body>GetProtocolInfoResponse>Sink"`
}

func FetchMediaProtocols(controlURL string) (map[string][]string, error) {
	body := `<u:GetProtocolInfo xmlns:u="urn:schemas-upnp-org:service:ConnectionManager:1"/>`

	resp, err := soapRequestRaw(
		controlURL,
		body,
		`"urn:schemas-upnp-org:service:ConnectionManager:1#GetProtocolInfo"`,
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var r protocolInfoResp
	if err := xml.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}

	media := make(map[string][]string)
	lines := strings.Split(r.Sink, ",")

	for _, l := range lines {
		parts := strings.Split(l, ":")
		if len(parts) >= 4 {
			mime := parts[2]
			profile := parts[3]
			media[mime] = append(media[mime], profile)
		}
	}

	return media, nil
}

func soapRequestRaw(controlURL, body, soapAction string) (*http.Response, error) {
	reqBody := DynamicBody(body)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	req, err := http.NewRequest("POST", controlURL, bytes.NewBufferString(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	req.Header.Set("SOAPAction", soapAction)

	return client.Do(req)
}

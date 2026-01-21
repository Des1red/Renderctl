package avtransport

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"renderctl/logger"
	"strings"
	"time"
)

type protocolInfoResp struct {
	Sink string `xml:"Body>GetProtocolInfoResponse>Sink"`
}

func soapRequest(controlURL string, body string, soapAction string) error {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	req, err := http.NewRequest("POST", controlURL, bytes.NewBufferString(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	req.Header.Set("SOAPAction", soapAction)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	logger.Info("Status: %d", resp.StatusCode)
	logger.Info("Response: %s", string(respBody))
	fmt.Println()

	return nil
}

func soapRequestRaw(controlURL, body, soapAction string) (*http.Response, error) {
	reqBody := dynamicBody(body)

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

func FetchMediaProtocols(connMgrURL string) (map[string][]string, error) {
	body := `<u:GetProtocolInfo xmlns:u="urn:schemas-upnp-org:service:ConnectionManager:1"/>`

	resp, err := soapRequestRaw(
		connMgrURL,
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

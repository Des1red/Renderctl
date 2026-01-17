package ssdp

import (
	"net"
	"renderctl/logger"
	"strings"
	"time"
)

type SSDPDevice struct {
	Location string
	Server   string
	USN      string
}

var ssdpSearches = []string{
	// --- Core ---
	"urn:schemas-upnp-org:device:MediaRenderer:1",
	"urn:schemas-upnp-org:device:MediaRenderer:2",

	// --- Services ---
	"urn:schemas-upnp-org:service:AVTransport:1",
	"urn:schemas-upnp-org:service:RenderingControl:1",
	"urn:schemas-upnp-org:service:ConnectionManager:1",

	// --- Smart TV ecosystems ---
	"urn:dial-multiscreen-org:service:dial:1",
	"urn:schemas-upnp-org:device:MediaServer:1",

	// --- Broad fallback ---
	"ssdp:all",
}

func ListenNotify(timeout time.Duration) ([]SSDPDevice, error) {
	logger.Notify("Listening for SSDP NOTIFY packets (%v)", timeout)

	addr, _ := net.ResolveUDPAddr("udp4", "239.255.255.250:1900")

	conn, err := net.ListenMulticastUDP("udp4", nil, addr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	_ = conn.SetDeadline(time.Now().Add(timeout))

	var devices []SSDPDevice
	buf := make([]byte, 2048)

	for {
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			break
		}

		resp := string(buf[:n])
		logger.Info("SSDP NOTIFY received (%d bytes)", n)
		if strings.Contains(resp, "ssdp:alive") {
			dev := parseSSDP(resp)
			if dev.Location != "" {
				logger.Success("SSDP NOTIFY device: %s", dev.Location)
				devices = append(devices, dev)
			}
		}
	}

	logger.Success("SSDP NOTIFY finished — %d device(s) found", len(devices))
	return devices, nil
}

func sendSearch(conn net.PacketConn, st string) error {
	msg := strings.Join([]string{
		"M-SEARCH * HTTP/1.1",
		"HOST: 239.255.255.250:1900",
		`MAN: "ssdp:discover"`,
		"MX: 3",
		"ST: " + st,
		"", "",
	}, "\r\n")

	dst, err := net.ResolveUDPAddr("udp4", "239.255.255.250:1900")
	if err != nil {
		return err
	}

	_, err = conn.WriteTo([]byte(msg), dst)
	return err
}

func Discover(timeout time.Duration) ([]SSDPDevice, error) {
	logger.Notify("Starting SSDP active discovery (%v)", timeout)
	conn, err := net.ListenPacket("udp4", ":0")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	devices := make(map[string]SSDPDevice) // dedupe by LOCATION
	buf := make([]byte, 2048)

	for _, st := range ssdpSearches {
		logger.Info("SSDP M-SEARCH for ST: %s", st)
		// Send search 2 times per ST
		for i := 0; i < 2; i++ {
			logger.Info("Sending SSDP M-SEARCH (%d/2) for %s", i+1, st)
			_ = sendSearch(conn, st)
			time.Sleep(150 * time.Millisecond)
		}

		_ = conn.SetDeadline(time.Now().Add(timeout))

		for {
			n, _, err := conn.ReadFrom(buf)
			if err != nil {
				break
			}

			resp := string(buf[:n])
			dev := parseSSDP(resp)
			if dev.Location != "" {
				logger.Success("SSDP response: %s", dev.Location)
				devices[dev.Location] = dev
			}
		}

		// Early exit if we found anything
		if len(devices) > 0 {
			logger.Notify("SSDP discovery early exit — device found")
			break
		}

	}

	// Convert map → slice
	var result []SSDPDevice
	for _, d := range devices {
		result = append(result, d)
	}

	logger.Success("SSDP discovery completed — %d unique device(s) found", len(result))
	return result, nil
}

func parseSSDP(resp string) SSDPDevice {
	lines := strings.Split(resp, "\r\n")
	var d SSDPDevice

	for _, l := range lines {
		l = strings.TrimSpace(l)
		switch {
		case strings.HasPrefix(strings.ToUpper(l), "LOCATION:"):
			d.Location = strings.TrimSpace(l[9:])
		case strings.HasPrefix(strings.ToUpper(l), "SERVER:"):
			d.Server = strings.TrimSpace(l[7:])
		case strings.HasPrefix(strings.ToUpper(l), "USN:"):
			d.USN = strings.TrimSpace(l[4:])
		}
	}
	logger.Result("Parsed SSDP headers: LOCATION=%s SERVER=%s USN=%s",
		d.Location, d.Server, d.USN)

	return d
}

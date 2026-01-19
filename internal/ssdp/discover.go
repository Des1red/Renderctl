package ssdp

import (
	"fmt"
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

func pickInterfaceByIP(localIP string) (*net.Interface, error) {
	ip := net.ParseIP(localIP)
	if ip == nil {
		return nil, fmt.Errorf("invalid local IP: %q", localIP)
	}

	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagMulticast == 0 {
			continue
		}
		if strings.Contains(iface.Name, "lo") {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, a := range addrs {
			var addrIP net.IP
			switch v := a.(type) {
			case *net.IPNet:
				addrIP = v.IP
			case *net.IPAddr:
				addrIP = v.IP
			}
			if addrIP != nil && addrIP.Equal(ip) {
				return &iface, nil
			}
		}
	}

	return nil, fmt.Errorf("no multicast interface found for local IP %s", localIP)
}

func LooksLikeTV(d SSDPDevice) bool {
	usn := strings.ToLower(d.USN)
	srv := strings.ToLower(d.Server)
	loc := strings.ToLower(d.Location)

	// HARD NOs — routers & infra
	if strings.Contains(usn, "internetgatewaydevice") {
		return false
	}
	if strings.Contains(usn, "wan") || strings.Contains(usn, "igd") {
		return false
	}
	if strings.Contains(loc, "igd") || strings.Contains(loc, "wps") {
		return false
	}

	// STRONG YES — real render paths
	if strings.Contains(usn, "mediarenderer") {
		return true
	}
	if strings.Contains(usn, "avtransport") {
		return true
	}
	if strings.Contains(usn, "renderingcontrol") {
		return true
	}

	// Vendor renderers (DLNA stacks lie here)
	if strings.Contains(srv, "samsung") ||
		strings.Contains(srv, "lg") ||
		strings.Contains(srv, "sony") ||
		strings.Contains(srv, "philips") ||
		strings.Contains(srv, "panasonic") {
		return true
	}

	// Chromecast / Android TV style
	if strings.Contains(usn, "mdx") || strings.Contains(usn, "dial") {
		return true
	}

	// LAST-RESORT rootdevice (ONLY if it smells like a TV)
	if strings.Contains(usn, "upnp:rootdevice") {
		if strings.Contains(srv, "tv") ||
			strings.Contains(srv, "dlna") ||
			strings.Contains(srv, "mediarenderer") {
			return true
		}
	}

	return false
}

func ListenNotify(timeout time.Duration, ip string) ([]SSDPDevice, error) {
	logger.Notify("Listening for SSDP NOTIFY packets (%v)", timeout)

	addr, _ := net.ResolveUDPAddr("udp4", "239.255.255.250:1900")

	// PICK A REAL INTERFACE (eth0 / wlan0)
	iface, err := pickInterfaceByIP(ip)
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenMulticastUDP("udp4", iface, addr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	_ = conn.SetReadBuffer(65536)
	_ = conn.SetDeadline(time.Now().Add(timeout))

	var devices []SSDPDevice
	buf := make([]byte, 8192)

	for {
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			break
		}

		resp := string(buf[:n])

		if strings.Contains(resp, "NOTIFY") &&
			(strings.Contains(resp, "ssdp:alive") ||
				strings.Contains(resp, "ssdp:byebye")) {
			dev := parseSSDP(resp)
			if dev.Location != "" || dev.USN != "" {
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

	// ONE deadline for the whole discovery window
	_ = conn.SetDeadline(time.Now().Add(timeout))

	devices := make(map[string]SSDPDevice)
	buf := make([]byte, 2048)

	for _, st := range ssdpSearches {
		logger.Info("SSDP M-SEARCH for ST: %s", st)

		// Send once, TVs often ignore rapid spam
		_ = sendSearch(conn, st)

		// Give TVs time to respond (Samsung needs this)
		time.Sleep(500 * time.Millisecond)
	}

	for {
		n, _, err := conn.ReadFrom(buf)
		if err != nil {
			break // deadline hit
		}

		resp := string(buf[:n])
		dev := parseSSDP(resp)

		if dev.Location != "" || dev.USN != "" {
			logger.Success("SSDP response: %s", dev.Location)
			devices[dev.Location] = dev
		}
	}

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

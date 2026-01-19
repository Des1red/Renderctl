package identity

import (
	"fmt"
	"net"
	"strings"
	"time"
)

const ssdpAddr = "239.255.255.250:1900"

func AnnounceMediaServer(uuid, location string) {
	msg := strings.Join([]string{
		"NOTIFY * HTTP/1.1",
		"HOST: 239.255.255.250:1900",
		"CACHE-CONTROL: max-age=1800",
		"NT: urn:schemas-upnp-org:device:MediaServer:1",
		fmt.Sprintf("USN: uuid:%s::urn:schemas-upnp-org:device:MediaServer:1", uuid),
		"NTS: ssdp:alive",
		fmt.Sprintf("LOCATION: %s", location),
		"SERVER: renderctl/1.0 UPnP/1.0",
		"",
		"",
	}, "\r\n")

	conn, err := net.Dial("udp4", ssdpAddr)
	if err != nil {
		return
	}
	defer conn.Close()

	// send a few times (UPnP norm)
	for i := 0; i < 3; i++ {
		conn.Write([]byte(msg))
		time.Sleep(300 * time.Millisecond)
	}
}

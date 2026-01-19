package identity

import "fmt"

func DeviceXML(uuid string) string {
	return fmt.Sprintf(`<?xml version="1.0"?>
<root xmlns="urn:schemas-upnp-org:device-1-0">
  <specVersion>
    <major>1</major>
    <minor>0</minor>
  </specVersion>

  <device>
    <deviceType>urn:schemas-upnp-org:device:MediaServer:1</deviceType>
    <friendlyName>renderctl Media Server</friendlyName>
    <manufacturer>renderctl</manufacturer>
    <modelName>renderctl</modelName>
    <modelNumber>1</modelNumber>
    <UDN>uuid:%s</UDN>

    <serviceList>
      <service>
        <serviceType>urn:schemas-upnp-org:service:ContentDirectory:1</serviceType>
        <serviceId>urn:upnp-org:serviceId:ContentDirectory</serviceId>
        <controlURL>/cd/control</controlURL>
        <eventSubURL>/cd/event</eventSubURL>
        <SCPDURL>/cd/scpd.xml</SCPDURL>
      </service>

      <service>
        <serviceType>urn:schemas-upnp-org:service:ConnectionManager:1</serviceType>
        <serviceId>urn:upnp-org:serviceId:ConnectionManager</serviceId>
        <controlURL>/cm/control</controlURL>
        <eventSubURL>/cm/event</eventSubURL>
        <SCPDURL>/cm/scpd.xml</SCPDURL>
      </service>
    </serviceList>
  </device>
</root>`, uuid)
}

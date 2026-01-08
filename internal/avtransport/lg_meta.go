package avtransport

func lgMetadata(t Target) string {
	return `<?xml version="1.0" encoding="utf-8"?>
<DIDL-Lite 
 xmlns="urn:schemas-upnp-org:metadata-1-0/DIDL-Lite/"
 xmlns:dc="http://purl.org/dc/elements/1.1/"
 xmlns:upnp="urn:schemas-upnp-org:metadata-1-0/upnp/">

  <item id="0" parentID="0" restricted="1">
    <dc:title>Video</dc:title>
    <upnp:class>object.item.videoItem.movie</upnp:class>
    <res protocolInfo="http-get:*:video/mp4:*">` + t.MediaURL + `</res>
  </item>

</DIDL-Lite>`
}

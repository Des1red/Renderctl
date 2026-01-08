## TV Controller

Simple TV controller for sending media via the UPnP AVTransport protocol.

This tool discovers UPnP/DLNA-compatible TVs on the local network, probes for valid AVTransport endpoints, and sends media to the TV using SOAP requests (SetAVTransportURI + Play). It supports automatic discovery, probing fallback, vendor-specific metadata handling, and local media serving.

## Features

# SSDP discovery

Listens for NOTIFY packets

Active M-SEARCH fallback

AVTransport probing

Validates endpoints using SOAP (GetTransportInfo)

Works even when SSDP fails

# Media playback

Sends media via SetAVTransportURI

Starts playback with Play

Handles vendor quirks (e.g. Samsung STOP-before-play)

 Vendor-aware metadata

Automatic metadata selection per vendor (Samsung, LG, Sony, Philips, generic)

Local media serving

Serves files over HTTP for TV access

# Caching

Stores discovered AVTransport endpoints per IP

Reuses cached endpoints on future runs

# Probe-only mode

Resolve and validate endpoints without sending media

# Colorful logging

Clear distinction between progress, notices, results, and errors

How it works (high level)

# SSDP phase

Listen for NOTIFY packets

If none found, perform active M-SEARCH

# Probe phase

Directly probe the TV IP for AVTransport endpoints

Accepts any valid SOAP response (200 / 500)

Enrichment (best-effort)

Attempts identity and capability discovery

Non-fatal if unavailable

Playback

STOP (vendor quirk handling)

SetAVTransportURI

Play

## Usage
Basic playback (auto mode)
go run main.go -Lf media.mp4 -Lip 192.168.1.110


Discovers the TV automatically

Serves the media locally

Sends it to the TV via AVTransport

Probe only (no playback)
go run main.go --probe-only -Tip 192.168.1.10


Probes the TV for a valid AVTransport endpoint

Does not send any media

Useful for debugging and testing

# Manual mode
go run main.go -mode manual -Tip 192.168.1.10 -Tport 9197 -Tpath /dmr/upnp/control/AVTransport1 -Lf media.mp4 -Lip 192.168.1.110


Skips discovery

Uses explicit control URL information

Command-line options
--probe-only        Probe AVTransport endpoint only (no playback)
-mode               Execution mode (auto | manual)

-Tip                TV IP address
-Tport              TV SOAP port
-Tpath              TV SOAP control path
-type               TV vendor (samsung, lg, sony, philips, generic)

-Lf                 Local media file
-Lip                Local IP for media serving

--list-cache        List cached AVTransport devices
--forget-cache      Interactive cache removal
--forget-cache <IP> Remove specific cached device
--forget-cache all  Clear cache

-h                  Show help

Supported vendors

Samsung

LG

Sony

Philips

Generic UPnP/DLNA renderers

Vendor handling mainly affects metadata generation and playback quirks.

## Notes & limitations

Identity enrichment is best-effort

Some TVs do not expose device descriptors outside SSDP

Not all TVs fully comply with UPnP/DLNA specs

Designed for local networks only

No authentication support (standard UPnP behavior)

## Project goal

#This project focuses on:

Correct protocol behavior

Minimal assumptions

Transparent debugging

Practical interoperability with real TVs

It is intentionally simple, explicit, and inspectable.

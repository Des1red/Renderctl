# TV Controller (tvctrl) — v2

Simple TV controller for sending media via the UPnP AVTransport protocol.

`tvctrl` discovers UPnP/DLNA-compatible TVs on the local network, resolves valid AVTransport endpoints, and sends media to the TV using SOAP requests (`SetAVTransportURI` + `Play`).  
It supports automatic discovery, probing fallback, **explicit cache selection**, vendor-aware metadata handling, and clean local media serving, and direct streaming mode.

---

## What’s new in v2

- Deterministic cache selection (`--select-cache`)
- Multiple cached targets (indexed & sorted)
- Clean HTTP server lifecycle (graceful shutdown)
- Explicit execution routing (no hidden mode hacks)
- Capability enrichment (actions + media protocols)
- Expanded AVTransport probing paths
- Built-in shell autocomplete (optional)
- Improved cache safety & control flow
- Dedicated STREAM mode (live + resolved streaming)

---

## Features

### SSDP discovery
- Listens for NOTIFY packets
- Active M-SEARCH fallback
- Extracts service descriptors when available

### AVTransport probing
- Probes common and vendor-specific AVTransport endpoints
- Validates endpoints using SOAP (`GetTransportInfo`, `Stop`, etc.)
- Works even when SSDP fails
- Optional deep-search mode for noisy networks

### Media playback
- Sends media via `SetAVTransportURI`
- Starts playback with `Play`
- Handles vendor quirks (e.g. Samsung STOP-before-play)

### Vendor-aware metadata
- Automatic metadata selection per vendor:
  - Samsung
  - LG
  - Sony
  - Philips
  - Generic
- Best-effort identity enrichment (non-fatal)

### Local media serving
- Serves files over HTTP for TV access
- Clean startup & shutdown using channels
- Server runs only when needed

### Caching
- Stores discovered AVTransport endpoints per IP
- Supports multiple cached devices
- Indexed, sorted cache entries
- Explicit cache selection (`--select-cache`)
- Safe reuse without re-probing

### Explicit cache selection (v2)
- Select a cached TV deterministically by index
- Skips SSDP, probing, and interactive prompts
- Runs playback immediately using cached ControlURL

### Probe-only mode
- Resolve and validate endpoints without sending media
- Useful for debugging and testing

### Colorful logging
- Clear distinction between:
  - Progress
  - Notices
  - Results
  - Errors

---

###Streaming mode (v2)

- Streaming mode is a dedicated execution path designed for reliable DLNA playback without screen mirroring.

- It streams raw media bytes to the TV via a local HTTP endpoint and instructs playback using AVTransport.

- Supported stream types

- tvctrl automatically classifies the -Lf input into one of three stream types:

1. Local file stream

- Example: movie.ts, video.mp4

- File is served locally over HTTP

- TV pulls the media directly

2. External media URL

- Example: https://example.com/video.mp4

- TV pulls the remote media through the local stream proxy

- No transcoding

3. Resolved stream (platforms)

- Example: YouTube URLs

- The URL does not point to raw media
 
- yt-dlp is used to extract audio/video
 
- Media is remuxed to MPEG-TS and streamed live
 
- The TV only ever sees a local /stream URL
 
- How streaming works
 
- Resolve AVTransport ControlURL (cache / SSDP / probe)
 
- Start internal HTTP stream server
 
- Expose /stream endpoint
 
- Send MediaURL to TV via SOAP:
 
- Stop
 
- SetAVTransportURI
 
- Play
 
- TV connects back and pulls the stream
 
- The stream server:
 
- Uses chunked transfer (no Content-Length)
 
- Dynamically observes TV behavior (HEAD / Range)
 
- Disables seeking for live / TS streams
 
- Keeps behavior predictable across vendors
 
- What streaming mode does NOT do
 
- No screen mirroring
 
- No HTML rendering
 
- No JavaScript execution
 
- No browser emulation
 
- No transcoding unless required for compatibility
 
- Streaming mode exists purely to deliver media bytes, not web content.
## How it works (high level)

### 1. SSDP phase (auto mode)
- Listen for NOTIFY packets
- Fallback to active M-SEARCH if needed

### 2. Cache resolution
- If `--select-cache` is used → **direct execution**
- Otherwise:
  - Try cached endpoint for known IP (interactive)
  - Skip if cache disabled

### 3. Probe phase
- Directly probe the TV IP for AVTransport endpoints
- Accepts valid SOAP responses (200 / 500)

### 4. Enrichment (best-effort)
- Identity discovery
- Capability discovery (actions + media protocols)
- Non-fatal if unavailable

### 5. Playback
- STOP (vendor quirk handling)
- `SetAVTransportURI`
- `Play`

---

## Usage

### Basic playback (auto mode)

- tvctrl -Lf media.mp4 -Lip 192.168.1.110

    Discovers the TV automatically

    Serves the media locally

    Sends it to the TV via AVTransport

### Explicit cache selection (v2)

- tvctrl --list-cache
- tvctrl --select-cache 0 -Lf media.mp4

    Uses the selected cached TV

    Skips discovery and probing

    No prompts, no guessing

- Probe only (no playback)

tvctrl --probe-only -Tip 192.168.1.10

    Probes the TV for a valid AVTransport endpoint

    Does not send any media

### Manual mode

- tvctrl -mode manual \
  -Tip 192.168.1.10 \
  -Tport 9197 \
  -Tpath /dmr/upnp/control/AVTransport1 \
  -Lf media.mp4 \
  -Lip 192.168.1.110

    Skips discovery

    Uses explicit control URL information

### Streaming mode
 --mode stream -Lf media.ts -Lip 192.168.1.110

 --mode stream -Lf https://www.youtube.com/watch?v=XXXX -Lip 192.168.1.110
     Starts a live stream server
     
     Sends /stream to the TV
     
     Handles platform resolution automatically
### Command-line options
## Execution

    --probe-only Probe AVTransport endpoint only

    --mode Execution mode (auto | manual | scan)

## Cache

    --auto-cache Skip cache save confirmation

    --no-cache Disable cache usage

    --list-cache List cached AVTransport devices (indexed)

    --select-cache <n> Select cached device by index

    --forget-cache Interactive cache removal

    --forget-cache <IP> Remove specific cached device

    --forget-cache all Clear cache

## Scan

    --deep-search Use expanded probing paths (slower, noisier)

    --subnet Scan subnet (e.g. 192.168.1.0/24)

    --ssdp Enable SSDP discovery

## TV

    --Tip TV IP address

    --Tport TV SOAP port

    --Tpath TV SOAP control path

    --type TV vendor

## Media

    --Lf Local media file or url(explicit to stream mode)

    --Lip Local IP for serving media

    --Ldir Local directory to serve

    --LPort Local HTTP port

# Shell autocomplete (optional)

One-time setup:

tvctrl install-completion
exec $SHELL

Enables tab completion for flags and commands.
### Supported vendors

    Samsung

    LG

    Sony

    Philips

    Generic UPnP / DLNA renderers

- Vendor handling mainly affects metadata generation and playback quirks.
### Notes & limitations

-    Identity enrichment is best-effort

-    Some TVs do not expose descriptors outside SSDP

-    Not all TVs fully comply with UPnP/DLNA specs

-    Designed for local networks only

-    No authentication support (standard UPnP behavior)

### Project goal

- This project focuses on:

    Correct protocol behavior

    Minimal assumptions

    Transparent debugging

    Practical interoperability with real TVs

It is intentionally simple, explicit, and inspectable.


### Attribution
This project was originally created by Des1red.
If you use or modify this software, attribution is appreciated.

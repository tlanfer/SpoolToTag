# SpoolToTag

Photograph a 3D printer filament spool label, extract the filament info with AI, and write an [OpenSpool](https://github.com/spuder/OpenSpool) NFC tag — all from your phone's browser.

## Features

- **AI-powered label reading** — snap a photo, get filament type, brand, color, and temperatures extracted automatically
- **NFC tag writing** — write OpenSpool-compatible NDEF tags using the Web NFC API (Chrome on Android)
- **Single binary** — Go server with embedded static files, no external dependencies
- **Mobile-first UI** — installable as a PWA, works great on phone home screens
- **Client-side image resizing** — photos are downsized before upload for faster analysis

## Requirements

- Go 1.22+ (to build)
- An OpenAI API key
- NTAG 215 NFC tags (NTAG 213 is too small)
- Chrome on Android (for Web NFC support)
- **HTTPS** — the Web NFC API requires a [secure origin](https://w3c.github.io/webappsec-secure-contexts/)

## Quick Start

```bash
go build -o spooltotag .
OPENAI_API_KEY=sk-... ./spooltotag
```

Or with Docker:

```bash
OPENAI_API_KEY=sk-... docker compose up --build
```

### HTTPS

The Web NFC API only works over HTTPS (secure origins). The easiest way to add HTTPS is with a reverse proxy like [caddy-docker-proxy](https://github.com/lucaslorentz/caddy-docker-proxy). For local development, `localhost` is treated as a secure origin by browsers.

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `OPENAI_API_KEY` | *(required)* | OpenAI API key |
| `LISTEN_ADDR` | `:8080` | Server listen address |
| `OPENAI_MODEL` | `gpt-5-nano` | OpenAI model for vision analysis |

## How It Works

1. Take a photo of the spool label
2. The image is resized and sent to the OpenAI vision API
3. AI extracts filament type, brand, color, and temperature range
4. Review and edit the extracted info
5. Hold an NFC tag to your phone to write the OpenSpool data

## License

Apache License 2.0 — see [LICENSE](LICENSE) for details.

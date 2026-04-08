# Spoolman Bambu Bridge

A bridge service that synchronizes filament and spool data between **Bambu Lab** 3D printers and **[Spoolman](https://github.com/Donkie/Spoolman)**.

It connects to your printers over MQTT, tracks filament usage in real time, and keeps Spoolman up to date automatically.

## Features

- **Multi-printer support** — monitor multiple Bambu Lab printers simultaneously
- **Real-time sync** — MQTT connection to each printer with automatic weight updates to Spoolman
- **Smart spool matching** — finds existing spools by NFC tag, filament type, or creates new ones automatically
- **Live dashboard** — responsive web UI with WebSocket updates showing all printers, AMS slots, and remaining filament
- **Resilient** — tolerates Spoolman or printer disconnections and reconnects automatically
- **Self-contained** — single binary with embedded frontend (Bulma CSS, no CDN required)

## Quick Start

### Prerequisites

- Go 1.24+ (for building from source)
- A running [Spoolman](https://github.com/Donkie/Spoolman) instance
- One or more Bambu Lab printers with LAN mode enabled

### Build & Run

```bash
# Clone the repository
git clone https://github.com/spoolman-bambu-bridge/spoolman-bambu-bridge.git
cd spoolman-bambu-bridge

# Copy and edit the config
cp config.example.yaml config.yaml
# Edit config.yaml with your Spoolman address, printer IPs, serials, and access codes

# Run
go run ./cmd/bridge
```

The dashboard is available at **http://localhost:8080**.

### Docker

```bash
# Build the image
docker build -t spoolman-bambu-bridge .

# Run with your config
docker run -d \
  --name spoolman-bambu-bridge \
  -p 8080:8080 \
  -v $(pwd)/config.yaml:/app/config.yaml \
  spoolman-bambu-bridge
```

## Configuration

Create a `config.yaml` based on the example:

```yaml
spoolman:
  address: "http://192.168.1.100:8000"

printers:
  - name: "X1C"
    ip: "192.168.1.50"
    serial: "01P00A00000001"
    access_code: "abcdefgh"
```

| Field | Description |
|---|---|
| `spoolman.address` | URL of your Spoolman instance |
| `printers[].name` | Display name (defaults to serial if omitted) |
| `printers[].ip` | Printer's LAN IP address |
| `printers[].serial` | Printer serial number (found in printer settings) |
| `printers[].access_code` | LAN access code (found in printer settings under Network) |

### Finding your printer's credentials

1. On your Bambu Lab printer, go to **Settings > Network**
2. **IP Address** — shown on the network screen
3. **Serial Number** — shown in **Settings > Device** or on the printer's label
4. **Access Code** — shown in **Settings > Network > LAN Mode**

Make sure **LAN Mode** is enabled on the printer.

## How It Works

### Initialization

1. Downloads the [Bambu-Spoolman filament database](https://github.com/piitaya/bambu-spoolman-db) (MFDB)
2. Ensures custom fields exist in Spoolman (`bambu_lab_id`, `bambu_lab_code` on filaments; `tag` on spools)
3. Syncs existing Spoolman filaments with MFDB data

### Spool Matching

When a filament change or weight update is detected, the bridge finds the corresponding spool in Spoolman using this priority:

1. **By tag** — matches the spool's `tag` field against the tray's NFC UUID (`tray_uuid`)
2. **By filament type** — finds a non-archived, untagged spool whose filament matches the Bambu filament ID
3. **Create new** — looks up the filament in MFDB, fetches it from Spoolman's external database if needed, creates the filament and spool

### Weight Updates

- The `remain` percentage from MQTT is converted to grams using `tray_weight`
- Updates are **debounced** (5 seconds) to avoid hammering Spoolman during rapid changes
- `last_used` is updated on every weight change; `first_used` is set once on first use

## Project Structure

```
cmd/bridge/          Main entrypoint
internal/
  bambu/             MQTT client for Bambu Lab printers
  bridge/            Core orchestration, spool matching, state management
  config/            YAML configuration loading
  filamentdb/        Bambu-Spoolman filament database (MFDB) fetcher
  spoolman/          Spoolman REST API client
backend/             HTTP server, WebSocket hub
frontend/
  templates/         Go HTML templates (dashboard)
  static/            Embedded assets (Bulma CSS)
```

## API

### `GET /api/v1/health`

Health check. Returns `{"status": "ok"}`.

### `GET /ws`

WebSocket endpoint. Streams the full bridge state as JSON on every change. Send any message to request the current state.

## Development

```bash
# Run tests
go test ./...

# Run with verbose logging
go run ./cmd/bridge

# Build binary
go build -o bridge ./cmd/bridge
```

## Acknowledgments

- [Spoolman](https://github.com/Donkie/Spoolman) — filament management system
- [Bambu-Spoolman DB](https://github.com/piitaya/bambu-spoolman-db) — Bambu Lab filament database for Spoolman

# 1. Context & Role
You are a Senior Go Developer and System Architect. Your goal is to build a bridge service that synchronizes filament data between Bambu Lab 3D Printers and Spoolman.

Reference Materials:
- Spoolman API: https://donkie.github.io/Spoolman/
- Matching filament Database (later named MFDB): https://github.com/piitaya/bambu-spoolman-db/blob/main/filaments.json

# 2. Technical Stack

- Backend: Go (Golang) 1.21+
- Frontend: Go HTML Templates + Bulma CSS (no CDN, the project should not require internet once started)
- Communication: MQTT over TLS (v3.1.1)
- Config: YAML
- Delivery: a docker container

# 3. Core Requirements & Logic

## A. Initialization Sequence
- MFDB Fetch:
  - Download the filaments.json from the reference link.

- Spoolman Setup:
  - Ensure custom fields bambu_lab_id (text) and bambu_lab_code (integer) exist in Spoolman via GET and POST /api/v1/field, at filament level, and also tag (text) at spool level
  - Download all the filament already setup in Spoolman using GET /api/v1/filament, filter the data to keep only the filament with an id in MFDB (field `spoolman_id`) 
  - Iterate over all filament entries kept, and ensure there's a filament registred in Spoolman, with the correct custom fields values
- String values for custom fields are json. So strings are escaped, for example "\"123456789AZ"\".
- Multi-Printer Management: Initialize a worker pool or goroutine for each printer defined in config.yaml.

## B. MQTT Integration

- Connection: connect to printer IP on port 8883.
- Security: use TLS with InsecureSkipVerify: true (printers use self-signed certs).
- Auth: username is bblp, password is the printer's access_code.
- Payload: you can find a sample payload in internal/bambu/sample_mqtt_payload.json.

## C. The Bridge Logic

- When a tag or weight changes, calculate the weight delta.
- Update the corresponding spool in Spoolman. Spoolman spool matching logic, in this order:
  - search for a spool in Spoolman with a matching `tag` (`tray_uuid` in the payload)
  - search in Spoolman for a not archived spool, without `tag` and with filament matching `bambu_lab_id` (`tray_id_name` in the payload).
  - create the spool:
    - search in MFDB for the entry with `id` matching with `tray_id_name`, take its `spoolman_id`
    - with the `spoolman_id`, ensure the filament exists in Spoolman, using GET /api/v1/filament. If it does not exists, use GET api/v1/external/filament to get all Spoolman external filament. There's a lot, but `spoolman_id` is the `id` field. Using the external filament, create the filament in Spoolman, and populate `bambu_lab_id` and `bambu_lab_code` fields from MFDB
    - now, create a spool in Spoolman with the filament we found/created, and add the corresponding `tag`
- Maintain a local state mapping of Printer -> Slot -> Spoolman_ID.

# 4. Implementation Plan

## Step 1: Configuration & Types

Define the Config struct (YAML) and the JSON structures for both the Filament DB and the Bambu MQTT messages.

## Step 2: Spoolman Client

Implement a Go client for Spoolman to:
- Create Custom Fields.
- Search for/Create Filaments.
- Search Filament by `bambu_lab_code` or `bambu_lab_id` custom field.
- Search for/Create/Update Spool.
- Search Spool by `tag` custom field.

Make sure the app calls the Spoolman API at startup to ensure the custom fields are present.

The app should be able to start even if the connection with Spoolman can't be established. The bridge should be resilient to Spoolman connection loss.

## Step 3: MQTT Worker

Implement the BambuClient that handles the persistent connection, re-connection logic, and message parsing.

Make sure MQTT workes starts with the app.

The app should be able to start even if the connection is lost/not established. The bridge should be resilient to MQTT connection loss.

## Step 4: Web Server & UI

Create a simple net/http server, the backend should also serve the frontend.

Use html/template to render a dashboard using Bulma CSS, as index.
The dashboard should be mobile/tablet friendly.

The dashboard must show each printer, its slots, and the current weight of the loaded spools.
Communication with the backend should be made using websocket a connection.

It should also indicates if it's connected to Spoolman.

## Step 5: Update Spool weight

When a weight change is received, update the spool (matching logic described in C. The Bridge Logic. Respect this matching logic and its order).
There are a lot of events, so make sure there's a debounce logic here also. In the event, `remain` is a percentage. Make sure you update the value only when there's a change.
When there's a weight update, also update the filament's `last_used` field.
If there's no `first_used` timestamp, add it.

## Step 6: Add a tag-scanned endpoint to the backend

When the /api/{version}/tag-scanned HTTP endpoint is called, it should give back the spool found in Spoolman, search by `tray_uuid` (`tag` custom field).

Input payload:
```json
{
  "tray_uuid": "<uuid>",
  "bambu_lab_code": "<code>",
  "create_if_missing": true|false,
  "remaining_weight": 0...1000
}
```
Case when the spool is not found:
- If `create_if_missing` is set to `false`, it should reply with `404`.
- If `create_if_missing` is set to `true`, the filament will be searched by `bambu_lab_code` custom field, and the corresponding spool created with the `tray_uuid` as `tag` custom field and remaining weight must be set. This field is optional, if not given, it's 1000. No `first_used` and `last_used` must be set here.

## Step 7: auto-archive spools when no remaining filament is left

In the config, add an optional `auto_archive` field in `spoolman` section, default set to true.
If set to true, when a spool is updated, if the remaining weight is 0, archive the spool in Spoolman.

## Step 8: location management

In the config, add an optional `printer_location` field in `spoolman` section, default set to false. If set to true:
- add a custom field at spool level, named `previous location`
- when the bridge init, add locations in Spoolman, corresponding to printers/AMS location. You will need to wait for MQTT events to populate this.
- when the bridge receive an MQTT event, spools location should be updated: save the previous location in `previous location` custom field, and update its new location.

# 5. Expected config.yaml Structure

```yaml
spoolman:
  address: "http://192.168.x.x:8000"
printers:
  - name: "X1C"
    ip: "192.168.1.50"
    serial: "01P..."
    access_code: "abcdefgh"
```

# 6. Development Instructions for Claude

Create unit tests for backend for critical features like:
- YAML parsing logic.
- Filament ID matching logic
- MQTT payload unmarshaling

The backend will also serve the frontend.

The API path will be prefixed with /api/{version} path. {version} is current v1. 

Keep the project structure flat and clean (standard Go project layout).

Handle errors gracefully; if one printer fails to connect, do not stop the service for others.

Ensure the UI is clean and responsive.

ALWAYS run the app using `go run ./cmd/bridge`
NEVER use pkill, killall, or any command that kills all go or bridge processes.
ALWAYS use port-specific termination and command name: `lsof -ti :8080 -a -c bridge | xargs kill -9`
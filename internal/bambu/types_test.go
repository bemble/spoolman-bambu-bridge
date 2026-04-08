package bambu

import (
	"encoding/json"
	"os"
	"testing"
)

func TestMessage_Unmarshal(t *testing.T) {
	payload := `{
		"print": {
			"command": "push_status",
			"gcode_state": "RUNNING",
			"ams": {
				"ams": [
					{
						"id": "0",
						"humidity": "2",
						"temp": "29.9",
						"tray": [
							{
								"id": "0",
								"tag_uid": "43BF663900000100",
								"tray_uuid": "7988E476536B437FA117F12A2D3D3DD3",
								"tray_id_name": "A01-W2",
								"tray_type": "PLA",
								"tray_sub_brands": "PLA Matte",
								"tray_color": "FFFFFFFF",
								"tray_weight": "1000",
								"remain": 74,
								"tray_info_idx": "GFA01"
							},
							{
								"id": "1",
								"tag_uid": "33FE980A00000100",
								"tray_uuid": "1412E6D286B3484B93F2A21CEFC3818D",
								"tray_id_name": "A00-B5",
								"tray_type": "PLA",
								"tray_sub_brands": "PLA Basic",
								"tray_color": "00B1B7FF",
								"tray_weight": "1000",
								"remain": 65,
								"tray_info_idx": "GFA00"
							}
						]
					}
				],
				"tray_now": "255",
				"tray_pre": "255"
			},
			"vt_tray": {
				"id": "255",
				"tray_type": "",
				"tray_uuid": "00000000000000000000000000000000",
				"remain": 0
			}
		}
	}`

	var msg Message
	if err := json.Unmarshal([]byte(payload), &msg); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if msg.Print == nil {
		t.Fatal("print is nil")
	}
	if msg.Print.GcodeState != "RUNNING" {
		t.Errorf("gcode_state = %q, want RUNNING", msg.Print.GcodeState)
	}
	if msg.Print.AMS == nil {
		t.Fatal("ams is nil")
	}

	units := msg.Print.AMS.AMS
	if len(units) != 1 {
		t.Fatalf("ams units = %d, want 1", len(units))
	}
	if units[0].Humidity != "2" {
		t.Errorf("humidity = %q, want 2", units[0].Humidity)
	}

	trays := units[0].Trays
	if len(trays) != 2 {
		t.Fatalf("trays = %d, want 2", len(trays))
	}

	tray0 := trays[0]
	if tray0.TrayType != "PLA" || tray0.Remain != 74 {
		t.Errorf("tray[0] type=%q remain=%d", tray0.TrayType, tray0.Remain)
	}
	if tray0.TrayUUID != "7988E476536B437FA117F12A2D3D3DD3" {
		t.Errorf("tray_uuid = %q", tray0.TrayUUID)
	}
	if tray0.TrayIDName != "A01-W2" {
		t.Errorf("tray_id_name = %q, want A01-W2", tray0.TrayIDName)
	}
	if tray0.TagUID != "43BF663900000100" {
		t.Errorf("tag_uid = %q", tray0.TagUID)
	}

	tray1 := trays[1]
	if tray1.TrayIDName != "A00-B5" || tray1.Remain != 65 {
		t.Errorf("tray[1] id_name=%q remain=%d", tray1.TrayIDName, tray1.Remain)
	}

	if msg.Print.AMS.TrayNow != "255" {
		t.Errorf("tray_now = %q, want 255", msg.Print.AMS.TrayNow)
	}

	vt := msg.Print.VTTray
	if vt == nil {
		t.Fatal("vt_tray is nil")
	}
	if vt.Remain != 0 {
		t.Errorf("vt_tray remain = %d, want 0", vt.Remain)
	}
}

func TestMessage_SamplePayload(t *testing.T) {
	data, err := os.ReadFile("sample_mqtt_payload.json")
	if err != nil {
		t.Fatalf("reading sample payload: %v", err)
	}

	// The sample is an array with one element containing message and topic
	var wrapper []struct {
		Message Message `json:"message"`
		Topic   string  `json:"topic"`
	}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if len(wrapper) != 1 {
		t.Fatalf("expected 1 message, got %d", len(wrapper))
	}

	msg := wrapper[0].Message
	if msg.Print == nil {
		t.Fatal("print is nil")
	}
	if msg.Print.Command != "push_status" {
		t.Errorf("command = %q, want push_status", msg.Print.Command)
	}
	if msg.Print.GcodeState != "FINISH" {
		t.Errorf("gcode_state = %q, want FINISH", msg.Print.GcodeState)
	}

	// Verify AMS units
	ams := msg.Print.AMS
	if ams == nil {
		t.Fatal("ams is nil")
	}
	if len(ams.AMS) != 2 {
		t.Fatalf("ams units = %d, want 2", len(ams.AMS))
	}

	// First AMS has 4 trays
	if len(ams.AMS[0].Trays) != 4 {
		t.Errorf("ams[0] trays = %d, want 4", len(ams.AMS[0].Trays))
	}
	// Second AMS has 1 tray
	if len(ams.AMS[1].Trays) != 1 {
		t.Errorf("ams[1] trays = %d, want 1", len(ams.AMS[1].Trays))
	}

	// Check a specific tray
	tray := ams.AMS[0].Trays[0]
	if tray.TrayIDName != "A01-W2" {
		t.Errorf("tray_id_name = %q, want A01-W2", tray.TrayIDName)
	}
	if tray.TrayUUID != "7988E476536B437FA117F12A2D3D3DD3" {
		t.Errorf("tray_uuid = %q", tray.TrayUUID)
	}
	if tray.Remain != 74 {
		t.Errorf("remain = %d, want 74", tray.Remain)
	}
	if tray.TrayWeight != "1000" {
		t.Errorf("tray_weight = %q, want 1000", tray.TrayWeight)
	}
}

func TestMessage_EmptyPayload(t *testing.T) {
	var msg Message
	if err := json.Unmarshal([]byte(`{}`), &msg); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if msg.Print != nil {
		t.Error("expected print to be nil for empty payload")
	}
}

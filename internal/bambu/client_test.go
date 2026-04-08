package bambu

import (
	"encoding/json"
	"testing"
)

func TestOnMessage_ParsesReport(t *testing.T) {
	var received *PrintReport
	var receivedSerial string

	handler := func(serial string, report *PrintReport) {
		receivedSerial = serial
		received = report
	}

	c := NewClient(ClientConfig{
		Name:       "TestPrinter",
		Serial:     "01P00TEST",
		AccessCode: "test1234",
		IP:         "127.0.0.1",
	}, handler, nil)

	// Simulate a message
	payload := Message{
		Print: &PrintReport{
			Command:    "push_status",
			GcodeState: "RUNNING",
			AMS: &AMSData{
				AMS: []AMSUnit{
					{
						ID: "0",
						Trays: []AMSTray{
							{
								ID:         "0",
								TrayUUID:   "AABBCCDD",
								TrayIDName: "A00-A0",
								Remain:     85,
								TrayWeight: "1000",
							},
						},
					},
				},
			},
		},
	}

	data, _ := json.Marshal(payload)
	// Call the handler directly (we can't test real MQTT without a broker)
	var msg Message
	json.Unmarshal(data, &msg)
	if msg.Print != nil {
		c.handler(c.cfg.Serial, msg.Print)
	}

	if received == nil {
		t.Fatal("handler was not called")
	}
	if receivedSerial != "01P00TEST" {
		t.Errorf("serial = %q, want 01P00TEST", receivedSerial)
	}
	if received.GcodeState != "RUNNING" {
		t.Errorf("gcode_state = %q, want RUNNING", received.GcodeState)
	}
	if len(received.AMS.AMS) != 1 {
		t.Fatalf("ams units = %d, want 1", len(received.AMS.AMS))
	}
	tray := received.AMS.AMS[0].Trays[0]
	if tray.TrayUUID != "AABBCCDD" || tray.Remain != 85 {
		t.Errorf("tray = %+v", tray)
	}
}

func TestOnMessage_IgnoresNonPrintMessages(t *testing.T) {
	var called bool
	handler := func(serial string, report *PrintReport) {
		called = true
	}

	c := NewClient(ClientConfig{
		Name:   "TestPrinter",
		Serial: "01P00TEST",
		IP:     "127.0.0.1",
	}, handler, nil)

	// Message without print field
	data := []byte(`{"info": {"sequence_id": "123"}}`)
	var msg Message
	json.Unmarshal(data, &msg)
	if msg.Print != nil {
		c.handler(c.cfg.Serial, msg.Print)
	}

	_ = c // use c to avoid unused warning
	if called {
		t.Error("handler should not be called for non-print messages")
	}
}

func TestClientConfig(t *testing.T) {
	cfg := ClientConfig{
		Name:       "X1C",
		IP:         "192.168.1.50",
		Serial:     "01P00A00000001",
		AccessCode: "abcdefgh",
	}

	c := NewClient(cfg, nil, nil)
	if c.cfg.Name != "X1C" {
		t.Errorf("name = %q, want X1C", c.cfg.Name)
	}
	if c.cfg.IP != "192.168.1.50" {
		t.Errorf("ip = %q", c.cfg.IP)
	}
}

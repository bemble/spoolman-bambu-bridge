package filamentdb

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEntry_Unmarshal(t *testing.T) {
	payload := `{
		"id": "A00-A0",
		"code": "10300",
		"material": "PLA Basic",
		"color_name": "Orange",
		"color_hex": "FF6A13",
		"spoolman_id": "bambulab_pla_orange_1000_175_n"
	}`

	var entry Entry
	if err := json.Unmarshal([]byte(payload), &entry); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if entry.ID != "A00-A0" {
		t.Errorf("id = %q, want A00-A0", entry.ID)
	}
	if entry.Code != "10300" {
		t.Errorf("code = %q, want 10300", entry.Code)
	}
	if entry.Material != "PLA Basic" {
		t.Errorf("material = %q, want PLA Basic", entry.Material)
	}
	if entry.ColorHex != "FF6A13" {
		t.Errorf("color_hex = %q, want FF6A13", entry.ColorHex)
	}
	if entry.SpoolmanID == nil || *entry.SpoolmanID != "bambulab_pla_orange_1000_175_n" {
		t.Errorf("spoolman_id = %v, want bambulab_pla_orange_1000_175_n", entry.SpoolmanID)
	}
}

func TestEntry_Unmarshal_NullSpoolmanID(t *testing.T) {
	payload := `{
		"id": "A09-A0",
		"code": "12002",
		"material": "PLA Tough",
		"color_name": "Orange",
		"spoolman_id": null
	}`

	var entry Entry
	if err := json.Unmarshal([]byte(payload), &entry); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if entry.SpoolmanID != nil {
		t.Errorf("spoolman_id should be nil, got %v", *entry.SpoolmanID)
	}
	if entry.ColorHex != "" {
		t.Errorf("color_hex should be empty, got %q", entry.ColorHex)
	}
}

func TestParse(t *testing.T) {
	data := `[
		{"id": "A00-A0", "code": "10300", "material": "PLA Basic", "color_name": "Orange", "color_hex": "FF6A13", "spoolman_id": null},
		{"id": "A00-B0", "code": "10301", "material": "PLA Basic", "color_name": "Red", "color_hex": "FF0000", "spoolman_id": null},
		{"id": "A09-A0", "code": "12002", "material": "PLA Tough", "color_name": "Orange", "spoolman_id": null}
	]`

	db, err := Parse(strings.NewReader(data))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if len(db.Entries) != 3 {
		t.Fatalf("entries = %d, want 3", len(db.Entries))
	}

	// Lookup by ID
	entry, ok := db.ByID["A00-A0"]
	if !ok {
		t.Fatal("expected to find A00-A0 in ByID")
	}
	if entry.Material != "PLA Basic" {
		t.Errorf("material = %q, want PLA Basic", entry.Material)
	}

	// Lookup by Code
	entry, ok = db.ByCode["12002"]
	if !ok {
		t.Fatal("expected to find 12002 in ByCode")
	}
	if entry.ID != "A09-A0" {
		t.Errorf("id = %q, want A09-A0", entry.ID)
	}

	// Missing lookup
	_, ok = db.ByID["ZZZZZ"]
	if ok {
		t.Error("expected no match for unknown ID")
	}
}

func TestFetch(t *testing.T) {
	data := `[
		{"id": "A00-A0", "code": "10300", "material": "PLA Basic", "color_name": "Orange", "spoolman_id": null}
	]`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(data))
	}))
	defer ts.Close()

	db, err := Fetch(ts.URL, ts.Client())
	if err != nil {
		t.Fatalf("fetch error: %v", err)
	}

	if len(db.Entries) != 1 {
		t.Fatalf("entries = %d, want 1", len(db.Entries))
	}
	if db.Entries[0].ID != "A00-A0" {
		t.Errorf("id = %q, want A00-A0", db.Entries[0].ID)
	}
}

func TestFetch_HTTPError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	_, err := Fetch(ts.URL, ts.Client())
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

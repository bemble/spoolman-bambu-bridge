package spoolman

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// newTestServer creates an httptest server and a Client pointing at it.
func newTestServer(handler http.HandlerFunc) (*httptest.Server, *Client) {
	ts := httptest.NewServer(handler)
	client := NewClient(ts.URL, ts.Client())
	return ts, client
}

func TestGetFields(t *testing.T) {
	ts, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/api/v1/field/filament" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		json.NewEncoder(w).Encode([]Field{
			{Key: "bambu_lab_code", Name: "Bambu Lab Code", FieldType: "text"},
		})
	})
	defer ts.Close()

	fields, err := client.GetFields("filament")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fields) != 1 || fields[0].Key != "bambu_lab_code" {
		t.Errorf("fields = %+v", fields)
	}
}

func TestCreateField(t *testing.T) {
	ts, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/api/v1/field/filament/bambu_lab_id" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var params FieldParams
		json.NewDecoder(r.Body).Decode(&params)
		if params.Name != "Bambu Lab ID" || params.FieldType != "text" {
			t.Errorf("params = %+v", params)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]Field{
			{Key: "bambu_lab_id", Name: "Bambu Lab ID", FieldType: "text"},
		})
	})
	defer ts.Close()

	fields, err := client.CreateField("filament", "bambu_lab_id", FieldParams{
		Name:      "Bambu Lab ID",
		FieldType: "text",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fields) != 1 || fields[0].Key != "bambu_lab_id" {
		t.Errorf("fields = %+v", fields)
	}
}

func TestEnsureFields_CreatesOnlyMissing(t *testing.T) {
	var createdKeys []string
	ts, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "GET" && r.URL.Path == "/api/v1/field/filament":
			// bambu_lab_code already exists, bambu_lab_id is missing
			json.NewEncoder(w).Encode([]Field{
				{Key: "bambu_lab_code", Name: "Bambu Lab Code", FieldType: "integer"},
			})
		case r.Method == "GET" && r.URL.Path == "/api/v1/field/spool":
			// tag is missing
			json.NewEncoder(w).Encode([]Field{})
		case r.Method == "POST":
			// Extract key from path: /api/v1/field/{entity}/{key}
			createdKeys = append(createdKeys, r.URL.Path)
			json.NewEncoder(w).Encode([]Field{})
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	})
	defer ts.Close()

	if err := client.EnsureFields(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should create bambu_lab_id (filament) and tag (spool)
	if len(createdKeys) != 2 {
		t.Errorf("expected 2 POSTs, got %d: %v", len(createdKeys), createdKeys)
	}
}

func TestGetFilaments(t *testing.T) {
	ts, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("vendor.name") != "Bambu Lab" {
			t.Errorf("expected vendor.name filter")
		}
		json.NewEncoder(w).Encode([]Filament{
			{ID: 1, Name: "PLA Basic", Material: "PLA"},
		})
	})
	defer ts.Close()

	filaments, err := client.GetFilaments(url.Values{"vendor.name": {"Bambu Lab"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(filaments) != 1 || filaments[0].Name != "PLA Basic" {
		t.Errorf("filaments = %+v", filaments)
	}
}

func TestCreateFilament(t *testing.T) {
	ts, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		var params FilamentParams
		json.NewDecoder(r.Body).Decode(&params)
		if params.Density != 1.24 || params.Diameter != 1.75 {
			t.Errorf("params = %+v", params)
		}
		json.NewEncoder(w).Encode(Filament{ID: 1, Name: params.Name, Density: params.Density, Diameter: params.Diameter})
	})
	defer ts.Close()

	vendorID := 1
	filament, err := client.CreateFilament(FilamentParams{
		Name:     "PLA Basic",
		VendorID: &vendorID,
		Material: "PLA",
		Density:  1.24,
		Diameter: 1.75,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filament.ID != 1 {
		t.Errorf("filament = %+v", filament)
	}
}

func TestGetSpools(t *testing.T) {
	ts, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]Spool{
			{ID: 1, Filament: Filament{ID: 1, Name: "PLA"}, RemainingWeight: 800},
		})
	})
	defer ts.Close()

	spools, err := client.GetSpools(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(spools) != 1 || spools[0].RemainingWeight != 800 {
		t.Errorf("spools = %+v", spools)
	}
}

func TestCreateSpool(t *testing.T) {
	ts, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		var params SpoolParams
		json.NewDecoder(r.Body).Decode(&params)
		if params.FilamentID != 1 {
			t.Errorf("filament_id = %d, want 1", params.FilamentID)
		}
		json.NewEncoder(w).Encode(Spool{ID: 1, Filament: Filament{ID: 1}})
	})
	defer ts.Close()

	spool, err := client.CreateSpool(SpoolParams{FilamentID: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if spool.ID != 1 {
		t.Errorf("spool = %+v", spool)
	}
}

func TestUpdateSpool(t *testing.T) {
	ts, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" || r.URL.Path != "/api/v1/spool/5" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		remaining := 750.0
		json.NewEncoder(w).Encode(Spool{ID: 5, RemainingWeight: remaining})
	})
	defer ts.Close()

	remaining := 750.0
	spool, err := client.UpdateSpool(5, SpoolParams{RemainingWeight: &remaining})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if spool.RemainingWeight != 750 {
		t.Errorf("remaining_weight = %f, want 750", spool.RemainingWeight)
	}
}

func TestUseSpool(t *testing.T) {
	ts, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" || r.URL.Path != "/api/v1/spool/3/use" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var params SpoolUseParams
		json.NewDecoder(r.Body).Decode(&params)
		if params.UseWeight != 10.5 {
			t.Errorf("use_weight = %f, want 10.5", params.UseWeight)
		}
		json.NewEncoder(w).Encode(Spool{ID: 3, RemainingWeight: 789.5})
	})
	defer ts.Close()

	spool, err := client.UseSpool(3, SpoolUseParams{UseWeight: 10.5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if spool.RemainingWeight != 789.5 {
		t.Errorf("remaining = %f, want 789.5", spool.RemainingWeight)
	}
}

func TestGetExternalFilaments(t *testing.T) {
	ts, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/external/filament" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode([]ExternalFilament{
			{ID: "GFL00", Name: "Bambu PLA Basic", Material: "PLA", Vendor: "Bambu Lab"},
		})
	})
	defer ts.Close()

	filaments, err := client.GetExternalFilaments()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(filaments) != 1 || filaments[0].ID != "GFL00" {
		t.Errorf("filaments = %+v", filaments)
	}
}

func TestFindFilamentByBambuID(t *testing.T) {
	ts, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]Filament{
			{ID: 1, Name: "PLA", Extra: map[string]string{"bambu_lab_id": ExtraString("A00-A0")}},
			{ID: 2, Name: "PETG", Extra: map[string]string{"bambu_lab_id": ExtraString("B00-B4")}},
		})
	})
	defer ts.Close()

	f, err := client.FindFilamentByBambuID("B00-B4")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f == nil || f.ID != 2 {
		t.Errorf("expected filament ID 2, got %+v", f)
	}

	f, err = client.FindFilamentByBambuID("ZZZZZ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f != nil {
		t.Errorf("expected nil for unknown ID, got %+v", f)
	}
}

func TestFindFilamentByBambuCode(t *testing.T) {
	ts, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]Filament{
			{ID: 1, Extra: map[string]string{"bambu_lab_code": "10300"}},
			{ID: 2, Extra: map[string]string{"bambu_lab_code": "40601"}},
		})
	})
	defer ts.Close()

	f, err := client.FindFilamentByBambuCode("40601")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f == nil || f.ID != 2 {
		t.Errorf("expected filament ID 2, got %+v", f)
	}
}

func TestFindSpoolByTag(t *testing.T) {
	ts, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]Spool{
			{ID: 1, Extra: map[string]string{"tag": ExtraString("0000000000000001")}},
			{ID: 2, Extra: map[string]string{"tag": ExtraString("0000000000000002")}},
		})
	})
	defer ts.Close()

	s, err := client.FindSpoolByTag("0000000000000002")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s == nil || s.ID != 2 {
		t.Errorf("expected spool ID 2, got %+v", s)
	}

	s, err = client.FindSpoolByTag("nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s != nil {
		t.Errorf("expected nil for unknown tag, got %+v", s)
	}
}

func TestErrorResponse(t *testing.T) {
	ts, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"internal error"}`))
	})
	defer ts.Close()

	_, err := client.GetSpools(nil)
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

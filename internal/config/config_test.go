package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadConfig_Valid(t *testing.T) {
	yaml := `
spoolman:
  address: "http://192.168.1.100:8000"
printers:
  - name: "X1C"
    ip: "192.168.1.50"
    serial: "01P00A00000001"
    access_code: "abcdefgh"
  - name: "P1S"
    ip: "192.168.1.51"
    serial: "01S00A00000002"
    access_code: "ijklmnop"
`
	cfg, err := LoadConfig(writeTempConfig(t, yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Spoolman.Address != "http://192.168.1.100:8000" {
		t.Errorf("spoolman address = %q, want %q", cfg.Spoolman.Address, "http://192.168.1.100:8000")
	}

	if len(cfg.Printers) != 2 {
		t.Fatalf("printers count = %d, want 2", len(cfg.Printers))
	}

	p := cfg.Printers[0]
	if p.Name != "X1C" || p.IP != "192.168.1.50" || p.Serial != "01P00A00000001" || p.AccessCode != "abcdefgh" {
		t.Errorf("printer[0] = %+v", p)
	}
}

func TestLoadConfig_DefaultName(t *testing.T) {
	yaml := `
spoolman:
  address: "http://localhost:8000"
printers:
  - ip: "192.168.1.50"
    serial: "01P00A00000001"
    access_code: "abcdefgh"
`
	cfg, err := LoadConfig(writeTempConfig(t, yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Printers[0].Name != "01P00A00000001" {
		t.Errorf("name = %q, want serial as default", cfg.Printers[0].Name)
	}
}

func TestLoadConfig_MissingSpoolmanAddress(t *testing.T) {
	yaml := `
printers:
  - ip: "192.168.1.50"
    serial: "01P00A00000001"
    access_code: "abcdefgh"
`
	_, err := LoadConfig(writeTempConfig(t, yaml))
	if err == nil {
		t.Fatal("expected error for missing spoolman address")
	}
}

func TestLoadConfig_NoPrinters(t *testing.T) {
	yaml := `
spoolman:
  address: "http://localhost:8000"
printers: []
`
	_, err := LoadConfig(writeTempConfig(t, yaml))
	if err == nil {
		t.Fatal("expected error for empty printers")
	}
}

func TestLoadConfig_MissingPrinterFields(t *testing.T) {
	tests := []struct {
		name string
		yaml string
	}{
		{"missing ip", `
spoolman:
  address: "http://localhost:8000"
printers:
  - serial: "01P00A00000001"
    access_code: "abcdefgh"
`},
		{"missing serial", `
spoolman:
  address: "http://localhost:8000"
printers:
  - ip: "192.168.1.50"
    access_code: "abcdefgh"
`},
		{"missing access_code", `
spoolman:
  address: "http://localhost:8000"
printers:
  - ip: "192.168.1.50"
    serial: "01P00A00000001"
`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := LoadConfig(writeTempConfig(t, tt.yaml))
			if err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	_, err := LoadConfig("/nonexistent/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

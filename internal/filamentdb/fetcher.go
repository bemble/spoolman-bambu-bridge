package filamentdb

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const DefaultURL = "https://raw.githubusercontent.com/piitaya/bambu-spoolman-db/main/filaments.json"

// DB holds the parsed filament database with lookup maps.
type DB struct {
	Entries  []Entry
	ByID     map[string]Entry // keyed by Entry.ID (e.g. "A00-A0")
	ByCode   map[string]Entry // keyed by Entry.Code (e.g. "10300")
}

// Fetch downloads and parses the filament database from the given URL.
func Fetch(url string, httpClient *http.Client) (*DB, error) {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetching filament db: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching filament db: status %d", resp.StatusCode)
	}

	return Parse(resp.Body)
}

// Parse reads and parses filament entries from a reader, building lookup maps.
func Parse(r io.Reader) (*DB, error) {
	var entries []Entry
	if err := json.NewDecoder(r).Decode(&entries); err != nil {
		return nil, fmt.Errorf("parsing filament db: %w", err)
	}

	db := &DB{
		Entries: entries,
		ByID:    make(map[string]Entry, len(entries)),
		ByCode:  make(map[string]Entry, len(entries)),
	}

	for _, e := range entries {
		if e.ID != "" {
			db.ByID[e.ID] = e
		}
		if e.Code != "" {
			db.ByCode[e.Code] = e
		}
	}

	return db, nil
}

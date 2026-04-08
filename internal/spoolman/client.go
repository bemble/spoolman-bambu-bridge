package spoolman

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// Client communicates with the Spoolman REST API.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new Spoolman API client.
func NewClient(baseURL string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{
		baseURL:    baseURL,
		httpClient: httpClient,
	}
}

// --- Custom Fields ---

// GetFields returns all custom fields for the given entity type.
func (c *Client) GetFields(entityType string) ([]Field, error) {
	var fields []Field
	err := c.get(fmt.Sprintf("/api/v1/field/%s", entityType), &fields)
	return fields, err
}

// CreateField creates a custom field on the given entity type.
func (c *Client) CreateField(entityType, key string, params FieldParams) ([]Field, error) {
	var fields []Field
	err := c.post(fmt.Sprintf("/api/v1/field/%s/%s", entityType, key), params, &fields)
	return fields, err
}

// EnsureFields ensures the required custom fields exist in Spoolman:
// - filament level: bambu_lab_id (text), bambu_lab_code (integer)
// - spool level: tag (text)
func (c *Client) EnsureFields() error {
	type fieldDef struct {
		entityType string
		key        string
		params     FieldParams
	}

	defs := []fieldDef{
		{"filament", "bambu_lab_id", FieldParams{Name: "Bambu Lab ID", FieldType: "text"}},
		{"filament", "bambu_lab_code", FieldParams{Name: "Bambu Lab Code", FieldType: "integer"}},
		{"spool", "tag", FieldParams{Name: "Tag", FieldType: "text"}},
	}

	for _, d := range defs {
		fields, err := c.GetFields(d.entityType)
		if err != nil {
			return fmt.Errorf("getting %s fields: %w", d.entityType, err)
		}

		exists := false
		for _, f := range fields {
			if f.Key == d.key {
				exists = true
				break
			}
		}

		if !exists {
			if _, err := c.CreateField(d.entityType, d.key, d.params); err != nil {
				return fmt.Errorf("creating field %s/%s: %w", d.entityType, d.key, err)
			}
		}
	}

	return nil
}

// --- Filaments ---

// GetFilaments returns filaments, optionally filtered by query parameters.
func (c *Client) GetFilaments(queryParams url.Values) ([]Filament, error) {
	path := "/api/v1/filament"
	if len(queryParams) > 0 {
		path += "?" + queryParams.Encode()
	}
	var filaments []Filament
	err := c.get(path, &filaments)
	return filaments, err
}

// CreateFilament creates a new filament.
func (c *Client) CreateFilament(params FilamentParams) (*Filament, error) {
	var filament Filament
	err := c.post("/api/v1/filament", params, &filament)
	return &filament, err
}

// FindFilamentByBambuID searches for a filament matching the given bambu_lab_id extra field.
// Returns nil if not found.
func (c *Client) FindFilamentByBambuID(bambuID string) (*Filament, error) {
	filaments, err := c.GetFilaments(nil)
	if err != nil {
		return nil, err
	}
	for _, f := range filaments {
		if ParseExtraString(f.Extra["bambu_lab_id"]) == bambuID {
			return &f, nil
		}
	}
	return nil, nil
}

// FindFilamentByBambuCode searches for a filament matching the given bambu_lab_code extra field.
// Returns nil if not found.
func (c *Client) FindFilamentByBambuCode(code string) (*Filament, error) {
	filaments, err := c.GetFilaments(nil)
	if err != nil {
		return nil, err
	}
	for _, f := range filaments {
		if f.Extra["bambu_lab_code"] == code {
			return &f, nil
		}
	}
	return nil, nil
}

// UpdateFilament updates a filament by ID.
func (c *Client) UpdateFilament(id int, params FilamentParams) (*Filament, error) {
	var filament Filament
	err := c.patch(fmt.Sprintf("/api/v1/filament/%d", id), params, &filament)
	return &filament, err
}

// --- Spools ---

// GetSpools returns spools, optionally filtered by query parameters.
func (c *Client) GetSpools(queryParams url.Values) ([]Spool, error) {
	path := "/api/v1/spool"
	if len(queryParams) > 0 {
		path += "?" + queryParams.Encode()
	}
	var spools []Spool
	err := c.get(path, &spools)
	return spools, err
}

// CreateSpool creates a new spool.
func (c *Client) CreateSpool(params SpoolParams) (*Spool, error) {
	var spool Spool
	err := c.post("/api/v1/spool", params, &spool)
	return &spool, err
}

// UpdateSpool updates a spool by ID.
func (c *Client) UpdateSpool(id int, params SpoolParams) (*Spool, error) {
	var spool Spool
	err := c.patch(fmt.Sprintf("/api/v1/spool/%d", id), params, &spool)
	return &spool, err
}

// FindSpoolByTag searches for a spool matching the given tag extra field.
// Returns nil if not found.
func (c *Client) FindSpoolByTag(tag string) (*Spool, error) {
	spools, err := c.GetSpools(nil)
	if err != nil {
		return nil, err
	}
	for _, s := range spools {
		if ParseExtraString(s.Extra["tag"]) == tag {
			return &s, nil
		}
	}
	return nil, nil
}

// UseSpool records filament usage on a spool.
func (c *Client) UseSpool(id int, params SpoolUseParams) (*Spool, error) {
	var spool Spool
	err := c.put(fmt.Sprintf("/api/v1/spool/%d/use", id), params, &spool)
	return &spool, err
}

// --- External DB ---

// GetExternalFilaments fetches the external filament database from Spoolman.
func (c *Client) GetExternalFilaments() ([]ExternalFilament, error) {
	var filaments []ExternalFilament
	err := c.get("/api/v1/external/filament", &filaments)
	return filaments, err
}

// --- HTTP helpers ---

func (c *Client) get(path string, result interface{}) error {
	resp, err := c.httpClient.Get(c.baseURL + path)
	if err != nil {
		return fmt.Errorf("GET %s: %w", path, err)
	}
	defer resp.Body.Close()
	return c.decodeResponse(resp, result)
}

func (c *Client) post(path string, body, result interface{}) error {
	return c.doJSON("POST", path, body, result)
}

func (c *Client) patch(path string, body, result interface{}) error {
	return c.doJSON("PATCH", path, body, result)
}

func (c *Client) put(path string, body, result interface{}) error {
	return c.doJSON("PUT", path, body, result)
}

func (c *Client) doJSON(method, path string, body, result interface{}) error {
	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest(method, c.baseURL+path, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%s %s: %w", method, path, err)
	}
	defer resp.Body.Close()
	return c.decodeResponse(resp, result)
}

func (c *Client) decodeResponse(resp *http.Response, result interface{}) error {
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	if result == nil {
		return nil
	}

	return json.NewDecoder(resp.Body).Decode(result)
}

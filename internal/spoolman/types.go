package spoolman

import "encoding/json"

// Field represents an extra/custom field in Spoolman.
type Field struct {
	Key       string `json:"key,omitempty"`
	Name      string `json:"name"`
	FieldType string `json:"field_type,omitempty"`
	Order     int    `json:"order,omitempty"`
}

// FieldParams is the request body for creating a custom field.
type FieldParams struct {
	Name      string `json:"name"`
	FieldType string `json:"field_type"`
	Order     int    `json:"order,omitempty"`
}

// Filament represents a filament in Spoolman.
type Filament struct {
	ID                   int               `json:"id"`
	Name                 string            `json:"name,omitempty"`
	Material             string            `json:"material,omitempty"`
	Density              float64           `json:"density,omitempty"`
	Diameter             float64           `json:"diameter,omitempty"`
	Weight               float64           `json:"weight,omitempty"`
	ExternalID           string            `json:"external_id,omitempty"`
	ColorHex             string            `json:"color_hex,omitempty"`
	SettingsExtruderTemp int               `json:"settings_extruder_temp,omitempty"`
	SettingsBedTemp      int               `json:"settings_bed_temp,omitempty"`
	Extra                map[string]string `json:"extra,omitempty"`
}

// FilamentParams is the request body for creating/updating a filament.
type FilamentParams struct {
	Name                 string            `json:"name,omitempty"`
	VendorID             *int              `json:"vendor_id,omitempty"`
	Material             string            `json:"material,omitempty"`
	Density              float64           `json:"density"`
	Diameter             float64           `json:"diameter"`
	Weight               float64           `json:"weight,omitempty"`
	ExternalID           string            `json:"external_id,omitempty"`
	ColorHex             string            `json:"color_hex,omitempty"`
	SettingsExtruderTemp int               `json:"settings_extruder_temp,omitempty"`
	SettingsBedTemp      int               `json:"settings_bed_temp,omitempty"`
	Extra                map[string]string `json:"extra,omitempty"`
}

// Spool represents a spool in Spoolman.
type Spool struct {
	ID              int               `json:"id"`
	Filament        Filament          `json:"filament"`
	RemainingWeight float64           `json:"remaining_weight,omitempty"`
	UsedWeight      float64           `json:"used_weight,omitempty"`
	InitialWeight   float64           `json:"initial_weight,omitempty"`
	FirstUsed       *string           `json:"first_used,omitempty"`
	LastUsed        *string           `json:"last_used,omitempty"`
	Archived        bool              `json:"archived"`
	Extra           map[string]string `json:"extra,omitempty"`
}

// SpoolParams is the request body for creating/updating a spool.
type SpoolParams struct {
	FilamentID      int               `json:"filament_id,omitempty"`
	RemainingWeight *float64          `json:"remaining_weight,omitempty"`
	InitialWeight   *float64          `json:"initial_weight,omitempty"`
	FirstUsed       *string           `json:"first_used,omitempty"`
	LastUsed        *string           `json:"last_used,omitempty"`
	Extra           map[string]string `json:"extra,omitempty"`
}

// SpoolUseParams is the request body for PUT /spool/{id}/use.
type SpoolUseParams struct {
	UseWeight float64 `json:"use_weight"`
}

// ExternalFilament represents a filament from the Spoolman external DB.
type ExternalFilament struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Material     string  `json:"material"`
	Vendor       string  `json:"manufacturer"`
	Density      float64 `json:"density,omitempty"`
	Diameter     float64 `json:"diameter,omitempty"`
	Weight       float64 `json:"weight,omitempty"`
	ColorHex     string  `json:"color_hex,omitempty"`
	ExtruderTemp int     `json:"settings_extruder_temp,omitempty"`
	BedTemp      int     `json:"settings_bed_temp,omitempty"`
}

// ExtraString encodes a string value for a Spoolman text extra field.
// Spoolman expects text values as JSON-encoded strings: "\"value\"".
func ExtraString(v string) string {
	b, _ := json.Marshal(v)
	return string(b)
}

// ExtraInt encodes an integer value for a Spoolman integer extra field.
// Spoolman expects integer values as plain digit strings: "40601".
func ExtraInt(v int) string {
	b, _ := json.Marshal(v)
	return string(b)
}

// ParseExtraString extracts a string from a Spoolman text extra field value.
func ParseExtraString(v string) string {
	var s string
	if err := json.Unmarshal([]byte(v), &s); err != nil {
		return v
	}
	return s
}

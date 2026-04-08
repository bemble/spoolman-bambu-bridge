package filamentdb

// Entry represents a single filament from the bambu-spoolman-db JSON.
type Entry struct {
	ID         string  `json:"id"`
	Code       string  `json:"code"`
	Material   string  `json:"material"`
	ColorName  string  `json:"color_name"`
	ColorHex   string  `json:"color_hex,omitempty"`
	SpoolmanID *string `json:"spoolman_id"`
}

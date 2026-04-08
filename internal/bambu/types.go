package bambu

// Message is the top-level MQTT report payload from a Bambu printer.
type Message struct {
	Print *PrintReport `json:"print,omitempty"`
}

// PrintReport contains the print status and AMS data.
type PrintReport struct {
	Command    string    `json:"command,omitempty"`
	GcodeState string    `json:"gcode_state,omitempty"`
	AMSStatus  int       `json:"ams_status,omitempty"`
	AMS        *AMSData  `json:"ams,omitempty"`
	VTTray     *AMSTray  `json:"vt_tray,omitempty"`
	VirSlot    []AMSTray `json:"vir_slot,omitempty"`
}

// AMSData holds the list of AMS units and metadata.
type AMSData struct {
	AMS         []AMSUnit `json:"ams"`
	TrayNow     string    `json:"tray_now,omitempty"`
	TrayPre     string    `json:"tray_pre,omitempty"`
	TrayTar     string    `json:"tray_tar,omitempty"`
	Version     int       `json:"version,omitempty"`
	InsertFlag  bool      `json:"insert_flag,omitempty"`
	PowerOnFlag bool      `json:"power_on_flag,omitempty"`
}

// AMSUnit represents a single AMS unit with up to 4 trays.
type AMSUnit struct {
	ID       string    `json:"id"`
	Humidity string    `json:"humidity,omitempty"`
	Temp     string    `json:"temp,omitempty"`
	Trays    []AMSTray `json:"tray"`
}

// AMSTray represents a single filament slot in an AMS unit or the external spool holder.
type AMSTray struct {
	ID            string `json:"id"`
	TagUID        string `json:"tag_uid,omitempty"`
	TrayUUID      string `json:"tray_uuid,omitempty"`
	TrayIDName    string `json:"tray_id_name,omitempty"`
	TrayInfoIdx   string `json:"tray_info_idx,omitempty"`
	TrayType      string `json:"tray_type,omitempty"`
	TraySubBrands string `json:"tray_sub_brands,omitempty"`
	TrayColor     string `json:"tray_color,omitempty"`
	TrayWeight    string `json:"tray_weight,omitempty"`
	TrayDiameter  string `json:"tray_diameter,omitempty"`
	Remain        int    `json:"remain"`
	Cols          []string `json:"cols,omitempty"`
	NozzleTempMin string `json:"nozzle_temp_min,omitempty"`
	NozzleTempMax string `json:"nozzle_temp_max,omitempty"`
	BedTemp       string `json:"bed_temp,omitempty"`
}

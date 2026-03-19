package models

// SystemConfig stores global key-value configuration directly in the database,
// such as the system's generated ZoneID.
type SystemConfig struct {
	ID    string `gorm:"primaryKey;type:varchar(100)" json:"id"` // The configuration KEY (e.g. "ZONE_ID")
	Value string `gorm:"type:text;not null" json:"value"`        // The configuration VALUE
}

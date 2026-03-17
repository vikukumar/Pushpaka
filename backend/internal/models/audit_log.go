package models

type AuditLog struct {
	ID         string `db:"id"          json:"id"`
	UserID     string `db:"user_id"     json:"user_id"`
	Action     string `db:"action"      json:"action"`
	Resource   string `db:"resource"    json:"resource"`
	ResourceID string `db:"resource_id" json:"resource_id"`
	Metadata   string `db:"metadata"    json:"metadata"` // raw JSON string
	IPAddr     string `db:"ip_addr"     json:"ip_addr"`
	UserAgent  string `db:"user_agent"  json:"user_agent"`
	CreatedAt  Time   `db:"created_at"  json:"created_at"`
}

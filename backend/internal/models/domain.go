package models

type Domain struct {
ID         string `db:"id"          json:"id"`
ProjectID  string `db:"project_id"  json:"project_id"`
UserID     string `db:"user_id"     json:"user_id"`
Domain     string `db:"domain"      json:"domain"`
Verified   bool   `db:"verified"    json:"verified"`
SSLEnabled bool   `db:"ssl_enabled" json:"ssl_enabled"`
CreatedAt  Time   `db:"created_at"  json:"created_at"`
UpdatedAt  Time   `db:"updated_at"  json:"updated_at"`
}

type AddDomainRequest struct {
ProjectID string `json:"project_id" binding:"required"`
Domain    string `json:"domain"     binding:"required"`
}

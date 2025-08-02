package models

import (
	"time"
	"gorm.io/gorm"
)

type AuditLog struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	UserID       uint           `json:"user_id" gorm:"not null;index"`
	Action       string         `json:"action" gorm:"not null;size:20"` // CREATE, UPDATE, DELETE
	TableName    string         `json:"table_name" gorm:"not null;size:100"`
	RecordID     uint           `json:"record_id" gorm:"not null;index"`
	OldValues    string         `json:"old_values" gorm:"type:text"`
	NewValues    string         `json:"new_values" gorm:"type:text"`
	IPAddress    string         `json:"ip_address" gorm:"size:45"`
	UserAgent    string         `json:"user_agent" gorm:"type:text"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	User User `json:"user" gorm:"foreignKey:UserID"`
}

// Audit Actions Constants
const (
	AuditActionCreate = "CREATE"
	AuditActionUpdate = "UPDATE"
	AuditActionDelete = "DELETE"
)

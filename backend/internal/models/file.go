package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

type File struct {
	ID        string          `json:"id"`
	UserID    string          `json:"user_id"`
	Key       string          `json:"key"`
	Filename  string          `json:"filename"`
	MimeType  string          `json:"mime_type"`
	Size      int64           `json:"size"`
	Metadata  JSONMap         `json:"metadata,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	DeletedAt *time.Time      `json:"-"`
}

// JSONMap is a custom type for JSONB fields
type JSONMap map[string]interface{}

// Value implements the driver.Valuer interface
func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface
func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(bytes, j)
}

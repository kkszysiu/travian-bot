package model

import (
	"database/sql/driver"
	"time"
)

// SQLiteTime wraps time.Time to support scanning from SQLite TEXT columns.
// The modernc.org/sqlite driver returns TEXT as string, which database/sql
// cannot automatically scan into time.Time.
type SQLiteTime struct {
	time.Time
}

// Scan implements the sql.Scanner interface.
func (t *SQLiteTime) Scan(value interface{}) error {
	if value == nil {
		t.Time = time.Time{}
		return nil
	}
	switch v := value.(type) {
	case string:
		if v == "" {
			t.Time = time.Time{}
			return nil
		}
		parsed, err := time.Parse(time.RFC3339, v)
		if err == nil {
			t.Time = parsed
			return nil
		}
		parsed, err = time.Parse("2006-01-02 15:04:05", v)
		if err == nil {
			t.Time = parsed
			return nil
		}
		t.Time = time.Time{}
		return nil
	case time.Time:
		t.Time = v
		return nil
	default:
		t.Time = time.Time{}
		return nil
	}
}

// Value implements the driver.Valuer interface.
func (t SQLiteTime) Value() (driver.Value, error) {
	if t.IsZero() {
		return "0001-01-01T00:00:00Z", nil
	}
	return t.Time.Format(time.RFC3339), nil
}

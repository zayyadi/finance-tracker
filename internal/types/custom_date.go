package types

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

// CustomDate is a custom type to handle YYYY-MM-DD date format in JSON.
type CustomDate struct {
	time.Time
}

const yyyyMMdd = "2006-01-02" // Go's reference layout for YYYY-MM-DD

// UnmarshalJSON implements the json.Unmarshaler interface for CustomDate.
func (cd *CustomDate) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	if s == "null" || s == "" {
		cd.Time = time.Time{} // Handle null or empty string as zero time
		return nil
	}
	t, err := time.Parse(yyyyMMdd, s)
	if err != nil {
		return fmt.Errorf("parsing time %q as \"%s\": %w", s, yyyyMMdd, err)
	}
	cd.Time = t
	return nil
}

// MarshalJSON implements the json.Marshaler interface for CustomDate.
func (cd CustomDate) MarshalJSON() ([]byte, error) {
	if cd.Time.IsZero() {
		return []byte("null"), nil // Marshal zero time as null, or `""` if you prefer an empty string
	}
	return []byte(fmt.Sprintf("\"%s\"", cd.Time.Format(yyyyMMdd))), nil
}

// Value implements the driver.Valuer interface for CustomDate.
// This allows GORM to store CustomDate as a time.Time in the database.
func (cd CustomDate) Value() (driver.Value, error) {
	if cd.Time.IsZero() {
		return nil, nil
	}
	return cd.Time, nil
}

// Scan implements the sql.Scanner interface for CustomDate.
// This allows GORM to read a time.Time from the database into CustomDate.
func (cd *CustomDate) Scan(value interface{}) error {
	if value == nil {
		cd.Time = time.Time{}
		return nil
	}
	if t, ok := value.(time.Time); ok {
		cd.Time = t
		return nil
	}
	return fmt.Errorf("failed to scan CustomDate: %v", value)
}

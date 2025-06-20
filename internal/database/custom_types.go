package database

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// CustomDate is a wrapper around time.Time to handle custom date formatting (YYYY-MM-DD).
type CustomDate struct {
	time.Time
}

// Layout defines the date format used for CustomDate.
const dateLayout = "2006-01-02"
const dateTimeLayout = "2006-01-02 15:04:05" // Used for broader compatibility if DB stores time

// MarshalJSON implements the json.Marshaler interface.
// This ensures that when a struct containing CustomDate is marshaled to JSON,
// the date is formatted as "YYYY-MM-DD".
func (cd *CustomDate) MarshalJSON() ([]byte, error) {
	if cd.Time.IsZero() {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf(`"%s"`, cd.Time.Format(dateLayout))), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
// This allows CustomDate to be correctly unmarshaled from a "YYYY-MM-DD" string in JSON.
func (cd *CustomDate) UnmarshalJSON(b []byte) error {
	s := string(b)
	if s == "null" || s == `""` { // Handle null or empty string from JSON
		cd.Time = time.Time{}
		return nil
	}
	// Remove quotes from JSON string
	if len(s) > 1 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
	}

	t, err := time.Parse(dateLayout, s)
	if err != nil {
		// Fallback: try parsing as a full timestamp if YYYY-MM-DD fails
		// This might happen if the input JSON has a full timestamp
		t, err = time.Parse(time.RFC3339, s)
		if err != nil {
			return fmt.Errorf("error parsing CustomDate from JSON string %s: %w", s, err)
		}
	}
	cd.Time = t
	return nil
}

// Scan implements the sql.Scanner interface.
// This method is used by GORM to read data from the database.
func (cd *CustomDate) Scan(value interface{}) error {
	if value == nil {
		cd.Time = time.Time{}
		return nil
	}

	var t time.Time
	var err error

	switch v := value.(type) {
	case time.Time:
		cd.Time = v
		return nil
	case []byte: // SQLite often returns dates/times as []byte (string)
		s := string(v)
		t, err = time.Parse(dateLayout, s) // Try YYYY-MM-DD first
		if err != nil {
			t, err = time.Parse(dateTimeLayout, s) // Fallback to YYYY-MM-DD HH:MM:SS
			if err != nil {
				t, err = time.Parse(time.RFC3339, s) // Fallback to RFC3339
				if err != nil {
					return fmt.Errorf("cannot scan CustomDate from []byte: %s, error: %w", s, err)
				}
			}
		}
		cd.Time = t
		return nil
	case string:
		t, err = time.Parse(dateLayout, v) // Try YYYY-MM-DD first
		if err != nil {
			t, err = time.Parse(dateTimeLayout, v) // Fallback to YYYY-MM-DD HH:MM:SS
			if err != nil {
				t, err = time.Parse(time.RFC3339, v) // Fallback to RFC3339
				if err != nil {
					return fmt.Errorf("cannot scan CustomDate from string: %s, error: %w", v, err)
				}
			}
		}
		cd.Time = t
		return nil
	default:
		return fmt.Errorf("cannot scan CustomDate: unsupported type %T", value)
	}
}

// Value implements the driver.Valuer interface.
// This method is used by GORM to write data to the database.
// It ensures the date is stored in "YYYY-MM-DD" format.
func (cd CustomDate) Value() (driver.Value, error) {
	if cd.Time.IsZero() {
		return nil, nil
	}
	// Store only the date part, as a string.
	// GORM with SQLite will handle this as TEXT affinity.
	return cd.Time.Format(dateLayout), nil
}

package models

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// Time wraps time.Time to ensure correct round-trip storage with both
// PostgreSQL (returns time.Time directly) and SQLite (returns a text string
// in Go's "2006-01-02 15:04:05.999999999 -0700 MST" format).
//
// Value() stores as RFC3339Nano text so both drivers read it back uniformly.
// Scan() accepts time.Time, RFC3339Nano strings, and modernc's time.String format.
type Time struct{ time.Time }

// NewTime wraps t into a models.Time (convenience helper).
func NewTime(t time.Time) Time {
	return Time{t.UTC()}
}

// NowUTC returns the current UTC time as models.Time.
func NowUTC() Time {
	return Time{time.Now().UTC()}
}

// Scan implements sql.Scanner.
func (t *Time) Scan(src interface{}) error {
	if src == nil {
		t.Time = time.Time{}
		return nil
	}
	switch v := src.(type) {
	case time.Time:
		t.Time = v.UTC()
		return nil
	case string:
		return t.parseString(v)
	case []byte:
		return t.parseString(string(v))
	default:
		return fmt.Errorf("cannot scan %T into models.Time", src)
	}
}

// Value implements driver.Valuer — stores as RFC3339Nano TEXT for both drivers.
func (t Time) Value() (driver.Value, error) {
	if t.IsZero() {
		return nil, nil
	}
	return t.UTC().Format(time.RFC3339Nano), nil
}

// MarshalJSON delegates to time.Time so JSON output is a plain RFC3339 string.
func (t Time) MarshalJSON() ([]byte, error) { return t.Time.UTC().MarshalJSON() }

// UnmarshalJSON delegates to time.Time.
func (t *Time) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}
	var tt time.Time
	if err := tt.UnmarshalJSON(data); err != nil {
		return err
	}
	t.Time = tt
	return nil
}

// parseString tries several time formats produced by different SQL drivers.
func (t *Time) parseString(s string) error {
	formats := []string{
		time.RFC3339Nano,
		time.RFC3339,
		// modernc.org/sqlite stores time.Time with time.String() format
		"2006-01-02 15:04:05.999999999 -0700 MST",
		"2006-01-02 15:04:05 -0700 MST",
		"2006-01-02T15:04:05",
		"2006-01-02",
	}
	for _, f := range formats {
		if parsed, err := time.Parse(f, s); err == nil {
			t.Time = parsed.UTC()
			return nil
		}
	}
	return fmt.Errorf("cannot parse time string %q", s)
}

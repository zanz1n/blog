package dto

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"time"
)

var (
	_ sql.Scanner   = &Timestamp{}
	_ driver.Valuer = &Timestamp{}
)

type Timestamp struct {
	time.Time
}

// Scan implements sql.Scanner.
func (t *Timestamp) Scan(src any) error {
	switch src := src.(type) {
	case int:
		*t = Timestamp{time.UnixMilli(int64(src))}
	case int64:
		*t = Timestamp{time.UnixMilli(src)}
	case uint:
		*t = Timestamp{time.UnixMilli(int64(src))}
	case uint64:
		*t = Timestamp{time.UnixMilli(int64(src))}
	default:
		return fmt.Errorf("Scan: unable to scan type %T into time.Time", src)
	}
	return nil
}

// Value implements driver.Valuer.
func (t Timestamp) Value() (driver.Value, error) {
	return t.UnixMilli(), nil
}

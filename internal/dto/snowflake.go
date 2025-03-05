package dto

import (
	"database/sql"
	"database/sql/driver"
	"encoding"
	"encoding/binary"
	"errors"
	"fmt"
	"math/rand/v2"
	"strconv"
	"time"

	"github.com/zanz1n/blog/internal/utils"
)

const (
	SnowflakeEpoch    = 1420070400000
	SnowflakeRandMask = 0x3FFFFF
	NullSnowflake     = Snowflake(0)
)

var (
	_nullSnowflake = Snowflake(0)

	_ sql.Scanner   = &_nullSnowflake
	_ driver.Valuer = _nullSnowflake
	_ fmt.Stringer  = _nullSnowflake

	_ encoding.TextMarshaler   = _nullSnowflake
	_ encoding.TextUnmarshaler = &_nullSnowflake
)

type Snowflake uint64

func NewSnowflake() Snowflake {
	return NewSnowflakeTime(time.Now())
}

func NewSnowflakeTime(t time.Time) Snowflake {
	return NewSnowflakeWith(t, rand.Uint32())
}

func NewSnowflakeWith(t time.Time, rand uint32) Snowflake {
	s := Snowflake(t.UnixMilli()-SnowflakeEpoch) << 22
	s |= Snowflake(rand) & SnowflakeRandMask
	return s
}

func (s Snowflake) Timestamp() time.Time {
	return time.UnixMilli(s.TimestampUnix())
}

func (s Snowflake) TimestampUnix() int64 {
	return int64((s >> 22) + SnowflakeEpoch)
}

func (s Snowflake) Rand() uint32 {
	return uint32(s & SnowflakeRandMask)
}

// Scan implements sql.Scanner.
func (s *Snowflake) Scan(src any) error {
	switch src := src.(type) {
	case int:
		*s = Snowflake(src)
		return nil
	case int64:
		*s = Snowflake(src)
		return nil
	case uint:
		*s = Snowflake(src)
		return nil
	case uint64:
		*s = Snowflake(src)
		return nil
	case string:
		if src == "" {
			*s = NullSnowflake
			return nil
		}

		v, err := strconv.ParseUint(src, 10, 0)
		if err != nil {
			return errors.Join(snowflakeScanErr(src), err)
		}

		*s = Snowflake(v)
		return nil
	case []byte:
		if len(src) == 0 {
			*s = NullSnowflake
			return nil
		} else if len(src) == 8 {
			return s.Scan(binary.LittleEndian.Uint64(src))
		} else {
			return s.Scan(utils.UnsafeString(src))
		}
	case nil:
		*s = NullSnowflake
		return nil
	default:
		return snowflakeScanErr(src)
	}
}

// String implements fmt.Stringer.
func (s Snowflake) String() string {
	return strconv.FormatUint(uint64(s), 10)
}

// Value implements driver.Valuer.
func (s Snowflake) Value() (driver.Value, error) {
	return int64(s), nil
}

// MarshalText implements encoding.TextMarshaler.
func (s Snowflake) MarshalText() (text []byte, err error) {
	return utils.UnsafeBytes(s.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (s *Snowflake) UnmarshalText(text []byte) error {
	v, err := strconv.ParseUint(utils.UnsafeString(text), 10, 0)
	if err != nil {
		return err
	}

	*s = Snowflake(v)
	return nil
}

func snowflakeScanErr(src any) error {
	return fmt.Errorf("Scan: unable to scan type %T into Snowflake", src)
}

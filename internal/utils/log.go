package utils

import (
	"log/slog"
	"time"
)

func DurationAttr(dur time.Duration) slog.Attr {
	return slog.Attr{
		Key:   "took",
		Value: slog.DurationValue(dur),
	}
}

func TookAttr(start time.Time, round time.Duration) slog.Attr {
	return DurationAttr(time.Since(start).Round(round))
}

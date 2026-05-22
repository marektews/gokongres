package helpers

import (
	"time"
)

func FormatTime(t *time.Time) string {
	if t != nil {
		return t.Format("15:04")
	}
	return ""
}

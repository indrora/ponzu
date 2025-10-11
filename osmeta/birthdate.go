//go:build !unix && !windows

package osmeta

import (
	"syscall"
	"time"
)

func getCreationTime(s *syscall.Stat_t) *time.Time {
	return nil
}

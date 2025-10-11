//go:build unix && !darwin

package osmeta

import (
	"syscall"
	"time"
)

func getCreationTime(s *syscall.Stat_t) *time.Time {
	ctime := time.Unix(s.Ctim.Sec, int64(s.Ctim.Nsec))
	return &ctime
}

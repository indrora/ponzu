//go:build darwin

package osmeta

import (
	"syscall"
	"time"
)

func getCreationTime(s *syscall.Stat_t) *time.Time {
	tTime := time.Unix(s.Birthtimespec.Sec, s.Birthtimespec.Nsec)
	return &tTime
}

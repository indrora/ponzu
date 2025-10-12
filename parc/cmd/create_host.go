package cmd

import (
	"runtime"

	"github.com/indrora/ponzu/ponzu/format"
)

func getmachineHostOS() string {
	switch runtime.GOOS {
	case "darwin":
		return format.HOST_OS_DARWIN
	case "linux":
		return format.HOST_OS_LINUX
	case "windows":
		return format.HOST_OS_NT
	case "dragonfly":
	case "freebsd":
	case "illumos":
	case "netbsd":
	case "openbsd":
	case "solaris":
		return format.HOST_OS_UNIX
	default:
		return format.HOST_OS_GENERIC
	}
	return format.HOST_OS_GENERIC
}

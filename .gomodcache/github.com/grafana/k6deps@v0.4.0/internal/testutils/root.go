package testutils

import "runtime"

// OSRoot returns a "root" directory dependending on the operating system
func OSRoot() string {
	if runtime.GOOS == "windows" {
		return "C:\\"
	}

	return "/"
}

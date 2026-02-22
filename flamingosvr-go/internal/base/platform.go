package base

import (
	"runtime"
	"syscall"
)

func GetOS() string {
	return runtime.GOOS
}

func GetArch() string {
	return runtime.GOARCH
}

func GetNumCPU() int {
	return runtime.NumCPU()
}

func SetMaxThreads(n int) {
	runtime.GOMAXPROCS(n)
}

func GetCurrentThreadID() int {
	return syscall.Getpid()
}

func IsWindows() bool {
	return runtime.GOOS == "windows"
}

func IsLinux() bool {
	return runtime.GOOS == "linux"
}

func IsMac() bool {
	return runtime.GOOS == "darwin"
}

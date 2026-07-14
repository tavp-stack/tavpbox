//go:build !windows
// +build !windows

package cmd

import (
	"os/exec"
)

func detachProcess(cmd *exec.Cmd) {
	// On Unix, the process is already detached by default
	// when started with cmd.Start()
}

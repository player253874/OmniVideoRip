//go:build !windows
// +build !windows

package extractors

import "os/exec"

func hideCmdWindow(cmd *exec.Cmd) {}

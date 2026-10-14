//go:build !windows

package core

import "os/exec"

// На остальных системах приложение не запускается, но код должен собираться
// для тестов и CI.
func hideConsole(_ *exec.Cmd) {}

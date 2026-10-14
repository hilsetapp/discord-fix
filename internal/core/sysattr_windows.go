//go:build windows

package core

import (
	"os/exec"
	"syscall"
)

// hideConsole прячет чёрное окно консоли у запускаемых движков -
// человек запускает приложение, а не две командные строки рядом.
func hideConsole(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
}

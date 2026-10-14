//go:build windows

package core

import (
	"fmt"
	"os"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

// IsElevated - правда ли процесс уже идёт от имени администратора.
func IsElevated() bool {
	var tok windows.Token
	if err := windows.OpenProcessToken(windows.CurrentProcess(), windows.TOKEN_QUERY, &tok); err != nil {
		return false
	}
	defer tok.Close()
	var elevation uint32
	var outLen uint32
	err := windows.GetTokenInformation(tok, windows.TokenElevation,
		(*byte)(unsafe.Pointer(&elevation)), uint32(unsafe.Sizeof(elevation)), &outLen)
	return err == nil && elevation != 0
}

// RelaunchElevated перезапускает текущий exe через UAC. Значит:
// окно разрешения покажет Windows, пользователь подтвердит - поднимется
// вторая копия уже с правами, а эта завершается в вызывающем коде.
// Ошибка 1223 = пользователь нажал «Нет», это обрабатываем отдельно.
func RelaunchElevated() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	var quoted []string
	for _, a := range os.Args[1:] {
		quoted = append(quoted, `"`+a+`"`)
	}
	params, err := windows.UTF16PtrFromString(strings.Join(quoted, " "))
	if err != nil {
		return err
	}
	verb, _ := windows.UTF16PtrFromString("runas")
	exePtr, _ := windows.UTF16PtrFromString(exe)

	shell32 := windows.NewLazySystemDLL("shell32.dll")
	proc := shell32.NewProc("ShellExecuteW")
	ret, _, _ := proc.Call(0,
		uintptr(unsafe.Pointer(verb)),
		uintptr(unsafe.Pointer(exePtr)),
		uintptr(unsafe.Pointer(params)),
		0, 1) // SW_NORMAL
	if ret <= 32 {
		return fmt.Errorf("ShellExecute runas вернул %d (1223 - это «нет» в UAC)", ret)
	}
	return nil
}

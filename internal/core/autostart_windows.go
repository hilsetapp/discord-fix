//go:build windows

package core

import (
	"golang.org/x/sys/windows/registry"
)

const runKeyPath = `Software\Microsoft\Windows\CurrentVersion\Run`
const runValueName = "DiscordFix"

// AutostartEnabled - есть ли наша запись в автозапуске пользователя.
func AutostartEnabled() bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, runKeyPath, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer k.Close()
	_, _, err = k.GetStringValue(runValueName)
	return err == nil
}

// SetAutostart включает или выключает запуск Discord Фикса вместе с Windows.
func SetAutostart(exePath string, enable bool) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, runKeyPath, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()
	if enable {
		return k.SetStringValue(runValueName, `"`+exePath+`" /minimized`)
	}
	return k.DeleteValue(runValueName)
}

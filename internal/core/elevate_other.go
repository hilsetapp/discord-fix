//go:build !windows

package core

// Вне Windows UAC нет, возвращаем "всё разрешено".
func IsElevated() bool { return true }

func RelaunchElevated() error { return nil }

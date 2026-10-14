//go:build !windows

package core

func AutostartEnabled() bool { return false }

func SetAutostart(_ string, _ bool) error { return nil }

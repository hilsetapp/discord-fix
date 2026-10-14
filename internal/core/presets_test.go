package core

import (
	"strings"
	"testing"
)

func TestArgsForZapret(t *testing.T) {
	p := FindPreset("std")
	if p == nil {
		t.Fatal("нет стратегии std")
	}
	exe, args := ArgsFor(p, "LIST")
	if exe != "winws.exe" {
		t.Fatalf("exe = %q", exe)
	}
	joined := strings.Join(args, "|")
	for _, want := range []string{
		"--hostlist=LIST",
		"--filter-tcp=443",
		"--filter-udp=443",
		"--filter-udp=50000-50099",
		"--new",
	} {
		if !strings.Contains(joined, want) {
			t.Errorf("в аргументах нет %s", want)
		}
	}
	// --new должно быть ровно два: tcp | udp quic | голос
	if n := strings.Count(joined, "--new"); n != 2 {
		t.Errorf("--new встречается %d раз, надо 2", n)
	}
}

func TestArgsForGdpi(t *testing.T) {
	p := FindPreset("gdpi")
	if p == nil {
		t.Fatal("нет стратегии gdpi")
	}
	exe, args := ArgsFor(p, "LIST")
	if exe != "goodbyedpi.exe" {
		t.Fatalf("exe = %q", exe)
	}
	if len(args) != 1 || args[0] != "-5" {
		t.Fatalf("странные аргументы gdpi: %v", args)
	}
}

func TestPresetsUnique(t *testing.T) {
	seen := map[string]bool{}
	for _, p := range Presets() {
		if seen[p.ID] {
			t.Errorf("дубликат id %s", p.ID)
		}
		seen[p.ID] = true
		if p.Name == "" {
			t.Errorf("у %s нет имени", p.ID)
		}
	}
}

func TestCheckResultString(t *testing.T) {
	r := CheckResult{SiteOK: true, CDNOK: false}
	if !r.OK() {
		t.Fatal("должно быть ок")
	}
	if !strings.Contains(r.String(), "сайт: ок") {
		t.Errorf("строка результата непонятная: %q", r.String())
	}
}

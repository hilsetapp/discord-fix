package core

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Хеши официальных бинарей, сняты при скачивании с релизов авторов
// (zapret v72.13 от bol-van, GoodbyeDPI 0.2.2 от ValdikSS).
// Это защита: если файл подменили, selftest сразу увидит.
var pinnedHashes = map[string]string{
	"cores/zapret/winws.exe":       "a14bff1df6234ea555d2e0c61b589f0707c0b12d6c9b7eeccda76012154996e8",
	"cores/zapret/WinDivert.dll":   "c1e060ee19444a259b2162f8af0f3fe8c4428a1c6f694dce20de194ac8d7d9a2",
	"cores/zapret/WinDivert64.sys": "8da085332782708d8767bcace5327a6ec7283c17cfb85e40b03cd2323a90ddc2",
	"cores/gdpi/goodbyedpi.exe":    "331ac6c1d22ba5a0a217f3f27d0d823051869cafc8b8ef7f2002fa2accebc74e",
	"cores/gdpi/WinDivert.dll":     "a97859785a2df1d4462e7d48d33ccbd89fedd40dac4970f4afd89e63f59ee1ec",
	"cores/gdpi/WinDivert64.sys":   "53ab28ec00be6e6f8aefa9ee76fc2735e94d7f3f9dbc06eb2b7ac8cd3084a6af",
}

func sha256Of(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// Selftest проверяет: бинари ядер на месте и с верными хешами,
// стратегии собираются, список доменов непустой. Код выхода 0 - всё хорошо.
func Selftest(root string) (ok bool, report string) {
	var b strings.Builder
	ok = true
	say := func(format string, a ...any) { fmt.Fprintf(&b, format+"\n", a...) }

	say("== selftest ==")

	// 1. ядра
	for rel, want := range pinnedHashes {
		full := filepath.Join(root, rel)
		got, err := sha256Of(full)
		switch {
		case err != nil:
			say("[ПЛОХО] %s нет файла: %v", rel, err)
			ok = false
		case got != want:
			say("[ПЛОХО] %s хеш не совпал: got %s", rel, got[:16])
			ok = false
		default:
			say("[ок] %s", rel)
		}
	}

	// 2. список доменов
	listBytes, err := os.ReadFile(filepath.Join(root, "lists", "discord.txt"))
	if err != nil {
		say("[ПЛОХО] lists/discord.txt: %v", err)
		ok = false
	} else {
		n := 0
		for _, line := range strings.Split(string(listBytes), "\n") {
			if strings.TrimSpace(line) != "" {
				n++
			}
		}
		say("[ок] lists/discord.txt: %d доменов", n)
		if n < 3 {
			say("[ПЛОХО] доменов подозрительно мало")
			ok = false
		}
	}

	// 3. стратегии собираются в аргументы
	for _, p := range Presets() {
		exe, args := ArgsFor(&p, "LISTS")
		if exe == "" || len(args) == 0 {
			say("[ПЛОХО] стратегия %s пустая", p.ID)
			ok = false
			continue
		}
		say("[ок] стратегия %-12s -> %s + %d аргументов", p.ID, exe, len(args))
	}

	if ok {
		say("итого: всё в порядке")
	} else {
		say("итого: ЕСТЬ ПРОБЛЕМЫ - смотри выше")
	}
	return ok, b.String()
}

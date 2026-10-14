package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/hilsetapp/discord-fix/internal/core"
	"github.com/hilsetapp/discord-fix/internal/gui"
)

var version = "dev"

func locateRoot() string {
	// Обычный случай: папки cores и lists лежат рядом с exe.
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		if dirHasCores(filepath.Join(dir, "cores")) {
			return dir
		}
	}
	// Запуск из исходников: go run ./cmd/discord-fix - cwd это корень репо
	// или cmd/discord-fix.
	if cwd, err := os.Getwd(); err == nil {
		for _, cand := range []string{cwd, filepath.Join(cwd, "..", "..")} {
			if dirHasCores(filepath.Join(cand, "cores")) {
				return cand
			}
		}
	}
	return "."
}

func dirHasCores(cores string) bool {
	st, err := os.Stat(cores)
	return err == nil && st.IsDir()
}

func main() {
	// Windows-юзеры привыкли к /minimized в автоке загрузке, принимаем и такой.
	minimized := false
	args := make([]string, 0, len(os.Args))
	for _, a := range os.Args {
		if strings.EqualFold(a, "/minimized") {
			minimized = true
			continue
		}
		args = append(args, a)
	}
	os.Args = args

	var doSelftest, doCheck bool
	var runPreset string
	var showVersion bool
	flag.BoolVar(&doSelftest, "selftest", false, "проверить бинари и сборку")
	flag.BoolVar(&doCheck, "check", false, "проверить, открывается ли Discord")
	flag.StringVar(&runPreset, "run", "", `запуск без окна: id стратегии или "auto"`)
	flag.BoolVar(&showVersion, "version", false, "печатать версию")
	flag.Parse()

	root := locateRoot()

	switch {
	case showVersion:
		fmt.Printf("discord-fix %s (%s/%s)\n", version, runtime.GOOS, runtime.GOARCH)
		return
	case doSelftest:
		ok, report := core.Selftest(root)
		fmt.Print(report)
		if !ok {
			os.Exit(1)
		}
		return
	case doCheck:
		st := core.CheckDiscord(6 * time.Second)
		fmt.Println(st)
		if st.OK() {
			return
		}
		os.Exit(1)
	case runPreset != "":
		headlessRun(root, runPreset)
		return
	}

	// Оконный режим: движкам нужны права администратора.
	// Не полагаемся только на манифест - перезапрашиваем сами через UAC,
	// если вдруг exe без прав (отменили, старая сборка, тихий антивирус).
	if !core.IsElevated() {
		if err := core.RelaunchElevated(); err == nil {
			return // перезапустились с правами, эта копия больше не нужна
		}
		// «Нет» в UAC или ошибка - пойдём дальше, программа сама напишет в журнале,
		// почему движки не поднимаются.
	}
	gui.Run(root, version, minimized)
}

// headlessRun - вариант для командной строки и отладки.
func headlessRun(root, preset string) {
	fmt.Println("discord-fix: работа без окна. Ctrl+C - остановить.")
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	events := make(chan core.Event, 64)
	go func() {
		for e := range events {
			fmt.Println("  >", e.Text)
		}
	}()

	ctl := core.NewController(root, events)
	ctl.Activate(preset)
	<-ctx.Done()
	ctl.Stop()
	fmt.Println("остановлено.")
}

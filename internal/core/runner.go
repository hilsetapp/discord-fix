package core

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"sync"
)

// Runner держит запущенный процесс движка (winws или goodbyedpi).
type Runner struct {
	Dir string // папка, где лежит exe (cores/zapret или cores/gdpi)

	mu      sync.Mutex
	cmd     *exec.Cmd
	cancel  context.CancelFunc
	running bool
	done    chan struct{} // закрывается, когда процесс завершился
	anyOut  bool          // написал ли процесс хоть строку
	Lines   chan string   // хвост логов для показа в приложении
}

// Start поднимает движок. Уже запущенный сначала глушим.
func (r *Runner) Start(exe string, args []string) error {
	r.Stop()
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, filepath.Join(r.Dir, exe), args...)
	cmd.Dir = r.Dir
	hideConsole(cmd)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return err
	}
	cmd.Stderr = cmd.Stdout

	r.mu.Lock()
	r.cmd, r.cancel, r.running = cmd, cancel, true
	r.done = make(chan struct{})
	r.anyOut = false
	done := r.done
	r.mu.Unlock()
	if r.Lines != nil {
		select {
		case r.Lines <- "запускаю " + exe:
		default:
		}
	}

	if err := cmd.Start(); err != nil {
		cancel()
		r.setRunning(false)
		return err
	}

	go func() {
		sc := bufio.NewScanner(stdout)
		sc.Buffer(make([]byte, 0, 64*1024), 256*1024)
		for sc.Scan() {
			line := sc.Text()
			r.mu.Lock()
			r.anyOut = true
			r.mu.Unlock()
			if r.Lines != nil {
				select {
				case r.Lines <- line:
				default: // не блокируемся, если никто не читает
				}
			}
		}
	}()

	go func() {
		err := cmd.Wait()
		r.setRunning(false)
		if r.Lines != nil {
			text := exe + " завершился"
			switch {
			case err == nil:
				text = exe + " завершился сам по себе (код 0)"
			default:
				if ee, ok := err.(*exec.ExitError); ok {
					text = fmt.Sprintf("%s завершился с ошибкой, код %d", exe, ee.ExitCode())
				} else {
					text = exe + " завершился: " + err.Error()
				}
			}
			select {
			case r.Lines <- text:
			default:
			}
		}
		close(done)
	}()
	return nil
}

// AnyOutput - написал ли процесс что-нибудь в stdout/stderr.
// Полезно для диагностики тихой смерти (антивирус и т.п.).
func (r *Runner) AnyOutput() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.anyOut
}

// Done закрывается после завершения последнего запущенного процесса.
func (r *Runner) Done() <-chan struct{} {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.done == nil {
		ch := make(chan struct{})
		close(ch)
		return ch
	}
	return r.done
}

// Stop глушит движок, если тот бежит.
func (r *Runner) Stop() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.cancel != nil {
		r.cancel()
		r.cancel = nil
	}
	r.running = false
}

// Running - правда ли, что процесс ещё жив.
func (r *Runner) Running() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.running
}

func (r *Runner) setRunning(v bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.running = v
}

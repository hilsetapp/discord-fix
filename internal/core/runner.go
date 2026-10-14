package core

import (
	"bufio"
	"context"
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
	Lines   chan string // хвост логов для показа в приложении
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
		for sc.Scan() {
			if r.Lines != nil {
				select {
				case r.Lines <- sc.Text():
				default: // не блокируемся, если никто не читает
				}
			}
		}
	}()

	go func() {
		_ = cmd.Wait()
		r.setRunning(false)
		if r.Lines != nil {
			select {
			case r.Lines <- exe + " завершился":
			default:
			}
		}
		close(done)
	}()
	return nil
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

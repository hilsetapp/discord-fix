package core

import (
	"context"
	"path/filepath"
	"sync"
	"time"
)

// Сколько даём каждой стратегии в режиме «Авто», прежде чем попробовать следующую.
const AutoPresetWait = 14 * time.Second

// Сколько ждём подъёма Discord при явном выборе стратегии.
const ManualWait = 20 * time.Second

// Event - сообщение контроллера в интерфейс.
// Kind: "status" (строка состояния), "log" (строчка в журнал),
// "run-started" (работаем, Text = название стратегии),
// "run-failed" (не пробилось, Text = причина),
// "died" (движок завершился сам).
type Event struct {
	Kind string
	Text string
}

// EngineDir - папка движка внутри корня приложения.
func EngineDir(root string, e Engine) string {
	return filepath.Join(root, "cores", string(e))
}

// Controller запускает и глушит движки по команде интерфейса.
// Весь долгий путь (перебор стратегий, ожидание сети) идёт в горутине,
// о ходе дел сообщает событиями.
type Controller struct {
	root   string
	events chan<- Event

	mu     sync.Mutex
	runner *Runner
	cancel context.CancelFunc
	busy   bool
}

func NewController(root string, events chan<- Event) *Controller {
	return &Controller{root: root, events: events}
}

func (c *Controller) send(kind, text string) {
	select {
	case c.events <- Event{Kind: kind, Text: text}:
	default:
	}
}

// Busy - идёт ли прямо сейчас перебор стратегий.
func (c *Controller) Busy() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.busy
}

// Active - работает ли сейчас какой-то движок.
func (c *Controller) Active() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.runner != nil && c.runner.Running()
}

// Activate поднимает стратегию по id, либо перебирает все, если передан "auto".
// Вызывать можно в любой момент: если уже идёт перебор, второй запуск игнорируется.
func (c *Controller) Activate(presetID string) {
	c.mu.Lock()
	if c.busy {
		c.mu.Unlock()
		return
	}
	if c.runner != nil {
		// что-то уже работает - сначала глушим, потом запускаем новое
		old := c.runner
		c.runner = nil
		old.Stop()
	}
	c.busy = true
	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel
	c.mu.Unlock()

	go c.work(ctx, presetID)
}

// Stop выключает всё: и перебор, и работающий движок.
func (c *Controller) Stop() {
	c.mu.Lock()
	cancel := c.cancel
	r := c.runner
	c.cancel = nil
	c.runner = nil
	c.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	if r != nil {
		r.Stop()
	}
}

// work - тело перебора. Заканчивается либо успехом (движок оставлен работать),
// либо событием run-failed, либо тихо, если юзер сам всё выключил.
func (c *Controller) work(ctx context.Context, presetID string) {
	defer func() {
		c.mu.Lock()
		c.busy = false
		c.cancel = nil
		c.mu.Unlock()
	}()

	var trials []*Preset
	if presetID == "auto" {
		all := Presets()
		for i := range all {
			if all[i].Engine == EngineZapret {
				trials = append(trials, &all[i])
			}
		}
	} else {
		p := FindPreset(presetID)
		if p == nil {
			c.send("run-failed", "стратегия "+presetID+" не найдена")
			return
		}
		trials = []*Preset{p}
	}

	wait := AutoPresetWait
	if presetID != "auto" {
		wait = ManualWait
	}
	listPath := filepath.Join(c.root, "lists", "discord.txt")

	for _, p := range trials {
		if ctx.Err() != nil {
			return
		}
		c.send("status", "пробую стратегию «"+p.Name+"»…")
		r := &Runner{Dir: EngineDir(c.root, p.Engine), Lines: make(chan string, 64)}
		go func() {
			for line := range r.Lines {
				c.send("log", line)
			}
		}()
		exe, args := ArgsFor(p, listPath)
		if err := r.Start(exe, args); err != nil {
			c.send("log", exe+" не запустился: "+err.Error())
			continue
		}

		// Ждём одновременно три исхода: Discord открылся, юзер отменил,
		// движок умер сразу после старта (обычно это значит, что
		// приложение запущено без прав администратора).
		reach := make(chan bool, 1)
		go func() {
			_, ok := WaitReachable(ctx, wait)
			reach <- ok
		}()
		select {
		case ok := <-reach:
			if !ok {
				r.Stop()
				if ctx.Err() != nil {
					return // юзер сам нажал «выключить» - выходим тихо
				}
				c.send("log", "за "+wait.Round(time.Second).String()+" не пробилась, идём дальше")
				continue
			}
		case <-ctx.Done():
			r.Stop()
			return
		case <-r.Done():
			if r.AnyOutput() {
				c.send("log", "движок выключился сразу после старта. Частая причина - приложение запущено без прав администратора")
			} else {
				if !IsElevated() {
					c.send("log", "движок замолчал сразу: приложение без прав администратора. Закрой и открой заново, разреши UAC")
				} else {
					c.send("log", "движок умер мгновенно и молча, даже с правами. Похоже на антивирус - загляни в карантин и добавь папку в исключения")
				}
			}
			continue
		}
		// успех: движок уже бежит, оставляем его жить
		c.mu.Lock()
		c.runner = r
		c.mu.Unlock()
		c.send("run-started", p.Name)
		go c.watch(r)
		return
	}
	c.send("run-failed", "ни одна стратегия не пробилась - попробуй позже или смени сеть")
}

// watch ждёт самопроизвольной смерти движка и сообщает об этом.
func (c *Controller) watch(r *Runner) {
	<-r.Done()
	c.mu.Lock()
	still := c.runner == r
	if still {
		c.runner = nil
	}
	c.mu.Unlock()
	if still {
		c.send("died", "движок завершился сам по себе - нажми кнопку, чтобы перезапустить")
	}
}

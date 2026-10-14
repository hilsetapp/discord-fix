// Тёмное окно Discord Фикса: одна большая кнопка, выбор стратегии, журнал.
// Всё рисуется Gio-юзер интерфейсом, долгие операции - в горутинах
// контроллера, который шлёт события в канал.
package gui

import (
	_ "embed"
	"fmt"
	"image"
	"image/color"
	"os"
	"time"

	"gioui.org/app"
	"gioui.org/f32"
	"gioui.org/font/opentype"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/hilsetapp/discord-fix/internal/core"
)

//go:embed InterVariable.ttf
var interFont []byte

var (
	colBg      = color.NRGBA{R: 17, G: 18, B: 23, A: 255}
	colPanel   = color.NRGBA{R: 27, G: 29, B: 37, A: 255}
	colPanelHi = color.NRGBA{R: 37, G: 40, B: 51, A: 255}
	colHover   = color.NRGBA{R: 46, G: 50, B: 64, A: 255}
	colText    = color.NRGBA{R: 232, G: 234, B: 240, A: 255}
	colMuted   = color.NRGBA{R: 146, G: 150, B: 162, A: 255}
	colAccent  = color.NRGBA{R: 88, G: 101, B: 242, A: 255} // blurple
	colAccent2 = color.NRGBA{R: 116, G: 128, B: 250, A: 255}
	colGreen   = color.NRGBA{R: 67, G: 181, B: 129, A: 255}
	colRed     = color.NRGBA{R: 237, G: 66, B: 69, A: 255}
	colYellow  = color.NRGBA{R: 250, G: 168, B: 26, A: 255}
	colWhite   = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
)

type appState int

const (
	stateIdle appState = iota
	stateSwitching
	stateOn
	stateFailed
)

type ui struct {
	th *material.Theme

	ctl *core.Controller

	presetItems []string
	presetIDs   []string
	presetSel   int
	presetOpen  bool
	presetClick widget.Clickable
	itemClick   []widget.Clickable

	toggle    widget.Clickable
	checkBtn  widget.Clickable
	autoBox   widget.Bool
	logToggle widget.Clickable
	logOpen   bool
	logList   widget.List

	state     appState
	status    string
	statusCol color.NRGBA
	logs      []string
	startedAt time.Time
	events    chan core.Event
}

// Run запускает окно и не возвращается, пока его не закроют.
func Run(root, version string, minimized bool) {
	_ = minimized
	go func() {
		w := new(app.Window)
		w.Option(app.Title("Discord Фикс"),
			app.Size(unit.Dp(480), unit.Dp(640)),
			app.MinSize(unit.Dp(430), unit.Dp(560)))
		if err := loop(w, root, version); err != nil {
			fmt.Fprintln(os.Stderr, "ошибка окна:", err)
			os.Exit(1)
		}
		os.Exit(0)
	}()
	app.Main()
}

func loop(w *app.Window, root, version string) error {
	u := newUI(root)

	var ops op.Ops
	for {
		switch e := w.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			u.drainEvents()
			u.handleClicks(gtx)
			u.layout(gtx, version)
			e.Frame(gtx.Ops)
		}
	}
}

func newUI(root string) *ui {
	th := material.NewTheme()
	if faces, err := opentype.ParseCollection(interFont); err == nil {
		th.Shaper = text.NewShaper(text.WithCollection(faces))
	}
	th.Palette.Bg = colBg
	th.Palette.Fg = colText

	events := make(chan core.Event, 128)
	u := &ui{
		th:        th,
		events:    events,
		state:     stateIdle,
		status:    "проверяю, открывается ли Discord…",
		statusCol: colMuted,
	}
	u.ctl = core.NewController(root, events)
	u.presetItems = []string{"Авто - сам подберёт рабочую"}
	u.presetIDs = []string{"auto"}
	for _, p := range core.Presets() {
		u.presetItems = append(u.presetItems, p.Name)
		u.presetIDs = append(u.presetIDs, p.ID)
	}
	u.itemClick = make([]widget.Clickable, len(u.presetItems))
	u.logList.Axis = layout.Vertical
	u.logList.ScrollToEnd = false
	u.autoBox.Value = core.AutostartEnabled()
	return u
}

// ----- события от контроллера -----

func (u *ui) drainEvents() {
	for {
		select {
		case ev := <-u.events:
			switch ev.Kind {
			case "log":
				u.addLog(ev.Text)
			case "status":
				u.status = ev.Text
				u.statusCol = colMuted
			case "run-started":
				u.state = stateOn
				u.status = "Обход работает: " + ev.Text
				u.statusCol = colGreen
				u.addLog("работает: " + ev.Text)
			case "run-failed":
				u.state = stateFailed
				u.status = "Не получилось: " + ev.Text
				u.statusCol = colRed
				u.addLog("не пробилось: " + ev.Text)
			case "died":
				u.state = stateIdle
				u.status = ev.Text
				u.statusCol = colYellow
				u.addLog(ev.Text)
			case "checked":
				if ev.Text == "ok" {
					if u.state == stateOn {
						// и так работаем, ничего не меняем
					} else {
						u.status = "Discord открывается даже без обхода - жать ничего не нужно"
						u.statusCol = colGreen
						if u.state != stateSwitching {
							u.state = stateIdle
						}
					}
				} else if u.state != stateOn && u.state != stateSwitching {
					u.status = "Discord не открывается - жми большую кнопку"
					u.statusCol = colYellow
					u.state = stateIdle
				}
			}
		default:
			return
		}
	}
}

func (u *ui) addLog(s string) {
	line := time.Now().Format("15:04:05") + "  " + s
	u.logs = append(u.logs, line)
	if len(u.logs) > 300 {
		u.logs = u.logs[len(u.logs)-300:]
	}
}

// ----- клики -----

func (u *ui) handleClicks(gtx layout.Context) {
	for u.presetClick.Clicked(gtx) {
		u.presetOpen = !u.presetOpen
	}
	for i := range u.itemClick {
		for u.itemClick[i].Clicked(gtx) {
			u.presetSel = i
			u.presetOpen = false
		}
	}
	if u.checkBtn.Clicked(gtx) {
		u.status = "проверяю связь с Discord…"
		u.statusCol = colMuted
		go func() {
			res := core.CheckDiscord(5 * time.Second)
			line := "проверка: " + res.String()
			kind := "bad"
			if res.OK() {
				kind = "ok"
			}
			u.events <- core.Event{Kind: "log", Text: line}
			u.events <- core.Event{Kind: "checked", Text: kind}
		}()
	}
	if u.logToggle.Clicked(gtx) {
		u.logOpen = !u.logOpen
	}
	if u.autoBox.Update(gtx) {
		if exe, err := os.Executable(); err == nil {
			if err := core.SetAutostart(exe, u.autoBox.Value); err != nil {
				u.addLog("автозапуск не переключился: " + err.Error())
				u.autoBox.Value = !u.autoBox.Value
			} else if u.autoBox.Value {
				u.addLog("буду запускаться вместе с Windows")
			} else {
				u.addLog("автозапуск выключен")
			}
		}
	}
	if u.toggle.Clicked(gtx) {
		u.onToggle()
	}
}

func (u *ui) onToggle() {
	switch u.state {
	case stateOn, stateSwitching:
		u.ctl.Stop()
		u.state = stateIdle
		u.status = "Выключено"
		u.statusCol = colMuted
		u.addLog("остановлено")
	default:
		u.state = stateSwitching
		u.startedAt = time.Now()
		id := u.presetIDs[u.presetSel]
		if id == "auto" {
			u.status = "Подбираю стратегию - обычно до минуты"
		} else {
			u.status = "Запускаю " + u.presetItems[u.presetSel] + "…"
		}
		u.statusCol = colAccent2
		u.addLog("запуск, стратегия: " + u.presetItems[u.presetSel])
		u.ctl.Activate(id)
	}
}

// ----- разметка -----

func (u *ui) layout(gtx layout.Context, version string) layout.Dimensions {
	paint.Fill(gtx.Ops, colBg)
	return layout.UniformInset(unit.Dp(20)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return u.header(gtx, version)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(14)}.Layout),
			layout.Rigid(u.statusCard),
			layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Center.Layout(gtx, u.bigToggle)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
			layout.Rigid(u.presetPicker),
			layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
			layout.Rigid(u.footerRow),
			layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
			layout.Flexed(1, u.logPane),
		)
	})
}

func (u *ui) header(gtx layout.Context, version string) layout.Dimensions {
	return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(u.logo),
		layout.Rigid(layout.Spacer{Width: unit.Dp(12)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(material.H6(u.th, "Discord Фикс").Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					t := material.Caption(u.th, "v"+version)
					t.Color = colMuted
					return t.Layout(gtx)
				}),
			)
		}),
	)
}

func (u *ui) logo(gtx layout.Context) layout.Dimensions {
	size := gtx.Dp(44)
	s := float32(size)
	defer clip.UniformRRect(image.Rect(0, 0, size, size), gtx.Dp(12)).Push(gtx.Ops).Pop()
	paint.Fill(gtx.Ops, colAccent)
	var p clip.Path
	p.Begin(gtx.Ops)
	p.MoveTo(f32.Pt(s*0.55, s*0.16))
	p.LineTo(f32.Pt(s*0.30, s*0.58))
	p.LineTo(f32.Pt(s*0.47, s*0.58))
	p.LineTo(f32.Pt(s*0.40, s*0.86))
	p.LineTo(f32.Pt(s*0.70, s*0.42))
	p.LineTo(f32.Pt(s*0.52, s*0.42))
	p.Close()
	defer clip.Outline{Path: p.End()}.Op().Push(gtx.Ops).Pop()
	paint.Fill(gtx.Ops, colWhite)
	return layout.Dimensions{Size: image.Pt(size, size)}
}

func (u *ui) statusCard(gtx layout.Context) layout.Dimensions {
	return card(gtx, colPanel, func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(unit.Dp(12)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					d := gtx.Dp(10)
					col := u.statusCol
					if u.state == stateSwitching {
						// мягкое мерцание во время подбора
						if time.Since(u.startedAt).Milliseconds()%1000 > 500 {
							col = colMuted
						}
						gtx.Execute(op.InvalidateCmd{})
					}
					defer clip.Ellipse{Max: image.Pt(d, d)}.Push(gtx.Ops).Pop()
					paint.Fill(gtx.Ops, col)
					return layout.Dimensions{Size: image.Pt(d, d)}
				}),
				layout.Rigid(layout.Spacer{Width: unit.Dp(10)}.Layout),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					t := material.Body1(u.th, u.status)
					t.TextSize = unit.Sp(14)
					return t.Layout(gtx)
				}),
			)
		})
	})
}

func (u *ui) bigToggle(gtx layout.Context) layout.Dimensions {
	size := gtx.Dp(124)
	return u.toggle.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		fill, ring := colAccent, colAccent2
		label, icon := "Включить", "play"
		switch u.state {
		case stateOn:
			fill, ring = colGreen, colGreen
			label, icon = "Выключить", "stop"
		case stateSwitching:
			fill, ring = colPanelHi, colAccent2
			label, icon = "отмена", "stop"
			gtx.Execute(op.InvalidateCmd{})
		case stateFailed:
			fill, ring = colAccent, colRed
		}
		if u.toggle.Hovered() && u.state == stateIdle {
			fill = colAccent2
		}
		pt := image.Pt(size, size)

		// кольцо
		func() {
			defer clip.Ellipse{Max: pt}.Push(gtx.Ops).Pop()
			paint.Fill(gtx.Ops, ring)
		}()
		// внутренний круг
		in := gtx.Dp(6)
		func() {
			defer clip.Ellipse{Min: image.Pt(in, in), Max: pt.Sub(image.Pt(in, in))}.Push(gtx.Ops).Pop()
			paint.Fill(gtx.Ops, fill)
		}()

		// иконка по центру верхней половины
		cx, cy := float32(pt.X)/2, float32(pt.Y)*0.40
		if icon == "stop" {
			s := float32(pt.X) * 0.12
			r := image.Rect(int(cx-s), int(cy-s), int(cx+s), int(cy+s))
			func() {
				defer clip.UniformRRect(r, gtx.Dp(4)).Push(gtx.Ops).Pop()
				paint.Fill(gtx.Ops, colWhite)
			}()
		} else {
			s := float32(pt.X) * 0.15
			var p clip.Path
			p.Begin(gtx.Ops)
			p.MoveTo(f32.Pt(cx-s*0.55, cy-s))
			p.LineTo(f32.Pt(cx+s, cy))
			p.LineTo(f32.Pt(cx-s*0.55, cy+s))
			p.Close()
			func() {
				defer clip.Outline{Path: p.End()}.Op().Push(gtx.Ops).Pop()
				paint.Fill(gtx.Ops, colWhite)
			}()
		}

		// подпись под иконкой
		macro := op.Record(gtx.Ops)
		lblDims := func(gtx layout.Context) layout.Dimensions {
			t := material.Body2(u.th, label)
			t.Color = colWhite
			t.TextSize = unit.Sp(12)
			return t.Layout(gtx)
		}(gtx)
		call := macro.Stop()
		off := image.Pt((pt.X-lblDims.Size.X)/2, int(float32(pt.Y)*0.62))
		op.Offset(off).Add(gtx.Ops)
		call.Add(gtx.Ops)
		return layout.Dimensions{Size: pt}
	})
}

func (u *ui) presetPicker(gtx layout.Context) layout.Dimensions {
	return card(gtx, colPanel, func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(unit.Dp(12)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					t := material.Caption(u.th, "СТРАТЕГИЯ ОБХОДА")
					t.Color = colMuted
					return t.Layout(gtx)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return u.presetClick.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						bg := colPanelHi
						if u.presetClick.Hovered() {
							bg = colHover
						}
						return card(gtx, bg, func(gtx layout.Context) layout.Dimensions {
							return layout.UniformInset(unit.Dp(10)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								return layout.Flex{Alignment: layout.Middle, Spacing: layout.SpaceBetween}.Layout(gtx,
									layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
										return material.Body1(u.th, u.presetItems[u.presetSel]).Layout(gtx)
									}),
									layout.Rigid(func(gtx layout.Context) layout.Dimensions {
										t := material.Body1(u.th, "▾")
										t.Color = colMuted
										return t.Layout(gtx)
									}),
								)
							})
						})
					})
				}),
				layout.Rigid(u.presetList),
			)
		})
	})
}

func (u *ui) presetList(gtx layout.Context) layout.Dimensions {
	if !u.presetOpen {
		return layout.Dimensions{}
	}
	children := make([]layout.FlexChild, 0, len(u.presetItems))
	for i, name := range u.presetItems {
		i, name := i, name
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return u.itemClick[i].Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				bg := colPanel
				switch {
				case i == u.presetSel:
					bg = colAccent
				case u.itemClick[i].Hovered():
					bg = colPanelHi
				}
				return card(gtx, bg, func(gtx layout.Context) layout.Dimensions {
					return layout.UniformInset(unit.Dp(8)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return material.Body2(u.th, name).Layout(gtx)
					})
				})
			})
		}))
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
}

func (u *ui) footerRow(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Alignment: layout.Middle, Spacing: layout.SpaceBetween}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			cb := material.CheckBox(u.th, &u.autoBox, "автозапуск с Windows")
			cb.Color = colMuted
			cb.IconColor = colAccent2
			cb.TextSize = unit.Sp(13)
			return cb.Layout(gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			btn := material.Button(u.th, &u.checkBtn, "Проверить связь")
			btn.Background = colPanelHi
			btn.Color = colText
			btn.CornerRadius = unit.Dp(10)
			btn.TextSize = unit.Sp(13)
			btn.Inset = layout.UniformInset(unit.Dp(8))
			return btn.Layout(gtx)
		}),
	)
}

func (u *ui) logPane(gtx layout.Context) layout.Dimensions {
	return card(gtx, colPanel, func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(unit.Dp(10)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return u.logToggle.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						chevron := "▸  Журнал"
						if u.logOpen {
							chevron = "▾  Журнал"
						}
						t := material.Caption(u.th, chevron)
						t.Color = colMuted
						return t.Layout(gtx)
					})
				}),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					if !u.logOpen || len(u.logs) == 0 {
						return layout.Dimensions{}
					}
					n := len(u.logs)
					lst := material.List(u.th, &u.logList)
					return lst.Layout(gtx, n, func(gtx layout.Context, i int) layout.Dimensions {
						t := material.Caption(u.th, u.logs[n-1-i])
						t.Color = colMuted
						return layout.UniformInset(unit.Dp(1)).Layout(gtx, t.Layout)
					})
				}),
			)
		})
	})
}

// card рисует скруглённую подложку под контентом.
func card(gtx layout.Context, bg color.NRGBA, content layout.Widget) layout.Dimensions {
	macro := op.Record(gtx.Ops)
	dims := content(gtx)
	call := macro.Stop()
	defer clip.UniformRRect(image.Rectangle{Max: dims.Size}, gtx.Dp(10)).Push(gtx.Ops).Pop()
	paint.Fill(gtx.Ops, bg)
	call.Add(gtx.Ops)
	return dims
}

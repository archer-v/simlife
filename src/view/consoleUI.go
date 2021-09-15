package view

import (
	"bytes"
	"fmt"
	"github.com/jroimartin/gocui"
	"github.com/logrusorgru/aurora"
	"log"
	"simlife/src/universe"
	"strings"
	"time"
)

type keyBindings struct {
	key      interface{}
	name     string
	descr    string
	handler  func(v *gocui.View) error
	viewName string
}

type ConsoleUI struct {
	u universe.Universe
	g *gocui.Gui
	k []keyBindings
	/*
		tProgressSpinner *pterm.SpinnerPrinter
	*/
	liveFiller string
	deadFiller string
}

var (
	runningStateDescr = map[universe.RunningState]string{
		universe.RUNNING_STATE_MANUAL:   aurora.Colorize("waiting", aurora.BlueFg).String(),
		universe.RUNNING_STATE_STEP:     "do the step",
		universe.RUNNING_STATE_RUN:      aurora.Colorize("running", aurora.CyanFg).String(),
		universe.RUNNING_STATE_FINISHED: aurora.Colorize("finished", aurora.RedFg).String(),
	}
)

func NewViewTerminal() *ConsoleUI {

	var err error
	t := ConsoleUI{
		liveFiller: aurora.Green("█").BgBrightGreen().String(),
		deadFiller: "░",
	}

	t.g, err = gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}

	t.g.Mouse = true
	//aurora.Green("The field size is larger than the viewing area").BgBlack().String())
	t.k = []keyBindings{
		{gocui.KeyCtrlC,
			"^C",
			"Exit",
			t.cmdQuit,
			""},
		{'n',
			"N",
			"Next step",
			t.cmdNextRound,
			""},
		{'r',
			"R",
			"Run",
			t.cmdRun,
			""},
		{'s',
			"S",
			"Stop",
			t.cmdStop,
			""},
		{'c',
			"C",
			"Clear",
			t.cmdClear,
			""},
		{'w',
			"W",
			"Settle with random",
			t.cmdSettleWithRandom,
			""},
		{gocui.MouseLeft,
			"MOUSE",
			"Settle the cell",
			t.cmdMouseClick,
			"battlefield"},
	}
	t.g.SetManagerFunc(t.layout)

	t.initKeyBindings(t.k)

	return &t
}

func (t *ConsoleUI) initKeyBindings(k []keyBindings) {
	for _, kb := range k {
		h := kb.handler
		if err := t.g.SetKeybinding(kb.viewName, kb.key, gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error { return h(view) }); err != nil {
			log.Panicln(err)
		}
	}
}

func (t *ConsoleUI) Register(u *universe.BaseUniverse) {
	t.u = u
}

func (t *ConsoleUI) Start() {
	if err := t.g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
	t.g.Close()
}

func (t *ConsoleUI) Refresh() {
	t.renderField(t.u.Area())
	t.renderConfiguration()
	t.renderStatus()
}

func (t *ConsoleUI) renderField(a universe.Area) {

	t.g.Update(func(g *gocui.Gui) error {
		v, e := g.View("battlefield")
		if e != nil {
			return e
		}
		//the entire field is redrawing at once now
		//this terminal driver allows to redraw only changed chars
		//there is an opportunity to speed up with a selective redraw
		v.Clear()

		crop := false
		maxW, maxH := v.Size()
		if a.Width > maxW || a.Height > maxH {
			crop = true
		}

		var b bytes.Buffer

		for i, l := range a.Entities {
			//discard the data outside the view area
			if i >= maxH {
				break
			}
			//line feed char
			if i != 0 {
				b.WriteByte(10)
			}
			if crop && i == (maxH-1) {
				b.WriteString(aurora.Red("The field size is larger than the viewing area").BgBlack().String())
				break
			}
			for j, e := range l {
				if j >= maxW {
					break
				}
				if bool(e) == true {
					b.WriteString(t.liveFiller)
				} else {
					b.WriteString(t.deadFiller)
				}
			}
		}
		_, _ = fmt.Fprint(v, b.String())
		return nil
	})
}

func (t *ConsoleUI) renderStatus() {
	s := t.u.Status()
	t.g.Update(func(g *gocui.Gui) error {
		if v, e := t.g.View("status"); e == nil {
			v.Clear()
			_, _ = fmt.Fprintln(v, t.renderProp("Step", "%v", s.IterationNum))
			_, _ = fmt.Fprintln(v, t.renderProp("Live Cells", "%v", s.LiveCells))
			_, _ = fmt.Fprintln(v, t.renderProp("Evaluation time", "%v", s.IterationTime.Round(time.Microsecond)))
			_, _ = fmt.Fprintln(v, t.renderProp("Mode", "%v", runningStateDescr[s.RunningMode]))
		}
		return nil
	})
}

func (t *ConsoleUI) renderConfiguration() {
	//it needs to call Update when calls from goroutine
	t.g.Update(func(g *gocui.Gui) error {
		c := t.u.Options()
		if v, e := g.View("configuration"); e == nil {
			v.Clear()
			_, _ = fmt.Fprintln(v, t.renderProp("Dimension", "%v x %v", c.Width, c.Height))
			_, _ = fmt.Fprintln(v, t.renderProp("Interval", "%v", c.Interval))
			_, _ = fmt.Fprintln(v, t.renderProp("Iterations", "%v steps", c.MaxSteps))
		}
		return nil
	})
}

func (t *ConsoleUI) renderProp(name string, valueformat string, values ...interface{}) string {
	return fmt.Sprintf(" "+aurora.Colorize(name, aurora.GreenFg).String()+": "+valueformat, values...)
}

func (t *ConsoleUI) layout(g *gocui.Gui) error {

	maxX, maxY := g.Size()
	leftColumnWidth := 28
	minWindowHeight := 20

	if maxY < minWindowHeight {
		if _, err := t.headerLayout(g, maxY, "Terminal height too small"); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
		}
		_ = g.DeleteView("configuration")
		_ = g.DeleteView("status")
		_ = g.DeleteView("battlefield")
		return nil

	} else {
		if _, err := t.headerLayout(g, 3, "This is \"The Life\" game simulation"); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
		}
	}

	if v, err := g.SetView("configuration", 0, 3, leftColumnWidth, 3+(maxY-5-3)/2); err != nil {
		if err != gocui.ErrUnknownView || v == nil {
			return err
		}
		v.Title = "Configuration"
		v.Frame = true
		t.renderConfiguration()
	}

	if v, err := g.SetView("status", 0, 3+(maxY-5-3)/2+1, leftColumnWidth, maxY-5); err != nil {
		if err != gocui.ErrUnknownView || v == nil {
			return err
		}
		v.Title = "Status"
		v.Frame = true
		t.renderStatus()
	}

	if v, err := g.SetView("battlefield", leftColumnWidth+1, 3, maxX-1, maxY-5); err != nil {
		if err != gocui.ErrUnknownView || v == nil {
			return err
		}
		v.Title = "Battle Field"
		v.Frame = true
		t.renderField(t.u.Area())
	} else {
		t.renderField(t.u.Area())
	}

	if v, err := g.SetView("help", -1, maxY-5, maxX, maxY-3); err != nil {
		if err != gocui.ErrUnknownView || v == nil {
			return err
		}
		v.Frame = false
		b := bytes.Buffer{}
		b.WriteString("KEYBINDINGS: ")
		for i, k := range t.k {
			if i != 0 {
				b.WriteString(", ")
			}
			b.WriteString(aurora.Green(k.name).String())
			b.WriteString(": ")
			b.WriteString(k.descr)
		}
		_, _ = fmt.Fprintln(v, b.String())
	}

	return nil
}

func (t *ConsoleUI) headerLayout(g *gocui.Gui, height int, text string) (v *gocui.View, err error) {
	maxX, _ := g.Size()
	if v, err = g.SetView("header", -1, -1, maxX+1, height); err != nil {
		if err == gocui.ErrUnknownView && v != nil {
			v.Frame = false
			v.BgColor = gocui.ColorCyan
			v.FgColor = gocui.ColorBlack
		}
	}
	if v != nil {
		v.Clear()
		if maxX < len(text) {
			panic(fmt.Sprintf("Terminal width is too small: %v", maxX))
		}
		_, _ = fmt.Fprintln(v, strings.Repeat("\n", height/2+1)+strings.Repeat(" ", (maxX-len(text))/2)+text)
	}
	return
}

func (t *ConsoleUI) cmdQuit(_ *gocui.View) error {
	return gocui.ErrQuit
}

func (t *ConsoleUI) cmdNextRound(_ *gocui.View) error {
	t.u.Step()
	return nil
}

func (t *ConsoleUI) cmdRun(_ *gocui.View) error {
	t.u.Run()
	return nil
}

func (t *ConsoleUI) cmdStop(_ *gocui.View) error {
	t.u.Stop()
	return nil
}

func (t *ConsoleUI) cmdClear(_ *gocui.View) error {
	t.u.Clear()
	return nil
}

func (t *ConsoleUI) cmdSettleWithRandom(_ *gocui.View) error {
	t.u.SettleWithRandomData()
	return nil
}

func (t *ConsoleUI) cmdMouseClick(v *gocui.View) error {
	cx, cy := v.Cursor()
	t.u.InverseCell(cx, cy)
	return nil
}

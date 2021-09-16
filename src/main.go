package main

import (
	"fmt"
	"github.com/integrii/flaggy"
	"simlife/src/universe"
	"simlife/src/view"
	"strings"
	"time"
)

var (
	testSample = [][]int{
		{1, 1}, {1, 2},
		{2, 1}, {2, 2},
		{3, 3},
		{4, 2},
		{4, 3},
		{5, 3},
	}

	engines = map[string]func(o *universe.Options, stateCh chan universe.Status) universe.Universe{
		"base": func(o *universe.Options, stateCh chan universe.Status) universe.Universe {
			return universe.NewBaseUniverse(o, stateCh)
		},
		"simple":        universe.NewSimpleUniverse,
		"smallBuff":     universe.NewSmallBuffUniverse,
		"multithreaded": universe.NewMultithreadedUniverse,
	}
)

type EnvOptions struct {
	interactive bool
	randomData  bool
	engine      string
}

func main() {
	eo, uo := initOptions()

	var stateCh chan universe.Status

	if !eo.interactive {
		stateCh = make(chan universe.Status, 10) //the buffered channel to getting the universe status
	}

	u := engines[eo.engine](uo, stateCh)

	u.AddTemplate(
		universe.Template{
			"testSample1",
			"the test sample with 3 stable patterns",
			testSample,
		})

	if eo.randomData {
		u.SettleWithRandomData()
	} else {
		u.SettleTemplate("testSample1")
	}

	if eo.interactive {
		v := view.NewConsoleUI()
		u.RegisterViewer(v)
		v.Start()
		u.Close()
	} else {
		v := view.NewConsoleOut()
		u.RegisterViewer(v)
		v.Start()
		u.Run()
		for {
			st := <-stateCh
			if st.RunningMode == universe.RUNNING_STATE_FINISHED {
				break
			}
		}
		u.Close()
		close(stateCh)
		//waiting for all final output printing
		time.Sleep(time.Millisecond * 200)
	}

}

func initOptions() (eo *EnvOptions, uo *universe.Options) {

	uo = &universe.DefaultUniverseOptions
	engineNames := make([]string, 0, len(engines))
	for k := range engines {
		engineNames = append(engineNames, k)
	}
	eo = &EnvOptions{engine: "base"}
	flaggy.DefaultParser.ShowHelpOnUnexpected = true

	runMode := flaggy.NewSubcommand("run")
	runMode.Description = "Run simulation with console output"

	uiMode := flaggy.NewSubcommand("ui")
	uiMode.Description = "Run with console UI"

	flaggy.AttachSubcommand(runMode, 1)
	flaggy.AttachSubcommand(uiMode, 1)

	flaggy.Int(&uo.Width, "x", "width", "Width of a simulation field")
	flaggy.Int(&uo.Height, "y", "height", "Height of a simulation field")
	flaggy.Duration(&uo.Interval, "i", "interval", "Simulation speed (interval between the steps) in format the number with 'ms' suffix, for example 150ms")
	flaggy.Int(&uo.MaxSteps, "s", "maxSteps", "Limit the simulation to maxSteps")
	flaggy.Bool(&eo.randomData, "r", "random", "Settle with random data")
	flaggy.String(&eo.engine, "e", "engine", "Engine to use ["+strings.Join(engineNames, "|")+"]")

	flaggy.Parse()

	eo.interactive = uiMode.Used
	if !uiMode.Used && !runMode.Used {
		flaggy.ShowHelpAndExit("Specify the running mode \"run\" or \"ui\"")
	}

	_, ok := engines[eo.engine]
	if !ok {
		flaggy.ShowHelpAndExit("unknown engine")
	}

	if eo.engine == "multithreaded" {
		fmt.Println("\nTo use multi-threading effectively set \"interval\" value to 0")
	}

	return
}

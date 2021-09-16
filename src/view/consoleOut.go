package view

import (
	"fmt"
	"simlife/src/universe"
	"sort"
	"time"
)

type ConsoleOut struct {
	u         universe.Universe
	startTime time.Time
}

func NewConsoleOut() *ConsoleOut {
	return &ConsoleOut{}
}

func (c *ConsoleOut) Refresh() {
	st := c.u.Status()
	if st.RunningMode == universe.RUNNING_STATE_FINISHED {
		totalTime := time.Since(c.startTime).Round(time.Millisecond)
		resultData := map[string]interface{}{
			"Last iteration": st.IterationNum,
			"Total time":     totalTime,
			"Live cells":     st.LiveCells,
		}
		fmt.Println("\nFinished:")
		c.printHashData(resultData)
	} else if st.RunningMode == universe.RUNNING_STATE_RUN {
		if st.IterationNum%10 == 0 {
			fmt.Printf("  Iterations done: %v\n", st.IterationNum)
		}
	}
}

func (c *ConsoleOut) Register(u *universe.BaseUniverse) {
	c.u = u
	o := c.u.Options()
	fmt.Println("Running configuration:")
	fmt.Printf("  Dimension: %v x %v\n", o.Width, o.Height)
	fmt.Printf("  Interval: %v\n", o.Interval)
	fmt.Printf("  Max iterations: %v steps\n", o.MaxSteps)
	c.printHashData(o.Advanced)
}

func (c *ConsoleOut) Start() {
	c.startTime = time.Now()
	fmt.Println("\nSimulation started...")
}

func (c *ConsoleOut) printHashData(d map[string]interface{}) {
	propNames := make([]string, 0, len(d))
	for k := range d {
		propNames = append(propNames, k)
	}
	sort.Strings(propNames)
	for _, propName := range propNames {
		fmt.Printf("  %s: %v\n", propName, d[propName])
	}
}

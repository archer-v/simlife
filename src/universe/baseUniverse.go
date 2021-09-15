package universe

import (
	"math/rand"
	"sync"
	"time"
)

type Cell bool

type Area struct {
	Width    int
	Height   int
	Entities [][]Cell
	sync.Mutex
}

//universe options
type Options struct {
	Width           int
	Height          int
	Interval        time.Duration
	MaxSteps        int
	MaxSkippedTicks int
}

//universe state
type Status struct {
	IterationNum  int
	RunningMode   RunningState
	LiveCells     int
	IterationTime time.Duration
}

//viewer
type Viewer interface {
	Refresh()
	Register(u *BaseUniverse)
	Start()
}

//universe seeding template
type Template struct {
	Name        string  //template name
	Descr       string  //template descr
	Coordinates [][]int //array of [x,y] coordinates
}

type RunningState int

const (
	DEF_SIMULATION_INTERVAL = time.Millisecond * 100
	DEF_MAX_STEPS           = 1000
	DEF_WIDTH               = 40
	DEF_HEIGHT              = 15
	DEF_MAX_SKIPPED_TICKS   = 5
)

const (
	RUNNING_STATE_MANUAL   = 0x0
	RUNNING_STATE_STEP     = 0x1
	RUNNING_STATE_RUN      = 0x2
	RUNNING_STATE_FINISHED = 0x3
)

var DefaultUniverseOptions = Options{
	Width:           DEF_WIDTH,
	Height:          DEF_HEIGHT,
	Interval:        DEF_SIMULATION_INTERVAL,
	MaxSteps:        DEF_MAX_STEPS,
	MaxSkippedTicks: DEF_MAX_SKIPPED_TICKS,
}

type BaseUniverse struct {
	options Options
	state   struct {
		Status
		sync.Mutex
	}
	stateCh       chan Status
	area          Area
	views         []Viewer
	templates     map[string]Template
	controlCh     chan func()
	closeCh       chan bool
	nextIteration func() (hasLiveEnitities bool, changed bool)
}

func NewBaseUniverse(o *Options, stateCh chan Status) *BaseUniverse {
	if o == nil {
		o = &DefaultUniverseOptions
	}
	u := BaseUniverse{
		options:   *o,
		controlCh: make(chan func(), 1),
		closeCh:   make(chan bool, 1),
		stateCh:   stateCh,
		templates: map[string]Template{},
	}
	//nextIteration can be implemented by successor
	u.nextIteration = u._nextIteration

	u.area = u.createArea(o.Width, o.Height)
	u.refreshView()
	go u.mainLoop()
	return &u
}

//add the seeding template
//the universe can be populated with this template by call SettleTemplate
func (u *BaseUniverse) AddTemplate(tmpl Template) {
	u.templates[tmpl.Name] = tmpl
}

//settle the universe with data
//vc - array of x,y coordinates
func (u *BaseUniverse) Settle(vc [][]int) {
	u.area.Lock()
	u.settle(vc, Cell(true))
	u.area.Unlock()
	u.refreshView()
}

//populate the universe with the seeding template
func (u *BaseUniverse) SettleTemplate(name string) {
	tmpl, ok := u.templates[name]
	if !ok {
		return
	}
	u.area.Lock()
	u.settle(tmpl.Coordinates, Cell(true))
	u.area.Unlock()
	u.state.LiveCells = u.liveCells()
	u.refreshView()
}

//populate the universe with random data
func (u *BaseUniverse) SettleWithRandomData() {
	if u.state.RunningMode == RUNNING_STATE_MANUAL || u.state.RunningMode == RUNNING_STATE_FINISHED {
		u.controlCh <- u.clear
		u.controlCh <- func() {
			u.area.Lock()
			for i := 0; i < u.area.Width*u.area.Height; i++ {
				u.settle([][]int{{rand.Intn(u.area.Width), rand.Intn(u.area.Height)}}, Cell(true))
			}
			u.area.Unlock()
			u.state.LiveCells = u.liveCells()
			u.refreshView()
		}
	}
}

//inverse the cell state at point x, y
func (u *BaseUniverse) InverseCell(x int, y int) {
	if x >= u.area.Width || y >= u.area.Height {
		return
	}
	u.area.Lock()
	u.area.Entities[y][x] = !u.area.Entities[y][x]
	u.area.Unlock()
	u.refreshView()
}

//register the viewer - the universe will call the viewer when the state is changed
func (u *BaseUniverse) RegisterViewer(v Viewer) {
	u.views = append(u.views, v)
	v.Register(u)
}

//return the channel with the universe's status updates
func (u *BaseUniverse) StateCh() chan Status {
	return u.stateCh
}

func (u *BaseUniverse) Status() Status {
	return u.state.Status
}

func (u *BaseUniverse) Options() Options {
	return u.options
}

func (u *BaseUniverse) Area() Area {
	return u.area
}

//run the universe simulation, returns immediately
func (u *BaseUniverse) Run() {
	u.controlCh <- u.run
}

//stop the universe sumulation, returns immediately
//the Status will be written the stateCh on finish
func (u *BaseUniverse) Stop() {
	u.controlCh <- u.stop
}

//do one simulation step, returns immediately
//the Status will be written to the stateCh on start and on finish
func (u *BaseUniverse) Step() {
	u.controlCh <- u.step
}

//clear the universe (kill all cells), returns immediately
//the Status will be written to the stateCh on finish
func (u *BaseUniverse) Clear() {
	u.controlCh <- u.clear
}

//stop the main loop, close the channels, returns immediately
func (u *BaseUniverse) Close() {
	u.closeCh <- true
}

//main loop, should start as a gouroutine
//waiting for command and execute
func (u *BaseUniverse) mainLoop() {
	var c = false
	for !c {
		select {
		case cmd := <-u.controlCh:
			cmd()
		case c = <-u.closeCh:

		}
	}
	close(u.closeCh)
	close(u.controlCh)
}

func (u *BaseUniverse) settle(vc [][]int, entity Cell) {
	for _, v := range vc {
		if v[0] >= u.area.Width || v[1] >= u.area.Height {
			continue
		}
		u.area.Entities[v[1]][v[0]] = entity
	}
}

//calculate the count of live cells
func (u *BaseUniverse) liveCells() int {
	liveCells := 0
	u.area.Lock()
	defer u.area.Unlock()
	u.walkArea(func(x int, y int, e Cell) {
		if bool(e) {
			liveCells++
		}
	})
	return liveCells
}

func (u *BaseUniverse) switchRunningState(to RunningState) {
	u.state.Lock()
	u.state.RunningMode = to
	st := u.state.Status
	u.state.Unlock()
	if u.stateCh != nil {
		u.stateCh <- st
	}
}

//start the universe simulation
//simulation will stop on Stop() calling or when the boundary conditions are reached
func (u *BaseUniverse) run() {
	go func() {
		u.switchRunningState(RUNNING_STATE_RUN)
		skipped := 0
		done := make(chan bool)
		defer close(done)
		for {
			mode := u.state.RunningMode
			if mode != RUNNING_STATE_RUN && mode != RUNNING_STATE_STEP {
				break
			}
			if skipped > u.options.MaxSkippedTicks {
				u.switchRunningState(RUNNING_STATE_FINISHED)
				//todo write the warning message
				break
			}
			//skip the tick if the universe is still in the calculation mode
			if mode != RUNNING_STATE_STEP {
				skipped = 0
				u.controlCh <- func() {
					u.step()
					done <- true
				}
				<-done
			} else {
				skipped++
			}
			if u.options.Interval > 0 {
				time.Sleep(u.options.Interval)
			}
		}

	}()
}

func (u *BaseUniverse) stop() {
	if u.state.RunningMode == RUNNING_STATE_RUN {
		u.switchRunningState(RUNNING_STATE_MANUAL)
	}
}

func (u *BaseUniverse) step() {

	finished := false
	rm := u.state.RunningMode
	maxIter := u.options.MaxSteps
	u.state.IterationNum++
	defer func() {
		if finished {
			u.switchRunningState(RUNNING_STATE_FINISHED)
		} else {
			u.switchRunningState(rm)
		}
		u.refreshView()
	}()

	if maxIter != 0 && u.state.IterationNum >= maxIter {
		finished = true
		return
	}
	u.switchRunningState(RUNNING_STATE_STEP)
	isAlive, changed := u.nextIteration()
	if !isAlive || !changed {
		finished = true
	}
	return
}

func (u *BaseUniverse) clear() {
	u.state.Lock()
	u.area.Lock()

	u.state.IterationNum = 0
	u.state.LiveCells = 0
	u.walkArea(func(x int, y int, e Cell) {
		u.area.Entities[y][x] = false
	})
	u.state.RunningMode = RUNNING_STATE_MANUAL
	u.area.Unlock()
	u.state.Unlock()
	u.switchRunningState(RUNNING_STATE_MANUAL)
	u.refreshView()

}

//do one simulation cycle
//walking the area and calculating the next state for the each cell
//the simplest implementation: creates the new area buffer with full size on each call
//All cells state is calculated to the new buffer and then this buffer is stored to the universe replacing the old one (by replacing the area pointer)
func (u *BaseUniverse) _nextIteration() (hasLiveEnitities bool, changed bool) {
	u.area.Lock()
	defer u.area.Unlock()
	start := time.Now()
	a := u.createArea(u.area.Width, u.area.Height)
	liveCellls := 0
	u.walkArea(func(x int, y int, e Cell) {
		nextState := u.cellNextState(x, y)
		hasLiveEnitities = hasLiveEnitities || nextState
		changed = changed || nextState != bool(e)
		a.Entities[y][x] = Cell(nextState)
		if nextState {
			liveCellls++
		}
	})
	u.area.Entities = a.Entities
	u.state.LiveCells = liveCellls
	u.state.IterationTime = time.Since(start)
	return
}

//allocate the area and return the pointer
func (u *BaseUniverse) createArea(width int, height int) Area {

	area := Area{Width: width, Height: height, Entities: make([][]Cell, height)}
	b := make([]Cell, width*height)
	for i := range area.Entities {
		start := width * i
		area.Entities[i] = b[start : start+width : start+width]
	}
	return area
}

//walk the entire area and call the cb function for each cell
func (u *BaseUniverse) walkArea(cb func(x int, y int, entity Cell)) {
	for y := range u.area.Entities {
		for x := range u.area.Entities[y] {
			cb(x, y, u.area.Entities[y][x])
		}
	}
}

//calculate the next state for the cell
func (u *BaseUniverse) cellNextState(x int, y int) (live bool) {
	//calculate neighbors
	liveNeighbours := 0
	area := u.area
	for i := -1; i < 2; i++ {
		for j := -1; j < 2; j++ {
			//skip my position
			if i == 0 && j == 0 {
				continue
			}
			nx := x + i
			ny := y + j
			//skip coordinates outside the area
			if nx < 0 || ny < 0 || nx >= area.Width || ny >= area.Height {
				continue
			}
			if area.Entities[ny][nx] {
				liveNeighbours++
			}
		}
	}

	if liveNeighbours < 2 {
		return false
	} else if liveNeighbours > 3 {
		return false
	} else if liveNeighbours == 3 {
		return true
	} else if liveNeighbours == 2 && area.Entities[y][x] {
		return true
	}

	return false
}

//calls Refresh for all registered views
func (u *BaseUniverse) refreshView() {
	for _, v := range u.views {
		v.Refresh()
	}
}

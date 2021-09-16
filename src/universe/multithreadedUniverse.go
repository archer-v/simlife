package universe

import (
	"sync"
	"time"
)

/*
	Universe implementation with multithreaded computation algorithm
	the field is splitted into the areas each of which is computed by individual goroutine
*/

const (
	DEF_WORKERS             = 10
	DEF_MIN_ROWS_PER_WORKER = 3
)

type MultithreadedUniverse struct {
	*BaseUniverse
	workers   int
	workAreas []workArea
}

type workArea struct {
	x1        int
	y1        int
	x2        int
	y2        int
	tmpBuff   Area
	liveCells int
	changed   bool
}

func newWorkArea(x1 int, y1 int, x2 int, y2 int) workArea {
	return workArea{
		x1,
		y1,
		x2,
		y2,
		createArea(x2-x1+1, y2-y1+1),
		0,
		false,
	}
}

func NewMultithreadedUniverse(o *Options, stateCh chan Status) Universe {
	mu := MultithreadedUniverse{BaseUniverse: NewBaseUniverse(o, stateCh)}
	//redefine the nextIteration
	mu.BaseUniverse.nextIteration = mu.nextIteration

	mu.workers = DEF_WORKERS
	linesPerWorker := mu.area.Height / mu.workers
	if linesPerWorker < DEF_MIN_ROWS_PER_WORKER {
		linesPerWorker = DEF_MIN_ROWS_PER_WORKER
	} else if linesPerWorker*mu.workers < mu.area.Height {
		linesPerWorker++
	}
	mu.workAreas = make([]workArea, 0, mu.workers)
	for y1 := 0; y1 < mu.area.Height; y1 += linesPerWorker {
		y2 := y1 + linesPerWorker - 1
		if y2 > mu.area.Height-1 {
			y2 = mu.area.Height - 1
		}
		mu.workAreas = append(mu.workAreas, newWorkArea(0, y1, mu.area.Width-1, y2))
	}
	mu.workers = len(mu.workAreas)
	mu.options.Advanced["engine"] = "multithreaded"
	mu.options.Advanced["Workers"] = mu.workers
	mu.options.Advanced["Rows per worker"] = linesPerWorker
	return &mu
}

func (mu *MultithreadedUniverse) nextIteration() (hasLiveEntities bool, changed bool) {
	mu.area.Lock()
	defer mu.area.Unlock()
	start := time.Now()
	liveCells := 0
	var waitGroup sync.WaitGroup
	for i := range mu.workAreas {
		workArea := &mu.workAreas[i]
		waitGroup.Add(1)
		go func() {
			mu.calcArea(workArea)
			waitGroup.Done()
		}()
	}
	waitGroup.Wait()
	for _, workArea := range mu.workAreas {
		mu.writeArea(workArea)
		liveCells += workArea.liveCells
		changed = changed || workArea.changed
	}
	mu.state.LiveCells = liveCells
	mu.state.IterationTime = time.Since(start)
	hasLiveEntities = liveCells > 0
	return
}

func (mu *MultithreadedUniverse) writeArea(wa workArea) {
	for y := range wa.tmpBuff.Entities {
		copy(mu.area.Entities[wa.y1+y][wa.x1:wa.x2], wa.tmpBuff.Entities[y])
	}
}

func (mu *MultithreadedUniverse) calcArea(wa *workArea) {
	wa.liveCells = 0
	wa.changed = false
	for y := wa.y1; y <= wa.y2; y++ {
		for x := wa.x1; x <= wa.x2; x++ {
			nextState := mu.cellNextState(x, y)
			if nextState {
				wa.liveCells++
			}
			wa.changed = wa.changed || nextState != bool(mu.area.Entities[y][x])
			wa.tmpBuff.Entities[y-wa.y1][x-wa.x1] = Cell(nextState)
		}
	}
}

package universe

import "time"

/*
	Simple Universe implementation with two buffers
	All cells state is calculated to the new buffer and then this buffer data is copied to the universe replacing the old one
*/
type SimpleUniverse struct {
	*BaseUniverse
	tmpBuff Area
}

func NewSimpleUniverse(o *Options, stateCh chan Status) Universe {
	su := SimpleUniverse{BaseUniverse: NewBaseUniverse(o, stateCh)}
	//redefine the nextIteration
	su.BaseUniverse.nextIteration = su.nextIteration
	su.tmpBuff = createArea(su.area.Width, su.area.Height)
	su.options.Advanced["engine"] = "simple"
	return &su
}

func (su *SimpleUniverse) nextIteration() (hasLiveEnitities bool, changed bool) {
	su.area.Lock()
	defer su.area.Unlock()
	start := time.Now()
	liveCells := 0
	for y := range su.area.Entities {
		for x := range su.area.Entities[y] {
			nextState := su.cellNextState(x, y)
			if nextState {
				liveCells++
			}
			changed = changed || nextState != bool(su.area.Entities[y][x])
			su.tmpBuff.Entities[y][x] = Cell(nextState)
		}
	}

	for y := range su.area.Entities {
		copy(su.area.Entities[y], su.tmpBuff.Entities[y])
	}

	su.state.LiveCells = liveCells
	su.state.IterationTime = time.Since(start)
	hasLiveEnitities = liveCells > 0
	return
}

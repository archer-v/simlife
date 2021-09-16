package universe

import "time"

/*
	Universe implementation with buffers optimization
	nextIteration uses small buffer to store the current and previous lines only.
    the first line of this buffer is copied to the main buffer as calculating moves to the next line
	also here we have small optimization to reduce memory copying
*/

type SmallBuffUniverse struct {
	*BaseUniverse
	tmpBuff Area
}

func NewSmallBuffUniverse(o *Options, stateCh chan Status) Universe {
	su := SmallBuffUniverse{BaseUniverse: NewBaseUniverse(o, stateCh)}
	//redefine the nextIteration
	su.BaseUniverse.nextIteration = su.nextIteration
	su.tmpBuff = createArea(su.area.Width, 2)
	su.options.Advanced["engine"] = "smallBuff"
	return &su
}

func (su *SmallBuffUniverse) nextIteration() (hasLiveEnitities bool, changed bool) {
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
			su.tmpBuff.Entities[1][x] = Cell(nextState)
		}
		if y-1 >= 0 {
			copy(su.area.Entities[y-1], su.tmpBuff.Entities[0])
		}
		su.tmpBuff.Entities[0], su.tmpBuff.Entities[1] = su.tmpBuff.Entities[1], su.tmpBuff.Entities[0]
	}
	copy(su.area.Entities[su.area.Height-1], su.tmpBuff.Entities[0])
	su.state.LiveCells = liveCells
	su.state.IterationTime = time.Since(start)
	hasLiveEnitities = liveCells > 0
	return
}

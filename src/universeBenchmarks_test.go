package main

import (
	"simlife/src/universe"
	"testing"
)

var (
	testTemplate = universe.Template{"ts1", "", [][]int{{1, 1}, {1, 2}, {2, 1}, {2, 2}, {3, 3}, {4, 2}, {4, 3}, {5, 3}}}
)

func universeStep(u universe.Universe, b *testing.B) {
	u.AddTemplate(testTemplate)
	stateCh := u.StateCh()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		u.Clear()
		<-stateCh //wait for finish
		u.SettleTemplate("ts1")
		b.StartTimer()
		u.Step()
		for {
			st := <-stateCh
			if st.RunningMode == universe.RUNNING_STATE_MANUAL {
				break
			}
		}
	}
	u.Close()
	close(stateCh)
}

func universeRun(u universe.Universe, b *testing.B) {
	u.AddTemplate(testTemplate)
	stateCh := u.StateCh()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		u.Clear()
		<-stateCh //wait for finish
		u.SettleTemplate("ts1")
		b.StartTimer()
		u.Run()
		for {
			//fmt.Printf("waiting iter: %v\n", i)
			st := <-stateCh
			//fmt.Printf("got iter: %v, state: %v, live cells: %v\n", i, st.RunningMode, st.LiveCells)
			if st.RunningMode == universe.RUNNING_STATE_FINISHED {
				break
			}
		}
	}
	u.Close()
	close(stateCh)
}

func newStateCh() chan universe.Status {
	return make(chan universe.Status, 10)
}

func newUniverseOptions() *universe.Options {
	o := universe.DefaultUniverseOptions
	o.Interval = 0
	return &o
}

func BenchmarkBaseUniverse_Step(b *testing.B) {
	//b.Skip()
	u := universe.NewBaseUniverse(newUniverseOptions(), newStateCh())
	universeStep(u, b)
}

func BenchmarkSimpleUniverse_Step(b *testing.B) {
	//b.Skip()
	u := universe.NewSimpleUniverse(newUniverseOptions(), newStateCh())
	universeStep(u, b)
}

func BenchmarkSmallBuffUniverse_Step(b *testing.B) {
	//b.Skip()
	u := universe.NewSmallBuffUniverse(newUniverseOptions(), newStateCh())
	universeStep(u, b)
}

func BenchmarkBaseUniverse_Run(b *testing.B) {

	u := universe.NewBaseUniverse(newUniverseOptions(), newStateCh())
	universeRun(u, b)
}

func BenchmarkSimpleUniverse_Run(b *testing.B) {
	//b.Skip()
	u := universe.NewSimpleUniverse(newUniverseOptions(), newStateCh())
	universeRun(u, b)
}

func BenchmarkSmallBuffUniverse_Run(b *testing.B) {
	//b.Skip()
	u := universe.NewSmallBuffUniverse(newUniverseOptions(), newStateCh())
	universeRun(u, b)
}

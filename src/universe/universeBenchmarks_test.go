package universe

import (
	"sort"
	"testing"
)

var (
	testTemplate = Template{"ts1", "", [][]int{{1, 1}, {1, 2}, {2, 1}, {2, 2}, {3, 3}, {4, 2}, {4, 3}, {5, 3}}}

	engines = map[string]func(o *Options, stateCh chan Status) Universe{
		"base": func(o *Options, stateCh chan Status) Universe {
			return NewBaseUniverse(o, stateCh)
		},
		"simple":        NewSimpleUniverse,
		"smallBuff":     NewSmallBuffUniverse,
		"multithreaded": NewMultithreadedUniverse,
	}
)

const (
	width  = 200
	height = 200
)

func universeStep(u Universe, b *testing.B) {
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
			if st.RunningMode == RunningStateManual {
				break
			}
		}
	}
	u.Close()
	close(stateCh)
}

func universeRun(u Universe, b *testing.B) {
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
			if st.RunningMode == RunningStateFinished {
				break
			}
		}
	}
	u.Close()
	close(stateCh)
}

func newStateCh() chan Status {
	return make(chan Status, 10)
}

func newUniverseOptions() *Options {
	o := DefaultUniverseOptions
	o.Interval = 0
	o.Width = width
	o.Height = height
	return &o
}

func engineNames() (engineNames []string) {
	engineNames = make([]string, 0, len(engines))
	for k := range engines {
		engineNames = append(engineNames, k)
	}
	sort.Strings(engineNames)
	return
}

func Benchmark_Step(b *testing.B) {
	for _, e := range engineNames() {
		b.Run(e, func(b *testing.B) {
			u := engines[e](newUniverseOptions(), newStateCh())
			universeStep(u, b)
		})
	}
}

func Benchmark_Universe(b *testing.B) {
	for _, e := range engineNames() {
		b.Run(e, func(b *testing.B) {
			u := engines[e](newUniverseOptions(), newStateCh())
			universeRun(u, b)
		})
	}
}

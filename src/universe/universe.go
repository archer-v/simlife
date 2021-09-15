package universe

type Universe interface {
	Status() Status
	Options() Options
	Area() Area
	StateCh() chan Status
	AddTemplate(tmpl Template)
	SettleTemplate(name string)
	SettleWithRandomData()
	Settle(vc [][]int)
	InverseCell(x int, y int)
	RegisterViewer(v Viewer)
	Run()
	Stop()
	Step()
	Clear()
	Close()
}

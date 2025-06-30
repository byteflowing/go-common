package service

import "sync"

type RoutineGroup struct {
	wg sync.WaitGroup
}

func NewRoutineGroup() *RoutineGroup {
	return &RoutineGroup{}
}

func (g *RoutineGroup) Run(fn func()) {
	g.wg.Add(1)

	go func() {
		defer g.wg.Done()
		fn()
	}()
}

func (g *RoutineGroup) Wait() {
	g.wg.Wait()
}

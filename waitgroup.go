package main

import "sync"

type WaitGroup struct {
	wg sync.WaitGroup
	p  chan struct{}
}

func NewWaitGroup(parallel int) (w *WaitGroup) {
	w = &WaitGroup{}
	if parallel <= 0 {
		return
	}
	w.p = make(chan struct{}, parallel)
	return
}

func (w *WaitGroup) AddDelta() {
	w.wg.Add(1)
	if w.p == nil {
		return
	}
	w.p <- struct{}{}
}

func (w *WaitGroup) Done() {
	w.wg.Done()
	if w.p == nil {
		return
	}
	<-w.p
}

func (w *WaitGroup) Wait() {
	w.wg.Wait()
}

func (w *WaitGroup) Parallel() int {
	return len(w.p)
}

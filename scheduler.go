// Copyright 2017 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package timer

type SimpleScheduler struct {
	dispatcher Dispatcher
}

func NewSimpleScheduler(dispatcher Dispatcher) Scheduler {
	return &SimpleScheduler{
		dispatcher: dispatcher,
	}
}
func (s *SimpleScheduler) Start() {
	s.dispatcher.Start()
}

func (s *SimpleScheduler) Stop() {
	s.dispatcher.Stop()
}

func (s *SimpleScheduler) Schedule(fn func()) {
	s.dispatcher.Dispatch(fn)
}

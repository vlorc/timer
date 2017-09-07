// Copyright 2017 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package timer

import (
	"errors"
)

type HashDispatcher struct {
	mask int
	count int
	table []Dispatcher
}

func NewHashDispatcher(args ...Dispatcher) Dispatcher {
	length := len(args)
	if 0 != (length&(length - 1)) {
		panic(errors.New("NewTimingWheel length"))
	}
	obj := &HashDispatcher{
		mask: length - 1,
		table: make([]Dispatcher,length),
	}

	copy(obj.table,args)
	return obj
}

func(h *HashDispatcher)Start(){
	for _,v := range h.table {
		if nil != v {
			v.Start()
		}
	}
}

func(h *HashDispatcher)Stop(){
	for _,v := range h.table {
		if nil != v {
			v.Stop()
		}
	}
}

func(h *HashDispatcher)Dispatch(fn func()){
	h.count++
	h.table[h.count & h.mask].Dispatch(fn)
}

type QueueDispatcher struct {
	request chan func()
	dispatcher Dispatcher
	length int
}

func NewQueueDispatcher(length int,dispatcher Dispatcher) Dispatcher {
	return &QueueDispatcher{
		length: length,
		dispatcher: dispatcher,
	}
}

func(q *QueueDispatcher)Start(){
	if nil != q.request {
		panic(errors.New("once start"))
	}

	q.dispatcher.Start()
	q.request = make(chan func(),q.length)
	go q.run()
}

func(q *QueueDispatcher)Stop(){
	close(q.request)
	q.dispatcher.Stop()
}

func(q *QueueDispatcher)Dispatch(fn func()){
	q.request <- fn
}

func(q *QueueDispatcher)run(){
	for fn := range q.request{
		q.dispatcher.Dispatch(fn)
	}
	q.request = nil
}

type SimpleDispatcher struct {}

func NewSimpleDispatcher() Dispatcher{
	return &SimpleDispatcher{}
}
func(s *SimpleDispatcher)Start(){
}
func(s *SimpleDispatcher)Stop(){
}
func(s *SimpleDispatcher)Dispatch(fn func()){
	fn()
}
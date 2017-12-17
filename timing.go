// Copyright 2017 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package timer

import (
	"errors"
	"fmt"
	"sync/atomic"
	"time"
)

type OperationCode int8

const (
	OP_REMOVE OperationCode = iota
	OP_AFTER
	OP_AT
)

type Operation struct {
	tm *Timer
	op OperationCode
}

type TimingWheel struct {
	ticker         *time.Ticker
	request        chan Operation
	interval       int64
	wheel          []*Wheel
	mask           int
	count          int64
	push           func(OperationCode, *Timer)
	scheduler      Scheduler
	schedulerReal  Scheduler
	quit           chan struct{}
	state          int32
}

func Default(interval ...time.Duration) *TimingWheel {
	val := time.Second
	if len(interval) > 0 && interval[0] > 0 {
		val = interval[0]
	}
	return NewTimingWheel(
		NewSimpleScheduler(NewQueueDispatcher(1024, NewSimpleDispatcher())),
		1024,
		val,
		64, 64, 128, 256)
}

func NewTimingWheel(scheduler Scheduler, request int, interval time.Duration, count ...int) *TimingWheel {
	length := len(count)
	if 0 != (length & (length - 1)) {
		panic(errors.New("NewTimingWheel illegal length"))
	}

	obj := &TimingWheel{
		interval:       int64(interval),
		wheel:          make([]*Wheel, len(count)),
		schedulerReal: scheduler,
		mask:           length - 1,
		request:        make(chan Operation, request),
		quit:           make(chan struct{}),
	}
	for i, v := range count {
		obj.wheel[i] = NewWheel(v)
	}
	obj.push = func(op OperationCode, tm *Timer) {
		obj.request <- Operation{
			op: op,
			tm: tm,
		}
	}

	return obj
}

func (t *TimingWheel) Count() int64 {
	return t.count
}

func (t *TimingWheel) Step() {
	t.count++
	for i, l := 0, 0; i <= l; i++ {
		index, slot := t.wheel[i].Step()
		if 0 == index {
			l = (l + 1) & t.mask
		}
		if nil != slot && slot.Len() > 0 {
			t.wheel[i].slot[index] = nil
			t.insertSlot(slot)
		}
	}
}

func (t *TimingWheel) Wait() {
	<-t.quit
}

func (t *TimingWheel) After(d time.Duration, fn func()) *Timer {
	return t.afterInterval(d, 0, fn)
}

func (t *TimingWheel) calc(d int64, v int64) int64 {
	d = d / t.interval
	if d <= 0 {
		d = v
	}
	return d
}

func (t *TimingWheel) atInterval(d time.Time, i int64, fn func()) (tm *Timer) {
	if d.IsZero() {
		return
	}

	tm = &Timer{
		count:    d.UnixNano(),
		value:    i,
		complete: toValue,
		fn:       fn,
	}
	t.push(OP_AT, tm)
	return tm
}

func (t *TimingWheel) At(d time.Time, fn func()) (tm *Timer) {
	return t.atInterval(d, 0, fn)
}

func (t *TimingWheel) AtInterval(d time.Time, i time.Duration, fn func()) (tm *Timer) {
	return t.atInterval(d, t.calc(int64(i), 1), fn)
}

func (t *TimingWheel) afterInterval(d time.Duration, i int64, fn func()) (tm *Timer) {
	if d <= 0 {
		return
	}

	tm = &Timer{
		count:    t.calc(int64(d), i),
		value:    i,
		complete: toValue,
		fn:       fn,
	}

	t.push(OP_AFTER, tm)
	return
}

func (t *TimingWheel) AfterInterval(d, i time.Duration, fn func()) *Timer {
	return t.afterInterval(d, t.calc(int64(i), 1), fn)
}

func (t *TimingWheel) Interval(i time.Duration, fn func()) *Timer {
	return t.afterInterval(i, t.calc(int64(i), 1), fn)
}

func (t *TimingWheel) Add(tm *Timer) {
	t.push(OP_AFTER, tm)
}

func (t *TimingWheel) Cancel(tm *Timer) {
	tm.Cancel()
	t.push(OP_REMOVE, tm)
	return
}

func (t *TimingWheel) Start() {
	if atomic.CompareAndSwapInt32(&t.state, 0, 1) {
		t.ready()
		go t.worker()
	}
}

func (t *TimingWheel) ready() {
	t.scheduler = t.schedulerReal
	t.scheduler.Start()
}

func (t *TimingWheel) Stop() {
	if !atomic.CompareAndSwapInt32(&t.state, 1, 0) {
		return
	}
	t.push = func(OperationCode, *Timer) {}
	close(t.request)
	t.scheduler = NewEmptyScheduler()
	t.schedulerReal.Stop()
}

func (t *TimingWheel) destroy() {
	close(t.quit)
	t.quit = nil
	t.ticker = nil
	t.request = nil
}

func (t *TimingWheel) worker() {
	defer t.destroy()

	if t.join() {
		t.loop()
	}
}

func (t *TimingWheel) join() bool {
	for {
		select {
		case req, ok := <-t.request:
			if !ok {
				return false
			}
			t.doRequest(req.op, req.tm)
		default:
			return true
		}
	}
}
func (t *TimingWheel) loop() {
	t.ticker = time.NewTicker(time.Duration(t.interval))
	defer t.ticker.Stop()

	for {
		select {
		case <-t.ticker.C:
			t.Step()
		case req, ok := <-t.request:
			if !ok {
				return
			}
			t.doRequest(req.op, req.tm)
		}
	}
}

func (t *TimingWheel) doRequest(op OperationCode, tm *Timer) {
	if nil == tm {
		return
	}
	switch op {
	case OP_REMOVE:
		tm.remove()
	case OP_AT:
		tm.count = (tm.count - time.Now().UnixNano()) / t.interval
		fallthrough
	default:
		t.insertTimer(tm)
	}
}

func (t *TimingWheel) insertSlot(slot *Slot) {
	node := slot.Front()
	node.prev.next = nil
	for {
		next := node.next
		t.insertTimer(node)
		if nil == next {
			break
		}
		node = next
	}
}

func (t *TimingWheel) insertTimer(tm *Timer) {
	if tm.count <= 0 {
		t.scheduler.Schedule(tm.fn)
		if tm.count = tm.complete(tm); tm.count <= 0 {
			tm.list = nil
			return
		}
	}

	value := tm.count
	bit := uint(0)
	for i, w := range t.wheel {
		value >>= w.bit
		if 0 == value {
			value = tm.count >> bit
			tm.count -= value << bit
			t.wheel[i].Push(int(value), tm)
			return
		}
		bit += w.bit
	}
	panic(fmt.Errorf("insert failed count(%d)", tm.count))
}

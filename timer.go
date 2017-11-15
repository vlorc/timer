// Copyright 2017 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package timer

import (
	"errors"
	"fmt"
	"time"
)

type OperationCode int

const (
	OP_REMOVE OperationCode = iota
	OP_AFTER
	OP_AT
)

type Operation struct {
	op OperationCode
	tm *Timer
}

type TimingWheel struct {
	ticker      *time.Ticker
	request     chan Operation
	quit        chan struct{}
	interval    int64
	wheel       []*Wheel
	mask        int
	count       int64
	id          int64
	request_max int

	push      	func(OperationCode,*Timer)
	scheduler      Scheduler
	scheduler_real Scheduler
}

func Default() *TimingWheel {
	return NewTimingWheel(
		NewSimpleScheduler(NewQueueDispatcher(256,NewSimpleDispatcher())),
		256,
		time.Second,
		32, 64, 128, 256)
}

func NewTimingWheel(scheduler Scheduler, request_max int, interval time.Duration, count ...int) *TimingWheel {
	length := len(count)
	if 0 != (length & (length - 1)) {
		panic(errors.New("NewTimingWheel length"))
	}

	obj := &TimingWheel{
		interval:       int64(interval),
		wheel:          make([]*Wheel, len(count)),
		scheduler_real: scheduler,
		mask:           length - 1,
		request_max:    request_max,
	}
	for i, v := range count {
		obj.wheel[i] = NewWheel(v)
	}

	return obj
}

func (t *TimingWheel) Start() {
	if nil != t.ticker {
		panic(errors.New("once start"))
	}

	t.start()
	go t.run()
}

func (t *TimingWheel) start() {
	t.scheduler = t.scheduler_real
	t.quit = make(chan struct{})
	t.request = make(chan Operation, t.request_max)
	t.ticker = time.NewTicker(time.Duration(t.interval))
	t.push = func(op OperationCode,tm *Timer) {
		t.request <- Operation{
			op: op,
			tm: tm,
		}
	}
	t.scheduler.Start()
}

func (t *TimingWheel) stop() {
	t.push = func(OperationCode,*Timer) {}
	t.scheduler = NewEmptyScheduler()
	t.scheduler_real.Stop()

	close(t.quit)
}

func (t *TimingWheel) Stop() {
	if nil != t.quit {
		t.stop()
	}
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
	<- t.quit
}

func (t *TimingWheel) After(d time.Duration, fn func()) (tm *Timer) {
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
		count:   d.UnixNano(),
		expires: i,
		fn:      fn,
	}
	t.push(OP_AT,tm)
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
		count:   t.calc(int64(d), i),
		expires: i,
		fn:      fn,
	}

	t.push(OP_AFTER,tm)
	return
}

func (t *TimingWheel) AfterInterval(d, i time.Duration, fn func()) (tm *Timer) {
	return t.afterInterval(d, t.calc(int64(i), 1), fn)
}

func (t *TimingWheel) Interval(i time.Duration, fn func()) (tm *Timer) {
	return t.afterInterval(i, t.calc(int64(i), 1), fn)
}

func (t *TimingWheel) Cancel(tm *Timer) {
	t.push(OP_REMOVE,tm)
	return
}

func (t *TimingWheel) destroy() {
	close(t.request)
	t.ticker.Stop()
	t.request = nil
	t.ticker = nil
	t.quit = nil
}

func (t *TimingWheel) run() {
	defer t.destroy()

	for {
		select {
		case <-t.ticker.C:
			t.Step()
		case d := <-t.request:
			t.doRequest(d.op, d.tm)
		case <-t.quit:
			return
		}
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
		t.id++
		tm.id = t.id
		t.insertTimer(tm)
	}
}

func (t *TimingWheel) insertTimer(tm *Timer) {
	if tm.count <= 0 {
		t.scheduler.Schedule(tm.fn)
		if tm.count = tm.expires; tm.count <= 0 {
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

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
	scheduler   Scheduler
}

type Timer struct {
	count int64
	fn    func()
	el    *Element
	id    int64
}

type Wheel struct {
	slot []*List
	pos  int
	mask int
	bit  uint
}

type Operation struct {
	op OperationCode
	tm *Timer
}

func popcount(x uint64) uint {
	const (
		m1  = 0x5555555555555555
		m2  = 0x3333333333333333
		m4  = 0x0f0f0f0f0f0f0f0f
		h01 = 0x0101010101010101
	)
	x -= (x >> 1) & m1
	x = (x & m2) + ((x >> 2) & m2)
	x = (x + (x >> 4)) & m4
	return uint((x * h01) >> 56)
}

func NewWheel(length int) *Wheel {
	if 0 != (length & (length - 1)) {
		panic(errors.New("NewWheel length"))
	}

	return &Wheel{
		slot: make([]*List, length),
		mask: length - 1,
		bit:  popcount(uint64(length - 1)),
	}
}

func (w *Wheel) Slot(i int) *List {
	slot := w.slot[i]
	if nil == slot {
		slot = NewList()
		w.slot[i] = slot
	}
	return slot
}

func (w *Wheel) Step() (int, *List) {
	w.pos++
	i := w.pos & w.mask
	return i, w.slot[i]
}

func NewTimingWheel(scheduler Scheduler, request_max int, interval time.Duration, count ...int) *TimingWheel {
	length := len(count)
	if 0 != (length & (length - 1)) {
		panic(errors.New("NewTimingWheel length"))
	}

	obj := &TimingWheel{
		interval:    int64(interval),
		wheel:       make([]*Wheel, len(count)),
		scheduler:   scheduler,
		mask:        length - 1,
		request_max: request_max,
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

	t.quit = make(chan struct{})
	t.request = make(chan Operation, t.request_max)
	t.ticker = time.NewTicker(time.Duration(t.interval))
	t.scheduler.Start()

	go t.run()
}

func (t *TimingWheel) Stop() {
	if nil != t.quit {
		close(t.quit)
		t.scheduler.Stop()
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
	<-t.quit
}

func (t *TimingWheel) After(d time.Duration, fn func()) (tm *Timer) {
	if d <= 0 {
		return
	}
	tm = &Timer{
		count: int64(d) / t.interval,
		fn:    fn,
	}

	t.request <- Operation{
		op: OP_AFTER,
		tm: tm,
	}
	return
}

func (t *TimingWheel) At(d time.Time, fn func()) (tm *Timer) {
	if d.IsZero() {
		return
	}

	tm = &Timer{
		count: d.UnixNano(),
		fn:    fn,
	}
	t.request <- Operation{
		op: OP_AT,
		tm: tm,
	}
	return
}

func (t *TimingWheel) Cancel(tm *Timer) {
	t.request <- Operation{
		op: OP_REMOVE,
		tm: tm,
	}
	return
}

func (t *TimingWheel) run() {
	for {
		select {
		case <-t.ticker.C:
			t.Step()
		case <-t.quit:
			t.ticker.Stop()
			break
		case d := <-t.request:
			t.doRequest(d.op, d.tm)
		}
	}

	close(t.request)
	t.request = nil
	t.ticker = nil
	t.quit = nil
}

func (t *TimingWheel) insertSlot(list *List) {
	for e := list.Front(); nil != e; e = e.Next() {
		t.insertTimer(e.Value.(*Timer))
	}
}

func (t *TimingWheel) doRequest(op OperationCode, tm *Timer) {
	switch op {
	case OP_REMOVE:
		if nil != tm.el {
			tm.el.Remove()
			tm.el = nil
		}
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
		return
	}

	value := tm.count
	bit := uint(0)
	for i, w := range t.wheel {
		value >>= w.bit
		if 0 == value {
			value = tm.count >> bit
			pos := int(value) + t.wheel[0].pos&t.wheel[0].mask
			tm.count -= value << bit
			tm.el = t.wheel[i].Slot(pos).Push(tm)
			return
		}
		bit += w.bit
	}
	panic(fmt.Errorf("insert failed count(%d)", tm.count))
}

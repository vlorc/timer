// Copyright 2017 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package timer

import (
	"errors"
)

type Wheel struct {
	slot []*Slot
	pos  int
	mask int
	bit  uint
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
		slot: make([]*Slot, length),
		mask: length - 1,
		bit:  popcount(uint64(length - 1)),
	}
}

func (w *Wheel) Push(i int, tm *Timer) {
	w.Slot(w.pos + i).Push(tm)
}

func (w *Wheel) Pop(i int) *Timer {
	if slot := w.slot[(w.pos+i)&w.mask]; nil == slot {
		return slot.Pop()
	}
	return nil
}

func (w *Wheel) Clear(i int) *Slot {
	i &= w.mask
	slot := w.slot[i]
	w.slot[i] = nil
	return slot
}

func (w *Wheel) Slot(i int) *Slot {
	i &= w.mask
	slot := w.slot[i]
	if nil == slot {
		slot = NewSlot()
		w.slot[i] = slot
	}
	return slot
}

func (w *Wheel) Step() (int, *Slot) {
	w.pos++
	i := w.pos & w.mask
	return i, w.slot[i]
}

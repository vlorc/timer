// Copyright 2017 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package timer

type Complete func(*Timer) int64

type Timer struct {
	count      int64
	value      int64
	complete   Complete
	fn         func()
	next, prev *Timer
	list       *Slot
}

func toZero(*Timer) int64 {
	return 0
}

func toValue(tm *Timer) int64 {
	return tm.value
}

func makeToTable(tab []int64) Complete {
	return func(tm *Timer) int64 {
		if tm.value++; tm.value < int64(len(tab)) {
			return tab[tm.value]
		}
		return 0
	}
}

func makeToTimes(times int64, complete Complete) Complete {
	return func(tm *Timer) int64 {
		if times--; times > 0 {
			return complete(tm)
		}
		return 0
	}
}

func NewTimer(fn func(), count int64) *Timer {
	return &Timer{
		count:    count,
		value:    count,
		complete: toValue,
		fn:       fn,
	}
}

func NewTimerTable(fn func(), count ...int64) *Timer {
	if 1 == len(count){
		return &Timer{
			count:    count[0],
			complete: toZero,
			fn:       fn,
		}
	}

	return &Timer{
		count:    count[0],
		complete: makeToTable(count),
		fn:       fn,
	}
}

func NewTimerPeriodic(fn func(), times, count int64) *Timer {
	return &Timer{
		count:    count,
		value:    count,
		complete: makeToTimes(times, toValue),
		fn:       fn,
	}
}

func NewTimerCustom(fn func(), complete Complete, count int64) *Timer {
	return &Timer{
		count:    count,
		complete: complete,
		fn:       fn,
	}
}

func (e *Timer) remove() *Timer {
	if nil != e.list {
		return e.list.Remove(e)
	}
	return nil
}

func (e *Timer) Cancel() {
	e.fn = func() {}
	e.complete = toZero
}

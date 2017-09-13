// Copyright 2017 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package timer

type Timer struct {
	count      int64
	expires    int64
	id         int64
	fn         func()
	next, prev *Timer
	list       *Slot
}

type Slot struct {
	root   *Timer
	length int
}

func NewSlot() *Slot {
	return &Slot{}
}

func (e *Timer) remove() *Timer {
	if nil != e.list {
		return e.list.Remove(e)
	}
	return nil
}

func (l *Slot) Front() *Timer {
	return l.root
}

func (l *Slot) Len() int {
	return l.length
}

func (l *Slot) Back() (el *Timer) {
	if nil != l.root {
		el = l.root.prev
	}
	return
}

func (l *Slot) Clear() {
	l.root = nil
	l.length = 0
}

func (l *Slot) Push(el *Timer) {
	if nil == l.root {
		el.prev = el
		el.next = el
		l.root = el
	} else {
		el.next = l.root
		el.prev = l.root.prev
		l.root.prev.next = el
		l.root.prev = el
	}

	el.list = l
	l.length++
	return
}

func (l *Slot) Pop() (el *Timer) {
	if nil == l.root {
		return
	}
	if l.root == l.root.next {
		el = l.root
		l.root = nil
	} else {
		el = l.root.prev
		el.prev.next = l.root
		l.root.prev = el.prev
	}

	el.list = nil
	l.length--
	return
}

func (l *Slot) Remove(el *Timer) *Timer {
	if nil == l.root || l != el.list {
		return nil
	}

	el.list = nil
	l.length--
	if l.root == el {
		if el == el.next {
			l.root = nil
			return nil
		}
		l.root = el.next
	}

	el.prev.next = el.next
	el.next.prev = el.prev
	return el.next
}

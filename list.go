// Copyright 2017 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package timer

type Element struct {
	next, prev *Element
	Value      interface{}
	list       *List
}

type List struct {
	root   *Element
	length int
}

func NewList() *List {
	return &List{}
}

func (e *Element) Next() *Element {
	if p := e.next; nil != e.list && e.list.root != p {
		return p
	}
	return nil
}

func (e *Element) Prev() *Element {
	if p := e.prev; nil != e.list && e.list.root != p {
		return p
	}
	return nil
}

func (e *Element) Remove() *Element {
	if nil != e.list {
		return e.list.Remove(e)
	}
	return nil
}

func (l *List) Front() *Element {
	return l.root
}

func (l *List) Len() int {
	return l.length
}

func (l *List) Back() (el *Element) {
	if nil != l.root {
		el = l.root.prev
	}
	return
}

func (l *List) Clear() {
	l.root = nil
	l.length = 0
}

func (l *List) Push(value interface{}) (el *Element) {
	el = &Element{
		Value: value,
		list:  l,
	}

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

	l.length++
	return
}

func (l *List) Pop() (el *Element) {
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

func (l *List) Remove(el *Element) *Element {
	if nil == l.root {
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

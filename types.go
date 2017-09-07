// Copyright 2017 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package timer

type Scheduler interface {
	Start()
	Stop()
	Schedule(func())
}

type Dispatcher interface {
	Start()
	Stop()
	Dispatch(func())
}

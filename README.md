# timing wheel

[![License](https://img.shields.io/:license-apache-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![codebeat badge](https://codebeat.co/badges/c41b426c-4121-4dc8-99c2-f1b60574be64)](https://codebeat.co/projects/github-com-vlorc-timer-master)
[![Go Report Card](https://goreportcard.com/badge/github.com/vlorc/timer)](https://goreportcard.com/report/github.com/vlorc/timer)
[![GoDoc](https://godoc.org/github.com/vlorc/timer?status.svg)](https://godoc.org/github.com/vlorc/timer)
[![Build Status](https://travis-ci.org/vlorc/timer.svg?branch=master)](https://travis-ci.org/vlorc/timer?branch=master)
[![Coverage Status](https://coveralls.io/repos/github/vlorc/timer/badge.svg?branch=master)](https://coveralls.io/github/vlorc/timer?branch=master)

golang timing wheel timer

## Examples

###  base timer
```golang
import (
	"fmt"
	"github.com/vlorc/timer"
)

func main() {
	timing := NewTimingWheel(
    		NewSimpleScheduler(NewSimpleDispatcher()),
    		64,
    		time.Second,
    		64,128,256,512)
    timing.Start()
    timing.After(26 * time.Second, func() {
    	fmt.Println("26 Second")
    	timing.Stop()
    })
    timing.Wait()
}
```
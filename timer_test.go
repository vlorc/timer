package timer

import (
	"testing"
	"time"
)

func Test_NewTimingWheel(t *testing.T) {

	timing := NewTimingWheel(
		NewSimpleScheduler(NewSimpleDispatcher()),
		64,
		time.Second,
		1, 2, 4, 8)

	begin := time.Now().Unix()

	t0 := timing.After(10*time.Second, func() {
		t.Error("no cancel")
	})
	timing.Add(NewTimerTable(func() {
		timing.Cancel(t0)
	}, 8))

	timing.After(28*time.Second, func() {
		if v := time.Now().Unix() - begin; 28 != v {
			t.Error("After Real: ", v)
		}
		timing.Stop()
	})

	interval := time.Now().Unix()
	timing.Interval(2*time.Second, func() {
		now := time.Now().Unix()
		if v := now - interval; 2 != v {
			t.Error("Interval Real: ", v)
		}
		interval = now
	})

	timing.Start()
	timing.Wait()
}

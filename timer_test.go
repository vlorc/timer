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
	timing.Start()

	begin := time.Now().Unix()

	t0 := timing.After(10*time.Second, func() {
		t.Error("no cancel")
	})
	timing.After(8*time.Second, func() {
		timing.Cancel(t0)
	})

	timing.After(26*time.Second, func() {
		if v := time.Now().Unix() - begin; 26 != v {
			t.Error("TIME REAL: ", v)
		}
		timing.Stop()
	})
	timing.Wait()
}

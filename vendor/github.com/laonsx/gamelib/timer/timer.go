package timer

import (
	"time"

	"github.com/laonsx/gamelib/g"
)

var timers []*time.Timer

func AfterFunc(delay time.Duration, count int, cb func(n int)) *time.Timer {

	t := time.NewTimer(delay)
	timers = append(timers, t)

	g.Go(func() {

		start(t, delay, count, cb)
	})

	return t
}

func start(t *time.Timer, delay time.Duration, count int, callback func(n int)) {

	defer func() {

		t.Stop()
	}()

	i := 1
	for {

		select {

		case <-t.C:

			callback(i)
			if count > 0 && i == count {

				return
			}

			t.Reset(delay)
			i++

		case <-g.Quit():

			return
		}
	}
}

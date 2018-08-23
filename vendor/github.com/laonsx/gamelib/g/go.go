package g

import (
	"sync"
)

var (
//	mux       sync.Mutex
//	pending   int
	quit      chan struct{}
	waitGroup sync.WaitGroup
)

func init() {

	quit = make(chan struct{})
}

func Go(f func()) {

	go run(f)
}

func Quit() chan struct{} {

	return quit
}

func Close() {

	close(quit)
	waitGroup.Wait()
}

func run(f func()) {

	waitGroup.Add(1)

	defer func() {

		waitGroup.Done()
//		mux.Lock()
//		pending--
//		mux.Unlock()
	}()

//	mux.Lock()
//	pending++
//	mux.Unlock()

	for {

		select {

		case <-quit:

			return

		default:

		}

		f()

		return
	}
}

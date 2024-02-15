package main

import (
	"asyncLogger/asyncLogger"
	"strconv"
	"sync"
	"time"
)

// you can make this public for package-wide use
var stdOutAsyncLogger asyncLogger.AsyncLogger

func main() {

	// this uses the builder pattern, which is
	// good style but also forces the users to
	// use a diffrente instance after each mutation.
	// For loggers, it is more suitable.
	ticker := *time.NewTicker(time.Second)
	channel := make(chan []byte, 3)

	config := asyncLogger.AsyncLoggerConfig{}.
		WithLoggerName("simpleLogger").
		WithLoggerSeverity(asyncLogger.INFO).
		WithTimeTick(ticker).
		WithBuffer(channel).
		WithAutoFlushSetTo(true).
		WithFlushTimeOut(100 * time.Millisecond)

	stdOutAsyncLogger = &asyncLogger.StdOutAsyncLogger{
		Config: config,
	}

	go func() {
		ticker.Reset(time.Second)
		stdOutAsyncLogger.Listen() // listen for new logs in the background
	}()

	// get the loggin handler
	handler := stdOutAsyncLogger.GetAsyncLoggerHandle()

	var waitGroup sync.WaitGroup
	for i := 0; i < 30; i++ {

		// simulate a latency
		if i%3 == 0 {
			time.Sleep(250 * time.Millisecond)
		}

		// register the go routine
		waitGroup.Add(1)

		go func(id int) {
			handler <- []byte("Hello from go routine n°" + strconv.Itoa(id))

			// on return, unregister self
			defer waitGroup.Done()
		}(i)
	}

	// wait for all the goroutines to end
	waitGroup.Wait()

	// just to make sure that everything is done
	// @todo: replace this with a quit channel
	// or a context deadline on the main func/app

	time.Sleep(5 * time.Second)
	stdOutAsyncLogger.Close()
}

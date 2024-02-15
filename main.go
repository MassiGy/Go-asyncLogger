package main

import (
	"asyncLogger/asyncLogger"
	"strconv"
	"sync"
	"time"
)

var stdOutAsyncLogger asyncLogger.AsyncLogger

func main() {

	config := asyncLogger.AsyncLoggerConfig{
		Name:          "simpleLogger",
		SeverityLevel: "info",
		Tick:          *time.NewTicker(time.Second),

		// you can make a buffred channel of cap <= 3 * 1 / (1/4)
		// since we tick every 1 second, and in every 1 second there
		// is 4 one-forth of a second, and in each one of these we
		// emit 3msg.
		// TL;DR: you can expand the length of this buffred channel up until 12 slots
		Buffer:       make(chan []byte),
		FlushTimeOut: 500 * time.Millisecond, // .5 second deadline
	}

	stdOutAsyncLogger = &asyncLogger.StdOutAsyncLogger{
		Config: config,
	}

	// flush on every tick
	stdOutAsyncLogger.SetAutoFlush()
	go stdOutAsyncLogger.Listen() // listen for new logs in the background

	// get the loggin handler
	handler := stdOutAsyncLogger.GetAsyncLoggerHandle()

	var waitGroup sync.WaitGroup
	for i := 0; i < 30; i++ {
		// register the go routine
		waitGroup.Add(1)

		if i%3 == 0 {
			time.Sleep(250 * time.Millisecond)
		}
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

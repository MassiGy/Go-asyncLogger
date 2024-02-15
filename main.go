package main

import (
	"asyncLogger/asyncLogger"
	"strconv"
	"sync"
	"time"
)

func main() {

	/*
		This is a program that is designed just to tinker around with
		channels, context and select statement

		Current setup:
			Async logger gets triggered every tick of 1 second
					For every tick, the logger has 500ms to log the buffred msgs

			The publishers are a set of 30 goroutines, fired in a loop
				Every 3 go routines, we sleep for 1/4 of a second
				(which means that every 1/4 second, 3 msg get bufferd)


		Quick rundown:	(small subset of a complete execution)

			We create our async logger
			We set the autoflush to flush in every tick
			we enter the loop i == 0,
				we sleep for 1/4 		(750ms remaining for the next tick)
				we quickly fire 3 msg
				we sleep for 1/4		(500ms remaining for the next tick)
				we quickly fire 3 msg
				we sleep for 1/4		(250ms remaining for the next tick)
				we quickly fire 3 msg
				we sleep for 1/4

				in parallel (
					first tick is fired
						we consume the 12 msg
						we wait for others until the deadline of 500 ms is met

					we quicly fire 3 msg
					we sleep for 1/4
				)

		During the last part (where the tick is fired) , we will intecept the
		newly fired 3 msgs before the 500ms deadline, and that is why in the
		console, the first tick contains 15msgs at once.

		An intersting question is that since in the 'in-parallel' part of the
		above quick run down we sleep for 1/4s and we have a 500ms deadline, why
		can't we squeeze in the extra 3messages (cuz 500ms = 2 * 1/4ms), is is because
		we are doing IO in this example or is it something else ?

	*/

	config := asyncLogger.AsyncLoggerConfig{
		Name:          "simpleLogger",
		SeverityLevel: "info",
		Tick:          *time.NewTicker(time.Second),
	}

	stdOutAsyncLogger := asyncLogger.StdOutAsyncLogger{
		// you can make a buffred channel of cap <= 3 * 1 / (1/4)
		// since we tick every 1 second, and in every 1 second there
		// is 4 one-forth of a second, and in each one of these we
		// emit 3msg.
		// TL;DR: you can expand the length of this buffred channel up until 12 slots
		Buffer:       make(chan string, 3),
		Config:       config,
		FlushTimeOut: 500 * time.Millisecond, // .5 second deadline
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
			handler <- "Hello from go routine n°" + strconv.Itoa(id)

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

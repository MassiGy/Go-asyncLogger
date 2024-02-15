package asyncLogger

import (
	"context"
	"fmt"
	"os"
	"time"
)

type AsyncLoggerConfig struct {

	// @todo: make the fields private, espeacially for the buffer
	//		  since the users of the logger only need the writing
	//        end of the buffer. Otherwise you can have a buffer leak
	//        where another part of your program consumes the buffer
	//        content.

	// set the name of the logger
	Name string

	// set the Severitylevel of the logger
	SeverityLevel string

	// set a ticker for autoflush behavior
	Tick time.Ticker

	Buffer       chan []byte
	IsAutoFlush  bool
	FlushTimeOut time.Duration
}

type AsyncLogger interface {
	GetConfig() AsyncLoggerConfig

	GetAsyncLoggerHandle() chan<- []byte
	SetAutoFlush() error
	Listen() error
	Flush(time.Time) error
	Close() error
}

type StdOutAsyncLogger struct {

	// config object, we can also embbed the type instead
	Config AsyncLoggerConfig
}

func (stdOutAsyncLogger StdOutAsyncLogger) GetConfig() AsyncLoggerConfig {
	return stdOutAsyncLogger.Config
}

// this will act as a middelware between the users of the channel and the logger
// internals. Hopefully this will make the users only use the channel as a write-to channel
// @todo: to fix this incertainty, make the buffer field private
func (stdOutAsyncLogger *StdOutAsyncLogger) GetAsyncLoggerHandle() chan<- []byte {
	return stdOutAsyncLogger.Config.Buffer
}

func (stdOutAsyncLogger *StdOutAsyncLogger) SetAutoFlush() error {
	stdOutAsyncLogger.Config.IsAutoFlush = true

	// @todo check if the ticker is proprely set

	return nil
}

func (stdOutAsyncLogger *StdOutAsyncLogger) Listen() error {
	var (
		err error
	)

	// for every tick, flush to logger sink
	for tick := range stdOutAsyncLogger.Config.Tick.C {

		// pass the time instant to flush
		err = stdOutAsyncLogger.Flush(tick)
		if err != nil {
			return err
		}
	}

	return nil
}

func (stdOutAsyncLogger *StdOutAsyncLogger) Flush(timeStamp time.Time) error {

	// FlushTimeOut is not really used as a time out since we will allways
	// meet that deadline, and that is due to the fact that listening to
	// the buffer is blocking, thus with the select statement we will be
	// able to cut it at that deadline
	ctx, cancel := context.WithTimeout(context.Background(), stdOutAsyncLogger.Config.FlushTimeOut)
	fmt.Fprintf(os.Stdout, "Start of tick ===============\n")

	// while true
	for {

		// either consume the msg, or quit
		select {
		case msg := <-stdOutAsyncLogger.Config.Buffer:
			{
				fmt.Fprintf(os.Stdout, "[Minute:%d, Second: %d, Milisecond:%d]\t", timeStamp.Minute(), timeStamp.Second(), timeStamp.UnixMilli())
				fmt.Fprintf(os.Stdout, "@%s:\t", stdOutAsyncLogger.Config.Name)
				fmt.Fprintf(os.Stdout, "(%s)\t", stdOutAsyncLogger.Config.SeverityLevel)
				fmt.Fprintf(os.Stdout, "%s\n", msg)
			}
		case <-ctx.Done():
			fmt.Fprintf(os.Stdout, "End of tick ===============\n")
			cancel()
			return nil
		}
	}
}

func (stdOutAsyncLogger *StdOutAsyncLogger) Close() error {

	// we need to close the buffer
	close(stdOutAsyncLogger.Config.Buffer)

	return nil
}

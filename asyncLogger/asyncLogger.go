package asyncLogger

import (
	"fmt"
	"os"
	"time"
)

// make the severityLevel type private, since this is the
// underlying type for our enum. This will prevent users from
// defining other entries of the same nature as our enums
type severityLevel int
type severityDict map[severityLevel]string

const (
	DEBUG = iota // starts at 0
	INFO
	WARNNING
	DANGER
	CRITICAL
)

var severityLevels severityDict = severityDict{
	DEBUG:    "DEBUG",
	INFO:     "INFO",
	WARNNING: "WARNNING",
	DANGER:   "DANGER",
	CRITICAL: "CRITICAL",
}

type AsyncLoggerConfig struct {

	// set the name of the logger
	name string

	// set the Severitylevel of the logger
	severityLevel severityLevel

	// set a ticker for autoflush behavior
	tick time.Ticker

	buffer       chan []byte
	isAutoFlush  bool
	flushTimeOut time.Duration
}

func (conf AsyncLoggerConfig) WithLoggerName(name string) AsyncLoggerConfig {
	conf.name = name
	return conf
}
func (conf AsyncLoggerConfig) WithLoggerSeverity(severity severityLevel) AsyncLoggerConfig {
	conf.severityLevel = severity
	return conf
}

func (conf AsyncLoggerConfig) WithTimeTick(tick time.Ticker) AsyncLoggerConfig {
	conf.tick = tick
	return conf
}

func (conf AsyncLoggerConfig) WithBuffer(buffer chan []byte) AsyncLoggerConfig {
	conf.buffer = buffer
	return conf
}
func (conf AsyncLoggerConfig) WithAutoFlushSetTo(isAutoFlush bool) AsyncLoggerConfig {
	conf.isAutoFlush = isAutoFlush
	return conf
}
func (conf AsyncLoggerConfig) WithFlushTimeOut(flushTimeOut time.Duration) AsyncLoggerConfig {
	conf.flushTimeOut = flushTimeOut
	return conf
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
func (stdOutAsyncLogger *StdOutAsyncLogger) GetAsyncLoggerHandle() chan<- []byte {
	return stdOutAsyncLogger.Config.buffer
}

func (stdOutAsyncLogger *StdOutAsyncLogger) SetAutoFlush() error {
	stdOutAsyncLogger.Config.isAutoFlush = true
	return nil
}

func (stdOutAsyncLogger *StdOutAsyncLogger) Listen() error {
	var (
		err error
	)

	// for every tick, flush to logger sink
	for tick := range stdOutAsyncLogger.Config.tick.C {

		// pass the time instant to flush
		err = stdOutAsyncLogger.Flush(tick)
		if err != nil {
			return err
		}
	}

	return nil
}

func (stdOutAsyncLogger *StdOutAsyncLogger) Flush(timeStamp time.Time) error {

	fmt.Fprintf(os.Stdout, "Start of tick ===============\n")

	// while true
	for {

		// either consume the msg, or quit
		select {
		case msg := <-stdOutAsyncLogger.Config.buffer:
			{
				fmt.Fprintf(os.Stdout, "[Minute:%d, Second: %d, Milisecond:%d]\t", timeStamp.Minute(), timeStamp.Second(), timeStamp.UnixMilli())
				fmt.Fprintf(os.Stdout, "@%s:\t", stdOutAsyncLogger.Config.name)
				fmt.Fprintf(os.Stdout, "(%s)\t", severityLevels[stdOutAsyncLogger.Config.severityLevel])
				fmt.Fprintf(os.Stdout, "%s\n", msg)
			}
		case <-time.After(stdOutAsyncLogger.Config.flushTimeOut):
			fmt.Fprintf(os.Stdout, "End of tick ===============\n")
			return nil
		}
	}
}

func (stdOutAsyncLogger *StdOutAsyncLogger) Close() error {

	// we need to close the buffer
	close(stdOutAsyncLogger.Config.buffer)

	return nil
}

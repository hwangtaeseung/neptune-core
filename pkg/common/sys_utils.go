package common

import (
	"os"
	"os/signal"
	"time"
)

func WaitForShutdown(shutdownCallback func()) {

	// set interrupt
	intChannel := make(chan os.Signal, 1)
	signal.Notify(intChannel, os.Interrupt)

	// wait
	<-intChannel

	// stop network
	shutdownCallback()
}

func GetCurrentTimeAsMilli() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func GetCurrentTimeAsSec() float64 {
	return float64(time.Now().UnixNano() / int64(time.Second))
}
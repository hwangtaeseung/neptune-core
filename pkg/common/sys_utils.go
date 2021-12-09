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

func GetCurrentTime() int64 {
	return time.Now().UnixNano() / 10000000
}
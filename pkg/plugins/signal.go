package plugins

import (
	"os"
	"os/signal"
	"syscall"
)

// OS listens for OS signals SIGINT and SIGTERM and writes to the channel if either of those scenarios are encountered, nil is sent over the channel.
func OS() chan error {
	c := make(chan os.Signal, 2)
	exit := make(chan error)
	go func() {
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		defer signal.Stop(c)
		<-c
		exit <- nil
	}()
	return exit
}

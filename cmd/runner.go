package main

import (
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
)

type Runner interface {
	Run() error
	Stop() error
}

func RunApp(app Runner) error {
	done := make(chan error, 3)

	cleanupOnInterrupt(done, app)
	recoveredRun(done, app)
	return <-done
}

func cleanupOnInterrupt(done chan error, app Runner) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		s := <-c
		log.Printf("[runner] got signal %v", s)
		done <- app.Stop()
	}()
}

func recoveredRun(done chan error, app Runner) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("stacktrace from panic: \n" + string(debug.Stack()))
			done <- app.Stop()
			return
		}
	}()

	if err := app.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Printf("[runner] got error after Run %v", err)
		err = app.Stop()
		done <- err
		return
	}
	done <- app.Stop()
}

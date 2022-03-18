package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mercari/modserver"
)

const (
	exitOk  = 0
	exitErr = 1
)

func main() {
	os.Exit(start(os.Args[1:]))
}

func start(args []string) int {
	if len(args) != 2 {
		fmt.Fprintln(os.Stderr, usage())
		return exitErr
	}

	cfg := &modserver.Config{
		ModDir: args[0],
	}
	h, err := modserver.NewHandler(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return exitErr
	}

	ctx, cancel := context.WithCancel(context.Background())
	srv := &http.Server{
		Addr:        args[1],
		Handler:     h,
		BaseContext: func(_ net.Listener) context.Context { return ctx },
	}
	srv.RegisterOnShutdown(cancel)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		err := srv.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(exitErr)
		}
	}()

	<-stop

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	err = srv.Shutdown(shutdownCtx)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return exitErr
	}

	return exitOk
}

func usage() string {
	return "usage: modserver [mod-directory] [address]"
}

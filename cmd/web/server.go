package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	defaultIdleTimeout    = time.Minute
	defaultReadTimeout    = 5 * time.Second
	defaultWriteTimeout   = 10 * time.Second
	defaultShutdownPeriod = 30 * time.Second
)

func (app *application) serveHTTP() error {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.httpPort),
		Handler:      app.routes(),
		ErrorLog:     slog.NewLogLogger(app.logger.Handler(), slog.LevelWarn),
		IdleTimeout:  defaultIdleTimeout,
		ReadTimeout:  defaultReadTimeout,
		WriteTimeout: defaultWriteTimeout,
	}

	internalSrv := &http.Server{
		Addr:         fmt.Sprintf("127.0.0.1:%d", app.config.internalHttpPort),
		Handler:      app.internalRoutes(),
		ErrorLog:     slog.NewLogLogger(app.logger.Handler(), slog.LevelWarn),
		IdleTimeout:  defaultIdleTimeout,
		ReadTimeout:  defaultReadTimeout,
		WriteTimeout: defaultWriteTimeout,
	}

	shutdownErrorChan := make(chan error)

	go func() {
		quitChan := make(chan os.Signal, 1)
		signal.Notify(quitChan, syscall.SIGINT, syscall.SIGTERM)
		<-quitChan

		ctx, cancel := context.WithTimeout(context.Background(), defaultShutdownPeriod)
		srvErr := srv.Shutdown(ctx)
		defer cancel()

		ctx2, cancel2 := context.WithTimeout(context.Background(), defaultShutdownPeriod)
		internalSrvErr := internalSrv.Shutdown(ctx2)
		defer cancel2()

		shutdownErrorChan <- errors.Join(srvErr, internalSrvErr)
	}()

	app.logger.Info("starting server", slog.Group("server", "addr", srv.Addr))
	app.logger.Info("starting internal server", slog.Group("internal-server", "addr", internalSrv.Addr))

	internalErrChan := make(chan error, 1)
	go func() {
		err := internalSrv.ListenAndServe()

		if !errors.Is(err, http.ErrServerClosed) {
			internalErrChan <- err
			return
		}
		internalErrChan <- nil
	}()

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	if err := <-shutdownErrorChan; err != nil {
		return err
	}

	if err := <-internalErrChan; err != nil {
		return err
	}

	app.logger.Info("stopped server", slog.Group("server", "addr", srv.Addr))
	app.logger.Info("stopped internal server", slog.Group("internal-server", "addr", internalSrv.Addr))

	app.wg.Wait()
	return nil
}

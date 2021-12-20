package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/hb-chen/tfstate/pkg/handler"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

func main() {
	e := echo.New()
	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Logger.SetLevel(log.DEBUG)

	e.Use(middleware.BasicAuth(func(u string, p string, context echo.Context) (b bool, err error) {
		// TODO
		log.Infof("basic auth: %v, %v", u, p)
		if u == p {
			return true, nil
		}

		return false, nil
	}))

	g := e.Group("/state/:stackId")

	state := handler.NewHandler()
	g.GET("", state.Get)
	// state update
	g.POST("", state.Update)

	// state lock
	g.POST("/lock", state.Lock)
	// state unlock
	g.POST("/unlock", state.Unlock)

	// Graceful Shutdown
	// Start server
	go func() {
		if err := e.Start(":8080"); err != nil {
			e.Logger.Errorf("Shutting down the server with error:%v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 10 seconds.
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}

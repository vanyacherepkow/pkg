package http

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func NewContext() context.Context {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("context is canceled")
				cancel()
				return
			}
		}
	}()
	return ctx
}

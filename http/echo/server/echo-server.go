package echoserver

import (
	"context"
	"fmt"
	"time"

	"github.com/labstack/echo/v4"
)

const (
	MaxHeaderBytes = 1 << 20
	ReadTimeout    = 15 * time.Second
	WriteTimeout   = 15 * time.Second
)

type EchoConfig struct {
	Port        string `json:"port"`
	Host        string `json:"host"`
	Development bool   `json:"development"`
	BasePath    string `json:"basePath"`
	Timeout     int    `json:"timeout"`
}

func NewEchoServer() *echo.Echo {
	e := echo.New()
	return e
}

func RunHttpServer(ctx context.Context, echo *echo.Echo, cfg *EchoConfig) error {
	echo.Server.ReadTimeout = ReadTimeout
	echo.Server.WriteTimeout = WriteTimeout
	echo.Server.MaxHeaderBytes = MaxHeaderBytes

	go func() {
		for {
			select {
			case <-ctx.Done():
				err := echo.Shutdown(ctx)
				if err != nil {
					fmt.Printf("(Shutdown) err: {%v}", err)
					return
				}
				fmt.Print("Server exited properly")
				return
			}
		}
	}()

	err := echo.Start(cfg.Port)
	return err
}

func ApplyVersioningFromHeader(echo *echo.Echo) {
	echo.Pre(apiVersion)
}

func apiVersion(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		req := c.Request()
		header := req.Header

		apiVersion := header.Get("version")

		req.URL.Path = fmt.Sprintf("/%s%s", apiVersion, req.URL.Path)

		return next(c)
	}
}

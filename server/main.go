// Package server はサーバーアプリの実装です。
package server

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/mjibson/goon"
)

var (
	e *echo.Echo
)

func init() {
	goon.LogErrors = false

	e := echo.New()

	e.Use(middleware.Recover())
	e.Use(middleware.Gzip())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost:4200"},
		AllowCredentials: true,
	}))

	setupEntityHandlers(e.Group("/entity"))

	http.Handle("/", e)
}

func setupEntityHandlers(g *echo.Group) {
	g.GET("/", handlerEntityListGet)
	g.POST("/", handlerEntityPost)
}

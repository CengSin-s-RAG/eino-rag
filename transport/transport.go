// Package transport Echo handler，仅做 HTTP 编解码
package transport

import (
	"agent.article.fp/service"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func Start(service *service.ChatService) error {
	e := echo.New()

	e.Use(middleware.CORS())

	v2 := e.Group("/v2")
	v2.POST("/:handle", Handle())

	return e.Start(":8086")
}

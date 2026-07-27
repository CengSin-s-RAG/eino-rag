package transport

import (
	"agent.article.fp/service"
	"context"
	"github.com/labstack/echo/v4"
	"net/http"
)

var (
	reqBindMap = make(map[string]service.Handle)
)

func init() {
	reqBindMap = map[string]service.Handle{
		"chat": func(ctx context.Context, req service.ChatRequest) (*service.ChatResponse, error) {
			return &service.ChatResponse{}, nil
		},
		"rewrite": func(ctx context.Context, req service.ChatRequest) (*service.ChatResponse, error) {
			return &service.ChatResponse{}, nil
		},
		"rerank": func(ctx context.Context, req service.ChatRequest) (*service.ChatResponse, error) {
			return &service.ChatResponse{}, nil
		},
	}
}

func Handle() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		handle, ok := reqBindMap[c.Param("handle")]
		if !ok {
			return echo.NewHTTPError(http.StatusBadRequest)
		}
		req := service.ChatRequest{}
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest)
		}

		resp, err := handle(ctx, req)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest)
		}

		return c.JSON(http.StatusOK, resp)
	}
}

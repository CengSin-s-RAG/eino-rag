// Package transport adapts HTTP requests to the framework-neutral service API.
package transport

import (
	"agent.article.fp/service"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Server struct {
	chat *service.ChatService
	aux  *service.AuxiliaryService
	echo *echo.Echo
}

func New(chat *service.ChatService, aux *service.AuxiliaryService) *Server {
	e := echo.New()
	e.Use(middleware.CORS())
	server := &Server{chat: chat, aux: aux, echo: e}
	v2 := e.Group("/v2")
	v2.POST("/chat", server.handleChat)
	v2.POST("/rerank", server.handleRerank)
	v2.POST("/rewrite", server.handleRewrite)
	sessions := v2.Group("/session")
	sessions.POST("/new", server.handleNewSession)
	sessions.GET("/list", server.handleListSessions)
	sessions.GET("/history", server.handleHistory)
	return server
}

func (s *Server) Start(address string) error {
	return s.echo.Start(address)
}

func (s *Server) Close() error { return s.echo.Close() }

func (s *Server) handleChat(c echo.Context) error {
	var request service.ChatRequest
	if err := c.Bind(&request); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid JSON request")
	}
	response, err := s.chat.Chat(c.Request().Context(), request)
	if err != nil {
		if c.Request().Context().Err() != nil {
			return c.Request().Context().Err()
		}
		return echo.NewHTTPError(http.StatusBadGateway, err.Error())
	}
	return c.JSON(http.StatusOK, response)
}

func (s *Server) handleRewrite(c echo.Context) error {
	var request service.RewriteRequest
	if err := c.Bind(&request); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid JSON request")
	}
	response, err := s.aux.Rewrite(c.Request().Context(), request)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadGateway, err.Error())
	}
	return c.JSON(http.StatusOK, response.Query)
}

func (s *Server) handleRerank(c echo.Context) error {
	var request service.RerankRequest
	if err := c.Bind(&request); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid JSON request")
	}
	response, err := s.aux.Rerank(c.Request().Context(), request)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadGateway, err.Error())
	}
	return c.JSON(http.StatusOK, response)
}

func (s *Server) handleNewSession(c echo.Context) error {
	session, err := s.aux.CreateSession(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, session)
}

func (s *Server) handleListSessions(c echo.Context) error {
	var query struct {
		Offset int64 `query:"offset"`
		Limit  int64 `query:"limit"`
	}
	if err := c.Bind(&query); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid query")
	}
	sessions, err := s.aux.ListSessions(c.Request().Context(), query.Offset, query.Limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]any{"sessions": sessions})
}

func (s *Server) handleHistory(c echo.Context) error {
	id := c.QueryParam("session_id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "session_id is required")
	}
	messages, err := s.aux.History(c.Request().Context(), id, 100)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, messages)
}

func notImplemented(c echo.Context) error {
	return echo.NewHTTPError(http.StatusNotImplemented, "this endpoint is not migrated yet")
}

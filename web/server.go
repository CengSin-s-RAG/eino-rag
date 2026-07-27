// Package web embeds and serves the standalone frontend with the application.
package web

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed index.html styles.css app.js
var assets embed.FS

type Server struct{ server *http.Server }

func New(address string) (*Server, error) {
	files, err := fs.Sub(assets, ".")
	if err != nil {
		return nil, err
	}
	return &Server{server: &http.Server{Addr: address, Handler: http.FileServer(http.FS(files))}}, nil
}

func (s *Server) Start() error { return s.server.ListenAndServe() }
func (s *Server) Close() error { return s.server.Close() }

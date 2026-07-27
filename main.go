package main

import (
	"agent.article.fp/bootstrap"
	"agent.article.fp/config"
	"agent.article.fp/transport"
	"agent.article.fp/web"
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg, err := config.Load("./config/config.yaml")
	if err != nil {
		log.Fatal("load config: ", err)
	}
	app, err := bootstrap.New(context.Background(), cfg)
	if err != nil {
		log.Fatal("build application: ", err)
	}
	defer func() {
		if err := app.Close(); err != nil {
			log.Printf("close application: %v", err)
		}
	}()

	server := transport.New(app.Chat, app.Aux)
	defer func() {
		if err := server.Close(); err != nil {
			log.Printf("close HTTP server: %v", err)
		}
	}()
	frontend, err := web.New(":3000")
	if err != nil {
		log.Fatal("create frontend server: ", err)
	}
	defer func() {
		if err := frontend.Close(); err != nil {
			log.Printf("close frontend server: %v", err)
		}
	}()
	errCh := make(chan error, 2)
	go func() { errCh <- server.Start(":8086") }()
	go func() { errCh <- frontend.Start() }()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	select {
	case err := <-errCh:
		if err != nil {
			log.Fatal("start HTTP server: ", err)
		}
	case <-signals:
		return
	}
}

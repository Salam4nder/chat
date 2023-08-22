package main

import (
	"log"
	"os"
	"time"

	"github.com/Salam4nder/chat/internal/chat"
	"github.com/Salam4nder/chat/internal/config"
	"github.com/Salam4nder/chat/internal/http"
	"github.com/rs/zerolog"
)

const (
	// ReadTimeout is the maximum duration for reading the entire
	// request, including the body.
	ReadTimeout = 10 * time.Second
	// WriteTimeout is the maximum duration before timing out
	// writes of the response. It is reset whenever a new
	// request's header is read.
	WriteTimeout = 10 * time.Second
)

func main() {
	chat.Rooms = make(map[string]*chat.Room)

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	config, err := config.New()
	fatalOnError(err)

	server := http.New().WithOptions(
		http.WithLogger(logger),
		http.WithAddr(config.HTTPServer.Addr()),
		http.WithHandler(nil),
		http.WithTimeout(ReadTimeout, WriteTimeout),
	)

	http.InitRoutes()

	if err := server.Serve(); err != nil {
		fatalOnError(err)
	}
}

func fatalOnError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

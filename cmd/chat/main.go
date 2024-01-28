package main

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Salam4nder/chat/internal/chat"
	"github.com/Salam4nder/chat/internal/config"
	"github.com/Salam4nder/chat/internal/db/cql"
	db "github.com/Salam4nder/chat/internal/db/keyspace/chat"
	"github.com/Salam4nder/chat/internal/event"
	"github.com/Salam4nder/chat/internal/http/handler/health"
	"github.com/Salam4nder/chat/internal/http/handler/websocket"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	// scyllaTimeout is the maximum duration to wait for a ScyllaDB connection.
	scyllaTimeout = 30 * time.Second
	// natsTimeout is the maximum duration to wait for a NATS connection.
	natsTimeout = 5 * time.Second
	// httpReadTimeout is the maximum duration for reading the entire
	// request, including the body.
	httpReadTimeout = 10 * time.Second
	// httpWriteTimeout is the maximum duration before timing out
	// writes of the response. It is reset whenever a new
	// request's header is read.
	httpWriteTimeout = 10 * time.Second
	// environmentDev is the development environment.
	environmentDev = "dev"
)

func main() {
	interruptCh := make(chan os.Signal, 1)
	signal.Notify(interruptCh, os.Interrupt, syscall.SIGTERM)

	// Config.
	// UNIX Time is faster and smaller than most timestamps
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	config, err := config.New()
	exitOnError(err)
	if config.Environment == environmentDev {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
	go config.Watch()

	// ScyllaDB.
	cluster := cql.NewClusterConfig(config.ScyllaDB)
	err = cluster.PingWithTimeout(scyllaTimeout, interruptCh)
	exitOnError(err)
	scyllaSession, err := cluster.Inner().CreateSession()
	exitOnError(err)

	// NATS.
	natsClient, err := nats.Connect(
		config.NATS.Addr(),
		nats.Timeout(natsTimeout),
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(20),
	)
	exitOnError(err)

	// Repos.
	// userRepo := db.NewScyllaUserRepository(scyllaSession)
	messageRepo := db.NewScyllaMessageRepository(scyllaSession)

	// In-memory event registry.
	eventRegistry := event.NewRegistry()

	// Services.
	messageService := chat.NewMessageService(messageRepo, natsClient)
	sessionService := chat.NewSessionService(natsClient, eventRegistry)

	// Concurrent-safe map of chat rooms.
	chat.Rooms = make(map[string]*chat.Room)

	// Subscribers.
	eventRegistry.Subscribe(chat.SessionConnectedEvent, sessionService.HandleSessionConnectedEvent)
	eventRegistry.Subscribe(chat.MessageCreatedInRoomEvent, messageService.HandleMessageCreatedInRoomEvent)

	natsChan := make(chan *nats.Msg, 64)
	messageSub, err := natsClient.ChanSubscribe(chat.MessageCreatedInRoomEvent, natsChan)
	exitOnError(err)

	go func(ch chan os.Signal) {
		for {
			select {
			case msg := <-natsChan:
				if msg == nil {
					return
				}
				var message chat.Message
				if err := gob.NewDecoder(bytes.NewReader(msg.Data)).
					Decode(&message); err != nil {
					log.Error().Err(err).Msg("failed to decode message")
				}
			case <-ch:
				return
			}
		}
	}(interruptCh)

	// HTTP server.
	server := &http.Server{
		Addr:         config.HTTPServer.Addr(),
		Handler:      nil,
		ReadTimeout:  httpReadTimeout,
		WriteTimeout: httpWriteTimeout,
	}
	healthHandler := health.NewHandler(scyllaSession)
	websocketHandler := websocket.NewHandler(eventRegistry)
	http.HandleFunc("/health", healthHandler.Health)
	http.HandleFunc("/chat", websocketHandler.HandleConnect)
	go func() {
		log.Info().
			Str("addr", config.HTTPServer.Addr()).
			Msg("main: serving http server...")

		if err := server.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				exitOnError(err)
			}
		}
	}()
	log.Info().Str("service", config.ServiceName).Send()

	<-interruptCh
	log.Info().Msg("main: cleaning up...")
	scyllaSession.Close()
	if err := messageSub.Unsubscribe(); err != nil {
		log.Error().Err(err).Msg("main: failed to unsubscribe from nats")
	}
	close(natsChan)
	natsClient.Close()
	if err = server.Shutdown(context.Background()); err != nil {
		log.Error().
			Err(err).
			Msg("main: failed to shutdown HTTP server")
	}
	log.Info().Msg("main: cleanup finished")

	os.Exit(0)
}

func exitOnError(err error) {
	if err != nil {
		log.Error().
			Err(err).
			Msg("main: failed to start service")
		os.Exit(1)
	}
}

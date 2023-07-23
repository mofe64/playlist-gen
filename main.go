package main

import (
	"context"
	"errors"
	"log"
	"mofe64/playlistGen/config"
	"mofe64/playlistGen/routes"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	router.Use(gin.Recovery())

	// Set up ping route
	router.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	// auth routes
	routes.AuthorizationRoute(router)

	// Create Custom Server
	server := &http.Server{
		Addr:    ":" + config.EnvHTTPPort(),
		Handler: router,
	}

	// Launch the server in a Goroutine (concurrent execution)
	go func() {
		log.Println("Starting server on port " + config.EnvHTTPPort())
		if err := server.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			log.Printf("Error listen %v\n", err)
		}
	}()

	// Create a channel to receive OS signals (e.g., interrupt or termination)
	c := make(chan os.Signal, 1)
	// Register OS signals to be captured by the channel
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, os.Kill)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	// Block until a signal is received.
	// Wait until a signal is received on the channel
	sig := <-c
	log.Println("Received signal:", sig)

	// Create a new context with a 30-second timeout for server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	// Ensure the context is canceled when the main function exits
	defer cancel()

	// Initiate a graceful shutdown of the server, allowing existing connections to complete
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
	log.Println("Server exiting ....")

}

package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/gorilla/websocket"
)

func main() {
	// Parse command-line flags
	// addr for address
	// n for number of connections
	addr := flag.String("addr", "localhost:8080", "http service address")
	numOfConnections := flag.Int("n", 1, "number of websocket connections")
	flag.Parse()
	log.SetFlags(0)

	// URL configuration of websocket endpoint
	u := url.URL{Scheme: "ws", Host: *addr, Path: "/goapp/ws"}
	log.Printf("Connecting to WebSocket URL: %s", u.String())

	// Channel to handle termination signals
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Context for handling graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Synchronize the shutdown process
	var wg sync.WaitGroup

	// Define a function to manage a single WebSocket connection
	connectWebSocket := func(connID int, u url.URL, ctx context.Context, wg *sync.WaitGroup) {
		defer wg.Done()

		// Set up a custom dialer with the Origin header
		dialer := websocket.DefaultDialer
		header := http.Header{}
		header.Add("Origin", "http://localhost:8080") // Replace with your server's allowed origin

		// Connect to the WebSocket server
		c, _, err := dialer.Dial(u.String(), header)
		if err != nil {
			log.Printf("conn #%d: failed to connect: %v", connID, err)
			return
		}
		defer c.Close()

		// Channel to signal when the connection should be closed
		done := make(chan struct{})

		// Read messages from the server
		go func() {
			defer close(done)
			for {
				select {
				case <-ctx.Done():
					return
				default:
					_, message, err := c.ReadMessage()
					if err != nil {
						log.Printf("conn #%d: read error: %v", connID, err)
						return
					}
					log.Printf("[conn #%d] response: %s", connID, message)
				}
			}
		}()

		// Handle graceful shutdown when context is canceled
		<-ctx.Done()
		log.Printf("conn #%d: received interrupt signal, closing connection", connID)
	}

	for i := 0; i < *numOfConnections; i++ {
		wg.Add(1)
		go connectWebSocket(i, u, ctx, &wg)
	}

	// Wait for an interrupt signal
	<-interrupt
	log.Println("Interrupt signal received, shutting down...")

	// Cancel the context to signal all goroutines to stop
	cancel()

	// Wait for all connections to close
	wg.Wait()

	log.Println("All connections closed. Exiting.")
}

// main is the entrypoint for the NetPilot application.
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yuzjing/netpilot/backend/internal/qos"
	"github.com/yuzjing/netpilot/backend/internal/server"
)

func main() {
	// 1. Create the core business logic component.
	qosManager := qos.NewManager()

	// 2. Create the HTTP server and inject the qosManager into it.
	httpServer := server.New(qosManager)

	// 3. Setup the server with graceful shutdown.
	// The variable is named 'srv'.
	srv := &http.Server{
		Addr:    ":8080",
		Handler: httpServer,
	}

	// Create a listener that supports both IPv4 and IPv6.
	lc := net.ListenConfig{}
	ln, err := lc.Listen(context.Background(), "tcp", srv.Addr)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	fmt.Println("NetPilot API server is running on http://localhost:8080")

	// Run server in a goroutine so it doesn't block.
	// We use 'srv' here.
	go func() {
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not serve: %v", err)
		}
	}()

	// Wait for an interrupt signal to gracefully shutdown the server.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// The context for the shutdown.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// We use 'srv' here as well.
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}

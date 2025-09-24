// main is the entrypoint for the NetPilot application.
package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yuzjing/netpilot/backend/qos"
	"github.com/yuzjing/netpilot/backend/server"
)

// The embed path is now simple and valid, relative to this file.
//
//go:embed all:frontend/build
var embeddedFiles embed.FS

func main() {
	// Create a sub-filesystem for our static assets.
	staticFS, err := fs.Sub(embeddedFiles, "frontend/build")
	if err != nil {
		log.Fatal(err)
	}

	// Create components and inject dependencies.
	qosManager := qos.NewManager()
	httpServer := server.New(qosManager, staticFS)

	// Setup and run the server.
	srv := &http.Server{
		Addr:    ":8080",
		Handler: httpServer,
	}

	lc := net.ListenConfig{}
	ln, err := lc.Listen(context.Background(), "tcp", srv.Addr)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	fmt.Println("NetPilot API server is running on http://localhost:8080")
	go func() {
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not serve: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
	log.Println("Server exiting")
}

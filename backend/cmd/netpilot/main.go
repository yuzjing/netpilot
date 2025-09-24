package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// QosRuleRequest represents the JSON body for applying a QoS rule.
type QosRuleRequest struct {
	Interface string `json:"interface"`
	Algorithm string `json:"algorithm"`
	Settings  struct {
		BandwidthMbit uint32 `json:"bandwidth_mbit"`
	} `json:"settings"`
}

// qosRuleHandler routes requests for /api/qos/rules based on the HTTP method.
func qosRuleHandler(w http.ResponseWriter, r *http.Request) {
	// Set common headers, including CORS for the frontend.
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle browser preflight requests.
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	switch r.Method {
	case "POST":
		handleApplyQoS(w, r)
	case "GET":
		handleGetQoS(w, r)
	case "DELETE":
		handleDeleteQoS(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleApplyQoS creates or updates a QoS rule.
func handleApplyQoS(w http.ResponseWriter, r *http.Request) {
	var req QosRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	switch req.Algorithm {
	case "cake":
		log.Printf("Applying CAKE to %s with bandwidth %d Mbit", req.Interface, req.Settings.BandwidthMbit)
		// TODO: Call the actual qos.ApplyCakeQoS implementation.
		// For example:
		// err := qos.Apply(req.Interface, req.Algorithm, req.Settings)
	default:
		http.Error(w, fmt.Sprintf("algorithm %s not supported", req.Algorithm), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "QoS rule applied successfully"})
}

// handleGetQoS returns the current QoS status.
func handleGetQoS(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement the logic to fetch the actual QoS rule from the system.
	mockResponse := map[string]interface{}{
		"interface": "eth0",
		"algorithm": "cake",
		"settings": map[string]string{
			"bandwidth": "500mbit",
		},
	}
	json.NewEncoder(w).Encode(mockResponse)
}

// handleDeleteQoS removes an existing QoS rule.
func handleDeleteQoS(w http.ResponseWriter, r *http.Request) {
	iface := r.URL.Query().Get("interface")
	if iface == "" {
		http.Error(w, "missing interface parameter", http.StatusBadRequest)
		return
	}

	log.Printf("Deleting QoS rule from %s", iface)
	// TODO: Call the actual implementation to delete the QoS rule.
	// For example:
	// err := qos.Delete(iface)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "QoS rule deleted successfully"})
}

// main sets up the routes and starts the server.
func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/qos/rules", qosRuleHandler)
	mux.HandleFunc("/api/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		fmt.Fprintln(w, "pong")
	})

	// Setup a server with graceful shutdown.
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// Create a listener that supports both IPv4 and IPv6.
	lc := net.ListenConfig{}
	ln, err := lc.Listen(context.Background(), "tcp", server.Addr)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	fmt.Println("NetPilot API server is running on http://localhost:8080")

	// Run server in a goroutine so it doesn't block.
	go func() {
		if err := server.Serve(ln); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not serve: %v", err)
		}
	}()

	// Wait for an interrupt signal to gracefully shutdown the server.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")

}

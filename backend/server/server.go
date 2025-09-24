// Package server handles all HTTP-related logic for the NetPilot application.
// It defines routes and connects them to the core QoS business logic.
package server

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"

	"github.com/yuzjing/netpilot/backend/qos"
)

// Server holds all dependencies for the HTTP server, like the qosManager.
type Server struct {
	qosManager *qos.Manager
	router     *http.ServeMux
	staticFS   fs.FS
}

// New creates a new Server instance and registers all its routes.
func New(qosManager *qos.Manager, staticFS fs.FS) *Server {
	s := &Server{
		qosManager: qosManager,
		router:     http.NewServeMux(),
		staticFS:   staticFS,
	}
	s.registerRoutes()
	return s
}

// ServeHTTP makes our Server compatible with the standard http.Server,
// allowing it to be used as a handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// registerRoutes is the central place where all API endpoints are defined.
func (s *Server) registerRoutes() {

	s.router.HandleFunc("/api/qos/rules", s.qosRuleHandler())
	s.router.HandleFunc("/api/ping", s.pingHandler())
	s.router.HandleFunc("/api/interfaces", s.interfacesHandler())

	if s.staticFS != nil {
		fileServer := http.FileServer(http.FS(s.staticFS))
		s.router.HandleFunc("/", fileServer.ServeHTTP)
	}
}

// qosRuleHandler is a master handler that routes different HTTP methods
// for the /api/qos/rules endpoint to their specific implementations.
func (s *Server) qosRuleHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set common headers, especially for CORS to allow the frontend to connect.
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle browser preflight requests for CORS.
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Route to the appropriate handler based on the request method.
		switch r.Method {
		case http.MethodPost:
			s.handleApplyQoS(w, r)
		case http.MethodGet:
			s.handleGetQoS(w, r)
		case http.MethodDelete:
			s.handleDeleteQoS(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// handleApplyQoS handles the POST request to create or update a QoS rule.
func (s *Server) handleApplyQoS(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Interface string `json:"interface"`
		Algorithm string `json:"algorithm"`
		Settings  struct {
			BandwidthMbit uint32 `json:"bandwidth_mbit"`
		} `json:"settings"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	rule := qos.Rule{
		Interface: req.Interface,
		Algorithm: qos.Algorithm(req.Algorithm),
		Settings: qos.Settings{
			"bandwidth_mbit": float64(req.Settings.BandwidthMbit),
		},
	}

	if err := s.qosManager.ApplyRule(rule); err != nil {
		log.Printf("Error applying QoS rule: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully applied %s rule to %s", rule.Algorithm, rule.Interface)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "QoS rule applied successfully"})
}

// handleGetQoS handles the GET request to fetch the current QoS rule for an interface.
func (s *Server) handleGetQoS(w http.ResponseWriter, r *http.Request) {
	iface := r.URL.Query().Get("interface")
	if iface == "" {
		http.Error(w, "missing interface parameter", http.StatusBadRequest)
		return
	}

	rule, err := s.qosManager.GetRule(iface)
	if err != nil {
		log.Printf("Error getting QoS rule: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if rule == nil {
		w.WriteHeader(http.StatusNoContent) // Use 204 No Content when no rule is found.
		return
	}

	json.NewEncoder(w).Encode(rule)
}

// handleDeleteQoS handles the DELETE request to remove a QoS rule from an interface.
func (s *Server) handleDeleteQoS(w http.ResponseWriter, r *http.Request) {
	iface := r.URL.Query().Get("interface")
	if iface == "" {
		http.Error(w, "missing interface parameter", http.StatusBadRequest)
		return
	}

	if err := s.qosManager.DeleteRule(iface); err != nil {
		log.Printf("Error deleting QoS rule: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully deleted QoS rule from %s", iface)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "QoS rule deleted successfully"})
}

// interfacesHandler handles the GET request to list available network interfaces.
func (s *Server) interfacesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		interfaces, err := net.Interfaces()
		if err != nil {
			log.Printf("Error fetching interfaces: %v", err)
			http.Error(w, "Failed to fetch network interfaces", http.StatusInternalServerError)
			return
		}
		var names []string
		for _, i := range interfaces {
			// Filter out loopback and non-running interfaces for a cleaner list.
			if i.Flags&net.FlagLoopback == 0 && i.Flags&net.FlagUp != 0 {
				names = append(names, i.Name)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		json.NewEncoder(w).Encode(names)
	}
}

// pingHandler is a simple health-check endpoint.
func (s *Server) pingHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		fmt.Fprintln(w, "pong")
	}
}

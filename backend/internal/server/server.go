// Package server handles all HTTP-related logic.
package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/yuzjing/netpilot/backend/internal/qos"
)

// Server holds all dependencies for the HTTP server.
type Server struct {
	qosManager *qos.Manager
	router     *http.ServeMux
}

// New creates a new Server and registers its routes.
func New(qosManager *qos.Manager) *Server {
	s := &Server{
		qosManager: qosManager,
		router:     http.NewServeMux(),
	}
	s.registerRoutes()
	return s
}

// ServeHTTP makes our Server compatible with http.ListenAndServe.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// registerRoutes sets up all the API endpoints.
func (s *Server) registerRoutes() {
	s.router.HandleFunc("/api/qos/rules", s.qosRuleHandler())
	s.router.HandleFunc("/api/ping", s.pingHandler())
}

// qosRuleHandler returns a handler function for the QoS endpoint.
func (s *Server) qosRuleHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set common headers for CORS
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		switch r.Method {
		case "POST":
			s.handleApplyQoS(w, r)
		// TODO: Add GET and DELETE cases
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// handleApplyQoS uses the qosManager to apply a rule.
func (s *Server) handleApplyQoS(w http.ResponseWriter, r *http.Request) {
	var qosRule qos.Rule
	if err := json.NewDecoder(r.Body).Decode(&qosRule); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 调用被注入的qosManager
	if err := s.qosManager.ApplyRule(qosRule); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "QoS rule applied successfully"})
}

// pingHandler is a simple health check endpoint.
func (s *Server) pingHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		fmt.Fprintln(w, "pong")
	}
}

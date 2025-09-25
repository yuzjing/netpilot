// Package server handles all HTTP-related logic for the NetPilot application.
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

// Server holds all dependencies for the HTTP server.
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

// ServeHTTP makes our Server compatible with the standard http.Server.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// registerRoutes sets up all API endpoints and the static file server.
func (s *Server) registerRoutes() {
	fileServer := http.FileServer(http.FS(s.staticFS))
	s.router.HandleFunc("/api/qos/rules", s.qosRuleHandler())
	s.router.HandleFunc("/api/ping", s.pingHandler())
	s.router.HandleFunc("/api/interfaces", s.interfacesHandler())
	s.router.HandleFunc("/", fileServer.ServeHTTP)
}

// --- Handler Implementations ---

func (s *Server) qosRuleHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
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
	rule := qos.Rule{Interface: req.Interface, Algorithm: qos.Algorithm(req.Algorithm), Settings: qos.Settings{"bandwidth_mbit": float64(req.Settings.BandwidthMbit)}}
	if err := s.qosManager.ApplyRule(rule); err != nil {
		log.Printf("Error applying QoS rule: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("Successfully applied %s rule to %s", rule.Algorithm, rule.Interface)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "QoS rule applied successfully"})
}

// 【核心修正】This function now correctly handles the *qos.Rule object.
func (s *Server) handleGetQoS(w http.ResponseWriter, r *http.Request) {
	iface := r.URL.Query().Get("interface")
	if iface == "" {
		http.Error(w, "missing interface parameter", http.StatusBadRequest)
		return
	}

	// 1. Call the manager, which returns a structured rule (*qos.Rule) or nil.
	rule, err := s.qosManager.GetRule(iface)
	if err != nil {
		log.Printf("Error getting QoS rule: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 2. If no rule is found, manager returns nil. We send 204 No Content.
	if rule == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// 3. If a rule IS found, we encode the ENTIRE rule object directly into JSON.
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(rule)
}

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
			if i.Flags&net.FlagLoopback == 0 && i.Flags&net.FlagUp != 0 {
				names = append(names, i.Name)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		json.NewEncoder(w).Encode(names)
	}
}

func (s *Server) pingHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		fmt.Fprintln(w, "pong")
	}
}

package api

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/tavp-stack/tavpbox/internal/config"
	"github.com/tavp-stack/tavpbox/internal/lxd"
)

type Server struct {
	socketPath string
	listener   net.Listener
}

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func NewServer() *Server {
	socketPath := filepath.Join(os.TempDir(), "tavpbox.sock")
	return &Server{socketPath: socketPath}
}

func (s *Server) Start() error {
	// Remove existing socket
	os.Remove(s.socketPath)

	listener, err := net.Listen("unix", s.socketPath)
	if err != nil {
		return fmt.Errorf("failed to start API server: %w", err)
	}
	s.listener = listener

	mux := http.NewServeMux()
	mux.HandleFunc("/api/boxes", s.handleBoxes)
	mux.HandleFunc("/api/boxes/", s.handleBox)
	mux.HandleFunc("/api/status", s.handleStatus)
	mux.HandleFunc("/api/plugins", s.handlePlugins)
	mux.HandleFunc("/api/init", s.handleInit)

	server := &http.Server{Handler: mux}
	go server.Serve(s.listener)

	fmt.Printf("API server listening on %s\n", s.socketPath)
	return nil
}

func (s *Server) Stop() {
	if s.listener != nil {
		s.listener.Close()
	}
	os.Remove(s.socketPath)
}

func (s *Server) handleBoxes(w http.ResponseWriter, r *http.Request) {
	client := lxd.New()

	switch r.Method {
	case "GET":
		containers, err := client.List()
		if err != nil {
			s.jsonResponse(w, Response{Success: false, Error: err.Error()})
			return
		}
		s.jsonResponse(w, Response{Success: true, Data: containers})

	default:
		s.jsonResponse(w, Response{Success: false, Error: "method not allowed"})
	}
}

func (s *Server) handleBox(w http.ResponseWriter, r *http.Request) {
	client := lxd.New()
	name := r.URL.Path[len("/api/boxes/"):]

	switch r.Method {
	case "GET":
		containerName := client.ContainerName(name)
		ip, _ := client.GetIP(containerName)
		s.jsonResponse(w, Response{Success: true, Data: map[string]string{
			"name": name,
			"ip":   ip,
		}})

	case "POST":
		action := r.URL.Query().Get("action")
		containerName := client.ContainerName(name)
		var err error
		switch action {
		case "start":
			err = client.Start(containerName)
		case "stop":
			err = client.Stop(containerName)
		default:
			s.jsonResponse(w, Response{Success: false, Error: "unknown action"})
			return
		}
		if err != nil {
			s.jsonResponse(w, Response{Success: false, Error: err.Error()})
			return
		}
		s.jsonResponse(w, Response{Success: true})

	case "DELETE":
		containerName := client.ContainerName(name)
		if err := client.Delete(containerName); err != nil {
			s.jsonResponse(w, Response{Success: false, Error: err.Error()})
			return
		}
		s.jsonResponse(w, Response{Success: true})

	default:
		s.jsonResponse(w, Response{Success: false, Error: "method not allowed"})
	}
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	client := lxd.New()
	containers, err := client.List()
	if err != nil {
		s.jsonResponse(w, Response{Success: false, Error: err.Error()})
		return
	}
	s.jsonResponse(w, Response{Success: true, Data: map[string]interface{}{
		"boxes": len(containers),
	}})
}

func (s *Server) handlePlugins(w http.ResponseWriter, r *http.Request) {
	home := config.HomeDir()
	stacks, _ := filepath.Glob(filepath.Join(home, "plugins", "stacks", "*.yml"))
	services, _ := filepath.Glob(filepath.Join(home, "plugins", "services", "*.yml"))

	s.jsonResponse(w, Response{Success: true, Data: map[string]interface{}{
		"stacks":   len(stacks),
		"services": len(services),
	}})
}

func (s *Server) handleInit(w http.ResponseWriter, r *http.Request) {
	s.jsonResponse(w, Response{Success: true, Data: map[string]string{
		"status": "ready",
	}})
}

func (s *Server) jsonResponse(w http.ResponseWriter, resp Response) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

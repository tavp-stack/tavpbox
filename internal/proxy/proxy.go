package proxy

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/tavp-stack/tavpbox/internal/certs"
)

type Route struct {
	Domain string `json:"domain"`
	IP     string `json:"ip"`
	Port   int    `json:"port"`
}

type Proxy struct {
	mu     sync.RWMutex
	routes []Route
	port   int
}

func New(port int) *Proxy {
	return &Proxy{port: port}
}

func (p *Proxy) routesFile() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".tavpbox", "proxy", "routes.json")
}

func (p *Proxy) loadRoutes() {
	data, err := os.ReadFile(p.routesFile())
	if err != nil {
		return
	}
	json.Unmarshal(data, &p.routes)
}

func (p *Proxy) saveRoutes() error {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".tavpbox", "proxy")
	os.MkdirAll(dir, 0755)
	data, _ := json.MarshalIndent(p.routes, "", "  ")
	return os.WriteFile(p.routesFile(), data, 0644)
}

func (p *Proxy) AddRoute(domain, ip string, port int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Load existing routes from disk first
	p.loadRoutesFromDisk()

	// Remove existing route for this domain
	var newRoutes []Route
	for _, r := range p.routes {
		if r.Domain != domain {
			newRoutes = append(newRoutes, r)
		}
	}
	newRoutes = append(newRoutes, Route{Domain: domain, IP: ip, Port: port})
	p.routes = newRoutes
	p.saveRoutes()
}

func (p *Proxy) loadRoutesFromDisk() {
	data, err := os.ReadFile(p.routesFile())
	if err != nil {
		return
	}
	if string(data) == "null" || len(strings.TrimSpace(string(data))) == 0 {
		return
	}
	json.Unmarshal(data, &p.routes)
}

func (p *Proxy) RemoveRoute(domain string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	var newRoutes []Route
	for _, r := range p.routes {
		if r.Domain != domain {
			newRoutes = append(newRoutes, r)
		}
	}
	p.routes = newRoutes
	p.saveRoutes()
}

func (p *Proxy) Routes() []Route {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.routes
}

func (p *Proxy) Start() error {
	p.loadRoutes()

	// Watch routes.json for changes
	go p.watchRoutes()

	mux := http.NewServeMux()
	mux.HandleFunc("/", p.handler)

	// Build TLS config using mkcert wildcard cert
	tlsConfig := &tls.Config{
		GetCertificate: func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
			certPath, keyPath := certs.GetWildcardCert("tavp.my.id")
			if certPath == "" {
				return nil, fmt.Errorf("no cert found, run: tavpbox setup")
			}
			cert, err := tls.LoadX509KeyPair(certPath, keyPath)
			if err != nil {
				return nil, err
			}
			return &cert, nil
		},
	}

	// Start HTTP on port 80
	go func() {
		fmt.Printf("TAVPBox proxy HTTP on :80\n")
		http.ListenAndServe(":80", mux)
	}()

	// Start HTTPS on port 443
	fmt.Printf("TAVPBox proxy HTTPS on :443\n")
	server := &http.Server{
		Addr:      ":443",
		Handler:   mux,
		TLSConfig: tlsConfig,
	}
	return server.ListenAndServeTLS("", "")
}

// watchRoutes periodically checks for changes to routes.json
func (p *Proxy) watchRoutes() {
	var lastMod time.Time
	routesFile := p.routesFile()

	for {
		time.Sleep(2 * time.Second)
		info, err := os.Stat(routesFile)
		if err != nil {
			continue
		}
		if info.ModTime() != lastMod {
			lastMod = info.ModTime()
			p.loadRoutes()
		}
	}
}

func (p *Proxy) handler(w http.ResponseWriter, r *http.Request) {
	host := r.Host
	if idx := strings.LastIndex(host, ":"); idx != -1 {
		host = host[:idx]
	}

	p.mu.RLock()
	var route *Route
	for _, rt := range p.routes {
		if rt.Domain == host {
			route = &rt
			break
		}
	}
	p.mu.RUnlock()

	if route == nil {
		http.Error(w, "TAVPBox — No project configured for "+host, http.StatusNotFound)
		return
	}

	target := fmt.Sprintf("http://%s:%d", route.IP, route.Port)
	proxyURL, err := url.Parse(target)
	if err != nil {
		http.Error(w, "Bad gateway", http.StatusBadGateway)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(proxyURL)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
	}
	proxy.ServeHTTP(w, r)
}

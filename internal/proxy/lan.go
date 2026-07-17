package proxy

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

const (
	LanPortMin = 8081
	LanPortMax = 8999
)

type LanPortEntry struct {
	Project string `json:"project"`
	Port    int    `json:"port"`
}

type LanPortManager struct {
	mu      sync.Mutex
	entries []LanPortEntry
	file    string
}

func NewLanPortManager() *LanPortManager {
	home, _ := os.UserHomeDir()
	return &LanPortManager{
		file: filepath.Join(home, ".tavpbox", "lan-ports.json"),
	}
}

func (m *LanPortManager) load() {
	data, err := os.ReadFile(m.file)
	if err != nil {
		return
	}
	json.Unmarshal(data, &m.entries)
}

func (m *LanPortManager) save() error {
	dir := filepath.Dir(m.file)
	os.MkdirAll(dir, 0755)
	data, _ := json.MarshalIndent(m.entries, "", "  ")
	return os.WriteFile(m.file, data, 0644)
}

// GetOrAssign returns the LAN port for a project, or assigns a new one
func (m *LanPortManager) GetOrAssign(project string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.load()

	// Check if already assigned
	for _, e := range m.entries {
		if e.Project == project {
			return e.Port, nil
		}
	}

	// Find next available port
	usedPorts := make(map[int]bool)
	for _, e := range m.entries {
		usedPorts[e.Port] = true
	}

	for port := LanPortMin; port <= LanPortMax; port++ {
		if !usedPorts[port] {
			m.entries = append(m.entries, LanPortEntry{Project: project, Port: port})
			if err := m.save(); err != nil {
				return 0, err
			}
			return port, nil
		}
	}

	return 0, fmt.Errorf("no available LAN ports (all %d-%d used)", LanPortMin, LanPortMax)
}

// Get returns the LAN port for a project (0 if not assigned)
func (m *LanPortManager) Get(project string) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.load()

	for _, e := range m.entries {
		if e.Project == project {
			return e.Port
		}
	}
	return 0
}

// Release removes the LAN port assignment for a project
func (m *LanPortManager) Release(project string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.load()

	var newEntries []LanPortEntry
	for _, e := range m.entries {
		if e.Project != project {
			newEntries = append(newEntries, e)
		}
	}
	m.entries = newEntries
	m.save()
}

// All returns all LAN port assignments
func (m *LanPortManager) All() []LanPortEntry {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.load()

	sort.Slice(m.entries, func(i, j int) bool {
		return m.entries[i].Port < m.entries[j].Port
	})
	return m.entries
}

// GetHostIP returns the first non-loopback IPv4 address
func GetHostIP() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "127.0.0.1"
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "127.0.0.1"
}
